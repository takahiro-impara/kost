# 技術調査: kost (Kubernetes Optimization & Sizing Tool) MVPコア機能

**Date**: 2026-01-14
**Feature**: kost (Kubernetes Optimization & Sizing Tool) MVPコア機能
**Phase**: Phase 0 - Research & Technology Selection

## 概要

本ドキュメントでは、kost (Kubernetes Optimization & Sizing Tool) の実装に必要な技術選定の調査結果をまとめます。各技術について、選定理由、代替案、およびトレードオフを記載します。

## 1. プログラミング言語

### Decision: Go 1.21以上

### Rationale:
- **Kubernetes エコシステムとの親和性**: client-goをはじめ、K8s関連ツールの大半がGoで実装されており、ネイティブな統合が可能
- **クロスコンパイルの容易さ**: 単一バイナリで Linux/macOS/Windows に配布可能、コンテナイメージも軽量
- **並行処理**: goroutineによる効率的な並行処理で、複数Deploymentの並列スキャンが容易
- **パフォーマンス**: 静的型付け・コンパイル言語として、大規模環境（1000+ Deployments）でも高速動作
- **コミュニティとライブラリ**: Kubernetes、Prometheus、LLM SDKの全てに成熟したGoライブラリが存在

### Alternatives Considered:
- **Python**: ライブラリは豊富だが、配布が複雑（依存関係管理）、パフォーマンスが劣る
- **Rust**: パフォーマンスは優れるが、開発速度が遅く、Kubernetesライブラリが未成熟
- **TypeScript/Node.js**: Prometheusクライアントは存在するが、CLIツールとしての配布が複雑

## 2. CLIフレームワーク

### Decision: Cobra (spf13/cobra v1.8+)

### Rationale:
- **業界標準**: kubectl、helm等のKubernetesツールで採用されており、ユーザーにとって馴染みやすいCLI体験を提供
- **サブコマンド対応**: scan/suggest/report の3コマンド構成を自然に実装可能
- **フラグ管理**: Persistent Flags、Local Flags、必須フラグのバリデーション等を標準サポート
- **ドキュメント自動生成**: man pages、Markdown、YAML形式のドキュメント生成機能
- **Viperとの統合**: 設定ファイル（config.yaml）との連携が容易

### Alternatives Considered:
- **urfave/cli**: シンプルだが、サブコマンドの階層化やフラグ管理が弱い
- **標準flagパッケージ**: 低レベルすぎて、複雑なCLI構造には不向き

## 3. 設定管理

### Decision: Viper (spf13/viper v1.18+)

### Rationale:
- **多様な設定ソース**: YAML、環境変数、CLIフラグの優先順位付き統合
- **Cobraとの統合**: Persistent Flagsと自動バインディング
- **環境変数の自動マッピング**: `FINOPS_PROMETHEUS_URL` → `prometheus.url` の自動変換
- **ホットリロード**: 設定ファイルの変更検知（将来の拡張に備える）
- **型安全**: `GetString()`, `GetInt()`, `GetDuration()` 等の型付きGetter

### Alternatives Considered:
- **標準encoding/json, encoding/yaml**: 低レベルすぎて、環境変数統合やフラグバインディングが手作業
- **koanf**: 柔軟性は高いが、Cobraとの統合がViperより複雑

## 4. Kubernetes クライアント

### Decision: client-go (k8s.io/client-go v0.29+)

### Rationale:
- **公式クライアント**: Kubernetesプロジェクト公式のGoクライアントライブラリ
- **完全なAPI対応**: Deployment、HPA、Pod等の全リソースタイプをサポート
- **型安全**: コード生成による型安全なAPIアクセス
- **認証統合**: kubeconfig、in-cluster config、トークン認証等を自動処理
- **バージョン互換性**: 複数のK8sバージョンに対応（v1.27〜v1.29想定）

### Alternatives Considered:
- **REST API直接呼び出し**: 低レベルすぎて、認証・型安全性・バージョン互換性の実装コストが高い
- **kubectl execの呼び出し**: 依存関係が増え、エラーハンドリングが複雑

## 5. Prometheusクライアント

### Decision: prometheus/client_golang v1.18+

### Rationale:
- **公式クライアント**: Prometheusプロジェクト公式のGoクライアントライブラリ
- **PromQL実行**: 完全なPromQLクエリ実行機能（Range Queries、Instant Queries）
- **型安全**: メトリクス結果の型安全なパース
- **HTTP クライアント統合**: タイムアウト、リトライ、TLS設定等を柔軟に制御
- **パフォーマンス**: 大量のメトリクスを効率的に処理

### Alternatives Considered:
- **HTTP API直接呼び出し**: JSON パースやエラーハンドリングが手作業で複雑
- **サードパーティライブラリ**: 公式以外は保守性に懸念

## 6. 統計処理

### Decision: montanaflynn/stats（パーセンタイル計算）

### Rationale:
- **パーセンタイル計算**: P50/P95/P99の計算を標準サポート
- **シンプルなAPI**: `stats.Percentile(data, 95.0)` のように直感的
- **軽量**: 依存関係が少なく、バイナリサイズへの影響が小さい
- **十分な精度**: MVP段階では厳密な統計処理は不要、十分な精度を提供

### Alternatives Considered:
- **gonum/stat**: より高度な統計機能を提供するが、MVPには過剰
- **自前実装**: ソート+インデックス計算で可能だが、テスト済みライブラリを使う方が安全

## 7. YAMLパッチ生成

### Decision: sigs.k8s.io/yaml（Strategic Merge Patch）

### Rationale:
- **Kubernetes標準**: kubectl applyと同じStrategic Merge Patchをサポート
- **人間可読**: 生成されるYAMLが読みやすく、GitOpsに適している
- **型安全**: Goの構造体から直接YAML生成、エラーリスクが低い
- **client-goとの統合**: K8sリソースの型定義を再利用可能

### Alternatives Considered:
- **JSON Patch**: 機械的で正確だが、人間が読みにくく、GitOps運用に不向き
- **標準encoding/yaml**: 低レベルすぎて、Strategic Mergeのセマンティクスを手実装する必要がある

## 8. Markdownレポート生成

### Decision: text/template（標準ライブラリ）

### Rationale:
- **標準ライブラリ**: 外部依存なしで、シンプルなテンプレートエンジン
- **十分な機能**: 変数展開、ループ、条件分岐等、レポート生成に必要な機能を全て提供
- **型安全**: テンプレートにGoの構造体を渡してレンダリング
- **パフォーマンス**: コンパイル済みテンプレートによる高速レンダリング

### Alternatives Considered:
- **html/template**: HTML用のエスケープ処理が入るため、Markdownには不向き
- **サードパーティテンプレートエンジン**: 標準ライブラリで十分な機能があり、依存を増やす必要がない

## 9. LLMプロバイダクライアント

### Decision:
- **OpenAI**: sashabaranov/go-openai（非公式だが広く使用されている）
- **Claude**: anthropics/anthropic-sdk-go（公式SDK、2024年リリース）

### Rationale:
- **OpenAI**: 成熟した非公式SDKで、Chat Completions APIを完全サポート、型安全、エラーハンドリングが充実
- **Claude**: Anthropic公式SDKで、Messages APIをネイティブサポート、将来の安定性が高い
- **抽象化レイヤー**: internal/llm/provider.go で統一インターフェースを定義し、プロバイダ切り替えを容易に
- **フォールバック**: LLM接続失敗時はルールベース説明を生成する仕組みで、可用性を担保

### Alternatives Considered:
- **REST API直接呼び出し**: HTTP リクエスト構築、レスポンスパース、エラーハンドリングが煩雑
- **単一プロバイダのみ**: ユーザー選択肢を狭め、ベンダーロックインのリスク

## 10. テストフレームワーク

### Decision:
- **標準testing**: Go標準のtestingパッケージ
- **testify/assert**: アサーションを読みやすく記述
- **testify/mock**: インターフェースモックの自動生成

### Rationale:
- **標準testing**: `go test` で即座に実行可能、CIとの統合が容易
- **testify**: 業界標準のアサーションライブラリで、テストコードの可読性向上
- **Fixture**: tests/fixtures/ に K8s/Prometheus のレスポンスJSONを配置し、再現性を確保
- **Table-Driven Tests**: Goの慣習に従い、複数のテストケースを効率的に記述

### Alternatives Considered:
- **ginkgo/gomega**: BDD形式だが、MVPには過剰、学習コストも高い
- **標準assertionのみ**: if文でのアサーションは冗長で読みにくい

## 11. CI/CD

### Decision: GitHub Actions

### Rationale:
- **GitHub統合**: リポジトリと同じプラットフォームで管理、PR連携が容易
- **無料枠**: 公開リポジトリは無料、プライベートでも月2,000分まで無料
- **豊富なアクション**: setup-go、golangci-lint、codecov等の公式アクションが充実
- **マトリックスビルド**: 複数のGoバージョン、複数のOSで並列テスト可能
- **シークレット管理**: LLM APIキー等の機密情報を安全に管理

### Alternatives Considered:
- **GitLab CI**: GitHub以外のプラットフォームが必要
- **CircleCI**: 設定が複雑、GitHub Actionsで十分

## 12. リント・フォーマット

### Decision:
- **gofmt**: 標準フォーマッタ
- **golangci-lint**: 統合リントツール（golint、go vet、staticcheck等を統合）

### Rationale:
- **gofmt**: Goコミュニティの標準、一貫したコードスタイル
- **golangci-lint**: 複数のリンターを一括実行、設定ファイル（.golangci.yml）で制御
- **CI統合**: GitHub Actionsで自動実行、PRでのコードレビューを効率化
- **高速**: キャッシュ機能により、大規模プロジェクトでも高速

### Alternatives Considered:
- **個別リンター**: golint、go vet等を個別実行するのは非効率
- **カスタムリンター**: MVP段階では標準ツールで十分

## 13. 依存関係管理

### Decision: Go Modules (go.mod/go.sum)

### Rationale:
- **Go標準**: Go 1.11以降の公式依存関係管理システム
- **バージョン固定**: go.sumで依存関係のハッシュを記録、再現性を担保
- **プライベートモジュール対応**: 将来の企業内利用にも対応
- **ベンダリング不要**: `go mod download` で依存を自動解決

### Alternatives Considered:
- **dep**: 非推奨、Go Modulesに移行済み
- **ベンダリング**: 依存をリポジトリに含めるのはサイズ増大の原因

## 14. セキュリティ - 入力検証

### Decision: 専用の internal/security パッケージ

### Rationale:
- **セキュリティレイヤーの分離**: 入力検証ロジックを一箇所に集約し、保守性を向上
- **再利用性**: ValidateNamespace(), ValidateOutputPath()等の関数を複数箇所で再利用
- **テスト容易性**: セキュリティロジックを独立してユニットテスト可能
- **脆弱性対策の一元管理**: PromQLインジェクション、パストラバーサル等の対策を一箇所で実装

### Functions:
- **ValidateNamespace**: Kubernetes命名規則（RFC 1123）に準拠しているか検証
- **ValidateLabelSelector**: インジェクション攻撃を防ぐ文字列検証
- **ValidateOutputPath**: パストラバーサル攻撃の防止（`..` を含むパスを拒否）
- **ValidatePrometheusURL**: 有効なHTTP/HTTPSスキームの検証
- **SanitizePromQLQuery**: PromQLインジェクション攻撃の防止
- **MaskSensitiveValue**: ログ出力時の機密情報マスキング

### Alternatives Considered:
- **各パッケージで個別に検証**: 重複コードが増え、セキュリティ対策の漏れが発生しやすい
- **サードパーティライブラリ**: Goの標準regexpと文字列処理で十分に実装可能

## 15. セキュリティ - 静的解析

### Decision: gosec (Static Security Scanner for Go)

### Rationale:
- **Go特化**: Go言語のセキュリティ脆弱性パターンを検出（G104: エラーチェック漏れ、G204: コマンドインジェクション等）
- **golangci-lint統合**: golangci-lintの一部として実行可能、CI/CDに容易に組み込める
- **OWASP Top 10対応**: SQL Injection、XSS、コマンドインジェクション等の一般的な脆弱性を検出
- **誤検知の制御**: 設定ファイル（.golangci.yml）で除外ルールを管理可能
- **高速**: 大規模プロジェクトでも数秒で完了

### Alternatives Considered:
- **Snyk Code**: 商用サービスで高機能だが、オープンソースプロジェクトにはgosecで十分
- **SonarQube**: 複数言語対応だが、セットアップが複雑でGo特化の検出精度は劣る

## 16. セキュリティ - 脆弱性スキャン

### Decision: govulncheck + Trivy

### Rationale:
- **govulncheck**:
  - **Go公式**: Go公式の脆弱性データベース（Go Vulnerability Database）を使用
  - **依存関係スキャン**: go.modの依存パッケージに既知の脆弱性が含まれているか検出
  - **コールグラフ解析**: 実際にコードで使用されている脆弱な関数のみを報告（誤検知を削減）
  - **CI統合**: `govulncheck ./...` で全パッケージをスキャン可能

- **Trivy**:
  - **多層スキャン**: コンテナイメージ、ファイルシステム、依存関係の全てをスキャン
  - **複数データベース**: NVD、GitHub Advisory、Alpine SecDB等の脆弱性データベースを統合
  - **高速**: キャッシュ機能により、大規模イメージも数秒でスキャン
  - **GitHub Actions統合**: aquasecurity/trivy-action で容易に統合

### Alternatives Considered:
- **Snyk**: 商用サービスで高機能だが、govulncheck + Trivyの組み合わせで十分
- **Grype**: Anchore社の脆弱性スキャナだが、Trivyの方がデータベースが豊富
- **Clair**: コンテナ脆弱性スキャナだが、Trivyの方が導入が容易

## 17. セキュリティ - TLS証明書検証

### Decision: Go標準の crypto/tls パッケージ

### Rationale:
- **デフォルト有効**: `http.Transport.TLSClientConfig` でTLS証明書検証がデフォルト有効
- **設定可能**: 開発環境で自己署名証明書を使用する場合、`InsecureSkipVerify` で無効化可能（非推奨）
- **カスタムCA対応**: 企業内CAを使用する場合、`RootCAs` で追加可能
- **Go標準**: 外部依存なしで実装可能

### Alternatives Considered:
- **証明書検証なし**: セキュリティリスクが高く、中間者攻撃（MITM）に脆弱
- **サードパーティライブラリ**: Go標準で十分な機能を提供

## まとめ

以上の技術選定により、以下のスタックで実装します：

| カテゴリ | 技術 | バージョン |
|----------|------|------------|
| 言語 | Go | 1.21+ |
| CLIフレームワーク | Cobra | v1.8+ |
| 設定管理 | Viper | v1.18+ |
| K8sクライアント | client-go | v0.29+ |
| Prometheusクライアント | client_golang | v1.18+ |
| 統計処理 | montanaflynn/stats | latest |
| YAMLパッチ | sigs.k8s.io/yaml | latest |
| Markdownレポート | text/template | 標準 |
| LLM (OpenAI) | go-openai | latest |
| LLM (Claude) | anthropic-sdk-go | latest |
| テスト | testing + testify | 標準 + v1.8+ |
| CI/CD | GitHub Actions | - |
| リント | golangci-lint | v1.55+ |
| 依存管理 | Go Modules | 標準 |
| **セキュリティ - 入力検証** | **internal/security** | **独自実装** |
| **セキュリティ - 静的解析** | **gosec** | **latest** |
| **セキュリティ - 脆弱性スキャン** | **govulncheck + Trivy** | **latest** |
| **セキュリティ - TLS検証** | **crypto/tls (標準)** | **標準** |

全ての技術選定は、Kubernetes エコシステムのベストプラクティスに従い、保守性、パフォーマンス、コミュニティサポート、セキュリティを重視しています。
