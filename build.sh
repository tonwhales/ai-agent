set -e
GOOS=darwin GOARCH=amd64 go build -o build/ai-macos-x64
GOOS=windows GOARCH=amd64 go build -o build/ai-windows-x64.exe
GOOS=linux GOARCH=amd64 go build -o build/ai-linux-x64
GOOS=linux GOARCH=arm GOARM=7 go build -o build/ai-linux-arm