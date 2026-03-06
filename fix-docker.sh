#!/bin/bash
sudo systemctl stop docker
sudo rm -rf /var/lib/docker/overlay2
sudo systemctl start docker
