#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: Apache-2.0

set -e

# Directory where images are cached
CACHE_DIR=~/docker-images
mkdir -p "$CACHE_DIR"

usage() {
  echo "Usage: $0 [load|save]"
  echo "  load: Load images from cache directory to Docker"
  echo "  save: Save all current Docker images to cache directory"
  exit 1
}

MODE=${1:-load}

if [ "$MODE" == "load" ]; then
  echo "Loading images from $CACHE_DIR..."
  if [ -d "$CACHE_DIR" ]; then
    shopt -s nullglob
    files=("$CACHE_DIR"/*.tar)
    if [ ${#files[@]} -eq 0 ]; then
      echo "No cached images found in $CACHE_DIR."
    else
      for filepath in "${files[@]}"; do
        echo "Loading $filepath..."
        docker load -i "$filepath" || echo "Warning: Failed to load $filepath"
      done
    fi
  fi

elif [ "$MODE" == "save" ]; then
  echo "Saving images to $CACHE_DIR..."

  # Get list of all images, excluding dangling ones
  mapfile -t IMAGES < <(docker images --format "{{.Repository}}:{{.Tag}}" | grep -v "<none>" || true)

  if [ ${#IMAGES[@]} -eq 0 ]; then
    echo "No images to save."
    exit 0
  fi

  echo "Found ${#IMAGES[@]} images to save."

  # Clear cache directory to remove old individual tars and avoid duplicates
  # Use :? to ensure CACHE_DIR is set
  rm -rf "${CACHE_DIR:?}"/*

  # Save all images to a single tarball for layer deduplication
  # We verify space separated list works for docker save
  echo "Saving ${#IMAGES[@]} images to ${CACHE_DIR}/all_images.tar..."
  docker save -o "${CACHE_DIR}/all_images.tar" "${IMAGES[@]}"

else
  usage
fi
