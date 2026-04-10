# ADR 006 — Database Backup Strategy

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-12 |

## Context

The application stores personal data under GDPR obligations. Art. 32 requires appropriate
technical measures to ensure recoverability. The entire dataset lives in a PostgreSQL volume
on a single OVH VPS. A disk failure, accidental volume deletion, or botched migration has no
recovery path without a backup strategy.

OHM's scale: ~120 members, low transaction rate, data changes slowly (events, RSVPs, fee
payments). The association is cost-sensitive.

---

## Alternatives Considered

**A — OVH Managed PostgreSQL (rejected)**

Automatic backups and point-in-time recovery (PITR) with zero operational burden. RPO:
minutes. Estimated cost: +€10–15/month — roughly doubling the current hosting cost. Rejected
on cost grounds; the operational benefit does not justify the expense at OHM's scale.

**B — WAL archiving + pg_basebackup (rejected)**

Continuous archiving to object storage; enables PITR. RPO: minutes to seconds. High setup
and maintenance complexity. Overkill for a system with low transaction rates and no
sub-hourly recovery requirement.

**C — OVH VPS volume snapshot (rejected as primary)**

OVH offers automated VPS snapshots (~€1–2/month). Volume-level, not database-aware: a
consistent snapshot requires the container to be stopped, and granular table-level restore is
not possible. Acceptable as a secondary safeguard but insufficient as the primary backup
mechanism.

**D — Daily `pg_dump` to OVH Object Storage (chosen)**

A cron job runs `pg_dump` daily, compresses the output, and uploads it to OVH Object Storage
(S3-compatible). RPO: 24 hours. Storage cost: negligible (a few MB per dump at
€0.01/GB/month). Restore is two commands. Well-understood, operationally simple, and
sufficient for OHM's tolerance for data loss.

---

## Decision

Daily `pg_dump` exported to OVH Object Storage. Retention: keep the last 30 dumps (rolling).

**Backup procedure (automated via cron on the VPS):**

```sh
#!/bin/sh
TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
DUMP_FILE="ohm-backup-${TIMESTAMP}.sql.gz"

docker compose exec -T db pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" \
  | gzip \
  | aws s3 cp - "s3://${BACKUP_BUCKET}/${DUMP_FILE}" \
      --endpoint-url "${OVH_S3_ENDPOINT}"

# Prune dumps older than 30 days
aws s3 ls "s3://${BACKUP_BUCKET}/" --endpoint-url "${OVH_S3_ENDPOINT}" \
  | awk '{print $4}' \
  | sort \
  | head -n -30 \
  | xargs -I{} aws s3 rm "s3://${BACKUP_BUCKET}/{}" \
      --endpoint-url "${OVH_S3_ENDPOINT}"
```

The script is stored in the repository under `scripts/backup.sh`. The cron entry runs daily
at 03:00 UTC (low-traffic window):

```
0 3 * * * /srv/ohm/scripts/backup.sh >> /var/log/ohm-backup.log 2>&1
```

**Required environment variables** (set in operator shell profile on the VPS, not committed):

| Variable | Purpose |
|----------|---------|
| `POSTGRES_USER` | PostgreSQL user |
| `POSTGRES_DB` | Database name |
| `BACKUP_BUCKET` | OVH Object Storage bucket name |
| `OVH_S3_ENDPOINT` | OVH Object Storage S3 endpoint URL |
| `AWS_ACCESS_KEY_ID` | OVH Object Storage access key |
| `AWS_SECRET_ACCESS_KEY` | OVH Object Storage secret key |

**Restore procedure:**

```sh
# 1. Download the desired dump
aws s3 cp "s3://${BACKUP_BUCKET}/ohm-backup-<TIMESTAMP>.sql.gz" - \
    --endpoint-url "${OVH_S3_ENDPOINT}" \
  | gunzip \
  | docker compose exec -T db psql -U "$POSTGRES_USER" "$POSTGRES_DB"

# 2. Re-run migrations if restoring to an older schema version
docker compose exec app buffalo pop migrate
```

---

## Trade-offs

| Concern | Impact |
|---------|--------|
| RPO = 24 hours | Up to one day of data loss on failure. Acceptable for OHM's activity cadence (weekly rehearsals, infrequent admin operations). |
| Manual restore | Restore requires SSH access to the VPS and knowledge of the procedure. Documented above; no automation required at this scale. |
| No PITR | Cannot restore to an arbitrary point in time. The 30-dump rolling window provides restore points at daily granularity. |
| Silent cron failure | If the cron job fails silently, backups stop without alerting anyone. Mitigate by checking the backup log periodically or adding a health check (alert if no new object appears in the bucket after 48 hours). |

---

## Consequences

- `scripts/backup.sh` must be created in the repository and deployed to the VPS.
- OVH Object Storage bucket must be provisioned before first deployment.
- Backup credentials must be added to the operator's environment on the VPS.
- Cron entry must be registered on the VPS as part of initial setup.
- Backup log should be reviewed periodically; no automated alerting in V1.
