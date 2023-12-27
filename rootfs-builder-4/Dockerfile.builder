FROM debian:stable

WORKDIR /workspace

RUN apt update && apt install --no-install-recommends -y \
  debootstrap \
  squashfs-tools

COPY create-rootfs.sh create-rootfs.sh
RUN chmod +x create-rootfs.sh

COPY prepare-rootfs.sh prepare-rootfs.sh
RUN chmod +x prepare-rootfs.sh

COPY agent.sh agent.sh
RUN chmod +x agent.sh

VOLUME /dist

ENTRYPOINT ["./create-rootfs.sh"]