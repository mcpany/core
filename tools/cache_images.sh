#!/bin/bash
set -e

# Directory for cached images
CACHE_DIR="docker-images"
mkdir -p "$CACHE_DIR"

if [ "$#" -gt 0 ]; then
  IMAGES=("$@")
else
  # Default list if no arguments provided
  IMAGES=(
    "redis:latest"
    "python:3.11-slim"
    "golang:latest"
    "node:lts-alpine"
    "gcr.io/distroless/base-debian12"
    "fullstorydev/grpcurl"
  )
fi

echo "Starting docker image caching/loading for: ${IMAGES[*]}"

for img in "${IMAGES[@]}"; do
  # Replace / and : with _ for filename
  sanitized_name=$(echo "$img" | tr '/:' '_')
  tarball="$CACHE_DIR/${sanitized_name}.tar"

  if [ -f "$tarball" ]; then
    echo "Found cached tarball for $img at $tarball. Loading..."
    docker load -i "$tarball"
  else
    echo "No cached tarball for $img. Checking if pulling is needed..."
    # Check if image exists locally (e.g. from previous steps, though unlikely in fresh runner)
    if docker image inspect "$img" > /dev/null 2>&1; then
        echo "Image $img already exists locally. Saving to cache..."
        docker save "$img" -o "$tarball"
    else
        echo "Pulling $img..."
        docker pull "$img"
        echo "Saving $img to $tarball..."
        docker save "$img" -o "$tarball"
    fi
  fi
done

echo "Docker image caching/loading complete."
