set -e
GOOS=linux GOARCH=arm GOARM=7 go build -o build/ai-linux-arm
set +e
ssh -i ../keys/id_rsa_openssh root@192.168.16.17 "killall -9 ai-linux-arm"
ssh -i ../keys/id_rsa_openssh root@192.168.16.17 "rm -fr /monad/ai/*"
ssh -i ../keys/id_rsa_openssh root@192.168.16.17 "mkdir -p /monad/ai/"
ssh -i ../keys/id_rsa_openssh root@192.168.16.17 "mkdir -p /monad/led_ctrl/"
set -e
scp -i ../keys/id_rsa_openssh ./build/ai-linux-arm root@192.168.16.17:/monad/ai/ai-linux-arm
scp -i ../keys/id_rsa_openssh ./bits/test5.bit root@192.168.16.17:/monad/led_ctrl/R2.bit
scp -i ../keys/id_rsa_openssh ./test_0001.hex root@192.168.16.17:/monad/ai/test_0001.hex
scp -i ../keys/id_rsa_openssh ./bits/r2_2.bit root@192.168.16.17:/monad/ai/ai.bit
# scp -i ../keys/id_rsa_openssh ./bits/i37.bit root@192.168.16.120:/monad/ai/ai.bit
scp -i ../keys/id_rsa_openssh ./test_launch.sh root@192.168.16.17:/monad/ai/test_launch.sh
ssh -i ../keys/id_rsa_openssh root@192.168.16.17 "sh /monad/ai/test_launch.sh"