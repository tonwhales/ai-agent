set +e
killall -9 miner
killall -9 utility
killall -9 ai-linux-arm
set -e
chmod +x ./start.sh
chmod +x ai-linux-arm