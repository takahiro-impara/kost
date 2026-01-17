# Quick Start Guide: kost (Kubernetes Optimization & Sizing Tool)

**Date**: 2026-01-14
**Updated**: 2026-01-15
**Feature**: kost (Kubernetes Optimization & Sizing Tool) MVP Core Features
**Phase**: Phase 1 - Quickstart Guide
**Status**: ✅ Implementation Complete (P1-P2 features)

## Overview

This guide explains how to go from installation to generating your first report with kost (Kubernetes Optimization & Sizing Tool) in under 10 minutes.

**Implemented Features**:
- ✅ Deployment resource optimization (CPU/Memory requests recommendations)
- ✅ Applicable YAML patch generation (Strategic Merge Patch format)
- ✅ Markdown reports and JSON summary generation
- ✅ labelSelector and namespace exclusion filters
- ✅ Comprehensive input validation and security features

## Prerequisites

Ensure the following environment is ready:

- **Go 1.25.5 or later**: For security vulnerability mitigation (if building from source)
- **Kubernetes cluster**: Access permissions to a running K8s cluster
- **kubectl**: `kubectl` command available and able to connect to the target cluster
- **Prometheus**: Prometheus deployed in the cluster with the following metrics available:
  - `container_cpu_usage_seconds_total`
  - `container_memory_working_set_bytes`
- **RBAC permissions**: Read-only permissions for Deployment/Pods
- **Metrics retention**: Prometheus retains at least 5 minutes of metrics (7 days recommended for production)

**Security Requirements**:
- API credentials managed via environment variables (not included in config files)
- Use read-only ServiceAccount

## Step 1: Installation

### Binary Download (Recommended)

```bash
# Download latest release
curl -LO https://github.com/takahiro-impara/kost/releases/latest/download/kost-linux-amd64

# Grant execute permission
chmod +x kost-linux-amd64

# Move to /usr/local/bin
sudo mv kost-linux-amd64 /usr/local/bin/kost

# Verify installation
kost version
```

### Build from Go Source

```bash
# Check Go version (1.25.5+ required)
go version

# Clone repository
git clone https://github.com/takahiro-impara/kost.git
cd kost

# Install dependencies
go mod download

# Build
go build -o bin/kost ./cmd/kost

# Or use Makefile
make build

# Verify binary
./bin/kost version
```

### Container Image

```bash
# Run with Docker
docker run --rm -v ~/.kube:/root/.kube \
  takahiro-impara/kost:latest version
```

## Step 2: RBAC Permission Setup

Apply the following RBAC resources to your target cluster:

```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: finops-advisor
  namespace: default

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: finops-advisor-reader
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["pods", "namespaces"]
  verbs: ["get", "list"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: finops-advisor-reader-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: finops-advisor-reader
subjects:
- kind: ServiceAccount
  name: finops-advisor
  namespace: default
```

Apply:

```bash
kubectl apply -f examples/rbac.yaml
```

## Step 3: Create Configuration File

Create `config.yaml`:

```yaml
# config.yaml
kube:
  context: ""  # Empty string uses default context

prometheus:
  url: "http://prometheus-operated.monitoring:9090"
  timeoutSeconds: 10

analysis:
  window: "7d"
  cpuPercentile: 0.95
  memPercentile: 0.95
  safetyFactor: 1.2
  minCpuMilli: 20
  minMemMi: 64

filters:
  namespacesExclude:
    - kube-system
    - monitoring
  labelSelector: ""

output:
  dir: "./out"
  format:
    - md
    - json
    - patch

llm:
  enabled: false  # Set to false for initial run
  provider: "openai"
  model: "gpt-4"
  maxTokens: 1200
  temperature: 0.2
```

## Step 4: Run Initial Scan

Run scan by specifying the target namespace:

```bash
# Scan production environment
kost scan --namespace prod --config config.yaml

# Output example:
# Scanning namespace: prod
# Found 15 Deployments
# Found 5 HPAs
# Scan completed in 2.3s
```

## Step 5: Generate Recommendations

Generate recommendations based on scan results:

```bash
# Generate recommendations
kost suggest --namespace prod --config config.yaml

# Output example:
# Analyzing 15 Deployments...
# Querying Prometheus for CPU/Memory metrics (window: 7d)...
# Calculating recommendations...
# Found 8 overprovisioned containers
# Found 2 underprovisioned containers
# Found 3 HPA optimization opportunities
# Suggestion completed in 12.5s
```

## Step 6: Generate Report

Generate report including recommendations:

```bash
# Generate report
kost report --namespace prod --config config.yaml

# Output example:
# Generating report...
# Writing report to: ./out/report.md
# Writing summary to: ./out/summary.json
# Writing patches to: ./out/patches/prod/
# Report generated successfully!
```

## Step 7: Review Report

Review generated reports:

```bash
# Markdown report
cat ./out/report.md

# JSON summary
cat ./out/summary.json | jq '.'

# List patch files
ls -la ./out/patches/prod/
# Output example:
# api-deployment.yaml
# api-deployment-hpa.yaml
# worker-deployment.yaml
```

## Step 8: Apply Patches (Optional)

Review recommendations and apply patches if acceptable:

```bash
# Review patch contents
cat ./out/patches/prod/api-deployment.yaml

# Verify with dry-run
kubectl apply -f ./out/patches/prod/api-deployment.yaml --dry-run=client

# Actually apply
kubectl apply -f ./out/patches/prod/api-deployment.yaml

# Also apply HPA recommendations
kubectl apply -f ./out/patches/prod/api-deployment-hpa.yaml
```

## LLM Option (Optional)

To enable LLM-powered explanation generation:

### Using OpenAI

```bash
# Set API key as environment variable
export OPENAI_API_KEY="sk-..."

# Enable LLM in config.yaml
# llm.enabled: true
# llm.provider: "openai"

# Generate report
kost report --namespace prod --config config.yaml --llm on
```

### Using Claude

```bash
# Set API key as environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Enable LLM in config.yaml
# llm.enabled: true
# llm.provider: "claude"
# llm.model: "claude-3-5-sonnet-20241022"

# Generate report
kost report --namespace prod --config config.yaml --llm on
```

## Security Best Practices

Recommendations for using kost safely:

### Protecting API Credentials

- **Manage with environment variables**: Always set API credentials via environment variables
  ```bash
  # Correct method
  export OPENAI_API_KEY="sk-..."

  # Incorrect method (do not write directly in config.yaml)
  # llm.apiKey: "sk-..." ← This is prohibited
  ```

- **Exclude from config files**: Add environment variable files (`.env`) to `.gitignore`
  ```bash
  echo ".env" >> .gitignore
  ```

- **Protect kubeconfig**: Properly protect `~/.kube/config` containing cluster access information
  ```bash
  chmod 600 ~/.kube/config
  ```

### Verify RBAC Permissions

Verify that kost has the minimum required permissions:

```bash
# Verify ServiceAccount permissions
kubectl auth can-i list deployments --as=system:serviceaccount:default:finops-advisor
kubectl auth can-i list pods --as=system:serviceaccount:default:finops-advisor
kubectl auth can-i list horizontalpodautoscalers --as=system:serviceaccount:default:finops-advisor

# Verify all return "yes"
```

### Input Validation

kost automatically validates inputs, but note the following:

- **Namespace names**: Use names compliant with Kubernetes naming conventions (lowercase, digits, hyphens, max 63 characters)
  ```bash
  # Correct example
  kost scan --namespace prod-app-v1

  # Incorrect example
  kost scan --namespace "PROD_APP"  # Uppercase and underscores not allowed
  ```

- **Output directory**: Use relative paths or safe absolute paths to prevent path traversal attacks
  ```bash
  # Correct examples
  output.dir: "./out"
  output.dir: "/tmp/kost-reports"

  # Incorrect examples (will be rejected)
  output.dir: "../../etc/passwd"
  output.dir: "/etc/sensitive"
  ```

### TLS Certificate Verification

TLS certificate verification for Prometheus connections is enabled by default:

```yaml
# Enable certificate verification in production (default)
prometheus:
  url: "https://prometheus.example.com"
  tlsVerify: true  # Default value

# Disable only for development with self-signed certificates
prometheus:
  url: "https://prometheus-dev.local"
  tlsVerify: false  # Not recommended for production
```

## Troubleshooting

### Cannot Connect to Prometheus

```bash
# Verify Prometheus endpoint
kubectl get svc -n monitoring prometheus-operated

# Verify connection with port-forward
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090

# Access http://localhost:9090 in browser to verify
```

### Required Metrics Not Available

```bash
# Verify metrics exist in Prometheus
curl -s 'http://prometheus:9090/api/v1/query?query=container_cpu_usage_seconds_total' | jq '.'

# Verify kube-state-metrics
kubectl get pods -n monitoring | grep kube-state-metrics
```

### RBAC Permission Error

```bash
# Verify current permissions
kubectl auth can-i list deployments --as=system:serviceaccount:default:finops-advisor

# Describe Role
kubectl describe clusterrole finops-advisor-reader
```

## Next Steps

- **Automation**: Run periodically in CI like GitHub Actions and save reports as artifacts
- **Multiple Namespaces**: Specify `--namespace` multiple times for batch analysis
- **GitOps Integration**: Automatically create PRs with generated patches (future feature)

## Summary

You have now learned the basics of using kost (Kubernetes Optimization & Sizing Tool). This tool:

- ✅ Completes installation to first report in under 10 minutes
- ✅ Optimizes both Deployment and HPA
- ✅ Automatically generates applicable YAML patches
- ✅ Natural language explanations with LLM (optional)

For detailed documentation, see [README.md](../../../README.md).
