# Deployment Setup

How production deploys work for Artmission, and how to (re)configure them from scratch. Backend deploys to **Cloud Run**, frontend deploys to **Vercel**, both triggered by **GitHub Actions** on push to `main`.

## Why GitHub Actions instead of native Git integrations

Neither Vercel's Git integration nor GCP Cloud Build's GitHub trigger can be used here: both require installing a GitHub App on this specific repository, which needs repo **owner** access.

GitHub Actions sidesteps this:

- **Vercel**: deployed via the Vercel CLI (`vercel build` / `vercel deploy`) using an API token. No Git integration, no GitHub App, no repo permission needed on Vercel's side at all.
- **GCP**: deployed via `gcloud`/`google-github-actions` using **Workload Identity Federation (WIF)** — GitHub's OIDC token exchanged for short-lived GCP credentials. No GitHub App, no stored JSON key.

The only unavoidable GitHub-side requirement is adding repo Secrets/Variables (Settings → Secrets and variables → Actions), which itself needs `admin` — ask the repo owner to paste the values below once, or to grant `Admin` so you can add them yourself.

### Credential design

GitHub Secrets hold every true secret: `VERCEL_TOKEN`, `DATABASE_DSN`, `AUTH_JWT_SECRET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`. Everything GCP-related is a non-sensitive identifier (GitHub Variables). The workflow injects the backend secrets directly as Cloud Run environment variables at deploy time; Google Secret Manager is not used.

## Prerequisites

- [`gcloud` CLI](https://docs.cloud.google.com/sdk/docs/install-sdk), authenticated as a GCP project owner/editor.
- [`vercel` CLI](https://vercel.com/docs/cli), authenticated to your Vercel account/team.
- Repo `admin` access on the repo to add Actions Secrets/Variables.

## Part 1 — GCP setup

Run once per environment (e.g. once for production).

```bash
PROJECT_ID=your-gcp-project-id
REGION=asia-southeast3          # pick your region
REPO_OWNER=AiSiriRak
REPO_NAME=Artmission

gcloud config set project "$PROJECT_ID"
gcloud auth application-default set-quota-project "$PROJECT_ID"

# 1. Enable required APIs
gcloud services enable run.googleapis.com artifactregistry.googleapis.com \
  iamcredentials.googleapis.com

# 2. Artifact Registry repo for backend Docker images
gcloud artifacts repositories create artmission \
  --repository-format=docker --location="$REGION" \
  --description="Artmission images"

# 3. Deploy service account — the identity GitHub Actions impersonates via WIF.
#    Can build/push images and deploy to Cloud Run.
gcloud iam service-accounts create gh-deployer --display-name="GitHub Actions deployer"
DEPLOY_SA="gh-deployer@$PROJECT_ID.iam.gserviceaccount.com"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$DEPLOY_SA" --role="roles/run.admin" --condition=None
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$DEPLOY_SA" --role="roles/artifactregistry.writer" --condition=None

# 4. Runtime service account — what the Cloud Run *service* runs as. It has no
#    roles: the backend needs no GCP APIs (DB and S3 are external), and a bare
#    SA avoids the default compute SA's broad Editor role.
gcloud iam service-accounts create artmission-backend-run \
  --display-name="Backend Cloud Run runtime"
RUN_SA="artmission-backend-run@$PROJECT_ID.iam.gserviceaccount.com"

# deploy SA needs permission to deploy the service AS the runtime SA
gcloud iam service-accounts add-iam-policy-binding "$RUN_SA" \
  --member="serviceAccount:$DEPLOY_SA" --role="roles/iam.serviceAccountUser" --condition=None

# 5. Workload Identity Federation — lets GitHub Actions authenticate with zero
#    stored keys. The attribute-condition restricts token exchange to this repo only.
gcloud iam workload-identity-pools create github-pool \
  --location="global" --display-name="GitHub Actions"

gcloud iam workload-identity-pools providers create-oidc github-provider \
  --location="global" --workload-identity-pool="github-pool" \
  --display-name="GitHub OIDC" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository=='$REPO_OWNER/$REPO_NAME'" \
  --issuer-uri="https://token.actions.githubusercontent.com"

PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")

# only this repo's workflows can impersonate the deploy SA
gcloud iam service-accounts add-iam-policy-binding "$DEPLOY_SA" \
  --role="roles/iam.workloadIdentityUser" --condition=None \
  --member="principalSet://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/github-pool/attribute.repository/$REPO_OWNER/$REPO_NAME"

# 6. Print the provider resource name -> this is the GCP_WIF_PROVIDER value
gcloud iam workload-identity-pools providers describe github-provider \
  --location=global --workload-identity-pool=github-pool --format="value(name)"
```

## Part 2 — Vercel setup

```bash
cd frontend
# Ensure you already installed vercel CLI
vercel login                 # your own account/team — repo access is irrelevant
vercel link                  # creates an empty project, prompts for scope + name
cat .vercel/project.json     # -> { "projectId": "...", "orgId": "..." }
```

Create a token: Vercel dashboard → Account Settings → **Tokens** → Create Token, named e.g. `gh-actions-artmission`. **Scope it to the team** — the scope dropdown also offers individual projects, but a project-scoped token cannot satisfy `vercel pull`'s internal team-info lookup and fails with a misleading `Could not retrieve Project Settings` error.

Once the backend is deployed and its Cloud Run URL is known, set it as a Vercel project env var:

```bash
vercel env add NEXT_PUBLIC_API_BASE_URL production
```

## Part 3 — GitHub repo configuration

Settings → Secrets and variables → Actions.

**Variables** tab:

| Name | Value |
|---|---|
| `GCP_PROJECT_ID` | your GCP project id |
| `GCP_REGION` | e.g. `asia-southeast3` |
| `GCP_WIF_PROVIDER` | output of GCP step 6, `projects/.../providers/github-provider` |
| `GCP_SERVICE_ACCOUNT` | `gh-deployer@PROJECT_ID.iam.gserviceaccount.com` |
| `GCP_RUN_SERVICE_ACCOUNT` | `artmission-backend-run@PROJECT_ID.iam.gserviceaccount.com` |
| `VERCEL_ORG_ID` | from `.vercel/project.json` |
| `VERCEL_PROJECT_ID` | from `.vercel/project.json` |
| `APP_ALLOWED_ORIGINS` | e.g. `https://your-app.vercel.app` |
| `AUTH_ACCESS_TOKEN_TTL` | e.g. `1h` |
| `AUTH_REFRESH_TOKEN_TTL` | e.g. `168h` |
| `AUTH_REFRESH_COOKIE_DOMAIN` | e.g. `.yourdomain.com` |
| `S3_PUBLIC_BUCKET_NAME` | e.g. `artmission-public` |
| `S3_PRIVATE_BUCKET_NAME` | e.g. `artmission-private` |
| `S3_PUBLIC_BASE_URL` | e.g. `https://<project>.supabase.co/storage/v1/object/public/artmission-public` |
| `S3_ENDPOINT` | e.g. `https://<project>.supabase.co/storage/v1/s3` |
| `S3_REGION` | e.g. `ap-southeast-1` |

**Secrets** tab:

| Name | Value |
|---|---|
| `VERCEL_TOKEN` | token created above |
| `DATABASE_DSN` | Postgres DSN, e.g. `postgres://user:pass@host:5432/db?sslmode=require` |
| `AUTH_JWT_SECRET` | long random string, e.g. `openssl rand -base64 48` |
| `S3_ACCESS_KEY_ID` | S3/Supabase storage access key id |
| `S3_SECRET_ACCESS_KEY` | S3/Supabase storage secret access key |


Optionally create a `production` [Environment](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment) (Settings → Environments) for required-reviewer gating before deploys run — both workflows already target `environment: production`. Without it configured, the reference is a no-op label.

## Part 4 — Workflows

- [`../.github/workflows/deploy-backend.yml`](../.github/workflows/deploy-backend.yml) — on push to `main` touching `backend/**` (or manual dispatch): authenticate via WIF, `docker build`/`push` to Artifact Registry, `deploy-cloudrun` with all env vars — plain values from repo Variables, secrets from repo Secrets — injected directly as Cloud Run environment variables.
- [`../.github/workflows/deploy-frontend.yml`](../.github/workflows/deploy-frontend.yml) — on push to `main` touching `frontend/**` (or manual dispatch): `vercel pull` → `vercel build --prod` → `vercel deploy --prebuilt --prod`.

Both are independent of `ci.yml`/`backend.yml` (which only run on pull requests) and only trigger on `main`.

## Troubleshooting

- **Backend crashes at startup with a config validation error**: a required secret (`DATABASE_DSN`, `AUTH_JWT_SECRET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`) is missing or empty in GitHub Secrets — an unset secret expands to an empty string. Check Settings → Secrets and variables → Actions → Secrets.
- **`deploy-cloudrun` fails parsing `env_vars`**: the action splits `env_vars` on commas/newlines. A secret value containing a comma or newline (e.g. a DB password) breaks it; use a value without them (URL-encode the password in the DSN).
- **WIF auth fails with `attribute condition was not met`**: the OIDC provider's `--attribute-condition` only allows `AiSiriRak/Artmission`; a fork or renamed repo needs the provider recreated with the new value.
- **`vercel pull` fails with `Could not retrieve Project Settings` despite correct/matching `VERCEL_ORG_ID`/`VERCEL_PROJECT_ID` and a valid, non-expired token**: run the pull step with `--debug` — it calls `GET /v2/user`, `GET /teams/:id`, and `GET /v9/projects/:id` separately, and the generic error fires if *any* of them fails, even if the project lookup itself succeeds. A `403` specifically on `GET /teams/:id` means the token was created with **Scope = a single project** (Account Settings → Tokens) instead of **Scope = the team**; project-scoped tokens can read that one project but can never satisfy `vercel pull`'s team-level lookup. Recreate the token scoped to the team itself.
- **Cloud Run service rejects traffic / 403**: confirm the `--allow-unauthenticated` flag deployed successfully; Cloud Run defaults new services to requiring IAM auth.
