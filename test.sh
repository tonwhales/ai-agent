set -e
GOOS=linux GOARCH=arm GOARM=7 go build -o build/ai-linux-arm
scp -i ../keys/id_rsa_openssh ./build/ai-linux-arm root@192.168.16.120:/sdcard/ai/ai-linux-arm
scp -i ../keys/id_rsa_openssh ./bits/i1.bit root@192.168.16.120:/sdcard/ai/i1.bit
scp -i ../keys/id_rsa_openssh ./bits/uart_test_1.bit root@192.168.16.120:/sdcard/ai/uart_test_1.bit