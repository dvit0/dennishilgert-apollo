#!/bin/bash

docker pull debian:stable
docker pull debian:bookworm-slim

docker build -t fc-rootfs-builder -f Dockerfile.builder .

docker run \
    --privileged \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $(pwd)/dist:/dist \
    fc-rootfs-builder