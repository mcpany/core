#!/bin/bash
export PATH=$PATH:~/.local/bin
bazelisk test //ui:lint
bazelisk test //server/...
