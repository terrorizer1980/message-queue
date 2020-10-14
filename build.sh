#!/bin/bash -xe

docker build -t quay.io/mullvad/message-queue:$(cat VERSION) .
