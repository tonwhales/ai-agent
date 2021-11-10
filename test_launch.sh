killall -9 ai-linux-arm
killall -9 miner
set -e
echo Uploading...
cd /monad/imperium/software
./utility upload /sdcard/ai/i3.bit
./utility reset
cd /sdcard/ai/
./ai-linux-arm --port /dev/ttyO1 --iterations 1000000000 --timeout 20