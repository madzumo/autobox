#!/bin/bash
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
        send "prime7\r"
        exp_continue
    }
    "Verify:" {
        send "prime7\r"
        exp_continue
    }
    "Would you like to enter a view-only password (y/n)?" {
        send "n\r"
    }
}'
sleep 2
vncserver -kill :1
echo "-------vnc shutdown------"
echo -e '#!/bin/bash\nxrdb $HOME/.Xresources\nstartxfce4 &\nbrave-browser --no-sandbox --new-window --start-maximized "$URL"' > ~/.vnc/xstartup
chmod +x ~/.vnc/xstartup
sleep 1
vncserver
#sleep 2
#curl -L https://azuredatastudio-update.azurewebsites.net/latest/linux-deb-x64/stable -o azure.deb
#azuredatastudio