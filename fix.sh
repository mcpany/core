#!/bin/bash
export PATH=$(pwd)/build/env/bin:$PATH
protoc-gen-go --version
protoc-gen-go-grpc --version
