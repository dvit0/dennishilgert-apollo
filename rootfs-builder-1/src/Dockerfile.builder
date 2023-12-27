FROM debian:stable

WORKDIR /workspace

RUN apt update && apt install -y \
  curl \
  build-essential \
  debootstrap \
  docker.io

RUN rm -rf /var/lib/apt/lists/*

COPY Dockerfile.target Dockerfile.target

COPY agent.sh agent.sh

COPY create-rootfs.sh create-rootfs.sh
RUN chmod +x create-rootfs.sh

COPY prepare-rootfs.sh prepare-rootfs.sh
RUN chmod +x prepare-rootfs.sh

VOLUME /output

ENTRYPOINT ["./create-rootfs.sh"]