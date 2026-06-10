// Package sandbox executes untrusted user code inside hardened, throwaway
// Docker containers. Every defence here is deliberate and interview-talkable:
//
//	no network namespace      no exfiltration / no inbound
//	read-only root + tmpfs    can't tamper with the image, /tmp is capped
//	memory + swap cap          OOM-kills memory bombs instead of the host
//	cpu quota                  one runaway loop can't starve the box
//	pids-limit (cgroup)        contains fork bombs, container-scoped
//	cap-drop ALL               no privileged kernel operations
//	no-new-privileges          setuid binaries can't escalate
//	seccomp allowlist          default-deny syscall filter (see seccomp.json)
//	non-root user 1000         no root even inside the namespace
//	hard wall-clock timeout    kills code that never terminates
//
// We intentionally do NOT use the `--ulimit nproc` from the original spec:
// nproc is enforced per-UID across the whole host, so uid 1000 trips it
// immediately on a busy machine. --pids-limit is the correct container-scoped
// control and is what actually contains a fork bomb.
package sandbox

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-units"
)

//go:embed seccomp.json
var seccompProfile string

const (
	memoryBytes = 128 * 1024 * 1024 // 128 MiB
	nanoCPUs    = 500_000_000       // 0.5 CPU
	pidsLimit   = 64
)

type Config struct {
	Image   string
	Timeout time.Duration
}

// RunResult summarises a finished (or killed) execution.
type RunResult struct {
	ExitCode int
	TimedOut bool
}

type Runner struct {
	cli     *client.Client
	image   string
	timeout time.Duration
	secOpt  []string
}

func NewRunner(cfg Config) (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	// Probe the daemon so we fail fast (and the server can disable the feature)
	// instead of erroring on the first user request.
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := cli.Ping(pingCtx); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("docker ping: %w", err)
	}

	image := cfg.Image
	if image == "" {
		image = "codeforge-runner-python"
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Runner{
		cli:     cli,
		image:   image,
		timeout: timeout,
		secOpt:  []string{"no-new-privileges", "seccomp=" + seccompProfile},
	}, nil
}

func (r *Runner) Close() error { return r.cli.Close() }

func (r *Runner) hostConfig() *container.HostConfig {
	pids := int64(pidsLimit)
	return &container.HostConfig{
		NetworkMode:    "none",
		ReadonlyRootfs: true,
		Tmpfs:          map[string]string{"/tmp": "size=64m,mode=1777"},
		CapDrop:        []string{"ALL"},
		SecurityOpt:    r.secOpt,
		Resources: container.Resources{
			Memory:     memoryBytes,
			MemorySwap: memoryBytes, // == Memory ⇒ swap disabled
			NanoCPUs:   nanoCPUs,
			PidsLimit:  &pids,
			Ulimits:    []*units.Ulimit{{Name: "nofile", Soft: 64, Hard: 64}},
		},
	}
}

// Warm creates and starts a container that blocks reading its stdin (the
// entrypoint's `cat`). Pre-warming pays the create+start cost up front so a
// pooled Exec only has to attach and stream — the bulk of cold-start latency.
func (r *Runner) Warm(ctx context.Context) (string, error) {
	cc := &container.Config{
		Image:        r.image,
		OpenStdin:    true,
		StdinOnce:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		User:         "1000:1000",
	}
	created, err := r.cli.ContainerCreate(ctx, cc, r.hostConfig(), nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	if err := r.cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		r.Remove(created.ID)
		return "", fmt.Errorf("start: %w", err)
	}
	return created.ID, nil
}

// Remove force-removes a container with a fresh context, so cleanup still runs
// after a timeout has cancelled the caller's context.
func (r *Runner) Remove(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = r.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

// Exec feeds code to an already-warmed container's stdin, streams its
// stdout/stderr to the given writers as it runs, and waits for exit (or the
// wall-clock timeout). The container is single-use; the caller removes it.
func (r *Runner) Exec(ctx context.Context, id, code string, stdout, stderr io.Writer) (RunResult, error) {
	att, err := r.cli.ContainerAttach(ctx, id, container.AttachOptions{
		Stream: true, Stdin: true, Stdout: true, Stderr: true,
	})
	if err != nil {
		return RunResult{}, fmt.Errorf("attach: %w", err)
	}
	defer att.Close()

	// Send the program, then half-close stdin so `cat` sees EOF and exec's python.
	go func() {
		_, _ = att.Conn.Write([]byte(code))
		_ = att.CloseWrite()
	}()

	// Demux the multiplexed attach stream into stdout/stderr as it arrives.
	copyDone := make(chan struct{})
	go func() {
		_, _ = stdcopy.StdCopy(stdout, stderr, att.Reader)
		close(copyDone)
	}()

	// Wait on a background context + our own timer, so a timeout is reported as
	// TimedOut rather than racing a context-cancelled wait error.
	statusCh, errCh := r.cli.ContainerWait(context.Background(), id, container.WaitConditionNotRunning)
	timer := time.NewTimer(r.timeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		r.kill(id)
		<-copyDone
		return RunResult{TimedOut: true, ExitCode: 137}, nil
	case <-ctx.Done():
		r.kill(id)
		<-copyDone
		return RunResult{}, ctx.Err()
	case err := <-errCh:
		<-copyDone
		return RunResult{}, fmt.Errorf("wait: %w", err)
	case st := <-statusCh:
		<-copyDone
		return RunResult{ExitCode: int(st.StatusCode)}, nil
	}
}

func (r *Runner) kill(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = r.cli.ContainerKill(ctx, id, "KILL")
}

// Run is the un-pooled path: warm a fresh container, exec, remove.
func (r *Runner) Run(ctx context.Context, code string, stdout, stderr io.Writer) (RunResult, error) {
	id, err := r.Warm(ctx)
	if err != nil {
		return RunResult{}, err
	}
	defer r.Remove(id)
	return r.Exec(ctx, id, code, stdout, stderr)
}
