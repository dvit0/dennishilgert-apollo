FROM debian:stable

WORKDIR /workspace

RUN apt update && apt install -y \
  curl \
  ca-certificates \
  gnupg \
  build-essential \
  git

# Install docker engine
RUN install -m 0755 -d /etc/apt/keyrings
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
RUN chmod a+r /etc/apt/keyrings/docker.gpg
RUN echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null
RUN apt update && apt install -y \
  docker-ce \
  docker-ce-cli \
  containerd.io \
  docker-buildx-plugin \
  docker-compose-plugin

# Remove apt caches
RUN rm -rf /var/lib/apt/lists/*

COPY Dockerfile.rootfs Dockerfile.rootfs

COPY scripts/create-rootfs.sh create-rootfs.sh
RUN chmod +x create-rootfs.sh

COPY scripts/copy-rootfs.sh copy-rootfs.sh
RUN chmod +x copy-rootfs.sh

COPY scripts/agent.sh agent.sh
RUN chmod +x agent.sh

VOLUME /dist

ENTRYPOINT [ "./create-rootfs.sh" ]