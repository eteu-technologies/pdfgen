#!/usr/bin/env bash
set -euo pipefail
: "${DOCKER_REGISTRY:=docker.io}"
: "${DOCKER_REPOSITORY:=eteu/pdfgen}"
: "${IMAGE_COMMIT_TAG:=commit-"$(git rev-parse HEAD || echo "0000000000000000000000000000000000000000")"}"

dockerfiles=(
    ./docker/Dockerfile.amd64
    ./docker/Dockerfile.arm64
)

built_imgs=()
for f in "${dockerfiles[@]}"; do
    f="$(basename -- "${f}")"
    arch="${f/Dockerfile\./}"
    echo ">>> Building image for '${arch}'"
    ./docker/build_docker.sh "${arch}" --load || {
        echo ">>> Failed to build image for arch '${arch}'"
        exit 1
    }

    # tag image with git commit id & latest tags
    tag="${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:${IMAGE_COMMIT_TAG}-${arch}"
    docker image tag "eteu/pdfgen:${arch}" "${tag}"
    built_imgs+=("${tag}")
done

# push all the images
for img in "${built_imgs[@]}"; do
    docker image push "${img}"
done

# create a manifrst
docker manifest create --insecure "${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:latest" \
    "${built_imgs[@]}"
docker manifest create --insecure "${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:${IMAGE_COMMIT_TAG}" \
    "${built_imgs[@]}"

# push
docker manifest push --purge --insecure "${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:latest"
docker manifest push --purge --insecure "${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:${IMAGE_COMMIT_TAG}"
