#!/bin/bash
set -e
bazelisk test //...
bazelisk run //:lint
