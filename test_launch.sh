killall -9 ai-linux-arm
killall -9 miner
set -e
echo Uploading...
cd /monad/imperium/software
./utility upload /monad/ai/ai.bit
./utility reset
cd /monad/ai/
# ./ai-linux-arm --port /dev/ttyO1 --supervised --dc test
./ai-linux-arm --port /dev/ttyO1 --iterations 1000000000 --timeout 20 --dc test --chip 1
# ./ai-linux-arm --port /dev/ttyO1 --chip 6 --iterations 160000000 --config './test_0001.hex'