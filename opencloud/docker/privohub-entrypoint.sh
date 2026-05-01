#!/bin/sh
# PrivoHub entrypoint: ensure a self-managed IDM admin password file exists on the data
# PVC before launching the OpenCloud server. The password is generated once on first boot
# and read by the IDM service via PRIVOHUB_IDM_ADMIN_PASSWORD_FILE on every subsequent
# start. The value is never echoed; we only stat/chmod the file.
set -eu

PRIV_DIR=/var/lib/opencloud/.privohub
PRIV_FILE="${PRIV_DIR}/idm-admin"

if [ ! -f "${PRIV_FILE}" ]; then
    mkdir -p "${PRIV_DIR}"
    chmod 0700 "${PRIV_DIR}"
    # od is in alpine coreutils; xxd is not. 32 bytes -> 64 hex chars.
    od -An -N32 -tx1 -v /dev/urandom | tr -d ' \n' > "${PRIV_FILE}"
    chmod 0600 "${PRIV_FILE}"
fi

exec /usr/bin/opencloud "$@"
