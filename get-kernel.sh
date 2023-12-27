#!/bin/bash

ARCH="$(uname -m)"
wget https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.7/${ARCH}/vmlinux-5.10.204