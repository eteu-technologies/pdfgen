#!/usr/bin/env bash
set -euo pipefail

export PROJ_VERSION="$(git describe --tags --abbrev 2>/dev/null || echo "0.0.0-dev")"
export DOCKER_BUILDKIT=1

docker build \
        --build-arg PROJ_VERSION="${PROJ_VERSION}" \
	"${@}" \
        -t eteu/pdfgen \
        -f docker/Dockerfile .
