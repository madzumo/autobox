#!/bin/bash
echo "Hello, DigitalOcean!" > /root/test.txt
sudo apt-get update
sudo apt-get install xfce4 xfce4-goodies tightvncserver -y
sudo apt-get install dbus-x11 -y
sudo apt-get install expect -y
curl -fsS https://dl.brave.com/install.sh | sh
sleep 2
expect -c '
spawn vncserver
expect {
    "Password:" {
        send "prime6996\r"
        exp_continue
    }
    "Verify:" {
        send "prime6996\r"
        exp_continue
    }
    "Would you like to enter a view-only password (y/n)?" {
        send "n\r"
    }
}'
sleep 2
vncserver -kill :1
echo -e '#!/bin/bash\nxrdb $HOME/.Xresources\nstartxfce4 &\nbrave-browser --no-sandbox --new-window --start-maximized "https://twitch.tv/vortix93"' > ~/.vnc/xstartup
chmod +x ~/.vnc/xstartup
vncserver