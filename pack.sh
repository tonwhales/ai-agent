set -e
cd build
rm -fr package
mkdir package
cp ai-linux-arm ./package/
cp ../bits/i2.bit ./package/ai.bit
cp ../scripts/install.sh ./package/install.sh
cp ../scripts/start.sh ./package/start.sh
cd package
zip output.zip *