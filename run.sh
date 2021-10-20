set -e
GOOS=darwin GOARCH=amd64 go build -o build/ai-macos-x64
./build/ai-macos-x64