#!/bin/bash

set -e

IMAGE_NAME="erb-backend-dev"
CONTAINER_NAME="erb-backend"

docker rm -f "$CONTAINER_NAME" 2>/dev/null || true

docker run -it \
  --name "$CONTAINER_NAME" \
  --env-file .env \
  -p 8080:8080 \
  -v "$(pwd)":/app \
  "$IMAGE_NAME"