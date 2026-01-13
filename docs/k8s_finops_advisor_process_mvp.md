# Kubernetes × FinOps × 生成AI OSS：開発プロセス & MVP設計（詳細版）

本ドキュメントは、**Kubernetes クラスタのリソース割当（requests/limits）最適化**を主題に、  
**可視化（OpenCost/Kubecost等）ではなく「改善提案（Actionable suggestions）」にフォーカス**した OSS を開発するための、実装に落とせるレベルの計画書です。

- 対象プロダクト（仮名）：`k8s-finops-advisor`
- 形式：CLI（将来 GitHub Action / GitOps PR 連携）
- コア価値：**「なぜ無駄か」→「どう直すか（具体パッチ）」→「効果見込み」**を一気通貫で出す

---

## 1. 目的・背景

### 1.1 目的（Goal）
Kubernetes の運用現場で頻発する課題の一つは、**リソース requests/limits の過剰設定**です。

- 過剰な requests は、スケジューリング効率の低下（ノードの無駄）につながる
- 過小な requests は、レイテンシ悪化・OOM・再起動などのリスクになる
- 「どこを」「どれだけ」直すべきかを判断するには、メトリクス + 文脈理解が必要

本 OSS は、メトリクス（例：Prometheus）とマニフェスト（K8s API）から情報を集め、  
**具体的な改善提案（推奨値、適用パッチ、説明）**を出力します。

### 1.2 非目的（Non-goals：MVPではやらない）
MVP のスコープを明確にするため、次は「やらない」と定義します。

- **完全なコスト計算 / コストアトリビューション**（OpenCost/Kubecost の領域）
- 自動適用（`kubectl apply` の自動実行）  
  - MVP では **提案・パッチ生成まで**。適用は人間の判断
- 全最適化領域の網羅（ノード最適化、スポット最適化、SLO最適化など）
- 複雑なワークロード種別の完全対応（まずは Deployment から）

---

## 2. 差別化（競合整理）

### 2.1 OpenCost / Kubecost
- 強い点：コスト可視化、割当て、レポーティング
- 弱い点：**「具体的に何をどう変えるか」**の提案生成は中心機能ではない

### 2.2 生成AIデバッグ系（例：K8sGPT 等）
- 強い点：トラブルシューティング支援、状態分析
- 弱い点：FinOps/リソース最適化の**定量推奨**と**適用パッチ**までの一体化は薄い

### 2.3 本 OSS の立ち位置
> **可視化**ではなく **改善提案**に特化し、  
> *推奨値（定量） + 根拠（統計） + パッチ（適用可能） + 説明（自然言語）* を提供する。

---

## 3. ターゲットユーザーとユースケース

### 3.1 想定ユーザー
- Kubernetes を運用する SRE / Platform Engineer
- 開発チームでクラスタコストを管理する担当者（FinOps）
- GitOps 運用を行うチーム（Argo CD / Flux 等）

### 3.2 主要ユースケース（MVP：一点突破）
**Namespace 単位で、Deployment の過剰/過小な requests を検出し、推奨値 + パッチ + レポートを出す。**

---

## 4. プロダクト概要（MVP）

### 4.1 MVP の一言定義
**「過剰/過小な requests/limits を検出し、推奨値・理由・YAMLパッチを出すCLI」**

### 4.2 入力（Inputs）
- kubeconfig / in-cluster config（K8s API）
- Prometheus endpoint（または in-cluster）
- 対象指定（namespace / label selector / deployment名）
- 集計期間（例：7d）と統計（P50/P95/P99）
- 任意：除外ルール（システム系 namespace、ジョブ系）

### 4.3 出力（Outputs）
- `report.md`：人間向けのサマリ（上位改善案、理由、注意点）
- `patches/`：適用候補パッチ（Deployment 単位）
  - `patches/<namespace>/<deployment>.yaml`
- `summary.json`：機械可読（将来 GitHub Action / PR 作成用）

---

## 5. 具体的なMVP仕様（コマンド、設定、アルゴリズム）

### 5.1 CLI コマンド設計（最小3コマンド）
MVP では 3コマンドを提供し、内部的には同じ実行パイプラインを分割します。

1) `scan`：対象収集（K8sからマニフェスト情報）
2) `suggest`：推奨値生成（統計・ルール）
3) `report`：レポート生成（Markdown / JSON、任意でAI要約）

例：
```bash
k8s-finops scan --namespace prod
k8s-finops suggest --namespace prod --window 7d --cpu p95 --mem p95
k8s-finops report --namespace prod --llm on
```

将来的には `run`（scan→suggest→report）を追加してもよいですが、MVP段階ではデバッグしやすい分割が推奨です。

### 5.2 設定ファイル（例：`config.yaml`）
```yaml
kube:
  context: ""
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
  namespacesExclude: ["kube-system", "monitoring"]
  labelSelector: ""
output:
  dir: "./out"
  format: ["md", "json", "patch"]
llm:
  enabled: true
  provider: "openai"
  model: "gpt-4.1-mini"
  maxTokens: 1200
  temperature: 0.2
```

> **注**：モデル名は例です。実装では環境変数で上書き可能にすると運用しやすいです。

### 5.3 取得するメトリクス（Prometheus）
MVPでは CPU/Memory utilization を中心にします。

- CPU 使用量（例）：`rate(container_cpu_usage_seconds_total[5m])`
- Memory 使用量（例）：`container_memory_working_set_bytes`

**重要**：node-exporter / cAdvisor / kube-state-metrics の導入状況でクエリが変わります。  
MVPでは「Prometheus から値を取れること」が前提になるため、docs に前提条件を明記します。

### 5.4 推奨値算出（AIに任せない：再現性と安全性）
推奨値は **統計 + ルール** で決定し、AIは説明・優先度付けに限定します。

#### 5.4.1 推奨 requests の基本式
- `recommended_cpu_request_m = max(Pxx_cpu_m * safetyFactor, minCpuMilli)`
- `recommended_mem_request_mi = max(Pxx_mem_mi * safetyFactor, minMemMi)`

例（CPU）：
- P95 = 120m、safetyFactor=1.2 → 推奨 = 144m（丸めて 150m）

#### 5.4.2 判定ルール（例）
- **Overprovisioned**：`current_request > Pxx * 2.0`
- **Underprovisioned**：`current_request < Pxx * 1.1`
- **Risky limit**（任意）：`limit >> request`（例：10倍超）

#### 5.4.3 期待効果（推定）
MVPでは「絶対のコスト」ではなく、**削減率**（request削減の割合）を提示します。

- `saving_cpu_request_ratio = (current - recommended) / current`（推奨が小さい場合）
- Memory も同様

> OpenCost と統合すれば金額換算も可能ですが、MVPではまず削減率と根拠を重視します。

### 5.5 パッチ生成（GitOps向け）
出力するパッチは最低限、Deployment の `resources.requests/limits` を更新できる形式にします。

- 推奨：**Strategic Merge Patch**（人が読みやすい）
- 代替：JSONPatch（機械的で安全だが読みづらい）

例（Strategic Merge Patch）：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
  namespace: prod
spec:
  template:
    spec:
      containers:
        - name: app
          resources:
            requests:
              cpu: "150m"
              memory: "256Mi"
```

---

## 6. 生成AIの使い方（MVPで価値が出る限定活用）

### 6.1 AIの役割（MVP）
- 提案の**優先順位付け**（影響大/リスク大から）
- 人間向けの**説明**（なぜ、どう直す、注意点）
- リスクの言語化（ピーク時、季節性、バッチ混在など）

### 6.2 AIに渡す入力（構造化が重要）
LLMには「ログ全文」ではなく、**要点を構造化して渡す**のが安定します。

- 現状 requests/limits
- P50/P95/P99 utilization（CPU/Memory）
- 推奨値（計算済み）
- 過剰/過小判定
- 過去イベント（OOMKilled、再起動回数）※取れれば
- 期待削減率（CPU/Memory）

### 6.3 例：LLMへのプロンプト方針（概要）
- System：FinOps + K8s SRE の観点で、事実に基づいて提案
- User：上記構造化データ（JSON）
- Output：Markdown のレポートセクション（短く、具体的に）

> 実装では、LLMレスポンスをそのまま貼るのではなく、テンプレートに差し込む形が安全です。

---

## 7. 開発プロセス（OSSとして回る手順）

### 7.1 フェーズ0：設計を固定（ブレ防止）
- README 冒頭に「目的」「非目的」「MVP」を明記
- 主要な I/O（入力→処理→出力）を図にする（ASCIIでもOK）
- 競合との差別化を 3行で書ける状態にする

成果物：
- `README.md` 初版
- `docs/design.md`（本ドキュメント）
- `LICENSE`

### 7.2 フェーズ1：リポジトリの骨格（Contributionしやすさ）
最低限、次を揃えて外部参加しやすくします。

- `README.md`
- `LICENSE`
- `CONTRIBUTING.md`
- `.github/ISSUE_TEMPLATE/*`
- `.github/pull_request_template.md`
- CI（lint / test / build）

推奨ディレクトリ例（言語により調整）：
```
.
├── cmd/                    # CLIエントリ（Goの場合）
├── internal/
│   ├── k8s/                # K8s APIアクセス
│   ├── metrics/            # Prometheusクエリ
│   ├── engine/             # 推奨値算出
│   ├── patch/              # パッチ生成
│   ├── report/             # md/json出力
│   └── llm/                # LLM I/F（任意）
├── docs/
├── examples/
└── out/                    # 出力（gitignore）
```

### 7.3 フェーズ2：MVPを小さく完成させる（マイルストーン）
**「小さい完成」**を積み重ねる設計です。

#### Milestone 0：Skeleton
- CLI雛形（サブコマンド、設定読み込み）
- ダミーデータで `report.md` を生成

#### Milestone 1：Data Acquisition
- K8s API から Deployment/Container の requests/limits 取得
- 対象は Deployment のみ（まずはこれだけ）

#### Milestone 2：Suggestion Engine（統計・ルール）
- Prometheus から CPU/Memory utilization を取得
- P95 等の統計に基づき推奨値算出
- 判定（over/under）と削減率を計算

#### Milestone 3：Patch / Report 出力
- `patches/` を生成
- `report.md` をテンプレートで生成
- `summary.json` を出力

#### Milestone 4：LLM Narrative（任意でMVPに含める）
- LLMは説明文生成・注意点に限定
- 失敗時はフォールバック（ルールベース文）に切替

#### Milestone 5：GitHub Action（MVP+）
- cronでレポート生成して artifact 出力
- 将来：PR作成（次フェーズ）

---

## 8. 品質・安全性（信頼されるOSSの条件）

### 8.1 安全設計
- デフォルトは `--dry-run`
- **自動適用しない**（MVP）
- 推奨値の根拠（P95、係数、丸め）を必ず出す

### 8.2 テスト方針
- engine（推奨値計算）はユニットテスト必須
- Prometheus/K8sレスポンスは fixture で再現性を確保
- 最低限：GitHub Actions で `lint + test + build`

### 8.3 例外・エラーハンドリング
- Prometheus から取れない場合：
  - エラーメッセージに「必要なメトリクス」「確認手順」を提示
- K8s権限不足：
  - 必要 RBAC（read-only）を docs に記載

---

## 9. まず作るIssue一覧（初期バックログ例）

### 9.1 Repo/CI
- [ ] 初版 README（目的、非目的、Quickstart、デモ）
- [ ] ライセンス追加
- [ ] CI（lint/test/build）
- [ ] Issue/PR テンプレ

### 9.2 機能
- [ ] `scan`：Deployment/Container の resources 取得
- [ ] `metrics`：Prometheus クエリ実装（CPU/Memory）
- [ ] `suggest`：推奨値算出（P95 + safety）
- [ ] `patch`：Strategic Merge Patch 生成
- [ ] `report`：Markdownテンプレート実装
- [ ] `summary.json` 生成（将来拡張用）
- [ ] LLM narrative（オプション、フォールバック付き）

### 9.3 Docs
- [ ] 前提条件（Prometheus/kube-state-metrics）
- [ ] RBAC例（read-only）
- [ ] FAQ（よくある失敗：メトリクス取れない 等）

---

## 10. 将来ロードマップ（MVP後）

### 10.1 Phase 2（MVPの次）
- OpenCost 連携で金額換算（節約見積り）
- GitHub Action（週次レポート）
- GitOps PR 作成（提案パッチを自動PR化）

### 10.2 Phase 3（高度化）
- HPA/VPA の提案
- ワークロード種別拡張（StatefulSet/Job）
- 変更の段階適用（safe rollout plan）
- SLO/エラーレートと連携（コストだけでなく信頼性も加味）

---

## 付録A：レポート（report.md）テンプレ案

```md
# k8s-finops-advisor report

対象: namespace=prod / window=7d / percentile=P95 / safety=1.2

## Top findings
1. Deployment `api` container `app`
   - 現状 request(cpu)=500m / P95=120m → 推奨=150m（-70%）
   - 理由: 平常時の使用率が低く、P95でも十分余裕あり
   - 注意: 月次バッチ等のピークがある場合は段階的に適用

## Suggested patches
- patches/prod/api.yaml

## Method
- CPU: P95(rate(container_cpu_usage_seconds_total[5m])) × safety
- Memory: P95(container_memory_working_set_bytes) × safety
```

---

## 付録B：MVPの「完了条件（Definition of Done）」

- 10分以内に導入できる（README Quickstart が通る）
- `scan → suggest → report` で `report.md` が生成される
- 少なくとも 1 Deployment で「根拠付き推奨」が出る
- `patches/` が生成され、適用可能な形式になっている
- CI が green（lint/test/build）
- `--dry-run` がデフォルトで安全

---

以上。
