#!/bin/bash

outDir=/dist
workDir=/tmp/rootfs-image-mnt

imageFileName=rootfs.ext4
imageFilePath=${outDir}/${imageFileName}

function buildImage() {
  imageTag=$1
  dockerfile=$2

  imageId=$(buildDockerImage ${imageTag} ${dockerfile})

  echo "built docker image ${imageId}"
}

function exportImage() {
  imageTag=$1

  prepareDirectories

  createEmptyImage
  mountImage

  containerId=$(createDockerContainer ${imageTag})
  startDockerContainer ${containerId}

  copyRootFs ${containerId}

  stopDockerContainer ${containerId}
  removeDockerContainer ${containerId}

  unmountImage

  checkImageFilesystem
  resizeImageToMinimumSize
}

function copyRootFs() {
  containerId=$1

  # If the directories are included in the LIST, then only create an empty directory on the target
  LIST="/boot /opt /proc /run /srv /sys /tmp"
  for d in $(docker exec ${containerId} find / -maxdepth 1 -type d); do
    if echo $LIST | grep -w $d > /dev/null; then
      echo "mkdir of ${d}"
      mkdir ${workDir}${d}
    fi
  done

  # If the directories are not included in the LIST, then copy the directories to the target
  LIST="/ /boot /opt /proc /run /srv /sys /tmp"
  for d in $(docker exec ${containerId} find / -maxdepth 1 -type d); do
    if echo $LIST | grep -v -w $d > /dev/null; then
      echo "copying content of ${d}"
      docker cp ${containerId}:${d} ${workDir}
    fi
  done
}

function copyRootFsOld() {
  containerId=$1

  excludeDirs="/boot /opt /proc /run /srv /sys /tmp"
  mkdirOnlyCommand="LIST=${excludeDirs}; for d in $(find / -maxdepth 1 -type d); do if echo $LIST | grep -w $d > /dev/null; then mkdir ${containerMountDir}${d}; fi; done; exit 0"
  copyDirsCommand="LIST=/ ${containerMountDir} ${excludeDirs}; for d in $(find / -maxdepth 1 -type d); do if echo $LIST | grep -v -w $d > /dev/null; then tar c \"$d\" | tar x -C ${containerMountDir}; fi; done; exit 0"
  cleanupCommand="rm -r ${containerMountDir}${containerMountDir}"

  echo "copying container rootfs to ${containerMountDir} ..."
  runInContainer ${containerId} "${mkdirOnlyCommand}"
  runInContainer ${containerId} "${copyDirsCommand}"
  runInContainer ${containerId} "${cleanupCommand}"
  echo "copied container rootfs successfully"
}

function copyRootFsToOutDir() {
  echo "copying exported rootfs to ${outDir} ..."
  cp ${imageFilePath} ${outDir}/${imageFileName}
  echo "copied exported rootfs successfully"
}

function prepareDirectories() {
  if [ -d ${workDir} ]; then
    rm -rf ${workDir}
  fi
  mkdir -p ${workDir}

  if [ -d ${imageDir} ]; then
    rm -rf ${imageDir}
  fi
  mkdir -p ${imageDir}
}

function createEmptyImage() {
  echo "creating empty image file ${imageFilePath} ..."
  dd if=/dev/zero of=${imageFilePath} bs=1M count=1000

  echo "creating ext4 filesystem in empty image file ${imageFilePath} ..."
  mkfs.ext4 ${imageFilePath}

  echo "created empty image file successfully"
}

function mountImage() {
  echo "mounting image ${imageFilePath} to ${workDir} ..."
  mount -o loop ${imageFilePath} ${workDir}
  echo "mounted image successfully"
}

function unmountImage() {
  echo "unmounting directory ${workDir} ..."
  umount ${workDir}
  echo "unmounted directory successfully"
}

function buildDockerImage() {
  imageTag=$1
  dockerfile=$2

  #echo "building docker image ${imageTag} from Dockerfile ${dockerfile} ..."
  imageId=$(docker build -t ${imageTag} -f ${dockerfile} .)
  #echo "built docker image successfully"

  echo ${imageId}
}

function createDockerContainer() {
  containerImage=$1

  #echo "creating docker container from image ${containerImage} ..."
  containerId=$(docker create ${containerImage})
  #echo "created docker container successfully"

  echo ${containerId}
}

function removeDockerContainer() {
  containerId=$1

  if [ "$(docker ps -aq -f id=${containerId})" ]; then
    if [ "$(docker ps -aq -f id=${containerId} -f status=exited)" ]; then
      echo "removing docker container ${containerId} ..."
      docker container rm ${containerId}
      echo "removed docker container successfully"
    else
      echo "docker container ${containerId} does not have status code 'exited'"
    fi
  else
    echo "docker container ${containerId} does not exist"
  fi
}

function runDockerContainer() {
  imageTag=$1

  echo "running docker container from image ${imageTag} ..."
  docker rm -f rootfs || true
  docker run -i --name rootfs -v ${workDir}:${containerMountDir} ${imageTag} bash # -s < ./copy-rootfs.sh
  echo "ran docker container successfully"
}

function waitForContainerToExit() {
  containerId=$1

  while [ ! "$(docker ps -aq -f status=exited -f id=${containerId})" ]
  do
    echo "docker container rootfs not exited yet, waiting"
    sleep 1
  done
}

function startDockerContainer() {
  containerId=$1

  echo "starting docker container ${containerId} ..."
  docker container start ${containerId}
  echo "started docker container successfully"
}

function stopDockerContainer() {
  containerId=$1

  echo "stopping docker container ${containerId} ..."
  docker container stop ${containerId}

  while [ ! "$(docker ps -aq -f status=exited -f id=${containerId})" ]
  do
    echo "docker container ${containerID} not exited yet, waiting"
    sleep 1
  done
  echo "stopped docker container successfully"
}

function runInContainer() {
  containerId=$1
  command=$2

  echo "running command ${command} in docker container ${containerId} ..."
  docker exec ${containerId} /bin/sh "${command}"
  echo "ran command in docker container successfully"
}

function resizeImageToMinimumSize() {
  echo "resizing image ${imageFilePath} with size $(du -sh ${imageFilePath}) to minimum size ..."
  resize2fs -M ${imageFilePath} 
  echo "resized image to size $(du -sh ${imageFilePath})"
}

function checkImageFilesystem() {
  echo "checking filesystem of image ${imageFilePath} ..."
  e2fsck -y -f ${imageFilePath}
  echo "check filesystem successfully"
}

buildImage fc-rootfs Dockerfile.rootfs
exportImage fc-rootfs
