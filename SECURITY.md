# Security Policy

`kubectl-plan` prioritizes safety, privacy, and system security above all. This document outlines our vulnerability reporting process and security posture.

## Supported Versions

Only the latest minor release receives security updates.

| Version | Supported |
| ------- | --------- |
| `< v0.1` | ❌ No     |
| `v0.1.x` |  Yes    |

## Security Architecture

### 1. Read-Only by Design
Through `v0.4`, `kubectl-plan` operates **exclusively in read-only mode**.
- The required RBAC `ClusterRole` contains **zero** mutating verbs (`create`, `update`, `patch`, `delete`).
- It only requests `get` and `list` permissions.
- You can safely run this tool in production environments knowing it is cryptographically and behaviorally prevented from modifying any cluster assets.

### 2. Privacy-first Telemetry (v1.0+)
When opted into community telemetry, data sanitization is performed locally inside the process **before** leaving your cluster:
- **Never Collected:** Cluster names, namespaces, resources names, IP addresses, hostnames, Prometheus values, or raw query strings.
- **Fingerprinting:** Same-cluster correlations are performed using a prefix SHA-256 hash of the cluster UID (`SHA256(cluster UID)[:16]`) which is mathematically non-reversible.
- Dedicating a custom test suite (`sanitizer_test.go`) explicitly asserts that no PII passes through.

---

## Reporting a Vulnerability

If you discover a security vulnerability, please **do not report it via public GitHub issues**. Instead, report it privately:

1. Send an email to **security@kubectl-plan.dev** (or open a private vulnerability report on GitHub if enabled).
2. Include a detailed description of the issue, steps to reproduce, and a proof of concept (PoC) if available.
3. We will acknowledge your report within 48 hours and coordinate a public disclosure window once a patch is prepared.
