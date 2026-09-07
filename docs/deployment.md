# Deployment Setup

How production deploys work for Artmission, and how to (re)configure them from scratch. Backend deploys to **Cloud Run**, frontend deploys to **Vercel**, both triggered by **GitHub Actions** on push to `main`.

## Why GitHub Actions instead of native Git integrations

Neither Vercel's Git integration nor GCP Cloud Build's GitHub trigger can be used here: both require installing a GitHub App on this specific repository, which needs repo **owner** access.

GitHub Actions sidesteps this:

- **Vercel**: deployed via the Vercel CLI (`vercel build` / `vercel deploy`) using an API token. No Git integration, no GitHub App, no repo permission needed on Vercel's side at all.
- **GCP**: deployed via `gcloud`/`google-github-actions` using **Workload Identity Federation (WIF)** — GitHub's OIDC token exchanged for short-lived GCP credentials. No GitHub App, no stored JSON key.

The only unavoidable GitHub-side requirement is adding repo Secrets/Variables (Settings → Secrets and variables → Actions), which itself needs `admin` — ask the repo owner to paste the values below once, or to grant `Admin` so you can add them yourself.

### Credential design

Only **one** value is a true secret in GitHub: `VERCEL_TOKEN`. Everything GCP-related is a non-sensitive identifier. Runtime secrets the backend actually needs at request time live in **Google Secret Manager** and are pulled at deploy time — they never touch GitHub.

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
  iamcredentials.googleapis.com secretmanager.googleapis.com

# 2. Artifact Registry repo for backend Docker images
gcloud artifacts repositories create artmission \
  --repository-format=docker --location="$REGION" \
  --description="Artmission images"

# 3. Deploy service account — the identity GitHub Actions impersonates via WIF.
#    Can build/push images and deploy to Cloud Run, but cannot read runtime secrets.
gcloud iam service-accounts create gh-deployer --display-name="GitHub Actions deployer"
DEPLOY_SA="gh-deployer@$PROJECT_ID.iam.gserviceaccount.com"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$DEPLOY_SA" --role="roles/run.admin" --condition=None
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$DEPLOY_SA" --role="roles/artifactregistry.writer" --condition=None

# 4. Runtime service account — what the Cloud Run *service* runs as.
#    Can read Secret Manager secrets, but cannot deploy anything (least privilege
#    split: a compromised deploy pipeline can't exfiltrate DB creds, and vice versa).
gcloud iam service-accounts create artmission-backend-run \
  --display-name="Backend Cloud Run runtime"
RUN_SA="artmission-backend-run@$PROJECT_ID.iam.gserviceaccount.com"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$RUN_SA" --role="roles/secretmanager.secretAccessor" --condition=None

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

# 7. Runtime secrets (values come from backend/.env.prod, never commit real values)
for name in DATABASE_DSN AUTH_JWT_SECRET S3_ACCESS_KEY_ID S3_SECRET_ACCESS_KEY; do
  gcloud secrets create "$name" --replication-policy="automatic"
done
echo -n "postgres://..."          | gcloud secrets versions add DATABASE_DSN --data-file=-
echo -n "<jwt secret>"            | gcloud secrets versions add AUTH_JWT_SECRET --data-file=-
echo -n "<s3 access key id>"      | gcloud secrets versions add S3_ACCESS_KEY_ID --data-file=-
echo -n "<s3 secret access key>" | gcloud secrets versions add S3_SECRET_ACCESS_KEY --data-file=-
```

To rotate a secret later: `echo -n "<new value>" | gcloud secrets versions add NAME --data-file=-`. The workflow always deploys `:latest`, so the next deploy picks it up — no redeploy of the pipeline itself required.

## Part 2 — Vercel setup

```bash
cd frontend
# Ensure you already installed vercel CLI
vercel login                 # your own account/team — repo access is irrelevant
vercel link                  # creates an empty project, prompts for scope + name
cat .vercel/project.json     # -> { "projectId": "...", "orgId": "..." }
```

Create a token: Vercel dashboard → Account Settings → **Tokens** → Create Token, scoped to the team, named e.g. `gh-actions-artmission`.

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

Optionally create a `production` [Environment](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment) (Settings → Environments) for required-reviewer gating before deploys run — both workflows already target `environment: production`. Without it configured, the reference is a no-op label.

## Part 4 — Workflows

- [`../.github/workflows/deploy-backend.yml`](../.github/workflows/deploy-backend.yml) — on push to `main` touching `backend/**` (or manual dispatch): authenticate via WIF, `docker build`/`push` to Artifact Registry, `deploy-cloudrun` with plain env vars from repo Variables and runtime secrets pulled live from Secret Manager.
- [`../.github/workflows/deploy-frontend.yml`](../.github/workflows/deploy-frontend.yml) — on push to `main` touching `frontend/**` (or manual dispatch): `vercel pull` → `vercel build --prod` → `vercel deploy --prebuilt --prod`.

Both are independent of `ci.yml`/`backend.yml` (which only run on pull requests) and only trigger on `main`.

## Troubleshooting

- **`deploy-cloudrun` step fails with a permission error on Secret Manager**: confirm `GCP_RUN_SERVICE_ACCOUNT` (not the deploy SA) has `roles/secretmanager.secretAccessor`, and that the `--service-account` flag in the workflow points at it.
- **WIF auth fails with `attribute condition was not met`**: the OIDC provider's `--attribute-condition` only allows `AiSiriRak/Artmission`; a fork or renamed repo needs the provider recreated with the new value.
- **Vercel build fails with missing project**: re-run `vercel link` locally and confirm `VERCEL_ORG_ID`/`VERCEL_PROJECT_ID` in GitHub match `.vercel/project.json` exactly.
- **Cloud Run service rejects traffic / 403**: confirm the `--allow-unauthenticated` flag deployed successfully; Cloud Run defaults new services to requiring IAM auth.
