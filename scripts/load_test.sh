#!/bin/bash

# Ensure the script exits immediately if any command fails.
set -e

echo "=========================================================="
echo " Starting High-Throughput Load Test with Vegeta"
echo " Target: http://localhost:8080/ (Load Balancer Entrance)"
echo "=========================================================="

# 1. Define targets formatting file for Vegeta.
echo "GET http://localhost:8080/" > targets.txt

# 2. Fire the attack.
# -rate=48000 specifies 48,000 requests per second.
# -duration=10s runs the test for 10 seconds (Total 480,000 requests)
echo "Attacking load balancer at 48,000 RPS..."
vegeta attack -targets=targets.txt -rate=48000 -duration=10s > results.bin

echo "Attack complete. Processing performace telemetry metrics..."

# 3. Generate textual metrics report focusing on Percentiles.
echo ""
echo "------------------- TELEMETRY REPORT -------------------"
vegeta report results.bin

# 4. Cleanup temporary files.
rm targets.txt
rm results.bin
echo "=========================================================="
