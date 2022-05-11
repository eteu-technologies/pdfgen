#!/usr/bin/env bash
set -euo pipefail

export PROJ_VERSION="$(git rev-parse HEAD || echo "0000000000000000000000000000000000000000")"

arch="${1:-}"
hostarch="$(uname -m)"
if [ -z "${arch:-}" ]; then
	arch="$(uname -m)"
else
	shift
fi

fixup_arch () {
	case "${1}" in
		x86_64)
			echo "amd64"
			;;
		armv8*)
			echo "arm64"
			;;
		aarch64)
			echo "arm64"
			;;
		*)
			echo "${1}"
			;;
	esac
}

arch="$(fixup_arch "${arch}")"
hostarch="$(fixup_arch "${hostarch}")"

docker buildx build \
	--platform "linux/${arch}" \
	--build-arg PROJ_VERSION="${PROJ_VERSION}" \
	"${@}" \
	-t eteu/pdfgen:"${arch}" \
	-f docker/Dockerfile .

if [ "${hostarch}" = "${arch}" ]; then
	docker tag eteu/pdfgen:"${arch}" eteu/pdfgen:latest
fi
