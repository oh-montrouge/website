#!/bin/sh
set -eu

# Backup-specific credentials live in /srv/ohm/.env.backup on the VPS.
# See docs/vps-setup.md for the required variables.
ENV_FILE="/srv/ohm/.env.backup"
if [ -f "$ENV_FILE" ]; then
    # shellcheck disable=SC1090
    . "$ENV_FILE"
fi

TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
DUMP_FILE="ohm-backup-${TIMESTAMP}.sql.gz"

docker compose -f /srv/ohm/docker-compose.yml exec -T postgres \
    pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" \
    | gzip \
    | aws s3 cp - "s3://${BACKUP_BUCKET}/${DUMP_FILE}" \
        --endpoint-url "${OVH_S3_ENDPOINT}"

# Keep the 30 most recent dumps; prune the rest
aws s3 ls "s3://${BACKUP_BUCKET}/" --endpoint-url "${OVH_S3_ENDPOINT}" \
    | awk '{print $4}' \
    | sort \
    | head -n -30 \
    | xargs -r -I{} aws s3 rm "s3://${BACKUP_BUCKET}/{}" \
        --endpoint-url "${OVH_S3_ENDPOINT}"
