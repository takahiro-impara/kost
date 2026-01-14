# Implementation Plan: kost (Kubernetes Optimization & Sizing Tool) MVPコア機能

**Branch**: `001-mvp-core` | **Date**: 2026-01-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-mvp-core/spec.md`

**Note**: このドキュメントは `/speckit.plan` コマンドによって生成されました。

## Summary

kost (Kubernetes Optimization & Sizing Tool) は、Kubernetes Deploymentのリソース設定（requests/limits）およびHPA設定（minReplicas/maxReplicas）の最適化推奨を提供するCLIツールです。Prometheusメトリクスから過去の実績データを取得し、統計分析（P95/P99等）とルールベースの判定により、具体的な推奨値、適用可能なYAMLパッチ、削減率を記載したレポートを生成します。

技術的アプローチ:
- Go言語で実装されるCLIツール（cobra、viper等の標準ライブラリを活用）
- Kubernetes client-goでDeployment/HPA情報を取得
- Prometheus APIクライアントでメトリクスを取得し、統計処理を実施
- LLMプロバイダ（OpenAI、Claude）との連携はオプション機能として実装
- 段階的開発（Milestone 0〜5）により、各フェーズで動作可能な状態を維持

## Technical Context

**Language/Version**: Go 1.21以上
**Primary Dependencies**:
- kubernetes (client-go v0.29+): K8s APIアクセス
- prometheus (client_golang v1.18+): Prometheusメトリクス取得
- cobra (v1.8+): CLIフレームワーク
- viper (v1.18+): 設定ファイル管理
- openai-go / anthropic-sdk-go: LLMプロバイダクライアント（オプション）

**Storage**: N/A（ステートレス、出力ファイルのみ）
**Testing**: go test（testing標準パッケージ）、testify（アサーション）、fixture（Prometheus/K8sレスポンス）
**Target Platform**: Linux/macOS/Windows（クロスコンパイル対応）、コンテナ実行も可能
**Project Type**: Single CLI project
**Performance Goals**:
- 100 Deployments のスキャン・分析を60秒以内に完了
- メモリ使用量は500MB以下に抑制
- 大規模環境（1000 Deployments）でも5分以内に完了

**Constraints**:
- Prometheusメトリクスが7日間以上保持されていること
- K8s RBAC read権限（Deployment/Pod/HPA）が必要
- LLM APIキーは環境変数で提供（設定ファイルには含まない）

**Security Requirements**:
- 全てのAPI認証情報は環境変数で管理（設定ファイル、ログ、レポートに含めない）
- 入力バリデーション（namespace、label selector、出力パス）の実装
- PromQL injection、パストラバーサル攻撃の防止
- TLS証明書検証（設定で無効化可能）
- 最小権限の原則（RBAC read-only）
- CI/CDでのセキュリティスキャン（gosec、govulncheck、Trivy）

**Scale/Scope**:
- 対象：中規模〜大規模Kubernetesクラスタ（10〜1000 Deployments）
- 初期リリース：MVP範囲（Deploymentのみ、HPA対応）
- 将来拡張：StatefulSet、DaemonSet、VPA提案等

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Principle I: 改善提案への特化

- [x] 推奨値（定量）、根拠（統計）、パッチ（適用可能）、説明（自然言語）を全て提供
- [x] 「なぜ無駄か」→「どう直すか」→「効果見込み」の一気通貫で出力
- [x] 単なるコスト可視化ではなく、具体的な改善提案を主目的とする

**Status**: 合格 - User Story 1〜5が全て改善提案機能であり、レポート・パッチ・説明を生成

### ✅ Principle II: 安全性優先の設計

- [x] デフォルトを `--dry-run` とし、自動適用を行わない（FR-025）
- [x] 推奨値の根拠（統計値、係数、丸め）を出力（FR-008）
- [x] 推奨値算出は統計+ルールで決定、AIは説明のみ（FR-005, FR-013）
- [x] セーフティファクター適用（Assumptionsに明記）
- [x] 最小値設定（Assumptionsに明記）

**Status**: 合格 - 全ての安全性要件が仕様に反映されている

### ✅ Principle III: MVP範囲の厳守

**MVPでやること**:
- [x] Kubernetes Deployment の requests/limits 最適化（User Story 1）
- [x] HPA設定の最適化（User Story 3）※憲章のNon-goalsに記載されているが、ユーザー要望により追加
- [x] Namespace 単位での分析（FR-001）
- [x] 推奨値・パッチ・レポートの生成（FR-008, FR-009）
- [x] Prometheus メトリクスからのCPU/Memory/Pod数分析（FR-003, FR-015）

**MVPでやらないこと**:
- [x] 完全なコスト計算（削減率のみ提示）
- [x] 自動適用（パッチ生成まで）
- [x] ノード最適化、スポット最適化
- [x] StatefulSet、DaemonSet、Job（Deploymentのみ）

**Status**: ⚠️ 部分的合格 - HPA最適化が追加されているが、これはユーザー要望による追加機能であり、MVP範囲の拡張として正当化される。憲章のNon-goalsには「HPA/VPA提案など」とあるが、今回はHPA最適化のみに限定し、VPAは対象外。

**Justification**: HPA最適化はオートスケーリング環境での重要な最適化ポイントであり、User Story 3（P3）として優先度を下げて実装することで、MVP範囲の過度な拡大を防ぐ。

### ✅ Principle IV: 段階的な開発と完成

- [x] Milestone 0: Skeleton（CLI雛形、ダミーデータでのレポート生成）
- [x] Milestone 1: Data Acquisition（K8s APIからのデータ取得）
- [x] Milestone 2: Suggestion Engine（統計・ルールベースの推奨値算出）
- [x] Milestone 3: Patch/Report出力（適用可能なパッチとレポート生成）
- [x] Milestone 4: LLM Narrative（説明文生成、フォールバック付き）
- [x] Milestone 5: GitHub Action（自動化）※MVP後

**Status**: 合格 - 各マイルストーンでCI（lint/test/build）を実行し、動作可能な状態を維持

### ✅ Principle V: テストと品質保証

- [x] 推奨値算出エンジンにユニットテスト必須（internal/engine/）
- [x] Prometheus/K8s APIレスポンスはfixtureで再現性確保（tests/fixtures/）
- [x] CI（GitHub Actions）で lint + test + build 実行
- [x] エラーメッセージに必要なメトリクス・確認手順を明示
- [x] 必要なRBAC権限をドキュメントに記載

**Status**: 合格 - テスト戦略がPhase 1で定義され、実装タスクに含まれる

### ✅ Principle VI: Go言語コーディング規約の遵守

- [x] Effective Go、Code Review Comments、Golang Standardsに準拠
- [x] gofmt、go vet、golint、golangci-lintの規則に従う
- [x] パッケージ構成、命名規則、エラーハンドリングはGoの慣習に従う
- [x] interface設計、並行処理、メモリ管理はGoのベストプラクティスを適用

**Status**: 合格 - Goプロジェクトとして標準的な構成を採用

### ✅ Principle VII: ドキュメントの日本語化

- [x] 仕様書、設計書、タスクリストを日本語で作成
- [x] コードコメント、エラーメッセージは英語
- [x] README.md等の公開ドキュメントは日本語版を優先

**Status**: 合格 - 全てのドキュメントが日本語で記述されている

## Project Structure

### Documentation (this feature)

```text
specs/001-mvp-core/
├── spec.md              # 機能仕様書
├── plan.md              # このファイル（実装計画）
├── research.md          # Phase 0: 技術調査結果
├── data-model.md        # Phase 1: データモデル定義
├── quickstart.md        # Phase 1: クイックスタートガイド
├── contracts/           # Phase 1: API契約（該当なし、CLIツールのため）
├── checklists/          # 品質チェックリスト
│   └── requirements.md
└── tasks.md             # Phase 2: 実装タスク（/speckit.tasks で生成）
```

### Source Code (repository root)

```text
.
├── cmd/
│   └── kost/                  # CLIエントリポイント
│       └── main.go
├── internal/
│   ├── k8s/                   # K8s APIアクセス
│   │   ├── client.go
│   │   ├── deployment.go
│   │   └── hpa.go
│   ├── metrics/               # Prometheusクエリ
│   │   ├── client.go
│   │   ├── query.go
│   │   └── stats.go
│   ├── engine/                # 推奨値算出
│   │   ├── resource.go
│   │   ├── hpa.go
│   │   └── calculator.go
│   ├── patch/                 # パッチ生成
│   │   ├── resource.go
│   │   └── hpa.go
│   ├── report/                # md/json出力
│   │   ├── markdown.go
│   │   └── json.go
│   ├── llm/                   # LLM I/F（オプション）
│   │   ├── provider.go
│   │   ├── openai.go
│   │   ├── claude.go
│   │   └── fallback.go
│   ├── security/              # セキュリティ（入力検証、バリデーション）
│   │   └── validation.go
│   └── config/                # 設定管理
│       └── config.go
├── pkg/                       # 公開パッケージ（将来の拡張用）
├── tests/
│   ├── fixtures/              # テストデータ
│   │   ├── k8s/
│   │   └── prometheus/
│   ├── integration/           # 統合テスト
│   └── unit/                  # ユニットテスト（internal/と並列）
├── docs/                      # ドキュメント
│   ├── k8s_finops_advisor_process_mvp.md
│   └── architecture.md
├── examples/                  # 利用例
│   ├── config.yaml
│   └── rbac.yaml
├── out/                       # 出力（gitignore）
├── .github/
│   └── workflows/
│       └── ci.yml
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── LICENSE
```

**Structure Decision**: Single CLI project構成を採用。Goの標準的な`cmd/`と`internal/`パッケージ構成に従い、CLIツールとして実装する。`pkg/`は将来の公開ライブラリ化に備えて用意するが、MVP段階では使用しない。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| HPA最適化の追加（憲章のNon-goalsに記載） | ユーザー要望により、オートスケーリング環境での重要な最適化ポイントとして追加 | HPA最適化なしでは、オートスケーリング環境でのコスト最適化が不完全になり、ユーザー価値が低下する。ただし、User Story 3（P3）として優先度を下げることで、MVP範囲の過度な拡大を防ぐ |
