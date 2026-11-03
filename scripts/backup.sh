#!/usr/bin/env sh
set -eu
: "${DATABASE_URL:?Set the source database connection securely}"
: "${BACKUP_FILE:?Set a private backup output path}"
umask 077
pg_dump --dbname="$DATABASE_URL" --format=custom --file="$BACKUP_FILE" --no-owner --no-acl
# Store this file in private encrypted storage; do not commit it.
