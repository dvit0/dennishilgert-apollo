#!/bin/bash

echo "pulling required images ..."
docker pull debian:stable
docker pull debian:bullseye-slim
echo "pulled required images successfully"

echo "building fc-rootfs-builder image ..."
docker build -t fc-rootfs-builder -f Dockerfile.builder .
echo "built fc-rootfs-builder image successfully"

echo "running fc-rootfs-builder container ..."
docker run \
    --privileged \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $(pwd)/dist:/dist \
    fc-rootfs-builder:latest
echo "ran fc-rootfs-builder container successfully"