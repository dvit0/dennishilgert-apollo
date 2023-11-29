#!/bin/bash

echo "Preparing"
# DNS requests are forwarded to the host. DHCP DNS options are ignored
#ln -sf /run/resolvconf/resolv.conf /etc/resolv.conf

echo "Setting up dns server"
rm /etc/resolv.conf
touch /etc/resolv.conf
echo "nameserver 1.1.1.1" > /etc/resolv.conf

#echo "Setting up hostname"
#echo "apollo-e-2sjd23" > /etc/hostname

#echo "Configuring network interface"
#cat <<EOF > /etc/netplan/99_config.yaml
#network:
#  version: 2
#  renderer: networkd
#  ethernets:
#    eth0:
#      dhcp4: no
#      addresses: [172.16.0.2/24]
#      gateway4: 172.16.0.1
#      nameservers:
#        addresses: [1.1.1.1,1.0.0.1]
#EOF
#netplan generate

echo "Removing root password"
passwd -d root

echo "Configuring auto login"
mkdir /etc/systemd/system/serial-getty@ttyS0.service.d/
cat <<EOF > /etc/systemd/system/serial-getty@ttyS0.service.d/autologin.conf
  [Service]
  ExecStart=
  ExecStart=-/sbin/agetty --autologin root -o '-p -- \\u' --keep-baud 115200,38400,9600 %I $TERM
EOF
