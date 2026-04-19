# VPS Setup and Deployment Runbook

One-time setup for the OVH VPS. Run these steps in order on first deployment.

---

## Prerequisites

On the VPS, install:
- Docker Engine (v24+) and the Compose plugin — https://docs.docker.com/engine/install/ubuntu/
- AWS CLI v2 (used by the backup script) — https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html

On your local machine:
- See "Local shell profile" section below for required environment variables.

---

## Use password authentication to add your SSH key

TODO: use password to connect to SSH, then create and register your key as described in the later "Onboarding a new deployer"
If all authorized machines are lost, we'll still be able to use OVH KVM console as a recovery path.

---

## SSH hardening

```sh
# On the VPS — disable password auth after confirming key-based login works
sudo sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo systemctl reload sshd
```

```sh
# Firewall: allow SSH, HTTP, HTTPS only
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable
```

---

## Repository and environment files

```sh
# Clone the repository
sudo mkdir -p /srv/ohm
sudo chown "$USER:$USER" /srv/ohm
git clone <repo-url> /srv/ohm
cd /srv/ohm
```

Create `/srv/ohm/.env` from the example — fill in all `CHANGE_ME` values:

```sh
cp .env.production.example .env
# Edit .env: set DATABASE_URL password, SESSION_SECRET, APP_DOMAIN, APP_IMAGE
```

Generate a strong SESSION_SECRET:
```sh
openssl rand -base64 32
```

Create `/srv/ohm/.env.backup` for the backup script (not committed, separate from `.env`):

```sh
cat > /srv/ohm/.env.backup << 'EOF'
POSTGRES_USER=ohm
POSTGRES_DB=ohm_production
BACKUP_BUCKET=ohm-website
OVH_S3_ENDPOINT=https://s3.eu-west-par.io.cloud.ovh.net
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
EOF
chmod 600 /srv/ohm/.env.backup
```

---

## Object Storage bucket

In the OVH console:
1. Create an Object Storage container (S3-compatible).
2. Set lifecycle policy to 30-day expiry, or rely on the script's manual pruning.
3. Create S3 credentials (access key + secret key) and record them in `.env.backup`.

---

## Local shell profile (required for deploy tooling)

`DEPLOY_HOST` and `IMAGE_REPO` are already set in `mise.toml` — no local config needed for those.

You only need a `GITHUB_TOKEN` to push images to the GitHub Container Registry.

**Create a GitHub classic token:**
1. Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Select scope: `write:packages` (includes `read:packages`)
4. Copy the token

Export it in your shell profile (`~/.bashrc`, `~/.zshrc`, or a `mise.local.toml`):

```toml
# mise.local.toml (gitignored — safe for secrets)
[env]
GITHUB_TOKEN = "ghp_your_token_here"
```

Log in to the registry once:
```sh
echo $GITHUB_TOKEN | docker login ghcr.io -u <your-github-username> --password-stdin
```

---

## First deployment

```sh
# On your local machine — build, tag, and push the image
mise run publish

# Deploy (SSH → pull → up → migrate)
mise run deploy
```

---

## Bootstrap the admin account

```sh
# On the VPS
cd /srv/ohm
docker compose exec app buffalo task db:seed:admin
```

Or via Mise from your local machine (requires DEPLOY_HOST set):
```sh
ssh $DEPLOY_HOST 'cd /srv/ohm && docker compose exec -T app buffalo task db:seed:admin'
```

---

## Backup cron job

Make the script executable and register the cron entry:

```sh
chmod +x /srv/ohm/scripts/backup.sh

# Add to crontab (runs daily at 03:00 UTC)
(crontab -l 2>/dev/null; echo "0 3 * * * /srv/ohm/scripts/backup.sh >> /var/log/ohm-backup.log 2>&1") | crontab -
```
Or `0 3 * * 0` if you want it weekly on Sunday.

Test the script runs without error:
```sh
/srv/ohm/scripts/backup.sh
```

Confirm an `.sql.gz` object appears in your OVH Object Storage bucket (AC-M5).

---

## Smoke test checklist

Run these after first deployment to confirm AC-M3 through AC-H3:

- [ ] `docker compose exec app buffalo-pop pop migrate status` shows all migrations applied (AC-M3)
- [ ] `curl -I https://{domain}/connexion` returns `HTTP/2 200` (AC-M4)
- [ ] `curl -I http://{domain}/connexion` redirects to `https://` (AC-M4)
- [ ] `bash scripts/backup.sh` exits 0 and an `.sql.gz` appears in the bucket (AC-M5)
- [ ] Log in with the seed-admin account, navigate two protected pages, log out — session holds (AC-H1)
- [ ] `docker run --rm $APP_IMAGE env` does not print `SESSION_SECRET` or `DATABASE_URL` (AC-H2)
- [ ] OVH Object Storage lifecycle policy shows 30-day retention (AC-H3)

---

## Day-2 operations

### Deploying a new feature

```sh
# 1. Build, tag :latest and :<git-sha>, push both
mise run publish

# 2. SSH to VPS, pull, restart, migrate
mise run deploy
```

`publish` always tags the image with the current git SHA. You can confirm what's running
in production at any time:
```sh
ssh $DEPLOY_HOST 'docker inspect $(docker compose -f /srv/ohm/docker-compose.yml ps -q app) --format "{{.Config.Image}}"'
```

### Rollback

**Current approach — revert code and republish:**

If the new version has a bug, revert the commit and re-run the deploy pipeline:
```sh
git revert HEAD          # or git reset, depending on the situation
mise run publish
mise run deploy
```

Since `docker compose up -d` pulls `:latest`, the VPS will pick up the reverted image.
If the migration from the bad deploy was additive and harmless (a new nullable column, a new
table), `buffalo-pop pop migrate` is a no-op on redeploy and nothing needs to be done about
the schema.

**Rollback to a specific previous SHA (without reverting code):**

Each `mise run publish` pushes `IMAGE_REPO:<sha>` alongside `:latest`. To roll back to a
specific version without touching the code:
```sh
# On the VPS — edit .env to pin APP_IMAGE to the old SHA
ssh $DEPLOY_HOST "sed -i 's|APP_IMAGE=.*|APP_IMAGE=${IMAGE_REPO}:<old-sha>|' /srv/ohm/.env"
ssh $DEPLOY_HOST 'cd /srv/ohm && docker compose pull && docker compose up -d'
# Note: skip migration step — schema already matches the old image
```

**Schema rollback (hard case):**

If the bad deploy included a destructive migration (dropped a column, changed a type),
rolling back the image is not enough — you also need `buffalo-pop pop migrate down` to undo
the schema change. This is risky (data loss possible) and must be done manually with a
verified backup in hand. This case is not scripted intentionally.

Future improvement: a `mise run rollback SHA=<sha>` task could automate the image-only
rollback path and refuse to run if a schema migration was part of the bad deploy.

---

### Onboarding a new deployer

A new team member who needs to deploy follows these steps:

**1. Generate an SSH key (if they don't have one):**
```sh
ssh-keygen -t ed25519 -C "their-email@example.com"
```

**2. Send their public key to the VPS admin:**
```sh
cat ~/.ssh/id_ed25519.pub
```

**3. VPS admin adds it (on the VPS):**
```sh
echo "<their-public-key>" >> ~/.ssh/authorized_keys
```

**4. Create a GitHub classic token:**
1. Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Select scope: `write:packages` (includes `read:packages`)
4. Copy the token

Export it in your shell profile (`~/.bashrc`, `~/.zshrc`, or a `mise.local.toml`):

```toml
# mise.local.toml (gitignored — safe for secrets)
[env]
GITHUB_TOKEN = "ghp_your_token_here"
```

**5. Log in to the Docker registry:**
```sh
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin
```

**6. Verify:**
```sh
ssh $DEPLOY_HOST 'echo "SSH OK"'
mise run publish  # should build and push without errors
```

**Troubleshooting — SSH still asks for a password after adding the key:**

SSH may not pick up a non-default key name automatically. Quick fix for the current session:
```sh
ssh-add ~/.ssh/id_ed25519
```

Permanent fix — add a `~/.ssh/config` entry so SSH always uses the right key for this host:
```
Host <vps-ip>
    IdentityFile ~/.ssh/id_ed25519
    AddKeysToAgent yes
```
