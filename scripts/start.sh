set -e
echo "Uploading..."
cd /monad/imperium/software/
./utility upload ./ai.bit

echo "Starting..."
/monad/work/ai-linux-arm --iterations 10000000 --timeout 10 --supervised