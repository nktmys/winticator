FROM golang:1.25

ARG USERNAME=develop
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN apt-get update && \
  apt-get install -y --no-install-recommends sudo \
    libgl1-mesa-dev xorg-dev libxkbcommon-dev xz-utils \
    apt-transport-https ca-certificates gnupg && \
  apt-get clean && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid $USER_GID $USERNAME  && \
  useradd --uid $USER_UID --gid $USER_GID -m $USERNAME && \
  echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME && \
  chmod 0440 /etc/sudoers.d/$USERNAME

  # Create and change to the app directory.
WORKDIR /app

# Install go packages for development.
RUN curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to /usr/local/bin

USER $USERNAME
