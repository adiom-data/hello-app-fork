#!/usr/bin/env bash
set -euo pipefail

git_commit="$(git rev-parse --short HEAD 2>/dev/null || printf 'local')"
registry_proxy="${REGISTRY_PROXY:-http://git-proxy:18089}"
registry_host="${registry_proxy#http://}"
registry_host="${registry_host#https://}"
repository_prefix="${REPOSITORY_PREFIX:-previews/4218776c-057b-44f8-aed6-238c199e8903/4e71d6f5-77b9-4d3a-8afd-dd261cc1b338}"
preview_prefix="${registry_host%/}/${repository_prefix#/}"
reference_prefix="${REFERENCE_PREFIX:-$preview_prefix}"
push_prefix="${PUSH_PREFIX:-$preview_prefix}"
artifact_prefix="${ARTIFACT_PREFIX:-$push_prefix}"
tag="${TAG:-session-4e71d6f5-77b9-4d3a-8afd-dd261cc1b338}"

printf 'STABLE_GIT_COMMIT %s\n' "$git_commit"
printf 'STABLE_REFERENCE_PREFIX %s\n' "$reference_prefix"
printf 'STABLE_PUSH_PREFIX %s\n' "$push_prefix"
printf 'STABLE_ARTIFACT_PREFIX %s\n' "$artifact_prefix"
printf 'STABLE_TAG %s\n' "$tag"
