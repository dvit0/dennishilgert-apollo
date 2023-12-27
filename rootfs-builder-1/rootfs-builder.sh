#!/bin/bash

docker pull debian:stable
docker pull debian:bookworm-slim

docker build -t fc-rootfs-builder -f src/Dockerfile.builder src

docker run \
    --privileged \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $(pwd)/output:/output \
    fc-rootfs-builder