#!/usr/bin/env bash

PROC_NAME="ballerina-lsp"

PID=$(pgrep -n "$PROC_NAME")

if [ -z "$PID" ]; then
  echo "No running process found for: $PROC_NAME"
  exit 1
fi

echo "Tracking process: $PROC_NAME"
echo "PID: $PID"
echo

max_rss=0
max_cpu="0.0"

while kill -0 "$PID" 2>/dev/null; do
  rss=$(ps -p "$PID" -o rss= | awk '{print $1}')
  cpu=$(ps -p "$PID" -o %cpu= | awk '{print $1}')
  elapsed=$(ps -p "$PID" -o etime= | awk '{$1=$1; print}')

  # Process may exit between kill and ps.
  if [ -z "$rss" ] || [ -z "$cpu" ]; then
    break
  fi

  if [ "$rss" -gt "$max_rss" ]; then
    max_rss=$rss
  fi

  max_cpu=$(awk -v cur="$cpu" -v max="$max_cpu" 'BEGIN {
    if (cur > max) print cur;
    else print max;
  }')

  printf "rss=%4d MB  max_rss=%4d MB  cpu=%6s%%  max_cpu=%6s%%  elapsed=%s\n" \
    "$((rss / 1024))" \
    "$((max_rss / 1024))" \
    "$cpu" \
    "$max_cpu" \
    "$elapsed"

  sleep 1
done

echo
echo "Process exited."
echo "Peak sampled RSS: $((max_rss / 1024)) MB"
echo "Peak sampled CPU: $max_cpu%"
