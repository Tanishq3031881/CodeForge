# Runner image for executing untrusted user Python in a locked-down container.
# Slim base, a non-root user, common libraries pre-installed, and NO pip at
# runtime (the container has no network anyway). The real isolation comes from
# the runtime flags the Go runner applies (no network, read-only rootfs, memory
# / pids / cpu limits, dropped capabilities, seccomp); this image just keeps the
# attack surface small.
FROM python:3.12-slim

RUN useradd -m -u 1000 -s /bin/bash runner && \
    pip install --no-cache-dir numpy requests && \
    rm -rf /root/.cache /var/lib/apt/lists/*

USER runner
WORKDIR /home/runner

COPY --chown=runner:runner entrypoint.sh /home/runner/entrypoint.sh
RUN chmod +x /home/runner/entrypoint.sh

ENTRYPOINT ["/home/runner/entrypoint.sh"]
