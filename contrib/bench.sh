#!/bin/bash
# Benchmarking script for PomoGo
set -e

# 1. Build binary
echo "Building pomogo..."
go build -o pomogo ./cmd/pomogo

# 2. Measure binary size
bin_size_bytes=$(wc -c < pomogo)
bin_size_mb_int=$((bin_size_bytes / 1048576))
bin_size_mb_frac=$(( (bin_size_bytes % 1048576) * 100 / 1048576 ))
bin_size_mb=$(printf "%d.%02d" $bin_size_mb_int $bin_size_mb_frac)

# 3. Measure cold start (average of 10 runs)
echo "Measuring cold-start time..."
total_time=0
for i in {1..10}; do
  start=$(date +%s%N)
  ./pomogo version > /dev/null
  end=$(date +%s%N)
  diff=$((end - start))
  total_time=$((total_time + diff))
done
avg_ns=$((total_time / 10))
avg_ms=$((avg_ns / 1000000))

# 4. Measure RSS and CPU %
echo "Measuring RSS and CPU% (running idle for 5s)..."
timeout 5 ./pomogo < /dev/null > /dev/null 2>&1 &
ppid_val=$!
sleep 3
child_pid=$(pgrep -P $ppid_val || true)

rss_kb=0
cpu_pct="0.0"
if [ -n "$child_pid" ] && ps -p $child_pid > /dev/null; then
  rss_kb=$(ps -p $child_pid -o rss= | tr -d ' ')
  cpu_pct=$(ps -p $child_pid -o %cpu= | tr -d ' ')
fi
kill -9 $child_pid $ppid_val > /dev/null 2>&1 || true

rss_mb_int=$((rss_kb / 1024))
rss_mb_frac=$(( (rss_kb % 1024) * 100 / 1024 ))
rss_mb=$(printf "%d.%02d" $rss_mb_int $rss_mb_frac)

# 5. Output results
echo "======================================"
echo "          POMOGO BENCHMARKS           "
echo "======================================"
echo "Binary Size:      $bin_size_mb MB ($bin_size_bytes bytes)"
echo "Cold Start:       $avg_ms ms"
echo "Idle RSS:         $rss_mb MB ($rss_kb KB)"
echo "Idle CPU:         $cpu_pct %"
echo "======================================"
