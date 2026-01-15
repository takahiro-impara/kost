# kost (Kubernetes Optimization & Sizing Tool)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25.5%2B-00ADD8?logo=go)](https://go.dev/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/lot-koichi/kost/pulls)

Kubernetes Deploymentのリソース設定（requests/limits）およびHPA設定（minReplicas/maxReplicas）を最適化するCLIツールです。

## 概要

kostは、Prometheusメトリクスから過去の実績データを取得し、統計分析（P95/P99等）とルールベースの判定により、具体的な推奨値、適用可能なYAMLパッチ、削減率を記載したレポートを生成します。

## 主な機能

- ✅ Deployment リソース最適化の分析と推奨値提示
- ✅ 適用可能なYAMLパッチの生成
- ✅ HPA設定の最適化推奨
- ✅ 複数LLMプロバイダ対応（OpenAI、Claude）
- ✅ 生成AIによる説明と優先順位付け

## クイックスタート

### インストール

```bash
# バイナリダウンロード
curl -LO https://github.com/lot-koichi/kost/releases/latest/download/kost-linux-amd64
chmod +x kost-linux-amd64
sudo mv kost-linux-amd64 /usr/local/bin/kost

# Go環境からビルド
git clone https://github.com/lot-koichi/kost.git
cd kost
make build
```

### 使い方

```bash
# 設定ファイルの作成
cp examples/config.yaml config.yaml

# Deploymentをスキャン
kost scan --namespace prod --config config.yaml

# 推奨値を生成
kost suggest --namespace prod --config config.yaml

# レポートを生成
kost report --namespace prod --config config.yaml
```

詳細は[quickstart.md](./specs/001-mvp-core/quickstart.md)を参照してください。

## Local環境での検証

minikubeまたはKindを使ってローカル環境でkostを検証する手順です。

### 方法1: minikubeを使用（推奨）

#### 1. minikubeのインストール

**macOS:**
```bash
brew install minikube
```

**Linux:**
```bash
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```

#### 2. minikubeクラスタの起動

```bash
# クラスタを起動（メトリクス収集を有効化）
minikube start --cpus=4 --memory=8192 --addons=metrics-server

# kubectlがminikubeを向いていることを確認
kubectl config current-context
# 出力: minikube
```

#### 3. Prometheusのインストール

```bash
# Helmのインストール（未インストールの場合）
brew install helm  # macOS
# または
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash  # Linux

# PrometheusをインストールKube-state-metricsと共に）
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false

# Prometheusの起動を待つ（2-3分）
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n monitoring --timeout=300s
```

#### 4. サンプルアプリケーションのデプロイ

```bash
# サンプルDeploymentを作成
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
            cpu: "500m"      # 過剰なリクエスト（検証用）
            memory: "512Mi"  # 過剰なリクエスト（検証用）
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

# Podが起動するまで待つ
kubectl wait --for=condition=ready pod -l app=nginx --timeout=60s

# 負荷をかける（メトリクスを生成）
kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://nginx-demo; done" &

# 5分間待って、メトリクスを蓄積
echo "Waiting for metrics to accumulate (5 minutes)..."
sleep 300
```

#### 5. kostの実行

```bash
# Prometheusにポートフォワード
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090 &

# kostのconfig.yamlを準備
cat <<EOF > config.yaml
kube:
  context: "minikube"

prometheus:
  url: "http://localhost:9090"
  timeoutSeconds: 30

analysis:
  window: "5m"  # ローカル検証では短い期間を使用
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

# kostでスキャン
./bin/kost scan -n default --config config.yaml

# 推奨値を生成
./bin/kost suggest -n default --config config.yaml

# レポートを生成
./bin/kost report -n default --config config.yaml

# レポートを確認
cat ./out/report.md
```

#### 6. クリーンアップ

```bash
# ポートフォワードを停止
pkill -f "port-forward"

# サンプルアプリを削除
kubectl delete deployment nginx-demo
kubectl delete hpa nginx-demo-hpa

# minikubeクラスタを削除（必要に応じて）
minikube delete
```

### 方法2: Kindを使用

#### 1. Kindのインストール

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

#### 2. Kindクラスタの作成

```bash
# クラスタ設定ファイルを作成
cat <<EOF > kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
EOF

# クラスタを作成
kind create cluster --name kost-test --config kind-config.yaml

# kubectlがkindを向いていることを確認
kubectl config current-context
# 出力: kind-kost-test
```

#### 3. Prometheusのインストール

```bash
# minikubeの場合と同じ手順
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false

kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n monitoring --timeout=300s
```

#### 4-6. サンプルアプリのデプロイ、kost実行、クリーンアップ

minikubeの場合と同じ手順で実行します。

```bash
# クラスタを削除（必要に応じて）
kind delete cluster --name kost-test
```

## 前提条件

- **Go 1.25.5以上**（セキュリティ脆弱性対策のため）
- Kubernetes クラスタへのアクセス権限
- Prometheusが導入済み（必要なメトリクスが取得可能）
- RBAC権限（Deployment/Pod/HPAの読み取り専用）

**重要**: Go 1.25.3以前のバージョンには既知のセキュリティ脆弱性（GO-2025-4175, GO-2025-4155）が存在します。Go 1.25.5以上へのアップグレードを強く推奨します。

## セキュリティ

kostはセキュリティを最優先に設計されています：

### API認証情報の管理

- **環境変数のみ**: 全てのAPI認証情報は環境変数で管理します
  - Kubernetes: kubeconfigファイルまたはin-cluster認証
  - LLM: `OPENAI_API_KEY` または `ANTHROPIC_API_KEY` 環境変数
- **設定ファイルには含めない**: API認証情報を `config.yaml` に記載しないでください
- **ログに出力しない**: エラーメッセージやレポートに認証情報は含まれません

### RBAC権限（最小権限の原則）

kostは以下の読み取り専用権限のみを必要とします：
- Deployments（apps/deployments）: `get`, `list`
- Pods（core/pods）: `get`, `list`
- HPAs（autoscaling/horizontalpodautoscalers）: `get`, `list`
- Namespaces（core/namespaces）: `get`, `list`

詳細は [examples/rbac.yaml](./examples/rbac.yaml) を参照してください。

### 入力検証

kostは以下の入力を検証し、セキュリティ攻撃を防止します：
- **Namespace名**: Kubernetes命名規則に準拠しているか検証
- **Label selector**: インジェクション攻撃を防ぐ文字列検証
- **Prometheus URL**: 有効なHTTP/HTTPSスキームの検証
- **出力パス**: パストラバーサル攻撃の防止
- **PromQLクエリ**: インジェクション攻撃の防止

### 脆弱性管理

kostのCI/CDパイプラインには以下のセキュリティスキャンが組み込まれています：
- **gosec**: Go言語の静的セキュリティ解析 ✅ クリーン（0件）
- **govulncheck**: Go依存関係の既知脆弱性スキャン ⚠️ Go標準ライブラリの既知脆弱性あり（対策方法記載）
- **Trivy**: コンテナイメージおよびファイルシステムの脆弱性スキャン

詳細なセキュリティテスト結果は [TEST_RESULTS.md](./TEST_RESULTS.md) を参照してください。

**セキュリティスキャンの実行方法**:
```bash
# gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### TLS証明書検証

- デフォルトでTLS証明書検証が有効です
- 設定で無効化可能ですが、本番環境では推奨しません

## ドキュメント

### 機能仕様
- [クイックスタートガイド](./specs/001-mvp-core/quickstart.md)
- [機能仕様書](./specs/001-mvp-core/spec.md)
- [データモデル](./specs/001-mvp-core/data-model.md)
- [技術調査](./specs/001-mvp-core/research.md)

### テストとセキュリティ
- [テスト結果レポート](./TEST_RESULTS.md) - セキュリティスキャン結果、E2Eテスト結果、既知の問題
- [セキュリティテスト計画](./SECURITY_TEST_PLAN.md) - セキュリティチェックリスト、ペネトレーションテスト
- [E2Eテスト計画](./E2E_TEST_PLAN.md) - 包括的なテストシナリオ

## ライセンス

Apache License 2.0
