#!/bin/sh
set -eu

WORKSPACE="/workspace"
SOURCE_FILE="${WORKSPACE}/Main.java"
TIMEOUT_SECONDS="${EXEC_TIMEOUT_SECONDS:-10}"

if [ -z "${CODE_B64:-}" ]; then
  echo "CODE_B64 is required" >&2
  exit 2
fi

mkdir -p "${WORKSPACE}"
printf '%s' "${CODE_B64}" | base64 -d > "${SOURCE_FILE}"

javac "${SOURCE_FILE}"

set +e
timeout "${TIMEOUT_SECONDS}" java -cp "${WORKSPACE}" Main
STATUS="$?"
set -e

if [ "${STATUS}" -eq 124 ]; then
  echo "execution timed out after ${TIMEOUT_SECONDS} seconds" >&2
fi

exit "${STATUS}"
