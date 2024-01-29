#!/bin/bash

WORKDIR=/tmp/protoc

echo "installing protoc ..."

# Create workdir
mkdir "$WORKDIR"

# Install zip to unzip the downloaded archive
sudo apt-get update && sudo apt-get install -y zip

# Download and install protoc
VERSION=25.2
FILENAME="protoc-$VERSION-linux-x86_64"
wget "https://github.com/protocolbuffers/protobuf/releases/download/v$VERSION/$FILENAME.zip" -P "$WORKDIR"
unzip "$WORKDIR/$FILENAME.zip" -d "$WORKDIR/$FILENAME"
sudo mv "$WORKDIR/$FILENAME/bin/protoc" /usr/local/bin/protoc
sudo mv "$WORKDIR/$FILENAME/include/google" /usr/local/include/
sudo chmod +x /usr/local/bin/protoc

# Cleanup
rm -rf "$WORKDIR"

echo "protoc has been installed successfully"