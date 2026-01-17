# kost (Kubernetes Optimization & Sizing Tool)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25.5%2B-00ADD8?logo=go)](https://go.dev/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/takahiro-impara/kost/pulls)

A CLI tool to optimize Kubernetes Deployment resource configurations (requests/limits) and HPA settings (minReplicas/maxReplicas).

## Overview

kost retrieves historical performance data from Prometheus metrics, performs statistical analysis (P95/P99, etc.) and rule-based evaluation, then generates reports containing specific recommendations, applicable YAML patches, and reduction rates.

## Key Features

- ✅ Analyze and recommend Deployment resource optimizations
- ✅ Generate applicable YAML patches
- ✅ Recommend HPA configuration optimizations
- ✅ Support multiple LLM providers (OpenAI, Claude)
- ✅ AI-powered explanations and prioritization

## Quick Start

### Installation

```bash
# Download binary
curl -LO https://github.com/takahiro-impara/kost/releases/latest/download/kost-linux-amd64
chmod +x kost-linux-amd64
sudo mv kost-linux-amd64 /usr/local/bin/kost

# Build from Go source
git clone https://github.com/takahiro-impara/kost.git
cd kost
make build
```

### Usage

```bash
# Create configuration file
cp examples/config.yaml config.yaml

# Scan Deployments
kost scan --namespace prod --config config.yaml

# Generate recommendations
kost suggest --namespace prod --config config.yaml

# Generate report
kost report --namespace prod --config config.yaml
```

For details, see [quickstart.md](./specs/001-mvp-core/quickstart.md).

## Local Environment Testing

Instructions for testing kost in a local environment using minikube or Kind.

### Method 1: Using minikube (Recommended)

#### 1. Install minikube

**macOS:**
```bash
brew install minikube
```

**Linux:**
```bash
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```

#### 2. Start minikube cluster

```bash
# Start cluster (with metrics collection enabled)
minikube start --cpus=4 --memory=8192 --addons=metrics-server

# Verify kubectl is pointing to minikube
kubectl config current-context
# Output: minikube
```

#### 3. Install Prometheus

```bash
# Install Helm (if not already installed)
brew install helm  # macOS
# or
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash  # Linux

# Install Prometheus (with kube-state-metrics)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false

# Wait for Prometheus to start (2-3 minutes)
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n monitoring --timeout=300s
```

#### 4. Deploy sample application

```bash
# Create sample Deployment
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-demo
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        resources:
          requests:
            cpu: "500m"      # Excessive request (for testing)
            memory: "512Mi"  # Excessive request (for testing)
          limits:
            cpu: "1000m"
            memory: "1Gi"
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: nginx-demo-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx-demo
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
EOF

# Wait for Pods to be ready
kubectl wait --for=condition=ready pod -l app=nginx --timeout=60s

# Generate load (to produce metrics)
kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://nginx-demo; done" &

# Wait 5 minutes for metrics to accumulate
echo "Waiting for metrics to accumulate (5 minutes)..."
sleep 300
```

#### 5. Run kost

```bash
# Port-forward to Prometheus
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090 &

# Prepare kost config.yaml
cat <<EOF > config.yaml
kube:
  context: "minikube"

prometheus:
  url: "http://localhost:9090"
  timeoutSeconds: 30

analysis:
  window: "5m"  # Use short period for local testing
  cpuPercentile: 0.95
  memPercentile: 0.95
  safetyFactor: 1.2
  minCpuMilli: 10
  minMemMi: 32

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
  enabled: false
EOF

# Scan with kost
./bin/kost scan -n default --config config.yaml

# Generate recommendations
./bin/kost suggest -n default --config config.yaml

# Generate report
./bin/kost report -n default --config config.yaml

# View report
cat ./out/report.md
```

#### 6. Cleanup

```bash
# Stop port-forward
pkill -f "port-forward"

# Delete sample application
kubectl delete deployment nginx-demo
kubectl delete hpa nginx-demo-hpa

# Delete minikube cluster (if needed)
minikube delete
```

### Method 2: Using Kind

#### 1. Install Kind

**macOS:**
```bash
brew install kind
```

**Linux:**
```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

#### 2. Create Kind cluster

```bash
# Create cluster configuration file
cat <<EOF > kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
EOF

# Create cluster
kind create cluster --name kost-test --config kind-config.yaml

# Verify kubectl is pointing to kind
kubectl config current-context
# Output: kind-kost-test
```

#### 3. Install Prometheus

```bash
# Same procedure as minikube
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false

kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n monitoring --timeout=300s
```

#### 4-6. Deploy sample app, run kost, cleanup

Follow the same steps as for minikube.

```bash
# Delete cluster (if needed)
kind delete cluster --name kost-test
```

## Prerequisites

- **Go 1.25.5 or later** (for security vulnerability mitigation)
- Access permissions to Kubernetes cluster
- Prometheus already deployed (with required metrics available)
- RBAC permissions (read-only for Deployment/Pod/HPA)

**Important**: Go 1.25.3 and earlier versions have known security vulnerabilities (GO-2025-4175, GO-2025-4155). Upgrading to Go 1.25.5 or later is strongly recommended.

## Security

kost is designed with security as the top priority:

### API Credential Management

- **Environment variables only**: All API credentials are managed via environment variables
  - Kubernetes: kubeconfig file or in-cluster authentication
  - LLM: `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` environment variables
- **Not in config files**: Do not include API credentials in `config.yaml`
- **Not logged**: Error messages and reports do not contain credentials

### RBAC Permissions (Principle of Least Privilege)

kost requires only the following read-only permissions:
- Deployments (apps/deployments): `get`, `list`
- Pods (core/pods): `get`, `list`
- HPAs (autoscaling/horizontalpodautoscalers): `get`, `list`
- Namespaces (core/namespaces): `get`, `list`

See [examples/rbac.yaml](./examples/rbac.yaml) for details.

### TLS Certificate Verification

- TLS certificate verification is enabled by default
- Can be disabled in configuration, but not recommended for production

## Documentation

- [Quick Start Guide](./specs/001-mvp-core/quickstart.md)
- [Feature Specification](./specs/001-mvp-core/spec.md)
- [Data Model](./specs/001-mvp-core/data-model.md)
- [Technical Research](./specs/001-mvp-core/research.md)

## License

Apache License 2.0
