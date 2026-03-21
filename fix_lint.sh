#!/bin/bash
export PATH=$PATH:~/.local/bin
cd server
bazelisk test //ui:lint
bazelisk test //server/...
