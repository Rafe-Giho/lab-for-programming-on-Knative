#!/bin/sh
set -eu

WORKSPACE="/workspace"
TMP_DIR="${WORKSPACE}/tmp"
SOURCE_FILE="${WORKSPACE}/main.cpp"
BINARY_FILE="${WORKSPACE}/main"
TIMEOUT_SECONDS="${EXEC_TIMEOUT_SECONDS:-10}"

if [ -z "${CODE_B64:-}" ]; then
  echo "CODE_B64 is required" >&2
  exit 2
fi

mkdir -p "${WORKSPACE}" "${TMP_DIR}"
export TMPDIR="${TMP_DIR}"
cd "${WORKSPACE}"
printf '%s' "${CODE_B64}" | base64 -d > "${SOURCE_FILE}"

g++ -O2 -pipe -std=c++17 "${SOURCE_FILE}" -o "${BINARY_FILE}"

set +e
timeout "${TIMEOUT_SECONDS}" "${BINARY_FILE}"
STATUS="$?"
set -e

if [ "${STATUS}" -eq 124 ]; then
  echo "execution timed out after ${TIMEOUT_SECONDS} seconds" >&2
fi

exit "${STATUS}"
