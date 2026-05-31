# Testing Guide

A complete, step-by-step walkthrough for testing `kubectl-plan` — from zero to verified output — covering both an existing cluster and a local Kind cluster.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Step 1 — Build the Binary](#step-1--build-the-binary)
- [Step 2 — Verify the Binary](#step-2--verify-the-binary)
- [Step 3 — Set Up Your Cluster](#step-3--set-up-your-cluster)
  - [Option A: You already have a cluster](#option-a-you-already-have-a-cluster)
  - [Option B: Spin up a local Kind cluster](#option-b-spin-up-a-local-kind-cluster)
- [Step 4 — Apply RBAC](#step-4--apply-rbac)
- [Step 5 — Run `kubectl plan doctor`](#step-5--run-kubectl-plan-doctor)
- [Step 6 — Run `kubectl plan scale`](#step-6--run-kubectl-plan-scale)
- [Step 7 — Run `kubectl plan restart`](#step-7--run-kubectl-plan-restart)
- [Step 8 — Run `kubectl plan why`](#step-8--run-kubectl-plan-why)
- [Step 9 — Optional: Namespace Criticality Profile](#step-9--optional-namespace-criticality-profile)
- [Step 10 — JSON Output (CI mode)](#step-10--json-output-ci-mode)
- [Step 11 — Run Unit Tests](#step-11--run-unit-tests)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

| Requirement | Minimum version | Check |
|---|---|---|
| Go | 1.22 | `go version` |
| kubectl | 1.24+ | `kubectl version --client` |
| Git | any | `git --version` |
| Kind _(Option B only)_ | 0.20+ | `kind version` |
| Docker _(Option B only)_ | any | `docker info` |

---

## Step 1 — Build the Binary

```bash
# Clone the repository
git clone https://github.com/samaasi/kubectl-plan.git
cd kubectl-plan
```

**Linux / macOS:**
```bash
go build -o kubectl-plan ./cmd/kubectl-plan
sudo mv kubectl-plan /usr/local/bin/kubectl-plan
```

**Windows (PowerShell):**
```powershell
go build -o kubectl-plan.exe ./cmd/kubectl-plan
# Move to a directory that is in your %PATH%, e.g.:
Move-Item kubectl-plan.exe "$env:USERPROFILE\bin\kubectl-plan.exe"
```

> If `~/bin` is not in your PATH on Windows, add it:
> `$env:PATH += ";$env:USERPROFILE\bin"` (add permanently via System → Advanced → Environment Variables)

---

## Step 2 — Verify the Binary

```bash
kubectl-plan version
```

Expected output:
```
kubectl-plan dev (commit: none, built: unknown)
```

Also verify kubectl discovers the plugin:
```bash
kubectl plan --help
```

Expected output (abbreviated):
```
kubectl-plan — operational decision support for Kubernetes

Usage:
  kubectl-plan [command]

Available Commands:
  scale       Analyse risk before scaling a workload
  restart     Analyse risk before rolling restart
  why         Show full risk score breakdown
  doctor      Diagnose analysis readiness

Flags:
  -n, --namespace string   Target namespace
  ...
```

If `kubectl plan --help` fails with "unknown command", check that the binary is named `kubectl-plan` (with a hyphen) and is on your `PATH`.

---

## Step 3 — Set Up Your Cluster

Choose one option.

---

### Option A: You already have a cluster

Check your current context:
```bash
kubectl config current-context
kubectl config get-contexts
```

Switch to the context you want to test against:
```bash
kubectl config use-context <your-context-name>
```

Verify connectivity:
```bash
kubectl cluster-info
kubectl get nodes
```

Identify a deployment to use in your tests:
```bash
# List all deployments across all namespaces
kubectl get deployments -A

# Note down a deployment name and namespace you want to analyse
# Prefer a non-critical workload for your first test
```

---

### Option B: Spin up a local Kind cluster

This option creates a local cluster with pre-wired test workloads that demonstrate cross-namespace dependencies, env var references, and Ingress routing — the full v0.1 dependency resolution pipeline.

**Install Kind** (if not already installed):
```bash
# macOS
brew install kind

# Linux
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Windows
choco install kind
```

**Create the cluster and load test workloads:**
```bash
./hack/test-cluster/setup.sh
```

Expected output:
```
Creating Kind cluster: kubectl-plan-test...
Applying test workloads...
deployment.apps/payment-api created
deployment.apps/checkout-service created
deployment.apps/billing-service created
deployment.apps/invoice-worker created
service/payment-api-svc created
service/checkout-svc created
service/billing-svc created
service/old-auth-svc created
ingress.networking.k8s.io/payment-ingress created

Cluster ready. Run commands with:
  kubectl plan scale deployment/payment-api --replicas=0 --context kind-kubectl-plan-test
  kubectl plan doctor --context kind-kubectl-plan-test
```

Verify the workloads are running:
```bash
kubectl get deployments -A --context kind-kubectl-plan-test
```

Expected:
```
NAMESPACE    NAME               READY   UP-TO-DATE   AVAILABLE
production   payment-api        3/3     3            3
production   checkout-service   2/2     2            2
production   invoice-worker     1/1     1            1
billing      billing-service    1/1     1            1
```

> Set a shell variable for convenience:
> ```bash
> export KUBE_CONTEXT=kind-kubectl-plan-test
> # then append to every command: --context $KUBE_CONTEXT
> ```

---

## Step 4 — Apply RBAC

`kubectl-plan` requires read-only access. Apply the bundled ClusterRole:

```bash
kubectl apply -f deploy/rbac/clusterrole.yaml
```

Expected:
```
clusterrole.rbac.authorization.k8s.io/kubectl-plan-reader created
```

Edit the ClusterRoleBinding to add your user or service account:
```bash
# Open and replace YOUR_USERNAME with your actual kubectl user
# (run `kubectl config view --minify` to see your current user)
nano deploy/rbac/clusterrolebinding.yaml
```

```yaml
subjects:
  - kind: User
    name: YOUR_USERNAME    # ← replace this
    apiGroup: rbac.authorization.k8s.io
```

Apply it:
```bash
kubectl apply -f deploy/rbac/clusterrolebinding.yaml
```

> **Kind cluster shortcut:** Kind uses a local certificate-based user. You can skip the ClusterRoleBinding for local testing and run commands directly as the admin user that Kind configured.

---

## Step 5 — Run `kubectl plan doctor`

Always run `doctor` first. It tells you what data sources are available and what confidence level to expect.

```bash
# Existing cluster
kubectl plan doctor

# Kind cluster
kubectl plan doctor --context kind-kubectl-plan-test
```

**Expected output (no Prometheus):**
```
ANALYSIS READINESS  [cluster: kind-kubectl-plan-test · namespace: default]

DATA SOURCES:
  ✓ Kubernetes API          reachable · 42 resources scanned
  ✗ Prometheus              not found
                            ⚠ Topology-only scoring active
                            Run with --prometheus-url to connect manually
  ✗ Istio / Service Mesh    not detected
  ✗ OpenTelemetry           not detected

NAMESPACE CRITICALITY PROFILE:
  ✗ No config found at ~/.kubectl-plan/criticality.yaml
  Using default heuristic: namespaces containing 'prod' → HIGH

ESTIMATED ANALYSIS CONFIDENCE:
  52%  █████░░░░░

TO IMPROVE CONFIDENCE:
  → Integrate Prometheus data source (v0.2)
  → Install Istio or Linkerd for traffic topology evidence (v0.3)
  → Create historical record store (v0.4)
```

**What this means:**
- **52% confidence** with topology-only is normal for v0.1 without Prometheus. This is expected.
- The tool is still fully functional — it will scan all Kubernetes resources and resolve the dependency graph.
- Confidence increases to ~90%+ when Prometheus is available (v0.2).

---

## Step 6 — Run `kubectl plan scale`

This is the primary command. It analyses what would happen if you scaled a workload to a given replica count.

**Kind cluster (using test workloads):**
```bash
kubectl plan scale deployment/payment-api --replicas=0 \
  --context kind-kubectl-plan-test \
  -n production
```

**Existing cluster:**
```bash
kubectl plan scale deployment/<your-deployment> --replicas=0 -n <your-namespace>
```

**Expected output:**
```
ACTION:     scale deployment/payment-api --replicas=0  [namespace: production]

RISK SCORE:       7.4 / 10  ███████░░░  HIGH
CONFIDENCE:        52%      █████░░░░░  (topology only)
UNCERTAINTY:       MEDIUM   (no Prometheus — some dependents inferred)

DEPENDENTS:
  ├─ checkout-service   DIRECT    [95%]
  │     Evidence: label selector {app: payment-api} matches pods
  │
  ├─ billing-service    DIRECT    [95%]
  │     Evidence: ingress/payment-ingress routes /api/pay → service/payment-api-svc
  │
  └─ invoice-worker     INDIRECT  [70%]
        Evidence: env.PAYMENT_URL=http://payment-api-svc matches service cluster DNS
       ~Uncertain: no Prometheus confirmation

UNKNOWN BLAST RADIUS:
  ⚠ No service mesh CRDs found — Consul/Istio/Envoy service relationships unknown
  ℹ Run `kubectl plan doctor` to see what instrumentation would increase confidence.

RISK CONTRIBUTORS:
  +2.2  production namespace   [criticality: HIGH, contains 'prod']
  +1.8  2 confirmed direct consumers
  +1.5  Cross-namespace impact
  +1.4  Ingress exposed (external traffic)
  ─────
  = 6.9 / 10

RECOMMENDATION:
  ⚠ Score 6.9 — review dependents before proceeding.
  → kubectl plan why deployment payment-api   for full scoring breakdown.
```

**What to look for:**
- `DEPENDENTS` — workloads that depend on `payment-api`, with their evidence type and confidence
- `UNKNOWN BLAST RADIUS` — honest list of what the tool cannot see
- `RISK CONTRIBUTORS` — the exact rules and weights that produced the score
- The score is deterministic: run it twice, get the same result

**Try scaling up:**
```bash
kubectl plan scale deployment/payment-api --replicas=5 \
  --context kind-kubectl-plan-test \
  -n production
```

Scaling up typically produces a lower risk score since you are increasing, not removing, capacity.

---

## Step 7 — Run `kubectl plan restart`

Analyses a rolling restart — the operation engineers run "without thinking".

```bash
# Kind cluster
kubectl plan restart deployment/checkout-service \
  --context kind-kubectl-plan-test \
  -n production

# Existing cluster
kubectl plan restart deployment/<your-deployment> -n <namespace>
```

Expected output structure (same format as scale, different action label):
```
ACTION:     restart deployment/checkout-service  [namespace: production]

RISK SCORE:       4.2 / 10  ████░░░░░░  MEDIUM
...
```

A rolling restart of `checkout-service` will typically score lower than a scale-to-zero of `payment-api` because `checkout-service` has fewer dependents in the test fixture.

---

## Step 8 — Run `kubectl plan why`

Shows the full, auditable score breakdown — every rule, its weight, its computed value, and its contribution.

```bash
# Kind cluster
kubectl plan why deployment/payment-api \
  --context kind-kubectl-plan-test \
  -n production

# Existing cluster
kubectl plan why deployment/<your-deployment> -n <namespace>
```

Expected output:
```
RISK SCORE BREAKDOWN: deployment/payment-api

Score:       6.9 / 10  ██████░░░░  HIGH
Confidence:  52%        █████░░░░░  (topology only)

CONTRIBUTORS:
  Rule                              Weight   Value   Contribution
  ─────────────────────────────── ──────   ─────   ────────────
  production namespace (HIGH)        20     0.80    +2.2
  Direct confirmed consumers (2)     30     0.40    +1.8
  Cross-namespace impact             10     1.00    +1.5
  Ingress exposed (external)         10     1.00    +1.4
  ─────────────────────────────── ──────   ─────   ────────────
  Total                                             6.9 / 10

CONFIDENCE SOURCES:
  ✓ Kubernetes topology    (label selectors, ingress routing)
  ? invoice-worker         (env var match only — no Prometheus)
```

**Verify the math:**

The score formula is:
```
score = Σ(weight × value) / Σ(active weights) × 10
```

You can verify manually:
- `(20×0.80) + (30×0.40) + (10×1.00) + (10×1.00)` = `16 + 12 + 10 + 10` = `48`
- Active weight sum = `20 + 30 + 10 + 10` = `70`
- `48 / 70 × 10` = `6.9` ✓

---

## Step 9 — Optional: Namespace Criticality Profile

Test how a criticality profile affects the score.

```bash
# Create the config directory
mkdir -p ~/.kubectl-plan

# Copy the example profile
cp config/criticality.example.yaml ~/.kubectl-plan/criticality.yaml
```

Edit it to add the `production` namespace (used in Kind test workloads):
```bash
cat > ~/.kubectl-plan/criticality.yaml << 'EOF'
profiles:
  - namespace: production
    level: CRITICAL
  - namespace: billing
    level: HIGH
  - namespace: staging
    level: LOW
EOF
```

Re-run doctor to confirm the profile loaded:
```bash
kubectl plan doctor --context kind-kubectl-plan-test
```

Expected change in output:
```
NAMESPACE CRITICALITY PROFILE:
  ✓ Config loaded: ~/.kubectl-plan/criticality.yaml
  production    → CRITICAL
  billing       → HIGH
```

Re-run the scale analysis:
```bash
kubectl plan scale deployment/payment-api --replicas=0 \
  --context kind-kubectl-plan-test \
  -n production
```

The risk score will be **higher** now because `production` is `CRITICAL` instead of the default `HIGH`. The `RISK CONTRIBUTORS` section will show the updated namespace multiplier.

---

## Step 10 — JSON Output (CI mode)

All commands support `--output json` for machine-readable output.

```bash
kubectl plan scale deployment/payment-api --replicas=0 \
  --context kind-kubectl-plan-test \
  -n production \
  --output json
```

Expected (abbreviated):
```json
{
  "action": "scale",
  "target": {
    "kind": "Deployment",
    "name": "payment-api",
    "namespace": "production"
  },
  "parameters": { "replicas": 0 },
  "risk_score": 6.9,
  "risk_level": "HIGH",
  "confidence": 0.52,
  "dependents": [
    {
      "name": "checkout-service",
      "namespace": "production",
      "relationship": "DIRECT",
      "confidence": 0.95,
      "evidence": [...]
    }
  ]
}
```

Pipe to `jq` to extract specific fields:
```bash
kubectl plan scale deployment/payment-api --replicas=0 \
  --context kind-kubectl-plan-test \
  -n production \
  --output json | jq '.risk_score'
```

---

## Step 11 — Run Unit Tests

The unit test suite does not require a running cluster — it uses YAML fixtures from `testdata/fixtures/`.

```bash
# Run all tests
go test ./... -race

# Run with coverage report
go test ./... -race -coverprofile=coverage.out
go tool cover -html=coverage.out   # opens browser

# Run a specific package
go test ./internal/risk/... -v
go test ./internal/dependency/... -v
```

Expected output:
```
ok  	github.com/samaasi/kubectl-plan/internal/analysis   0.312s
ok  	github.com/samaasi/kubectl-plan/internal/criticality  0.089s
ok  	github.com/samaasi/kubectl-plan/internal/dependency  0.201s
ok  	github.com/samaasi/kubectl-plan/internal/k8s        0.145s
ok  	github.com/samaasi/kubectl-plan/internal/risk        0.178s
ok  	github.com/samaasi/kubectl-plan/pkg/version         0.041s
```

---

## Troubleshooting

### `kubectl plan: command not found`

The binary must be named `kubectl-plan` (with a hyphen) and be on your `PATH`.

```bash
which kubectl-plan          # should print the binary path
kubectl-plan --help         # should work directly
kubectl plan --help         # kubectl discovers it via PATH
```

### `Error: no kubeconfig file found`

kubectl-plan uses your current kubeconfig. Make sure `kubectl` itself works first:
```bash
kubectl cluster-info
```

If it works, `kubectl plan` will work too.

### Score is `0.0 / 10`

The target deployment was not found in the specified namespace.

```bash
# Verify the deployment exists
kubectl get deployment payment-api -n production

# Check you're using the right context
kubectl config current-context
```

### `Error: forbidden: User cannot list resource`

The RBAC ClusterRole is not bound to your user. Re-apply the ClusterRoleBinding with your correct username:

```bash
kubectl config view --minify -o jsonpath='{.users[0].name}'
```

Use that value in `deploy/rbac/clusterrolebinding.yaml` and re-apply.

### Kind cluster pods not starting

```bash
kubectl get pods -A --context kind-kubectl-plan-test
```

If pods are in `Pending`, the Kind cluster may need more memory. Ensure Docker has at least 4 GB allocated (Docker Desktop → Settings → Resources).

### Low dependency count (fewer dependents than expected)

This is expected on a real cluster where workloads don't reference each other via env vars. The test fixtures are specifically designed to produce a rich dependency graph. On your own cluster, the results reflect your actual workload topology.

---

## Tear Down (Kind only)

```bash
kind delete cluster --name kubectl-plan-test
```
