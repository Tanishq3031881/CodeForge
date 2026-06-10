package sandbox

import (
	"context"
	"io"
	"log"
	"time"
)

// Pool keeps a buffer of pre-warmed containers so a run only pays attach +
// stream latency, not create + start. Cold start is typically several hundred
// ms; a pooled run is a fraction of that. Each container is single-use: popped,
// fed code, then discarded, and the pool refills in the background.
type Pool struct {
	runner  *Runner
	idle    chan string
	closing chan struct{}
}

func NewPool(runner *Runner, size int) *Pool {
	if size < 0 {
		size = 0
	}
	p := &Pool{
		runner:  runner,
		idle:    make(chan string, size),
		closing: make(chan struct{}),
	}
	for range size {
		go p.refill()
	}
	return p
}

// refill warms one container and parks it on the idle channel. If the pool is
// closing (or already full and closing) the fresh container is removed.
func (p *Pool) refill() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	id, err := p.runner.Warm(ctx)
	if err != nil {
		log.Printf("sandbox pool: warm failed: %v", err)
		return
	}
	select {
	case p.idle <- id:
	case <-p.closing:
		p.runner.Remove(id)
	}
}

// Run executes code in a pooled container if one is ready, otherwise warms one
// on demand (overflow under burst). Either way the container is removed after.
func (p *Pool) Run(ctx context.Context, code string, stdout, stderr io.Writer) (RunResult, error) {
	var id string
	select {
	case id = <-p.idle:
		// Took one from the pool — kick off a replacement.
		go p.refill()
	default:
		// Pool empty: warm synchronously (this request eats the cold start).
		w, err := p.runner.Warm(ctx)
		if err != nil {
			return RunResult{}, err
		}
		id = w
	}
	defer p.runner.Remove(id)
	return p.runner.Exec(ctx, id, code, stdout, stderr)
}

// Close drains and removes any idle containers.
func (p *Pool) Close() {
	close(p.closing)
	for {
		select {
		case id := <-p.idle:
			p.runner.Remove(id)
		default:
			_ = p.runner.Close()
			return
		}
	}
}
