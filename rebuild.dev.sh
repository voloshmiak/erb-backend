#!/bin/bash

set -e

docker build -f Dockerfile.dev -t "erb-backend-dev" .