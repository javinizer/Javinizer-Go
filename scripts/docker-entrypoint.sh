#!/bin/sh
set -e

container_uid="$(id -u)"
container_gid="$(id -g)"
requested_uid="${PUID:-${USER_ID:-}}"
requested_gid="${PGID:-${GROUP_ID:-}}"

if [ -n "${requested_uid}" ] && [ "${requested_uid}" != "${container_uid}" ]; then
    echo "WARNING: Requested UID (${requested_uid}) does not match runtime UID (${container_uid})."
    echo "         Ensure container is started with matching user mapping."
fi
if [ -n "${requested_gid}" ] && [ "${requested_gid}" != "${container_gid}" ]; then
    echo "WARNING: Requested GID (${requested_gid}) does not match runtime GID (${container_gid})."
    echo "         Ensure container is started with matching user mapping."
fi

# Preflight write checks for mounted state directory.
if [ ! -d "/javinizer" ]; then
    echo "ERROR: /javinizer does not exist. Check your volume mapping."
    exit 1
fi

if ! mkdir -p /javinizer/logs /javinizer/cache; then
    echo "ERROR: Unable to create /javinizer/logs or /javinizer/cache."
    echo "       Container is running as uid=${container_uid} gid=${container_gid}."
    echo "       On Unraid, set PUID/PGID (or USER_ID/GROUP_ID) to match share ownership."
    exit 1
fi

javinizer_probe="/javinizer/.javinizer-write-test.$$"
if ! (umask 077 && : > "${javinizer_probe}") 2>/dev/null; then
    echo "ERROR: /javinizer is not writable by uid=${container_uid} gid=${container_gid}."
    echo "       Fix directory ownership/permissions or adjust PUID/PGID."
    exit 1
fi
rm -f "${javinizer_probe}" 2>/dev/null || true

# /media may be intentionally read-only for scan-only usage. Warn instead of failing.
if [ -d "/media" ]; then
    media_writable=0
    media_probe="/media/.javinizer-write-test.$$"
    if (umask 077 && : > "${media_probe}") 2>/dev/null; then
        rm -f "${media_probe}" 2>/dev/null || true
        media_writable=1
    else
        # Some environments keep /media root read-only while allowing writes
        # in specific subdirectories. Check existing children before warning.
        for media_dir in /media/*; do
            [ -d "${media_dir}" ] || continue
            media_probe="${media_dir}/.javinizer-write-test.$$"
            if (umask 077 && : > "${media_probe}") 2>/dev/null; then
                rm -f "${media_probe}" 2>/dev/null || true
                media_writable=1
                break
            fi
        done
    fi

    if [ "${media_writable}" -eq 0 ]; then
        echo "WARNING: /media is not writable by uid=${container_uid} gid=${container_gid}."
        echo "         Scan/review works, but organize/move/copy operations may fail."
    fi
fi

# Copy default config if it doesn't exist
if [ ! -f "/javinizer/config.yaml" ]; then
    echo "No config file found, copying default configuration..."
    cp /app/config/config.yaml.default /javinizer/config.yaml
    echo "✓ Default configuration created at /javinizer/config.yaml"
fi

# Execute the main command
exec "$@"
