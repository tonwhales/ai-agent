set -e
GOOS=linux GOARCH=arm GOARM=7 go build -o build/ai-linux-arm
set +e
ssh -i ../keys/id_rsa_openssh root@192.168.16.120 "killall -9 ai-linux-arm"
ssh -i ../keys/id_rsa_openssh root@192.168.16.120 "rm -fr /sdcard/ai/*"
set -e
scp -i ../keys/id_rsa_openssh ./build/ai-linux-arm root@192.168.16.120:/sdcard/ai/ai-linux-arm
scp -i ../keys/id_rsa_openssh ./bits/uart_chain_9.bit root@192.168.16.120:/sdcard/ai/ai.bit
scp -i ../keys/id_rsa_openssh ./test_launch.sh root@192.168.16.120:/sdcard/ai/test_launch.sh
ssh -i ../keys/id_rsa_openssh root@192.168.16.120 "sh /sdcard/ai/test_launch.sh"