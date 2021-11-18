set -e

# Build
./build.sh

# Prepare
./prepare.js

# Package
cd build
rm -fr package
mkdir package
cp ai-linux-arm ./package/
cp ../bits/i33.bit ./package/ai.bit
cp ../scripts/install.sh ./package/install.sh
cp ../scripts/start.sh ./package/start.sh
cd package
zip output.zip *

# Move
cd ../../
export VERSION=$(cat VERSION)
mv ./build/package/output.zip ./build/${VERSION}.zip