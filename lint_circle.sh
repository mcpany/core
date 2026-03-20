#!/bin/bash
export PATH=$PWD/build/env/bin:$PATH
make -C server lint
