<!--
Sync Impact Report:
Version: 0.0.0 → 1.0.0
Rationale: Initial constitution establishment for kost (Kubernetes Optimization & Sizing Tool) project
Modified Principles: N/A (new constitution)
Added Sections: All sections newly created
Removed Sections: None
Templates requiring updates:
  ✅ plan-template.md - Compatible with current principles
  ✅ spec-template.md - Compatible with current principles
  ✅ tasks-template.md - Compatible with current principles
Follow-up TODOs: None
-->

# kost (Kubernetes Optimization & Sizing Tool) プロジェクト憲章

## コア原則

### I. 改善提案への特化（非交渉原則）

本プロジェクトは「可視化」ではなく「改善提案」に特化する。

- **MUST**: 推奨値（定量）、根拠（統計）、パッチ（適用可能）、説明（自然言語）の全てを提供すること
- **MUST**: 具体的な改善提案を「なぜ無駄か」→「どう直すか（具体パッチ）」→「効果見込み」の一気通貫で出力すること
- **MUST NOT**: 単なるコスト可視化やレポーティングを主目的としないこと

**根拠**: OpenCost/Kubecost等の既存ツールとの差別化であり、本プロジェクトの核となる価値提供である。「具体的に何をどう変えるか」の提案こそが、現場のSREが求める実用的な価値である。

### II. 安全性優先の設計（非交渉原則）

リソース最適化は本番環境の安定性に直結するため、安全性を最優先する。

- **MUST**: デフォルトを `--dry-run` とし、自動適用を行わないこと
- **MUST**: 推奨値の根拠（統計値、係数、丸め）を必ず出力すること
- **MUST**: 推奨値算出は「統計 + ルール」で決定し、AIには説明・優先度付けのみを委ねること
- **MUST**: セーフティファクター（例：1.2倍）を適用し、過小設定によるリスクを軽減すること
- **MUST**: 最小値（minCpuMilli, minMemMi）を設定し、極端な推奨値を防ぐこと

**根拠**: 本番環境でのリソース不足はレイテンシ悪化・OOM・再起動など直接的な障害を招く。AIの判断に全てを委ねず、再現性と検証可能性を担保することで、SREが信頼して運用できるツールとなる。

### III. MVP範囲の厳守（非交渉原則）

MVP（Minimum Viable Product）では実装範囲を明確に限定し、段階的な価値提供を行う。

**MVPでやること**:
- Kubernetes Deployment の requests/limits 最適化
- Namespace 単位での分析
- 推奨値・パッチ・レポートの生成（CLIベース）
- Prometheus メトリクスからのCPU/Memory utilization分析

**MVPでやらないこと（Non-goals）**:
- 完全なコスト計算・コストアトリビューション
- 自動適用（`kubectl apply` の自動実行）
- 全最適化領域の網羅（ノード最適化、スポット最適化、SLO最適化など）
- 複雑なワークロード種別の完全対応（StatefulSet、Job、HPA/VPA提案など）

**根拠**: 一点突破による早期の価値提供とフィードバック獲得を重視する。全機能を一度に実装すると開発期間が長期化し、実際のユーザーニーズとの乖離が発生するリスクが高まる。

### IV. 段階的な開発と完成（非交渉原則）

「小さい完成」を積み重ねることで、各段階で動作するプロダクトを維持する。

- **MUST**: マイルストーン単位で独立して動作可能な状態を作ること
  - Milestone 0: Skeleton（CLI雛形、ダミーデータでのレポート生成）
  - Milestone 1: Data Acquisition（K8s APIからのデータ取得）
  - Milestone 2: Suggestion Engine（統計・ルールベースの推奨値算出）
  - Milestone 3: Patch/Report出力（適用可能なパッチとレポート生成）
  - Milestone 4: LLM Narrative（説明文生成、フォールバック付き）
  - Milestone 5: GitHub Action（自動化）
- **MUST**: 各マイルストーンで CI（lint/test/build）がグリーンであること
- **MUST**: 各段階で実際に動作するデモを提供できること

**根拠**: 段階的な開発により、各フェーズでの検証とフィードバックが可能になる。特にOSS開発では外部貢献者が参加しやすい環境を維持するため、常に動作する状態を保つことが重要である。

### V. テストと品質保証

信頼性の高いOSSとして、適切なテスト戦略を実施する。

- **MUST**: 推奨値算出エンジン（engine）にはユニットテスト必須
- **MUST**: Prometheus/K8s APIレスポンスはfixtureで再現性を確保すること
- **MUST**: CI（GitHub Actions等）で lint + test + build を実行すること
- **MUST**: エラーメッセージには「必要なメトリクス」「確認手順」を明示すること
- **MUST**: 必要なRBAC権限（read-only）をドキュメントに記載すること

**根拠**: 運用ツールとして、動作の再現性と検証可能性が不可欠である。特にPrometheusやK8s APIとの連携部分は環境依存が大きいため、fixtureによる分離が重要である。

### VI. Go言語コーディング規約の遵守

Go言語でのベストプラクティスに厳格に従う。

- **MUST**: Effective Go、Code Review Comments、Golang Standardsに準拠すること
- **MUST**: gofmt、go vet、golint、golangci-lintの規則に従うこと
- **MUST**: パッケージ構成、命名規則、エラーハンドリングはGoの慣習に従うこと
- **MUST**: interface設計、並行処理、メモリ管理はGoのベストプラクティスを適用すること

**根拠**: Kubernetes エコシステムでは Go が標準言語であり、Goの慣習に従うことでコミュニティからの貢献や統合が容易になる。また、保守性と可読性の向上にも繋がる。

### VII. ドキュメントの日本語化

プロジェクト内の全てのドキュメントは日本語で記述する。

- **MUST**: 仕様書、設計書、タスクリスト、レポート等の全ドキュメントを日本語で作成すること
- **MUST**: コードコメント、エラーメッセージは英語とすること（国際化を考慮）
- **MUST**: README.md等の公開ドキュメントは日本語版を優先し、必要に応じて英語版を追加すること

**根拠**: プロジェクト初期段階では日本語でのコミュニケーションと仕様明確化を優先する。ただし、OSSとしての国際展開を見据え、コードとインターフェースは英語を維持する。

## プロジェクト構造の原則

### ディレクトリ構成（推奨）

```
.
├── cmd/                    # CLIエントリポイント
├── internal/
│   ├── k8s/                # K8s APIアクセス
│   ├── metrics/            # Prometheusクエリ
│   ├── engine/             # 推奨値算出
│   ├── patch/              # パッチ生成
│   ├── report/             # md/json出力
│   └── llm/                # LLM I/F（任意）
├── docs/                   # ドキュメント
├── examples/               # 利用例
└── out/                    # 出力（gitignore）
```

### CLIコマンド設計

MVP段階では以下の3コマンドを提供する：

1. `scan`: 対象収集（K8sからマニフェスト情報）
2. `suggest`: 推奨値生成（統計・ルール）
3. `report`: レポート生成（Markdown/JSON、任意でAI要約）

## 生成AIの使用方針

### AIの役割（限定的活用）

- **許可される用途**:
  - 提案の優先順位付け（影響大/リスク大から）
  - 人間向けの説明文生成（なぜ、どう直す、注意点）
  - リスクの言語化（ピーク時、季節性、バッチ混在など）

- **禁止される用途**:
  - 推奨値の計算（これは統計+ルールベースで実施）
  - 安全性判断の委譲
  - 根拠不明なブラックボックス推奨

### AIへの入力仕様

LLMには構造化されたデータのみを渡す：
- 現状 requests/limits
- P50/P95/P99 utilization（CPU/Memory）
- 推奨値（計算済み）
- 過剰/過小判定
- 過去イベント（OOMKilled、再起動回数）
- 期待削減率（CPU/Memory）

**根拠**: ログ全文等の非構造化データを渡すと結果の安定性と検証可能性が損なわれる。構造化データにより、AIの役割を「説明生成」に限定し、安全性を担保する。

## OSSとしての運用原則

### リポジトリの必須要素

- **MUST**: README.md（目的、非目的、Quickstart、デモ）
- **MUST**: LICENSE
- **MUST**: CONTRIBUTING.md
- **MUST**: .github/ISSUE_TEMPLATE/*
- **MUST**: .github/pull_request_template.md
- **MUST**: CI（lint / test / build）

### MVP完了定義（Definition of Done）

以下の全てが満たされた時、MVPは完了とする：

- 10分以内に導入できる（README Quickstartが通る）
- `scan → suggest → report` で `report.md` が生成される
- 少なくとも1 Deploymentで「根拠付き推奨」が出る
- `patches/` が生成され、適用可能な形式になっている
- CI が green（lint/test/build）
- `--dry-run` がデフォルトで安全

## ガバナンス

### 憲章の適用

- 本憲章は全ての開発プラクティスに優先する
- 全てのPR・レビューは本憲章への準拠を検証しなければならない
- 複雑性の導入には明確な正当化が必要である

### 憲章の改定

- 憲章の改定にはドキュメント化、承認、移行計画が必要
- バージョニング規則:
  - MAJOR: 後方互換性のない原則の削除または再定義
  - MINOR: 新原則・セクションの追加または実質的な拡張
  - PATCH: 明確化、文言修正、誤字修正、非意味的な改善

### コンプライアンスレビュー

- 各マイルストーン完了時に憲章準拠を確認すること
- 原則違反が発見された場合、即座に修正または正当化を行うこと

**バージョン**: 1.0.0 | **批准日**: 2026-01-14 | **最終改定日**: 2026-01-14
