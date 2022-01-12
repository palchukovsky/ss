# Copyright 2021-2022, the SS project owners. All rights reserved.
# Please see the OWNERS and LICENSE files for details.

FROM docker

ARG CONFIG

WORKDIR /usr/src/app

COPY . .

# Upgrades the system as no specific version of OS is set.
RUN \
  apk update && \
  apk upgrade && \
  apk add build-base curl git go

ENV PATH=${PATH}:/root/go/bin
