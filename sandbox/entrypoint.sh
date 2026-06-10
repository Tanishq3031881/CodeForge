#!/bin/bash
# Read the user's program from stdin, then run it.
set -e
cat > /tmp/code.py
exec python3 /tmp/code.py
