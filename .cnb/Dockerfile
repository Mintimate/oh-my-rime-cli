FROM node:24-bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    git \
    jq \
    minisign \
  && rm -rf /var/lib/apt/lists/*

RUN npm install --global @cnbcool/cnb-cli@1.10.9
