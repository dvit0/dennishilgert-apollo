#!/bin/bash

# Trap the SIGINT signal and exit
trap 'echo "\nExiting ..."; exit' INT

# Custom script code goes here
echo 'This is intended to be the agent executable'
echo 'It is executed by the tini init system'
echo ''
echo 'Use CTRL + C to exit'

# Loop infinitely
while true; do
  sleep 1
done