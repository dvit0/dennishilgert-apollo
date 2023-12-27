#!/bin/bash

# Credit https://github.com/anyfiddle/firecracker-rootfs-builder

workDir=/tmp/rootfs
outDir=/dist

#outputImageFilename=${2:-rootfs.ext4}
outputImageFilename=rootfs.ext4
imageFile=${outDir}/${outputImageFilename}

function prepareoutDir() {
    mkdir -p ${outDir}
    rm -rf ${imageFile}
}

function createEmptyImage() {
    echo "Creating rootfs image as ${imageFile}"

    # Truncate the image file to desired size
    truncate -s 5G ${imageFile}

    # Make image file
    mkfs.ext4 ${imageFile}

    #mksquashfs ${workDir} ${outputImageFilename} -noappend
}

function mountImage() {
    #Create temp mount directory
    rm -rf ${workDir}
    mkdir ${workDir}

    # Mount the image as a loop device (Virtual drive kind of)
    echo "Mounting rootfs image to ${workDir}"
    mount -o loop ${imageFile} ${workDir}
}

function unmountImage() {
    umount ${workDir} || true
}

function createRootFsWithScript() {
    echo "Downloading debian root filesystem using debootstrap"
    # udev is necessary for auto login
    # others are optional but useful
    debootstrap --arch=amd64 --variant=minbase --include=udev,tini,iproute2,iputils-ping bullseye ${workDir} http://deb.debian.org/debian/
    #debootstrap --arch=amd64 --variant=minbase --include=udev,systemd,systemd-sysv,procps,haveged,iproute2,iputils-ping bullseye ${workDir} http://deb.debian.org/debian/
}

function runOnRootFs() {
    prepareScript=$1
    agent=$2

    echo "Change to mounted rootfs using chroot"
    mount -t proc /proc ${workDir}/proc/
    mount -t sysfs /sys ${workDir}/sys/
    mount -o bind /dev ${workDir}/dev/

    echo "Copying prepare script to new filesystem"
    cp ${prepareScript} ${workDir}/${prepareScript}
    chmod +x ${workDir}/${prepareScript}

    echo "Copying agent to new filesystem"
    cp ${agent} ${workDir}/usr/bin/${agent}
    chmod +x ${workDir}/usr/bin/${agent}

    echo "Customizing image with prepare script"
    chroot ${workDir} /bin/sh ${prepareScript}

    rm ${workDir}/${prepareScript}

    echo "Unmounting"
    umount ${workDir}/dev
    umount ${workDir}/proc
    umount ${workDir}/sys
}

function prepareFilesystem() {
    rm -rf ${workDir}/var/cache/apt/archives \
	       ${workDir}/usr/share/doc \
	       ${workDir}/var/lib/apt/lists
	mkdir -p /{dev,bin,etc,etc/init.d,tmp,var,run,proc,sys}
}

function checkImageFilesystem() {
    # Check image image file system
    e2fsck -y -f ${imageFile}
}

function getMinimumFilesizeForImage() {
    # Get minimum size of the image
    resize2fs -P ${imageFile}
}

function resizeImageToMinimumSize() {
    resize2fs -M ${imageFile} 
}

prepareoutDir
createEmptyImage
mountImage
createRootFsWithScript
runOnRootFs prepare-rootfs.sh agent.sh
prepareFilesystem
unmountImage
resizeImageToMinimumSize
checkImageFilesystem
