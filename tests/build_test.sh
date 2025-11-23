#!/bin/bash
# Copyright 2024 Author(s) of MCP Any
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

# Get the short hash of the current git commit.
GIT_COMMIT=$(git rev-parse --short HEAD)
export GIT_COMMIT

# The name of the image to be built.
IMAGE_NAME="mcpany-server-build-test"
IMAGE_TAG=$GIT_COMMIT
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

echo "Building mcpany-server image with tag: $IMAGE_TAG"

# Build the mcpany-server image using docker-compose.
docker compose build mcpany-server

# Check if the image with the git commit tag was created.
echo "Checking for the existence of the built image: $FULL_IMAGE_NAME"
if docker images | grep -q "$IMAGE_NAME\s*$IMAGE_TAG"; then
  echo "Build test passed! Image $FULL_IMAGE_NAME found."
  # Clean up the built image.
  docker rmi "$FULL_IMAGE_NAME"
  echo "Cleaned up the built image."
else
  echo "Build test failed! Image $FULL_IMAGE_NAME not found."
  exit 1
fi
