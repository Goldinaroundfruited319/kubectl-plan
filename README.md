# kubectl-plan

> **Terraform has `plan`. Kubernetes should too.**

[![CI](https://github.com/samaasi/kubectl-plan/actions/workflows/ci.yml/badge.svg)](https://github.com/samaasi/kubectl-plan/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/samaasi/kubectl-plan)](https://goreportcard.com/report/github.com/samaasi/kubectl-plan)
[![License](https://img.shields.io/github/license/samaasi/kubectl-plan)](LICENSE)
[![Release](https://img.shields.io/github/v/release/samaasi/kubectl-plan)](https://github.com/samaasi/kubectl-plan/releases/latest)

`kubectl-plan` is an **operational decision support CLI plugin** for Kubernetes. It bridges the gap between observability (which tells you what *happened*) and execution (which acts without foresight), answering the ultimate pre-flight question:

**"What will happen if I perform this operation?"**

```
Prometheus tells you what is happening right now.
Grafana shows you what happened over time.
kubectl-plan answers: "Is it safe to do this right now?"
```

---

## Table of Contents

- [Why kubectl-plan?](#why-kubectl-plan)
- [Key Features](#key-features)
- [Sample Output](#sample-output)
- [Installation](#installation)
- [Quickstart](#quickstart)
- [Building from Source](#building-from-source)
- [Usage — v0.1 Commands](#usage--v01-commands)
- [Configuration](#configuration)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

---

## Why kubectl-plan?

Traditional observability tools are retrospective. They cannot evaluate prospective changes.

| Tool | Question answered |
|---|---|
| Prometheus | "What is the current error rate of payment-api?" |
| Grafana | "What was the traffic pattern over the last 7 days?" |
| Jaeger / Tempo | "Which services were called during this request?" |
| kubectl-graph | "What does the dependency graph look like?" |
| **kubectl-plan** | **"Is it safe to scale payment-api to zero right now?"** |

No existing tool answers the last question. That gap is where this project lives.

---

## Key Features

- **Confidence-based decision support** — calculates confirmed dependents using API references, label selectors, and Ingress routing, then computes an uncertainty index so you know exactly what is unknown
- **Outage prevention** — pre-flight checks for the commands engineers run without thinking: `scale`, `restart`, `delete`
- **Auditable scores** — inspect the deterministic mathematical scoring breakdown using `kubectl plan why`
- **Readiness diagnostics** — diagnose exactly how ready your environment is to provide high-confidence checks with `kubectl plan doctor`
- **Read-only by design** — `kubectl-plan` never mutates your cluster through `v0.4` (see [SECURITY.md](SECURITY.md))

---

## Sample Output

```
ACTION:     scale deployment/payment-api --replicas=0  [namespace: production]

RISK SCORE:       8.7 / 10  ████████░░  HIGH
CONFIDENCE:        94%      █████████░  (topology + Prometheus)
UNCERTAINTY:       LOW      (well-instrumented service)

DEPENDENTS:
  ├─ checkout-service   DIRECT    [99%]
  │     Evidence: 18,234 req/24h · destination_service=payment-api · Prometheus
  │
  ├─ billing-service    DIRECT    [99%]
  │     Evidence: 4,102 req/24h  · destination_service=payment-api · Prometheus
  │
  └─ invoice-worker     INDIRECT  [70%]
        Evidence: env.PAYMENT_URL matches service DNS · no traffic observed
       ~Uncertain: no Prometheus confirmation

UNKNOWN BLAST RADIUS:
  ⚠ Cannot detect: Kafka consumers, external HTTP clients, Consul-registered services
  ℹ Run `kubectl plan doctor` to see what instrumentation would increase confidence.

RISK CONTRIBUTORS:
  +3.0  production-payments namespace   [criticality: CRITICAL]
  +2.4  Ingress exposed (external traffic)
  +1.8  3 confirmed direct consumers
  +1.5  Cross-namespace impact
  ─────
  = 8.7 / 10

RECOMMENDATION:
  ⚠ Do not proceed during peak traffic.
  → kubectl plan why deployment payment-api   for full scoring breakdown.
```

---

## Installation

### Via Krew (Recommended)

```bash
kubectl krew install plan
```

Once installed, kubectl discovers the plugin automatically:

```bash
kubectl plan --help
```

### Pre-built Binary

Download the latest binary from the [Releases](https://github.com/samaasi/kubectl-plan/releases/latest) page:

```bash
# Linux / macOS
curl -Lo kubectl-plan \
  https://github.com/samaasi/kubectl-plan/releases/latest/download/kubectl-plan_linux_amd64
chmod +x kubectl-plan
sudo mv kubectl-plan /usr/local/bin/
```

Windows: download `kubectl-plan_windows_amd64.exe`, rename to `kubectl-plan.exe`, place in a directory on `%PATH%`.

---

## Quickstart


Two paths depending on what you have available.

### Path A — You have a running cluster

If you already have kubectl connected to a cluster (EKS, GKE, AKS, k3s, Minikube, etc.):

```bash
# 1. Build and install
git clone https://github.com/samaasi/kubectl-plan.git
cd kubectl-plan
go build -o kubectl-plan ./cmd/kubectl-plan
sudo mv kubectl-plan /usr/local/bin/   # Windows: copy to a dir in %PATH%

# 2. Apply read-only RBAC (only needed once per cluster)
kubectl apply -f deploy/rbac/clusterrole.yaml
# Edit clusterrolebinding.yaml to set your username, then:
kubectl apply -f deploy/rbac/clusterrolebinding.yaml

# 3. Diagnose your environment first
kubectl plan doctor

# 4. Pick any deployment in your cluster and analyse it
kubectl plan scale deployment/<your-deployment> --replicas=0 -n <namespace>
kubectl plan why deployment/<your-deployment> -n <namespace>
```

> **Not sure which deployment to try?** Pick something non-critical in a staging namespace.
> `kubectl get deployments -A` lists everything available.

---

### Path B — No cluster yet (local Kind)

Requires: [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) and Docker.

```bash
git clone https://github.com/samaasi/kubectl-plan.git
cd kubectl-plan

# Build the binary
go build -o kubectl-plan ./cmd/kubectl-plan
sudo mv kubectl-plan /usr/local/bin/

# Spin up a local cluster pre-loaded with test workloads
./hack/test-cluster/setup.sh

# The script prints the context name when done. Then:
kubectl plan doctor --context kind-kubectl-plan-test

kubectl plan scale deployment/payment-api --replicas=0 \
  --context kind-kubectl-plan-test -n production

kubectl plan why deployment/payment-api \
  --context kind-kubectl-plan-test -n production

kubectl plan restart deployment/checkout-service \
  --context kind-kubectl-plan-test -n production
```

The test workloads (`payment-api`, `checkout-service`, `billing-service`) are pre-wired with env var references and cross-namespace dependencies so you see a realistic dependency graph on the first run.

---

### What to expect

`kubectl plan doctor` output tells you the confidence level of your environment:

```
DATA SOURCES:
  ✓ Kubernetes API    reachable · N resources scanned
  ✗ Prometheus        not found — topology-only scoring active

ESTIMATED ANALYSIS CONFIDENCE:
  52%  █████░░░░░

TO IMPROVE CONFIDENCE:
  → Integrate Prometheus data source (v0.2)
```

Topology-only (no Prometheus) is fully functional for v0.1 — you get dependency graph analysis, risk scoring, and recommendations. Prometheus adds real traffic evidence in v0.2.

---

## Building from Source

### Prerequisites

- Go `>= 1.22`
- `kubectl` configured with a valid `kubeconfig`
- A Kubernetes cluster (local Kind/Minikube or remote — read-only access is sufficient)

### Quick build

```bash
git clone https://github.com/samaasi/kubectl-plan.git
cd kubectl-plan

# Build the binary
go build -o kubectl-plan ./cmd/kubectl-plan

# Install into PATH
mv kubectl-plan /usr/local/bin/   # Linux / macOS
# or: copy kubectl-plan.exe to a directory in %PATH%  # Windows
```

### Using Make

```bash
make build    # compile → ./kubectl-plan
make test     # run all unit tests with -race
make clean    # remove binary
```

### Verify the build

```bash
kubectl-plan version
# kubectl-plan dev (commit: none, built: unknown)
```

---

## Usage — v0.1 Commands

All commands require a working `kubeconfig`. By default they target the current context and namespace.

### Global flags

| Flag | Default | Description |
|---|---|---|
| `-n`, `--namespace` | current context | Target namespace |
| `--context` | current context | Override kubeconfig context |
| `-o`, `--output` | `terminal` | Output format: `terminal` \| `json` |
| `--all-namespaces` | false | Include cross-namespace dependency scanning |
| `--ascii` | false | Disable unicode box drawing |
| `--no-color` | false | Disable ANSI color (also respects `NO_COLOR` env) |

---

### `kubectl plan scale`

Analyse risk before scaling a workload.

```bash
# Simulate scaling to zero — highest-risk routine operation
kubectl plan scale deployment/payment-api --replicas=0

# Scale up analysis
kubectl plan scale deployment/payment-api --replicas=5 -n production

# JSON output for CI pipelines
kubectl plan scale deployment/payment-api --replicas=0 --output json
```

**What it checks:**
- All confirmed direct and indirect dependents (label selectors, Ingress routing, env var references)
- HPA presence (may auto-recover after scale)
- PodDisruptionBudget constraints
- Cross-namespace impact
- Namespace criticality profile

---

### `kubectl plan restart`

Analyse risk before rolling restart — the command engineers run "without thinking".

```bash
kubectl plan restart deployment/payment-api
kubectl plan restart statefulset/postgres -n data
```

**Why this matters:** Rolling restarts cascade. Services that appear independent share a common dependency. `kubectl plan restart` surfaces the blast radius before the pods start terminating.

---

### `kubectl plan why`

Inspect the full, auditable scoring breakdown for any workload.

```bash
kubectl plan why deployment/payment-api
kubectl plan why deployment/payment-api -n production
```

Output shows each scoring rule, its weight, its computed value, and its contribution to the final score. Nothing is hidden.

---

### `kubectl plan doctor`

Diagnose why confidence scores are low and get actionable improvement steps.

```bash
kubectl plan doctor
kubectl plan doctor --namespace production
kubectl plan doctor --output json
```

**Checks performed:**
- Kubernetes API reachability and resource count
- Prometheus availability and service coverage
- Service mesh / Istio detection
- OpenTelemetry collector presence
- Namespace criticality profile load status
- Historical record count

Run this first when you get a low confidence score.

---

## Configuration

### Namespace Criticality Profiles

By default, any namespace containing `prod` is treated as `HIGH` criticality. Override this with a YAML profile:

```bash
# Copy the example config
cp config/criticality.example.yaml ~/.kubectl-plan/criticality.yaml
```

```yaml
# ~/.kubectl-plan/criticality.yaml
profiles:
  - namespace: production-payments
    level: CRITICAL   # +30 score multiplier
  - namespace: production-checkout
    level: HIGH       # +20 score multiplier
  - namespace: production-marketing
    level: MEDIUM     # +10 score multiplier
  - namespace: staging
    level: LOW        # no multiplier
```

Criticality levels affect the `namespace_criticality` rule weight in the risk scoring engine. See [docs/criticality-profiles.md](docs/criticality-profiles.md).

### RBAC — Minimum Required Permissions

`kubectl-plan` is **read-only**. Apply the bundled ClusterRole:

```bash
kubectl apply -f deploy/rbac/clusterrole.yaml
kubectl apply -f deploy/rbac/clusterrolebinding.yaml
```

Or grant permissions manually — see [docs/installation.md](docs/installation.md) for the full permission matrix.

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    kubectl-plan                      │
│                    (CLI plugin)                      │
└───────────────────┬─────────────────────────────────┘
                    │
          ┌─────────▼──────────┐
          │   Command Layer     │  Verb + resource parsing
          │   (Cobra CLI)       │  Flag normalization
          └─────────┬───────────┘
                    │
          ┌─────────▼───────────────┐
          │   Analysis Engine        │  Orchestrates pipeline
          │   internal/analysis/     │  Returns AnalysisResult
          └─────────┬───────────────┘
                    │
        ┌───────────▼──────────────────────────┐
        │        Dependency Engine              │
        │        internal/dependency/           │
        │  Builds confidence-weighted graph     │
        │  Tags every edge with Evidence        │
        │  Computes per-edge uncertainty        │
        └───────────┬──────────────────────────┘
                    │
   ┌────────────────▼───────────────────────────┐
   │           Data Source Adapters              │
   │  ┌────────────┐  ┌──────────────────────┐  │
   │  │ Kubernetes │  │   Prometheus         │  │
   │  │    API     │  │   (optional)         │  │
   │  └────────────┘  └──────────────────────┘  │
   │  ┌────────────┐  ┌──────────────────────┐  │
   │  │  History   │  │  Criticality Profile │  │
   │  │  Store     │  │  (~/.kubectl-plan/)  │  │
   │  └────────────┘  └──────────────────────┘  │
   └────────────────┬───────────────────────────┘
                    │
          ┌─────────▼──────────────┐
          │   Risk Scorer           │  Deterministic weighted rules
          │   + Uncertainty Scorer  │  Separate axis from risk
          │   internal/risk/        │
          └─────────┬───────────────┘
                    │
          ┌─────────▼──────────┐
          │  Output Renderer    │  Terminal / JSON / CI
          └────────────────────┘
```

**Resolution algorithm (v0.1 — K8s API only):**

1. `ownerReferences` — authoritative parent/child links (confidence: 1.00)
2. Service label selectors matching pod labels (confidence: 0.95)
3. Ingress backends routing to matched Services (confidence: 0.95)
4. NetworkPolicy ingress selectors (confidence: 0.80)
5. Env var values matching service name or cluster DNS (confidence: 0.70)
6. DNS pattern matching in string values (confidence: 0.65)
7. ConfigMap/Secret volume mounts (confidence: 0.60)
8. CronJob URL pattern matching (confidence: 0.50)

**Risk scoring formula:**

```
risk_score = Σ (rule_weight × rule_value) / Σ active_rule_weights × 10
```

Fully deterministic. No ML. Reproducible given the same cluster state. Full documentation in [docs/risk-model.md](docs/risk-model.md).

---

## Project Structure

```
kubectl-plan/
├── cmd/
│   └── kubectl-plan/
│       └── main.go              # Entrypoint, Cobra root command
│
├── internal/
│   ├── analysis/                # Orchestration engine (fetch → graph → score → render)
│   ├── dependency/              # Confidence-weighted dependency graph + evidence
│   ├── risk/                    # Weighted scoring, uncertainty, recommender, why-cmd
│   ├── k8s/                     # client-go wrapper, parallel resource fetcher
│   ├── criticality/             # Namespace criticality profile loader
│   └── output/                  # Terminal / JSON / CI renderer
│
├── pkg/
│   └── version/                 # Build-time version injection
│
├── config/
│   └── criticality.example.yaml # Example namespace criticality profile
│
├── deploy/
│   └── rbac/                    # ClusterRole + ClusterRoleBinding manifests
│
├── testdata/
│   ├── fixtures/                # Kubernetes resource YAML fixtures for unit tests
│   └── golden/                  # Golden output files for renderer regression tests
│
├── docs/                        # Extended documentation
│   ├── installation.md
│   ├── risk-model.md
│   ├── confidence-model.md
│   ├── criticality-profiles.md
│   └── examples/
│
├── hack/
│   ├── build.sh                 # Build helper script
│   ├── install-krew.sh          # Krew plugin install helper
│   └── test-cluster/            # Local Kind cluster setup for integration tests
│
├── .github/workflows/
│   ├── ci.yml                   # Test on push/PR to master and develop
│   └── release.yml              # GoReleaser — master-branch tags only
│
├── Makefile
├── .goreleaser.yaml
├── CONTRIBUTING.md
├── SECURITY.md
└── LICENSE
```

---

## Testing

### Run all unit tests

```bash
go test ./... -race
```

### Run with coverage

```bash
go test ./... -race -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test fixtures

Kubernetes resource fixtures live in [`testdata/fixtures/`](testdata/fixtures/). Unit tests in `internal/dependency` and `internal/risk` load these YAML files to build realistic dependency graphs without a live cluster.

Golden output files in [`testdata/golden/`](testdata/golden/) are used by the renderer tests to catch regressions in terminal output formatting.

### Integration testing with a local cluster

```bash
# Spin up a local Kind cluster with test workloads
./hack/test-cluster/setup.sh

# Run all commands against it
kubectl plan scale deployment/payment-api --replicas=0
kubectl plan restart deployment/payment-api
kubectl plan why deployment/payment-api
kubectl plan doctor
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full setup instructions.

---

## Roadmap

Each milestone ships independently useful functionality. The tool is usable today at `v0.1`.

### ✅ v0.1 — Core _(current)_

> A working `kubectl plan` plugin that delivers risk output in seconds with zero external dependencies.

| Capability | Status |
|---|---|
| `kubectl plan scale` | ✅ Shipped |
| `kubectl plan restart` | ✅ Shipped |
| `kubectl plan why` | ✅ Shipped |
| `kubectl plan doctor` | ✅ Shipped |
| `kubectl plan delete` | 🔜 In progress |
| Dependency engine (K8s API — 8 resolution steps) | ✅ Shipped |
| Risk scoring (deterministic weighted rules) | ✅ Shipped |
| Uncertainty score (separate axis from risk) | ✅ Shipped |
| Namespace criticality profiles | ✅ Shipped |
| Terminal / JSON output renderer | ✅ Shipped |
| RBAC manifests (read-only ClusterRole) | ✅ Shipped |
| GoReleaser multi-platform distribution | ✅ Shipped |

`v0.1` produces no writes to the cluster. Every command is a read-only analysis.

---

### 🔄 v0.2 — Observability Integration

> Replace topological inference with real traffic evidence from Prometheus.

- Auto-discover Prometheus in cluster (flag → env var → K8s API scan)
- Named PromQL builders for traffic, error rate, P99 latency
- Evidence enrichment: upgrade topology edges with observed traffic (confidence → 0.99)
- Discover Prometheus-only dependencies invisible to topology analysis
- 3 new risk rules: `live_request_rate`, `error_rate_elevated`, `p99_latency_high`
- Graceful degradation: topology-only mode when Prometheus is absent

---

### 🔄 v0.3 — GitOps Integration

> Shift risk analysis left into PR workflows and manifest diffs.

- `kubectl plan manifest ./k8s/` — diff manifests vs live cluster, run analysis per changed resource
- ArgoCD PreSync resource hook + PR comment posting
- GitHub Actions integration (`kubectl-plan/action@v1`)
- Flux notification provider
- JSON output schema for CI consumption

---

### 🔄 v0.4 — Historical Impact Memory

> Stop inferring. Start remembering.

- Append-only local history store (`~/.kubectl-plan/history.jsonl`)
- Record every plan run with risk score, confidence, and cluster ID
- Outcome recording: manual and automatic via Prometheus polling
- `kubectl plan history deployment/payment-api` — surface past operations on same target
- Historical evidence in risk output: "Previous scale 3→1 caused +32% latency"

---

### 🔄 v1.0 — Stable + Admission Controller _(opt-in)_

> Enforce risk thresholds at the API server level for teams that require it.

- `ValidatingAdmissionWebhook` server
- Configurable risk threshold + bypass label
- cert-manager integration for TLS
- Optional — designed for regulated or high-stakes environments only
- Stability guarantee: API compatibility from this release forward

---

## Contributing

We welcome contributions of all kinds — bug reports, documentation, test fixtures, and new features.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development environment setup (Go, Kind, kubectl)
- How to run the test suite
- Branch and PR workflow (`develop` → `master`)
- Code style and commit message conventions

---

## Security

`kubectl-plan` is **read-only through v0.4**. It never creates, patches, or deletes any Kubernetes resource.

See [SECURITY.md](SECURITY.md) for the full security policy, including the telemetry data sanitization commitment and how to report vulnerabilities.

---

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
