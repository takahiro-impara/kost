# E2Eテスト計画

## 概要
minikube環境を使用した包括的なE2Eテスト計画。仕様書の全機能を網羅的にテストする。

## テスト環境

### 前提条件
- minikube v1.30+
- kubectl v1.28+
- Helm v3.12+
- Go 1.21+

### セットアップ
```bash
# minikubeクラスタ起動
minikube start --cpus=4 --memory=8192 --addons=metrics-server

# Prometheusインストール
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false

# 起動確認
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n monitoring --timeout=300s
```

## テストシナリオ

### シナリオ1: 基本的なスキャンと推奨値生成

#### 1.1 単一Deploymentのスキャン
```bash
# テストデータ作成
kubectl create namespace test-basic
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-overprovisioned
  namespace: test-basic
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx-test
  template:
    metadata:
      labels:
        app: nginx-test
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        resources:
          requests:
            cpu: "500m"      # 意図的に過剰
            memory: "512Mi"  # 意図的に過剰
          limits:
            cpu: "1000m"
            memory: "1Gi"
EOF

# メトリクス蓄積のため5分待機
sleep 300

# スキャン実行
./bin/kost scan -n test-basic --config config.yaml

# 期待結果: 1 Deployment検出
```

**検証項目**:
- [ ] Deploymentが正しく検出される
- [ ] コンテナ情報が取得できる
- [ ] エラーが発生しない

#### 1.2 推奨値生成
```bash
# 推奨値生成
./bin/kost suggest -n test-basic --config config.yaml

# 期待結果:
# - Total containers analyzed: 1
# - Overprovisioned containers: 1
# - CPU/Memory推奨値が生成される
```

**検証項目**:
- [ ] メトリクスが正しく取得される
- [ ] P95統計値が計算される
- [ ] Safety factor (1.2) が適用される
- [ ] オーバープロビジョニングが検出される

#### 1.3 レポート生成
```bash
# レポート生成
./bin/kost report -n test-basic --config config.yaml

# レポート確認
cat out/report.md
cat out/summary.json
cat out/patches/test-basic/nginx-overprovisioned.yaml
```

**検証項目**:
- [ ] Markdownレポートが生成される
- [ ] JSONサマリーが生成される
- [ ] YAMLパッチが生成される
- [ ] パッチがkubectl dry-runでvalidationを通る

### シナリオ2: 複数Deploymentの同時分析

#### 2.1 テストデータ作成
```bash
kubectl create namespace test-multi

# Deployment 1: オーバープロビジョニング
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-overprovisioned
  namespace: test-multi
  labels:
    tier: frontend
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
        tier: frontend
    spec:
      containers:
      - name: app
        image: nginx:latest
        resources:
          requests:
            cpu: "200m"
            memory: "256Mi"
EOF

# Deployment 2: 適切なプロビジョニング
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-appropriate
  namespace: test-multi
  labels:
    tier: backend
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
        tier: backend
    spec:
      containers:
      - name: app
        image: nginx:latest
        resources:
          requests:
            cpu: "10m"
            memory: "32Mi"
EOF

# Deployment 3: リソース未設定
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-no-resources
  namespace: test-multi
  labels:
    tier: worker
spec:
  replicas: 1
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
        tier: worker
    spec:
      containers:
      - name: app
        image: nginx:latest
        # resourcesを設定しない
EOF

sleep 300
```

#### 2.2 全Deploymentスキャン
```bash
./bin/kost scan -n test-multi --config config.yaml

# 期待結果: Found 3 Deployments
```

**検証項目**:
- [ ] 3つのDeploymentすべてが検出される
- [ ] 各Deploymentのコンテナ情報が取得できる

#### 2.3 推奨値生成と判定
```bash
./bin/kost suggest -n test-multi --config config.yaml

# 期待結果:
# - Total containers: 3
# - Overprovisioned: 1
# - Appropriate: 1
# - Not set: 1
```

**検証項目**:
- [ ] オーバープロビジョニングの判定が正しい
- [ ] 適切なプロビジョニングの判定が正しい
- [ ] リソース未設定の検出が正しい

### シナリオ3: Label Selectorフィルタ

#### 3.1 labelSelectorによるフィルタリング
```bash
# config.yamlを編集
cat > config-filtered.yaml <<EOF
kube:
  context: "minikube"

prometheus:
  url: "http://localhost:9090"
  timeoutSeconds: 30

analysis:
  window: "5m"
  cpuPercentile: 0.95
  memPercentile: 0.95
  safetyFactor: 1.2
  minCpuMilli: 10
  minMemMi: 32

filters:
  namespacesExclude:
    - kube-system
    - monitoring
  labelSelector: "tier=frontend"  # frontendのみ

output:
  dir: "./out"
  format:
    - md
    - json
    - patch

llm:
  enabled: false
EOF

# フィルタ付きスキャン
./bin/kost scan -n test-multi --config config-filtered.yaml

# 期待結果: Found 1 Deployment (app-overprovisioned のみ)
```

**検証項目**:
- [ ] labelSelectorが正しく動作する
- [ ] 条件に合うDeploymentのみが対象になる

### シナリオ4: Namespace除外フィルタ

#### 4.1 除外リストのテスト
```bash
# 複数namespaceにDeployment作成
kubectl create namespace test-excluded
kubectl apply -n test-excluded -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: should-be-excluded
spec:
  replicas: 1
  selector:
    matchLabels:
      app: excluded
  template:
    metadata:
      labels:
        app: excluded
    spec:
      containers:
      - name: app
        image: nginx:latest
EOF

# 除外設定
cat > config-exclude.yaml <<EOF
kube:
  context: "minikube"

prometheus:
  url: "http://localhost:9090"
  timeoutSeconds: 30

analysis:
  window: "5m"
  cpuPercentile: 0.95
  memPercentile: 0.95
  safetyFactor: 1.2
  minCpuMilli: 10
  minMemMi: 32

filters:
  namespacesExclude:
    - kube-system
    - monitoring
    - test-excluded  # 除外
  labelSelector: ""

output:
  dir: "./out"
  format:
    - md

llm:
  enabled: false
EOF

# スキャン
./bin/kost scan -n test-excluded --config config-exclude.yaml

# 期待結果: namespaceがフィルタされることを確認
```

**検証項目**:
- [ ] 除外リストが正しく動作する
- [ ] kube-system, monitoringが除外される

### シナリオ5: エラーハンドリング

#### 5.1 Prometheus接続エラー
```bash
# ポートフォワードを停止
pkill -f "port-forward.*prometheus"

# 実行
./bin/kost suggest -n test-basic --config config.yaml

# 期待結果: エラーメッセージが表示される
# "failed to create prometheus client"
# "Prometheus URL is accessible"などのヒントが表示される
```

**検証項目**:
- [ ] 適切なエラーメッセージが表示される
- [ ] トラブルシューティングのヒントが提供される
- [ ] クラッシュせずに終了する

#### 5.2 メトリクス不足
```bash
# 新規Deploymentを作成（メトリクスなし）
kubectl apply -n test-basic -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: no-metrics
spec:
  replicas: 1
  selector:
    matchLabels:
      app: no-metrics
  template:
    metadata:
      labels:
        app: no-metrics
    spec:
      containers:
      - name: app
        image: nginx:latest
EOF

# すぐに実行（メトリクス蓄積前）
./bin/kost suggest -n test-basic --config config.yaml

# 期待結果:
# "Warning: No metrics found for ..."
# トラブルシューティングヒント表示
```

**検証項目**:
- [ ] メトリクス不足時の警告が表示される
- [ ] 他のDeploymentの分析は継続される
- [ ] 有用なヒントが提供される

#### 5.3 不正な設定ファイル
```bash
# 不正なconfig.yaml
cat > config-invalid.yaml <<EOF
kube:
  context: "minikube"

prometheus:
  url: "ftp://invalid-url"  # 不正なスキーム

analysis:
  window: "5m"
  cpuPercentile: 0.95
  memPercentile: 0.95
  safetyFactor: -1.0  # 不正な値
  minCpuMilli: 10
  minMemMi: 32
EOF

# 実行
./bin/kost suggest -n test-basic --config config-invalid.yaml

# 期待結果: バリデーションエラー
```

**検証項目**:
- [ ] 設定ファイルのバリデーションが動作する
- [ ] 不正な値が拒否される
- [ ] エラーメッセージが具体的

### シナリオ6: 多様なワークロードパターン

#### 6.1 CPU intensive workload
```bash
kubectl create namespace test-cpu-intensive
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cpu-intensive
  namespace: test-cpu-intensive
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cpu-test
  template:
    metadata:
      labels:
        app: cpu-test
    spec:
      containers:
      - name: stress
        image: polinux/stress
        command: ["stress"]
        args: ["--cpu", "1", "--timeout", "600s"]
        resources:
          requests:
            cpu: "100m"
            memory: "64Mi"
EOF

# 10分待機してメトリクス蓄積
sleep 600

# 分析
./bin/kost suggest -n test-cpu-intensive --config config.yaml

# 期待結果: CPUがunderprovisioned判定される
```

**検証項目**:
- [ ] CPU使用率が高いワークロードを正しく検出
- [ ] アンダープロビジョニング判定が機能する
- [ ] 適切な推奨値が提示される

#### 6.2 Memory intensive workload
```bash
kubectl create namespace test-memory-intensive
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-intensive
  namespace: test-memory-intensive
spec:
  replicas: 1
  selector:
    matchLabels:
      app: memory-test
  template:
    metadata:
      labels:
        app: memory-test
    spec:
      containers:
      - name: stress
        image: polinux/stress
        command: ["stress"]
        args: ["--vm", "1", "--vm-bytes", "100M", "--timeout", "600s"]
        resources:
          requests:
            cpu: "50m"
            memory: "50Mi"  # 不足
EOF

sleep 600

./bin/kost suggest -n test-memory-intensive --config config.yaml

# 期待結果: Memoryがunderprovisioned判定される
```

**検証項目**:
- [ ] メモリ使用率が高いワークロードを正しく検出
- [ ] アンダープロビジョニング判定が機能する

### シナリオ7: HPA連携テスト

#### 7.1 HPA設定のあるDeployment
```bash
kubectl create namespace test-hpa
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-with-hpa
  namespace: test-hpa
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hpa-test
  template:
    metadata:
      labels:
        app: hpa-test
    spec:
      containers:
      - name: app
        image: nginx:latest
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app-with-hpa
  namespace: test-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app-with-hpa
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

sleep 300

./bin/kost scan -n test-hpa --config config.yaml
```

**検証項目**:
- [ ] HPAがあるDeploymentを正しく検出
- [ ] HPA設定が取得できる（将来の機能拡張用）

### シナリオ8: 統計計算の検証

#### 8.1 P50/P95/P99パーセンタイルの計算
```bash
# 異なるパーセンタイル設定でテスト
cat > config-p99.yaml <<EOF
kube:
  context: "minikube"

prometheus:
  url: "http://localhost:9090"
  timeoutSeconds: 30

analysis:
  window: "5m"
  cpuPercentile: 0.99  # P99に変更
  memPercentile: 0.99  # P99に変更
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

llm:
  enabled: false
EOF

./bin/kost suggest -n test-basic --config config-p99.yaml
```

**検証項目**:
- [ ] P99設定が正しく適用される
- [ ] 推奨値がP95より高くなる

#### 8.2 Safety Factor適用確認
```bash
# Safety Factor 1.5でテスト
cat > config-sf15.yaml <<EOF
kube:
  context: "minikube"

prometheus:
  url: "http://localhost:9090"
  timeoutSeconds: 30

analysis:
  window: "5m"
  cpuPercentile: 0.95
  memPercentile: 0.95
  safetyFactor: 1.5  # 1.5倍
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
    - json

llm:
  enabled: false
EOF

./bin/kost suggest -n test-basic --config config-sf15.yaml

# JSONを確認してsafetyFactorが適用されているか検証
cat out/summary.json | grep safetyFactor
```

**検証項目**:
- [ ] Safety Factorが正しく適用される
- [ ] 推奨値 = P95 × SafetyFactor になる

### シナリオ9: 出力形式の検証

#### 9.1 全出力形式の生成
```bash
cat > config-all-formats.yaml <<EOF
kube:
  context: "minikube"

prometheus:
  url: "http://localhost:9090"
  timeoutSeconds: 30

analysis:
  window: "5m"
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

./bin/kost report -n test-multi --config config-all-formats.yaml

# ファイル確認
ls -lh out/
ls -lh out/patches/test-multi/
```

**検証項目**:
- [ ] report.md が生成される
- [ ] summary.json が生成される
- [ ] patches/**/*.yaml が生成される
- [ ] 各ファイルが有効な形式である

#### 9.2 パッチのバリデーション
```bash
# 全パッチファイルのバリデーション
for patch in out/patches/test-multi/*.yaml; do
  echo "Validating: $patch"
  kubectl apply -f "$patch" --dry-run=client
done
```

**検証項目**:
- [ ] 全パッチがkubectl validationを通過する
- [ ] Strategic Merge Patch形式が正しい

### シナリオ10: パフォーマンステスト

#### 10.1 大量Deploymentの処理
```bash
kubectl create namespace test-performance

# 50個のDeployment作成
for i in {1..50}; do
  kubectl apply -n test-performance -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-$i
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app-$i
  template:
    metadata:
      labels:
        app: app-$i
    spec:
      containers:
      - name: app
        image: nginx:latest
        resources:
          requests:
            cpu: "10m"
            memory: "32Mi"
EOF
done

sleep 300

# パフォーマンス測定
time ./bin/kost suggest -n test-performance --config config.yaml
```

**検証項目**:
- [ ] 大量Deploymentを処理できる
- [ ] メモリリークが発生しない
- [ ] 処理時間が妥当（50個で数秒以内）

## クリーンアップ

```bash
# 全テストnamespaceを削除
kubectl delete namespace test-basic
kubectl delete namespace test-multi
kubectl delete namespace test-excluded
kubectl delete namespace test-cpu-intensive
kubectl delete namespace test-memory-intensive
kubectl delete namespace test-hpa
kubectl delete namespace test-performance

# ポートフォワードを停止
pkill -f "port-forward"

# minikubeクラスタ削除（オプション）
minikube delete
```

## テスト実行チェックリスト

### 基本機能
- [ ] シナリオ1: 単一Deployment
- [ ] シナリオ2: 複数Deployment
- [ ] シナリオ3: labelSelectorフィルタ
- [ ] シナリオ4: namespace除外フィルタ

### エラーハンドリング
- [ ] シナリオ5.1: Prometheus接続エラー
- [ ] シナリオ5.2: メトリクス不足
- [ ] シナリオ5.3: 不正な設定

### ワークロードパターン
- [ ] シナリオ6.1: CPU intensive
- [ ] シナリオ6.2: Memory intensive
- [ ] シナリオ7: HPA連携

### 統計・計算
- [ ] シナリオ8.1: パーセンタイル計算
- [ ] シナリオ8.2: Safety Factor適用

### 出力形式
- [ ] シナリオ9.1: 全出力形式
- [ ] シナリオ9.2: パッチバリデーション

### パフォーマンス
- [ ] シナリオ10: 大量Deployment処理

## 成功基準

1. **全シナリオがパス**: 全10シナリオがエラーなく完了すること
2. **判定精度**: オーバー/アンダー/適切の判定が正確であること
3. **エラーハンドリング**: エラー時に有用なメッセージが表示されること
4. **パフォーマンス**: 50 Deploymentの処理が10秒以内に完了すること
5. **出力品質**: 生成されたパッチがすべてvalidであること

## テスト結果記録

各シナリオの実行結果を記録：

| シナリオ | ステータス | 実行時間 | 備考 |
|---------|-----------|---------|------|
| 1.1 単一Deployment | - | - | |
| 1.2 推奨値生成 | - | - | |
| 1.3 レポート生成 | - | - | |
| ... | - | - | |
