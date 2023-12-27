#!/bin/bash

# Hide ^C
#stty -echoctl

function trap_exit() {
  echo 'Trapped CTRL + C'
  exit
}

# Trap the SIGINT signal and exit
trap trap_exit INT
trap trap_exit SIGINT

# Custom script code goes here
echo 'This is intended to be the agent executable'
echo 'It is executed by the tini init system'

echo 'Ping host machine ...'
ping -c 5 172.16.0.1
echo ''

echo 'Ping google.com ...'
ping -c 5 google.com
echo ''

echo 'Press <any key> to exit'
read -s -n 1 key

echo $PATH
which shutdown
which reboot
reboot now