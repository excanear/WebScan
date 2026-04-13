#!/usr/bin/env bash
set -euo pipefail

export TERM=xterm-256color
echo "Running headless UI tests (TestRunHeadless) under Xvfb if available"

if command -v xvfb-run >/dev/null 2>&1; then
  xvfb-run -s "-screen 0 1024x768x24" go test ./... -run TestRunHeadless -v
else
  go test ./... -run TestRunHeadless -v
fi
