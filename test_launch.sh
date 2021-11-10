killall -9 ai-linux-arm
killall -9 miner
set -e
echo Uploading...
cd /monad/imperium/software
./utility upload /sdcard/ai/ai.bit
./utility reset
cd /sdcard/ai/
./ai-linux-arm --port /dev/ttyO1 --iterations 20000000 --timeout 60