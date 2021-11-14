killall -9 ai-linux-arm
killall -9 miner
set -e
echo Uploading...
cd /monad/imperium/software
./utility upload /monad/ai/ai.bit
./utility reset
cd /monad/ai/
# ./ai-linux-arm --port /dev/ttyO2 --iterations 800000000 --timeout 15 --dc test
./ai-linux-arm --supervised