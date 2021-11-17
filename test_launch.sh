killall -9 ai-linux-arm
killall -9 miner
set -e
echo Uploading...
cd /monad/imperium/software
./utility upload /monad/ai/ai.bit
./utility reset
cd /monad/ai/
./ai-linux-arm --port /dev/ttyO1 --supervised --dc test
# ./ai-linux-arm --port /dev/ttyO1 --iterations 80000000 --timeout 10 --dc test --chip 1
# ./ai-linux-arm --port /dev/ttyO1 --test --chip 6