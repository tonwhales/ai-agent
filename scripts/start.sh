set -e
echo "Uploading..."
cd /monad/imperium/software/
./utility upload /monad/imperium/software/work/ai.bit

echo "Starting..."
cd /monad/imperium/software/work
./ai-linux-arm --iterations 10000000 --timeout 10 --supervised --port /dev/ttyO1