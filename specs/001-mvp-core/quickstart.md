# クイックスタートガイド: kost (Kubernetes Optimization & Sizing Tool)

**Date**: 2026-01-14
**Updated**: 2026-01-15
**Feature**: kost (Kubernetes Optimization & Sizing Tool) MVPコア機能
**Phase**: Phase 1 - Quickstart Guide
**Status**: ✅ 実装完了（P1-P2機能）

## 概要

このガイドでは、kost (Kubernetes Optimization & Sizing Tool) のインストールから最初のレポート生成までを10分以内に完了する手順を説明します。

**実装済み機能**:
- ✅ Deploymentリソース最適化（CPU/Memory requests推奨値）
- ✅ 適用可能なYAMLパッチ生成（Strategic Merge Patch形式）
- ✅ Markdownレポート、JSONサマリー生成
- ✅ labelSelectorとnamespace除外フィルタ
- ✅ 包括的な入力検証とセキュリティ機能

## 前提条件

以下の環境が整っていることを確認してください：

- **Go 1.25.5以上**: セキュリティ脆弱性対策のため（ビルドする場合）
- **Kubernetes クラスタ**: 稼働中のK8sクラスタへのアクセス権限
- **kubectl**: `kubectl` コマンドが使用可能で、対象クラスタに接続できること
- **Prometheus**: クラスタ内にPrometheusが導入済みで、以下のメトリクスが取得可能
  - `container_cpu_usage_seconds_total`
  - `container_memory_working_set_bytes`
- **RBAC権限**: Deployment/Podの読み取り権限（read-only）
- **メトリクス保持期間**: Prometheusが最低5分間のメトリクスを保持していること（本番環境では7日間推奨）

**セキュリティ要件**:
- API認証情報は環境変数で管理（設定ファイルに含めない）
- read-only権限のServiceAccountを使用

## Step 1: インストール

### バイナリダウンロード（推奨）

```bash
# 最新リリースをダウンロード
curl -LO https://github.com/lot-koichi/kost/releases/latest/download/kost-linux-amd64

# 実行権限を付与
chmod +x kost-linux-amd64

# /usr/local/binに配置
sudo mv kost-linux-amd64 /usr/local/bin/kost

# インストール確認
kost version
```

### Go環境からのビルド

```bash
# Go version確認（1.25.5以上必要）
go version

# リポジトリをクローン
git clone https://github.com/lot-koichi/kost.git
cd kost

# 依存関係のインストール
go mod download

# ビルド
go build -o bin/kost ./cmd/kost

# または Makefileを使用
make build

# バイナリを確認
./bin/kost version
```

### コンテナイメージ

```bash
# Dockerで実行
docker run --rm -v ~/.kube:/root/.kube \
  lot-koichi/kost:latest version
```

## Step 2: RBAC権限の設定

対象クラスタに以下のRBACリソースを適用します：

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

適用：

```bash
kubectl apply -f examples/rbac.yaml
```

## Step 3: 設定ファイルの作成

`config.yaml` を作成します：

```yaml
# config.yaml
kube:
  context: ""  # 空文字列でデフォルトコンテキストを使用

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
  enabled: false  # 初回はfalseで実行
  provider: "openai"
  model: "gpt-4"
  maxTokens: 1200
  temperature: 0.2
```

## Step 4: 最初のスキャン実行

対象Namespaceを指定してスキャンを実行します：

```bash
# 本番環境をスキャン
kost scan --namespace prod --config config.yaml

# 出力例:
# Scanning namespace: prod
# Found 15 Deployments
# Found 5 HPAs
# Scan completed in 2.3s
```

## Step 5: 推奨値の生成

スキャン結果を基に推奨値を生成します：

```bash
# 推奨値を生成
kost suggest --namespace prod --config config.yaml

# 出力例:
# Analyzing 15 Deployments...
# Querying Prometheus for CPU/Memory metrics (window: 7d)...
# Calculating recommendations...
# Found 8 overprovisioned containers
# Found 2 underprovisioned containers
# Found 3 HPA optimization opportunities
# Suggestion completed in 12.5s
```

## Step 6: レポート生成

推奨値を含むレポートを生成します：

```bash
# レポートを生成
kost report --namespace prod --config config.yaml

# 出力例:
# Generating report...
# Writing report to: ./out/report.md
# Writing summary to: ./out/summary.json
# Writing patches to: ./out/patches/prod/
# Report generated successfully!
```

## Step 7: レポートの確認

生成されたレポートを確認します：

```bash
# Markdownレポート
cat ./out/report.md

# JSONサマリ
cat ./out/summary.json | jq '.'

# パッチファイル一覧
ls -la ./out/patches/prod/
# 出力例:
# api-deployment.yaml
# api-deployment-hpa.yaml
# worker-deployment.yaml
```

## Step 8: パッチの適用（オプション）

推奨値を確認し、問題なければパッチを適用します：

```bash
# パッチ内容を確認
cat ./out/patches/prod/api-deployment.yaml

# dry-runで確認
kubectl apply -f ./out/patches/prod/api-deployment.yaml --dry-run=client

# 実際に適用
kubectl apply -f ./out/patches/prod/api-deployment.yaml

# HPA推奨値も適用
kubectl apply -f ./out/patches/prod/api-deployment-hpa.yaml
```

## LLMオプション（オプション）

LLMによる説明生成を有効化する場合：

### OpenAI使用時

```bash
# APIキーを環境変数に設定
export OPENAI_API_KEY="sk-..."

# config.yamlでLLMを有効化
# llm.enabled: true
# llm.provider: "openai"

# レポート生成
kost report --namespace prod --config config.yaml --llm on
```

### Claude使用時

```bash
# APIキーを環境変数に設定
export ANTHROPIC_API_KEY="sk-ant-..."

# config.yamlでLLMを有効化
# llm.enabled: true
# llm.provider: "claude"
# llm.model: "claude-3-5-sonnet-20241022"

# レポート生成
kost report --namespace prod --config config.yaml --llm on
```

## セキュリティのベストプラクティス

kostを安全に利用するための推奨事項：

### API認証情報の保護

- **環境変数で管理**: API認証情報は必ず環境変数で設定してください
  ```bash
  # 正しい方法
  export OPENAI_API_KEY="sk-..."

  # 間違った方法（config.yamlに直接記載しない）
  # llm.apiKey: "sk-..." ← これは禁止
  ```

- **設定ファイルから除外**: `.gitignore` に環境変数ファイル（`.env`）を追加してください
  ```bash
  echo ".env" >> .gitignore
  ```

- **kubeconfigの保護**: クラスタアクセス情報を含む `~/.kube/config` を適切に保護してください
  ```bash
  chmod 600 ~/.kube/config
  ```

### RBAC権限の検証

kostが必要とする最小権限が付与されているか確認してください：

```bash
# ServiceAccountの権限確認
kubectl auth can-i list deployments --as=system:serviceaccount:default:finops-advisor
kubectl auth can-i list pods --as=system:serviceaccount:default:finops-advisor
kubectl auth can-i list horizontalpodautoscalers --as=system:serviceaccount:default:finops-advisor

# 全て "yes" が返ることを確認
```

### 入力バリデーション

kostは自動的に入力を検証しますが、以下の点に注意してください：

- **Namespace名**: Kubernetes命名規則に準拠した名前を使用してください（小文字、数字、ハイフン、最大63文字）
  ```bash
  # 正しい例
  kost scan --namespace prod-app-v1

  # 間違った例
  kost scan --namespace "PROD_APP"  # 大文字とアンダースコアは不可
  ```

- **出力ディレクトリ**: パストラバーサル攻撃を防ぐため、相対パスまたは安全な絶対パスを使用してください
  ```bash
  # 正しい例
  output.dir: "./out"
  output.dir: "/tmp/kost-reports"

  # 間違った例（拒否されます）
  output.dir: "../../etc/passwd"
  output.dir: "/etc/sensitive"
  ```

### TLS証明書検証

Prometheus接続時のTLS証明書検証はデフォルトで有効です：

```yaml
# 本番環境では証明書検証を有効化（デフォルト）
prometheus:
  url: "https://prometheus.example.com"
  tlsVerify: true  # デフォルト値

# 開発環境で自己署名証明書を使用する場合のみ無効化
prometheus:
  url: "https://prometheus-dev.local"
  tlsVerify: false  # 本番環境では非推奨
```

## トラブルシューティング

### Prometheusに接続できない

```bash
# Prometheusエンドポイントを確認
kubectl get svc -n monitoring prometheus-operated

# ポートフォワードで接続確認
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090

# ブラウザで http://localhost:9090 にアクセスして確認
```

### 必要なメトリクスが取得できない

```bash
# Prometheusでメトリクスの存在確認
curl -s 'http://prometheus:9090/api/v1/query?query=container_cpu_usage_seconds_total' | jq '.'

# kube-state-metricsの確認
kubectl get pods -n monitoring | grep kube-state-metrics
```

### RBAC権限エラー

```bash
# 現在の権限を確認
kubectl auth can-i list deployments --as=system:serviceaccount:default:finops-advisor

# Roleの確認
kubectl describe clusterrole finops-advisor-reader
```

## 次のステップ

- **自動化**: GitHub Actions等のCIで定期実行し、レポートをArtifactとして保存
- **複数Namespace**: `--namespace` を複数回指定して一括分析
- **GitOps統合**: 生成されたパッチをPRとして自動作成（将来機能）

## まとめ

以上で、kost (Kubernetes Optimization & Sizing Tool) の基本的な使い方を習得しました。このツールは：

- ✅ 10分以内にインストールから最初のレポート生成まで完了
- ✅ Deployment と HPA の両方を最適化
- ✅ 適用可能なYAMLパッチを自動生成
- ✅ LLMによる自然言語説明（オプション）

詳細なドキュメントは [README.md](../../../README.md) を参照してください。
