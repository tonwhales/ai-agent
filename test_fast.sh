GOOS=linux GOARCH=arm GOARM=7 go build -o build/ai-linux-arm
scp -i ../keys/id_rsa_openssh ./build/ai-linux-arm root@192.168.16.120:/monad/ai/ai-linux-arm
scp -i ../keys/id_rsa_openssh ./test_launch.sh root@192.168.16.120:/monad/ai/test_launch.sh
ssh -i ../keys/id_rsa_openssh root@192.168.16.120 "/monad/ai/ai-linux-arm --port /dev/ttyO1 --chip 6"
# ssh -i ../keys/id_rsa_openssh root@192.168.16.120 "/monad/ai/ai-linux-arm --supervised --dc test"