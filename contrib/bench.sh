#!/usr/bin/env bash
# Benchmarking script for PomoGo
set -euo pipefail

bin_path="${TMPDIR:-/tmp}/pomogo-bench-$$"
preview_file="$(mktemp "${TMPDIR:-/tmp}/pomogo-preview.XXXXXX")"
trap 'rm -f "$bin_path" "$preview_file"' EXIT

avg_command_ms() {
  total_time=0
  for _ in {1..10}; do
    start=$(date +%s%N)
    "$@" > /dev/null
    end=$(date +%s%N)
    diff=$((end - start))
    total_time=$((total_time + diff))
  done
  avg_ns=$((total_time / 10))
  echo $((avg_ns / 1000000))
}

# 1. Build binary
echo "Building pomogo..."
go build -o "$bin_path" ./cmd/pomogo

# 2. Measure binary size
bin_size_bytes=$(wc -c < "$bin_path")
bin_size_mb_int=$((bin_size_bytes / 1048576))
bin_size_mb_frac=$(( (bin_size_bytes % 1048576) * 100 / 1048576 ))
bin_size_mb=$(printf "%d.%02d" $bin_size_mb_int $bin_size_mb_frac)

# 3. Measure cold start (average of 10 runs)
echo "Measuring cold-start time..."
avg_ms=$(avg_command_ms "$bin_path" version)

# 4. Measure preview render cost with effects off and on
echo "Measuring preview render time..."
"$bin_path" screenshot-preview --width 80 --height 24 --layout monolith --theme high-contrast --effects none > "$preview_file"
if [ ! -s "$preview_file" ]; then
  echo "Preview render is empty"
  exit 1
fi
preview_none_ms=$(avg_command_ms "$bin_path" screenshot-preview --width 80 --height 24 --layout monolith --theme high-contrast --effects none)
preview_effects_ms=$(avg_command_ms "$bin_path" screenshot-preview --width 80 --height 24 --layout monolith --theme high-contrast --effects rain)

# 5. Measure RSS and CPU %
echo "Measuring RSS and CPU% (running idle for 5s)..."
timeout 5 "$bin_path" < /dev/null > /dev/null 2>&1 &
ppid_val=$!
sleep 3
child_pid=$(pgrep -P $ppid_val || true)

rss_display="n/a"
cpu_pct="n/a"
if [ -n "$child_pid" ] && ps -p $child_pid > /dev/null; then
  rss_kb=$(ps -p $child_pid -o rss= | tr -d ' ')
  cpu_pct=$(ps -p $child_pid -o %cpu= | tr -d ' ')
  rss_mb_int=$((rss_kb / 1024))
  rss_mb_frac=$(( (rss_kb % 1024) * 100 / 1024 ))
  rss_display=$(printf "%d.%02d MB (%d KB)" $rss_mb_int $rss_mb_frac $rss_kb)
fi
kill -9 $child_pid $ppid_val > /dev/null 2>&1 || true

# 6. Output results
echo "======================================"
echo "          POMOGO BENCHMARKS           "
echo "======================================"
echo "Binary Size:      $bin_size_mb MB ($bin_size_bytes bytes)"
echo "Cold Start:       $avg_ms ms"
echo "Preview Render:   $preview_none_ms ms (effects off)"
echo "Preview Effects:  $preview_effects_ms ms (rain)"
echo "Idle RSS:         $rss_display"
echo "Idle CPU:         $cpu_pct %"
echo "======================================"
