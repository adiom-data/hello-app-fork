#!/usr/bin/env bash
set -euo pipefail

git_commit="$(git rev-parse --short HEAD 2>/dev/null || printf 'local')"
preview_reference_prefix="${STABLE_PREVIEW_REFERENCE_PREFIX:-${PREVIEW_REFERENCE_PREFIX:-}}"
preview_tag="${STABLE_PREVIEW_TAG:-${PREVIEW_TAG:-}}"

if [[ -z "$preview_reference_prefix" ]]; then
  preview_reference_prefix="{STABLE_PREVIEW_REFERENCE_PREFIX}"
fi

if [[ -z "$preview_tag" ]]; then
  preview_tag="{STABLE_PREVIEW_TAG}"
fi

printf 'STABLE_GIT_COMMIT %s\n' "$git_commit"
printf 'STABLE_PREVIEW_REFERENCE_PREFIX %s\n' "$preview_reference_prefix"
printf 'STABLE_PREVIEW_TAG %s\n' "$preview_tag"
