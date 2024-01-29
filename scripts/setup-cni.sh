#!/bin/bash

# References:
# https://github.com/firecracker-microvm/firecracker-go-sdk/blob/main/README.md
# https://combust-labs.github.io/firebuild-docs/installation/cni_plugins/
# https://gvisor.dev/docs/tutorials/cni/
# https://gruchalski.com/posts/2021-02-17-bridging-the-firecracker-network-gap/

if [[ $GOPATH == "" ]]
then
  echo "Please make sure that the GOPATH environment variable is set properly and try again"
  exit
fi

# Create directory for CNI plugin binaries and config files
mkdir -p /opt/cni/bin
mkdir -p /etc/cni/conf.d/

# Install the CNI plugins binaries to the binaries directory
CNI_PLUGINS_VERSION=v1.4.0
wget https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGINS_VERSION/cni-plugins-linux-amd64-$CNI_PLUGINS_VERSION.tgz

# Build and install the tc-redirect-tap plugin binary
mkdir -p $GOPATH/src/github.com/awslabs/tc-redirect-tap
cd $GOPATH/src/github.com/awslabs/tc-redirect-tap
git clone https://github.com/awslabs/tc-redirect-tap.git .
make install

# Create the config file for the firecracker network
sudo sh -c 'cat > /etc/cni/conf.d/fcnet.conflist << EOF
{
  "name": "fcnet",
  "cniVersion": "0.4.0",
  "plugins": [
    {
      "type": "bridge",
      "bridge": "fcnetbridge0",
      "isDefaultGateway": true,
      "hairpinMode": true,
      "ipMasq": true,
      "ipam": {
        "type": "host-local",
        "subnet": "10.6.0.0/24",
        "rangeStart": "10.6.0.5",
        "gateway": "10.6.0.1"
      },
      "dns": {
        "nameservers": [ "10.6.0.1", "1.1.1.1" ]
      }
    },
    {
      "type": "firewall"
    },
    {
      "type": "tc-redirect-tap"
    }
  ]
}
EOF'