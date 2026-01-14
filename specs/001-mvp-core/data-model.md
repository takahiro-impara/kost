# データモデル: kost (Kubernetes Optimization & Sizing Tool) MVPコア機能

**Date**: 2026-01-14
**Feature**: kost (Kubernetes Optimization & Sizing Tool) MVPコア機能
**Phase**: Phase 1 - Data Model Design

## 概要

本ドキュメントでは、kost (Kubernetes Optimization & Sizing Tool) の内部データモデルを定義します。各エンティティの構造、関係、バリデーションルールを記載し、実装時の型定義の基礎とします。

## Core Entities

### 1. Config (設定)

**Purpose**: アプリケーション全体の設定を保持

**Fields**:
- `kubeContext` (string, optional): Kubernetes context名（空文字列の場合はデフォルト）
- `prometheusURL` (string, required): PrometheusエンドポイントURL
- `prometheusTimeoutSeconds` (int, default: 10): Prometheusクエリタイムアウト
- `analysisWindow` (string, default: "7d"): 分析期間（例: "7d", "14d"）
- `cpuPercentile` (float64, default: 0.95): CPU推奨値のパーセンタイル
- `memPercentile` (float64, default: 0.95): Memory推奨値のパーセンタイル
- `safetyFactor` (float64, default: 1.2): requests/limits推奨値の安全係数
- `minCpuMilli` (int, default: 20): 最小CPU requests（ミリコア）
- `minMemMi` (int, default: 64): 最小Memory requests（MiB）
- `hpaMinReplicasFactor` (float64, default: 0.8): HPA minReplicas推奨値の係数
- `hpaMaxReplicasFactor` (float64, default: 1.3): HPA maxReplicas推奨値の係数
- `namespacesExclude` ([]string, default: ["kube-system", "monitoring"]): 除外するNamespace
- `labelSelector` (string, optional): Deploymentフィルタ用のラベルセレクタ
- `outputDir` (string, default: "./out"): 出力ディレクトリ
- `outputFormats` ([]string, default: ["md", "json", "patch"]): 出力形式
- `llmEnabled` (bool, default: false): LLM機能の有効/無効
- `llmProvider` (string, default: "openai"): LLMプロバイダ（"openai" or "claude"）
- `llmModel` (string, optional): LLMモデル名（例: "gpt-4", "claude-3-5-sonnet-20241022"）
- `llmMaxTokens` (int, default: 1200): LLMレスポンスの最大トークン数
- `llmTemperature` (float64, default: 0.2): LLM Temperature

**Validation**:
- `prometheusURL`: 必須、有効なHTTP/HTTPS URL（`security.ValidatePrometheusURL()`）
- `cpuPercentile`, `memPercentile`: 0.0 < x <= 1.0
- `safetyFactor`: x >= 1.0
- `hpaMinReplicasFactor`: 0.0 < x <= 1.0
- `hpaMaxReplicasFactor`: x >= 1.0
- `llmProvider`: "openai" or "claude"
- `outputDir`: パストラバーサル検証（`security.ValidateOutputPath()`）
- `labelSelector`: インジェクション攻撃防止（`security.ValidateLabelSelector()`）

**Security**:
- API認証情報（LLM APIキー等）は環境変数のみから取得し、設定ファイルに含めない
- 環境変数: `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`

**Source**: config.yaml + 環境変数（環境変数が優先）

---

### 2. Deployment (Kubernetesワークロード)

**Purpose**: K8s Deployment情報を保持

**Fields**:
- `namespace` (string, required): Deployment が属する Namespace
- `name` (string, required): Deployment 名
- `containers` ([]Container, required): コンテナリスト
- `hasHPA` (bool): HPAが設定されているか
- `hpaRef` (*HPA, optional): 関連するHPA（hasHPA=true の場合のみ）

**Relationships**:
- 1 Deployment → N Containers
- 1 Deployment → 0..1 HPA

**Validation**:
- `namespace`: 必須、Kubernetes命名規則に準拠（`security.ValidateNamespace()`）
- `name`: 必須、非空文字列
- `containers`: 1つ以上のコンテナが必要

**Security**:
- Namespace名はRFC 1123準拠（小文字英数字、ハイフン、最大63文字）を検証

**Source**: Kubernetes API（client-go, Deployments.List）

---

### 3. Container (コンテナ)

**Purpose**: Deployment内の個別コンテナ情報を保持

**Fields**:
- `name` (string, required): コンテナ名
- `cpuRequestMilli` (int, optional): 現在のCPU requests（ミリコア）
- `memRequestMi` (int, optional): 現在のMemory requests（MiB）
- `cpuLimitMilli` (int, optional): 現在のCPU limits（ミリコア）
- `memLimitMi` (int, optional): 現在のMemory limits（MiB）

**Validation**:
- `name`: 必須、非空文字列
- `cpuRequestMilli`, `memRequestMi`: nil（未設定）または正の整数
- limits >= requests（設定されている場合）

**Source**: Kubernetes API（Deployment.Spec.Template.Spec.Containers）

---

### 4. HPA (Horizontal Pod Autoscaler)

**Purpose**: HPA設定情報を保持

**Fields**:
- `namespace` (string, required): HPA が属する Namespace
- `name` (string, required): HPA 名
- `targetDeployment` (string, required): 対象Deployment名
- `minReplicas` (int, required): 現在のminReplicas
- `maxReplicas` (int, required): 現在のmaxReplicas
- `targetCPUUtilization` (int, optional): targetCPUUtilizationPercentage
- `targetMemoryUtilization` (int, optional): targetMemoryUtilizationPercentage（K8s 1.23+）

**Validation**:
- `namespace`, `name`, `targetDeployment`: 必須、非空文字列
- `minReplicas` >= 1
- `maxReplicas` >= `minReplicas`

**Source**: Kubernetes API（client-go, HorizontalPodAutoscalers.List）

---

### 5. Metrics (メトリクス)

**Purpose**: Prometheusから取得したメトリクスデータを保持

**Fields**:
- `deploymentName` (string, required): 対象Deployment名
- `containerName` (string, optional): 対象コンテナ名（Container metricsの場合）
- `metricType` (string, required): メトリクス種別（"cpu", "memory", "replicas"）
- `values` ([]float64, required): 時系列データ
- `p50` (float64): P50パーセンタイル
- `p95` (float64): P95パーセンタイル
- `p99` (float64): P99パーセンタイル
- `min` (float64): 最小値
- `max` (float64): 最大値
- `avg` (float64): 平均値

**Validation**:
- `values`: 1つ以上のデータポイントが必要
- `metricType`: "cpu", "memory", "replicas" のいずれか

**Source**: Prometheus API（QueryRange）

**Queries**:
- **CPU**: `rate(container_cpu_usage_seconds_total{namespace="<ns>",pod=~"<deployment>-.*",container="<container>"}[5m])`
- **Memory**: `container_memory_working_set_bytes{namespace="<ns>",pod=~"<deployment>-.*",container="<container>"}`
- **Replicas**: `kube_deployment_status_replicas{namespace="<ns>",deployment="<deployment>"}`

**Security**:
- PromQLクエリは`security.SanitizePromQLQuery()`で検証し、インジェクション攻撃を防止
- 禁止文字: `;`, `--`, `/*`, `*/`, `\`, `\n`, `\r`

---

### 6. ResourceRecommendation (リソース推奨値)

**Purpose**: requests/limitsの推奨値と判定結果を保持

**Fields**:
- `deployment` (string, required): 対象Deployment名
- `container` (string, required): 対象コンテナ名
- `currentCPURequestMilli` (int, optional): 現在のCPU requests
- `recommendedCPURequestMilli` (int, required): 推奨CPU requests
- `currentMemRequestMi` (int, optional): 現在のMemory requests
- `recommendedMemRequestMi` (int, required): 推奨Memory requests
- `cpuJudgement` (string, required): CPU判定（"overprovisioned", "underprovisioned", "appropriate"）
- `memJudgement` (string, required): Memory判定（"overprovisioned", "underprovisioned", "appropriate"）
- `cpuSavingRatio` (float64): CPU削減率（0.0〜1.0、負の値は増加を意味）
- `memSavingRatio` (float64): Memory削減率（0.0〜1.0、負の値は増加を意味）
- `rationale` (Rationale, required): 推奨値の根拠

**Sub-Type: Rationale**:
- `cpuP95` (float64): CPU P95値（ミリコア）
- `memP95` (float64): Memory P95値（MiB）
- `safetyFactor` (float64): 適用された安全係数
- `minCpuMilli` (int): 適用された最小CPU値
- `minMemMi` (int): 適用された最小Memory値

**Validation**:
- `recommendedCPURequestMilli` >= Config.minCpuMilli
- `recommendedMemRequestMi` >= Config.minMemMi
- `cpuJudgement`, `memJudgement`: "overprovisioned", "underprovisioned", "appropriate" のいずれか

**Calculation**:
```
recommendedCPURequestMilli = max(cpuP95 * safetyFactor, minCpuMilli)
recommendedMemRequestMi = max(memP95 * safetyFactor, minMemMi)

cpuJudgement =
  if currentCPURequestMilli > cpuP95 * 2.0 then "overprovisioned"
  elif currentCPURequestMilli < cpuP95 * 1.1 then "underprovisioned"
  else "appropriate"

cpuSavingRatio = (currentCPURequestMilli - recommendedCPURequestMilli) / currentCPURequestMilli
```

---

### 7. HPARecommendation (HPA推奨値)

**Purpose**: HPA minReplicas/maxReplicasの推奨値と判定結果を保持

**Fields**:
- `deployment` (string, required): 対象Deployment名
- `hpaName` (string, required): 対象HPA名
- `currentMinReplicas` (int, required): 現在のminReplicas
- `recommendedMinReplicas` (int, required): 推奨minReplicas
- `currentMaxReplicas` (int, required): 現在のmaxReplicas
- `recommendedMaxReplicas` (int, required): 推奨maxReplicas
- `minJudgement` (string, required): minReplicas判定（"too_high", "too_low", "appropriate"）
- `maxJudgement` (string, required): maxReplicas判定（"too_high", "too_low", "appropriate"）
- `rationale` (HPARationale, required): 推奨値の根拠

**Sub-Type: HPARationale**:
- `replicasP5` (float64): Replicas P5値
- `replicasP99` (float64): Replicas P99値
- `minReplicasFactor` (float64): minReplicas係数
- `maxReplicasFactor` (float64): maxReplicas係数

**Validation**:
- `recommendedMinReplicas` >= 1
- `recommendedMaxReplicas` >= `recommendedMinReplicas`

**Calculation**:
```
recommendedMinReplicas = max(ceil(replicasP5 * minReplicasFactor), 1)
recommendedMaxReplicas = max(ceil(replicasP99 * maxReplicasFactor), recommendedMinReplicas)

minJudgement =
  if currentMinReplicas > replicasP5 * 1.5 then "too_high"
  elif currentMinReplicas < replicasP5 * 0.8 then "too_low"
  else "appropriate"
```

---

### 8. LLMProvider (LLMプロバイダ)

**Purpose**: LLMプロバイダの抽象化

**Fields**:
- `name` (string, required): プロバイダ名（"openai", "claude"）
- `apiKey` (string, required): API認証キー（環境変数から取得）
- `model` (string, required): モデル名
- `endpoint` (string, optional): カスタムエンドポイント（通常は不要）

**Interface** (Go):
```go
type LLMProvider interface {
    GenerateNarrative(ctx context.Context, input StructuredInput) (string, error)
}
```

**Validation**:
- `name`: "openai" or "claude"
- `apiKey`: 必須、非空文字列
- `model`: 必須、非空文字列

**Security**:
- APIキーは環境変数（`OPENAI_API_KEY`, `ANTHROPIC_API_KEY`）からのみ取得
- ログ出力時は`security.MaskSensitiveValue()`でマスキング（例: "sk-1234...abcd"）
- 設定ファイル、レポート、エラーメッセージには含めない

**Source**: Config + 環境変数（`OPENAI_API_KEY`, `ANTHROPIC_API_KEY`）

---

### 9. Report (レポート)

**Purpose**: 最終出力物を保持

**Fields**:
- `namespace` (string, required): 対象Namespace
- `window` (string, required): 分析期間（例: "7d"）
- `percentile` (string, required): 使用したパーセンタイル（例: "P95"）
- `safetyFactor` (float64, required): 適用された安全係数
- `resourceRecommendations` ([]ResourceRecommendation, required): リソース推奨値リスト
- `hpaRecommendations` ([]HPARecommendation, required): HPA推奨値リスト
- `narrative` (string, optional): LLM生成の説明文
- `generatedAt` (time.Time, required): レポート生成日時

**Outputs**:
- **Markdown**: `{outputDir}/report.md`
- **JSON**: `{outputDir}/summary.json`
- **Patches**: `{outputDir}/patches/{namespace}/{deployment}.yaml`, `{outputDir}/patches/{namespace}/{deployment}-hpa.yaml`

**Markdown Structure**:
```markdown
# kost report

対象: namespace={namespace} / window={window} / percentile={percentile} / safety={safetyFactor}

## Top findings
1. Deployment `{deployment}` container `{container}`
   - 現状 request(cpu)={current}m / P95={p95}m → 推奨={recommended}m（{saving}%）
   - 理由: {rationale}
   - 注意: {注意点}

## HPA recommendations
1. HPA `{hpa}` for Deployment `{deployment}`
   - 現状 minReplicas={current} / P5={p5} → 推奨={recommended}
   - 現状 maxReplicas={current} / P99={p99} → 推奨={recommended}

## Suggested patches
- patches/{namespace}/{deployment}.yaml
- patches/{namespace}/{deployment}-hpa.yaml

## Method
- CPU: P95(rate(container_cpu_usage_seconds_total[5m])) × safety
- Memory: P95(container_memory_working_set_bytes) × safety
- HPA min: P5(kube_deployment_status_replicas) × factor
- HPA max: P99(kube_deployment_status_replicas) × factor
```

---

## State Transitions

本アプリケーションはステートレスなCLIツールのため、永続的な状態遷移はありません。実行フローは以下の通りです：

1. **Init**: 設定ファイル読み込み → Config生成
2. **Scan**: K8s API → Deployment/HPA情報取得
3. **Metrics**: Prometheus API → Metrics取得・統計計算
4. **Suggest**: Metrics → ResourceRecommendation/HPARecommendation生成
5. **Report**: Recommendation → Report生成（Markdown/JSON/Patches）
6. **Terminate**: 出力完了、プロセス終了

---

## Data Volume Assumptions

- **小規模環境**: 10 Deployments × 2 Containers/Deployment = 20 Containers
- **中規模環境**: 100 Deployments × 3 Containers/Deployment = 300 Containers
- **大規模環境**: 1000 Deployments × 3 Containers/Deployment = 3000 Containers

**メトリクスデータポイント**:
- 7日間、5分間隔 → 2,016 データポイント/Container
- 大規模環境: 3000 Containers × 2,016 = 6,048,000 データポイント
- メモリ使用量推定: 6M × 8 bytes (float64) ≈ 48 MB（十分に許容範囲）

---

## 10. SecurityLayer (セキュリティレイヤー)

**Purpose**: 入力検証とセキュリティ機能の提供

kostは `internal/security` パッケージで統一的なセキュリティレイヤーを実装します。このパッケージは全ての入力検証とセキュリティ関連の処理を集約し、脆弱性の混入を防ぎます。

### Security Functions

#### ValidateNamespace(namespace string) error
- **目的**: Kubernetes Namespace名の検証
- **検証内容**:
  - RFC 1123準拠（小文字英数字、ハイフン）
  - 最大63文字
  - 正規表現: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`

#### ValidateLabelSelector(selector string) error
- **目的**: Kubernetes Label Selectorの検証
- **検証内容**:
  - インジェクション攻撃を防ぐ文字列検証
  - 禁止文字: `'`, `;`, `"`, `\`

#### ValidateOutputPath(outputPath string) error
- **目的**: 出力ディレクトリパスの検証
- **検証内容**:
  - パストラバーサル攻撃の防止（`..` を含むパスを拒否）
  - センシティブディレクトリへのアクセス防止（`/etc`, `/root`, `/sys`, `/proc`, `/dev`）

#### ValidatePrometheusURL(url string) error
- **目的**: Prometheus URLの検証
- **検証内容**:
  - HTTP/HTTPSスキームの検証
  - 禁止文字: スペース、タブ、改行、`"`, `'`, `<`, `>`

#### SanitizePromQLQuery(query string) (string, error)
- **目的**: PromQLクエリの検証
- **検証内容**:
  - インジェクション攻撃の防止
  - 禁止文字: `;`, `--`, `/*`, `*/`, `\`, `\n`, `\r`

#### MaskSensitiveValue(value string) string
- **目的**: 機密情報のマスキング
- **処理内容**:
  - 8文字以下: `***`
  - 9文字以上: 先頭4文字と末尾4文字を表示、中間を `...` でマスク
  - 例: `"sk-1234567890abcdef"` → `"sk-1...cdef"`

### Security Design Principles

1. **最小権限の原則**: Kubernetes RBAC権限は読み取り専用（`get`, `list`）のみ
2. **入力検証の一元化**: 全ての外部入力は `internal/security` パッケージで検証
3. **API認証情報の分離**: 環境変数のみで管理、設定ファイルには含めない
4. **ログのサニタイズ**: 全てのログ出力で機密情報をマスキング
5. **脆弱性スキャン**: CI/CDでgosec、govulncheck、Trivyを実行

### Integration Points

- `internal/config/config.go`: 設定値の検証（Prometheus URL、出力パス、Label selector）
- `cmd/kost/scan.go`: Namespace名の検証
- `internal/metrics/query.go`: PromQLクエリのサニタイズ
- `internal/llm/`: APIキーのマスキング

---

## まとめ

以上のデータモデルに基づき、internal/ パッケージ配下に以下のGoパッケージを実装します：

- `internal/config`: Config構造体と設定管理
- `internal/k8s`: Deployment, Container, HPA構造体とK8s APIクライアント
- `internal/metrics`: Metrics構造体とPrometheus APIクライアント
- `internal/engine`: ResourceRecommendation, HPARecommendation構造体と推奨値計算ロジック
- `internal/llm`: LLMProvider interface と各プロバイダ実装
- `internal/report`: Report構造体とMarkdown/JSON/Patch生成ロジック
- `internal/security`: セキュリティレイヤー（入力検証、サニタイズ、マスキング）
