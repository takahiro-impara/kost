# 機能仕様書: kost (Kubernetes Optimization & Sizing Tool) MVPコア機能

**Feature Branch**: `001-mvp-core`
**Created**: 2026-01-14
**Updated**: 2026-01-14
**Status**: Draft
**Input**: User description: "kost (Kubernetes Optimization & Sizing Tool) MVPコア機能の実装（HPA最適化、Claude Code LLM対応を含む）"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Deploymentリソース最適化の分析と推奨値提示 (Priority: P1)

SRE/Platform Engineerは、本番環境のKubernetesクラスタでDeploymentのリソース設定（requests/limits）が過剰または過小になっていないかを確認したい。過剰な設定はコストの無駄につながり、過小な設定はパフォーマンス低下やOOMのリスクを招く。

ユーザーは対象のNamespaceを指定してスキャンを実行し、メトリクスに基づいた統計分析により、各Deploymentの現状と推奨値を比較したレポートを受け取る。

**Why this priority**: これがMVPの中核機能であり、「具体的な改善提案」という本ツールの差別化要素の実現に必要不可欠である。この機能単体でも価値を提供できる。

**Independent Test**: 対象NamespaceにDeploymentが1つ以上存在し、Prometheusからメトリクスが取得できる状態で、`scan`→`suggest`→`report`コマンドを順次実行し、レポートファイルに推奨値と根拠が記載されていることを確認する。

**Acceptance Scenarios**:

1. **Given** 本番Namespaceに複数のDeploymentが稼働している、**When** ユーザーがNamespaceを指定してスキャンを実行する、**Then** 全てのDeploymentとコンテナのリソース設定情報が収集される
2. **Given** スキャンが完了している、**When** 7日間のメトリクスを基にP95統計で推奨値生成を実行する、**Then** 各Deploymentに対する推奨CPU/Memoryリクエスト値が算出され、過剰/過小の判定が行われる
3. **Given** 推奨値が生成されている、**When** レポート生成を実行する、**Then** Markdownレポートに上位の改善案、根拠（統計値、係数）、削減率が記載される

---

### User Story 2 - 適用可能なYAMLパッチの生成 (Priority: P2)

SREは推奨値を確認した後、実際にKubernetesマニフェストに適用できる形でパッチファイルを取得したい。手動で値をコピー＆ペーストするのではなく、GitOps運用に組み込める形式で提供されることが望ましい。

ユーザーはレポート生成時に、推奨値を反映したStrategic Merge Patch形式のYAMLファイルを取得する。

**Why this priority**: 推奨値の提示だけでなく、実際に適用可能なパッチを提供することで、改善提案の実行可能性が高まる。ただし、パッチ適用自体は人間の判断に委ねるため、P1よりは優先度が低い。

**Independent Test**: User Story 1の完了後、`patches/`ディレクトリ配下にNamespaceとDeployment名ごとのYAMLファイルが生成され、その内容が推奨値を反映したStrategic Merge Patch形式になっていることを確認する。

**Acceptance Scenarios**:

1. **Given** 推奨値が生成されている、**When** レポート生成を実行する、**Then** `patches/<namespace>/<deployment>.yaml`の形式でパッチファイルが生成される
2. **Given** パッチファイルが生成されている、**When** ファイル内容を確認する、**Then** Deployment名、Namespace、コンテナ名、推奨requests値が正しく記載されている
3. **Given** パッチファイルが生成されている、**When** kubectl applyでパッチを適用する、**Then** Deploymentのresources.requestsが推奨値に更新される（手動テスト）

---

### User Story 3 - HPA設定の最適化推奨 (Priority: P3)

SREはDeploymentにHPA（Horizontal Pod Autoscaler）が設定されている場合、過去のリソース利用傾向から適切なminReplicas/maxReplicasの値を知りたい。現在の設定が過剰（maxReplicasが大きすぎる）または過小（minReplicasが小さすぎてスパイク対応できない）な場合、コストとパフォーマンスの両面で問題が生じる。

ユーザーはスキャン時にHPAが設定されたDeploymentを検出し、過去のPod数の変動傾向（最小/最大/P95等）に基づいて、推奨minReplicas/maxReplicas値をレポートで受け取る。

**Why this priority**: requests/limitsの最適化（P1）と比較すると、HPA最適化は追加的な価値提供である。ただし、オートスケーリング環境では重要な最適化ポイントとなるため、P3として含める。

**Independent Test**: 対象NamespaceにHPA設定済みDeploymentが1つ以上存在し、過去7日間のPod数メトリクス（kube_deployment_status_replicas）が取得できる状態で、`scan`→`suggest`→`report`コマンドを実行し、レポートにHPA推奨値（minReplicas/maxReplicas）と根拠が記載されていることを確認する。

**Acceptance Scenarios**:

1. **Given** DeploymentにHPAが設定されている、**When** スキャンを実行する、**Then** 現在のHPA設定（minReplicas/maxReplicas/targetCPU等）が収集される
2. **Given** HPA設定が収集されている、**When** 7日間のPod数メトリクスを基に推奨値生成を実行する、**Then** 推奨minReplicas（P5統計+余裕）、推奨maxReplicas（P99統計+余裕）が算出される
3. **Given** HPA推奨値が生成されている、**When** レポート生成を実行する、**Then** レポートに現在値、推奨値、根拠（統計値、余裕係数）、削減/増加効果が記載される

---

### User Story 4 - 複数LLMプロバイダ対応（Claude Code含む） (Priority: P4)

SREは組織の方針や予算に応じて、複数のLLMプロバイダを選択できることを望む。特に、Claude Code（Anthropic Claude API）を使用したい場合がある。

ユーザーは設定ファイルでLLMプロバイダ（OpenAI、Claude Code等）を選択し、それぞれのAPI仕様に応じた説明生成を受け取る。

**Why this priority**: LLM機能自体が付加価値（P1, P2はLLMなしでも動作）であり、プロバイダの選択肢拡大はさらなる付加価値のため、P4とする。ただし、柔軟性の観点から実装価値は高い。

**Independent Test**: 設定ファイルでLLMプロバイダを「openai」または「claude」に設定し、それぞれでレポート生成を実行し、適切なAPI呼び出しが行われ、説明が生成されることを確認する。

**Acceptance Scenarios**:

1. **Given** 設定ファイルでLLMプロバイダが「openai」に設定されている、**When** レポート生成を実行する、**Then** OpenAI APIが呼び出され、説明が生成される
2. **Given** 設定ファイルでLLMプロバイダが「claude」に設定されている、**When** レポート生成を実行する、**Then** Claude API（Anthropic）が呼び出され、説明が生成される
3. **Given** いずれかのLLMプロバイダが設定されている、**When** API接続に失敗する、**Then** ルールベースのフォールバック説明が生成される

---

### User Story 5 - 生成AIによる説明と優先順位付け (Priority: P5)

SREは多数のDeploymentがある場合、どこから改善すべきかの優先順位と、なぜその推奨が行われているのかの自然言語による説明を受け取りたい。統計値だけでなく、文脈を含めた説明があると判断がしやすくなる。

ユーザーはレポート生成時にLLMオプションを有効化し、推奨の優先順位付けと自然言語での説明をレポートに含める。

**Why this priority**: AIによる説明は付加価値ではあるが、推奨値自体は統計とルールで算出されるため、必須ではない。MVP段階では「あると便利」な機能として位置付ける。User Story 3, 4がHPA最適化とLLMプロバイダ選択に再配置されたため、優先度をP5に変更。

**Independent Test**: User Story 1, 2の完了後、LLMオプション有効でレポート生成を実行し、レポートに優先順位付けされた改善案と自然言語での説明（理由、注意点）が含まれていることを確認する。LLM接続に失敗した場合、フォールバックとしてルールベースの説明が出力されることを確認する。

**Acceptance Scenarios**:

1. **Given** LLMオプションが有効化されている、**When** レポート生成を実行する、**Then** レポートに「影響大」「リスク大」などの優先順位付けが記載される
2. **Given** LLMオプションが有効化されている、**When** レポート生成を実行する、**Then** 各推奨に対して「なぜ無駄か」「どう直すか」「注意点」の説明が自然言語で記載される
3. **Given** LLM接続に失敗した、**When** レポート生成を実行する、**Then** ルールベースのフォールバック説明がレポートに記載される

---

### Edge Cases

- **Prometheusメトリクスが取得できない場合**: システムはエラーメッセージを表示し、必要なメトリクス（container_cpu_usage_seconds_total, container_memory_working_set_bytes, kube_deployment_status_replicas）と確認手順（Prometheus接続先、クエリの検証方法）を提示する
- **対象Namespaceに一切のDeploymentが存在しない場合**: スキャン結果は空となり、レポートにも「対象リソースなし」と記載される
- **現在のrequestsが未設定（null）の場合**: システムは「未設定」として扱い、推奨値のみを提示する（過剰/過小判定は行わない）
- **P95統計値が極端に小さい場合**: 最小値（minCpuMilli=20m, minMemMi=64Mi）が適用され、推奨値がこれを下回らないようにする
- **RBAC権限不足でK8s APIにアクセスできない場合**: システムはエラーメッセージを表示し、必要な権限（Deployment/Pod/HPA/Namespaceの読み取り専用）をドキュメントに記載する
- **HPAが設定されていないDeploymentの場合**: HPA推奨値のセクションはスキップされ、requests/limitsの推奨のみが提示される
- **HPA用のPod数メトリクスが不足している場合**: HPA推奨は「データ不足」として扱い、現在値のみを記載する
- **LLMプロバイダのAPI認証情報が未設定の場合**: LLM機能は無効化され、ルールベース説明のみでレポートが生成される
- **未対応のLLMプロバイダが指定された場合**: エラーメッセージを表示し、対応プロバイダ（openai, claude）をリストする
- **悪意のある入力（PromQL injection試行）が検出された場合**: システムは入力を拒否し、安全なクエリのみを実行する
- **出力ディレクトリパスにパストラバーサルが含まれる場合**: システムはエラーを返し、相対パス（../等）を含むパスを拒否する
- **TLS証明書検証に失敗した場合**: システムは接続を拒否し、証明書検証を無効化するオプション（--insecure-skip-tls-verify）を提供する

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: システムはKubernetes APIに接続し、指定されたNamespace内の全Deploymentリソースの情報を取得できなければならない
- **FR-002**: システムは各Deploymentの全コンテナに設定されたCPU/Memory requests/limitsを抽出できなければならない
- **FR-003**: システムはPrometheusエンドポイントに接続し、指定期間（デフォルト7日間）のCPU使用率とMemory使用量メトリクスを取得できなければならない
- **FR-004**: システムは取得したメトリクスからP50/P95/P99パーセンタイル統計を算出できなければならない
- **FR-005**: システムは統計値とセーフティファクター（デフォルト1.2）を用いて、推奨CPU/Memory requests値を算出できなければならない
- **FR-006**: システムは現在のrequests値と推奨値を比較し、過剰（current > Pxx * 2.0）または過小（current < Pxx * 1.1）を判定できなければならない
- **FR-007**: システムは推奨値と現在値の差から、CPU/Memory削減率（%）を計算できなければならない
- **FR-008**: システムはMarkdown形式のレポートファイル（report.md）を生成し、上位の改善案、理由、削減率、注意点を記載できなければならない
- **FR-009**: システムはStrategic Merge Patch形式のYAMLファイルを`patches/<namespace>/<deployment>.yaml`の構造で生成できなければならない
- **FR-010**: システムは機械可読なJSON形式のサマリファイル（summary.json）を生成し、将来の自動化連携に備えなければならない
- **FR-011**: システムは設定ファイル（config.yaml）から、Kubernetes context、Prometheusエンドポイント、分析期間、パーセンタイル、セーフティファクター、最小値、除外Namespace、出力形式を読み込めなければならない
- **FR-012**: ユーザーはCLIコマンドでスキャン（scan）、推奨値生成（suggest）、レポート生成（report）を個別に実行できなければならない
- **FR-013**: システムはKubernetes APIに接続し、指定されたNamespace内のHPA（HorizontalPodAutoscaler）リソースの情報を取得できなければならない
- **FR-014**: システムは各HPAに対して、現在のminReplicas、maxReplicas、targetCPUUtilizationPercentage（または他のメトリクス）を抽出できなければならない
- **FR-015**: システムはPrometheusエンドポイントに接続し、指定期間（デフォルト7日間）の実際のPod数（kube_deployment_status_replicas）メトリクスを取得できなければならない
- **FR-016**: システムは取得したPod数メトリクスからP5/P50/P99パーセンタイル統計を算出できなければならない
- **FR-017**: システムはPod数統計と余裕係数（例：minは0.8倍、maxは1.3倍）を用いて、推奨minReplicas/maxReplicasを算出できなければならない
- **FR-018**: システムは現在のHPA設定と推奨値を比較し、過剰（maxが大きすぎる）または過小（minが小さすぎる）を判定できなければならない
- **FR-019**: システムはHPA推奨値をレポート（report.md）とパッチファイル（patches/<namespace>/<deployment>-hpa.yaml）に含めて出力できなければならない
- **FR-020**: システムは設定ファイルまたは環境変数からLLMプロバイダ（openai, claude等）を読み込めなければならない
- **FR-021**: システムは選択されたLLMプロバイダに応じて、適切なAPI仕様（OpenAI Chat Completions、Anthropic Messages等）で構造化データを送信できなければならない
- **FR-022**: システムはLLMプロバイダがOpenAIの場合、OpenAI Chat Completions API（GPT-4等）を呼び出し、優先順位付けと自然言語説明を取得できなければならない
- **FR-023**: システムはLLMプロバイダがClaudeの場合、Anthropic Messages API（Claude 3.5 Sonnet等）を呼び出し、優先順位付けと自然言語説明を取得できなければならない
- **FR-024**: システムはLLM接続に失敗した場合（ネットワークエラー、認証失敗、API制限等）、ルールベースのフォールバック説明を生成しなければならない
- **FR-025**: システムはデフォルトでdry-runモードとし、自動的にKubernetesリソースを変更してはならない

### Security Requirements

- **FR-026**: システムは全てのAPI認証情報（LLM APIキー、Kubernetesトークン等）を環境変数から読み込み、設定ファイル、ログ、レポート出力に含めてはならない
- **FR-027**: システムはPrometheusへのクエリ実行時、ユーザー入力を適切にエスケープし、PromQL injectionを防止しなければならない
- **FR-028**: システムはファイル出力時、パストラバーサル攻撃を防ぐためパス検証を行い、指定された出力ディレクトリ外への書き込みを禁止しなければならない
- **FR-029**: システムはKubernetes/Prometheus APIへの接続時、TLS証明書の検証を行い（設定で無効化可能）、中間者攻撃のリスクを軽減しなければならない
- **FR-030**: システムは最小権限の原則に従い、Kubernetes APIへのアクセスは読み取り専用（get, list）権限のみを要求し、書き込み権限（create, update, delete）を必要としてはならない
- **FR-031**: システムは依存関係の脆弱性を定期的にスキャンし、既知の脆弱性を含むライブラリの使用を避けなければならない
- **FR-032**: システムはエラーメッセージやログに、APIキー、トークン、パスワード等の機密情報を含めてはならない

### Key Entities

- **Deployment**: Kubernetesワークロードの実行単位。Namespace、名前、コンテナリスト、各コンテナのresources設定、関連HPA（任意）を持つ
- **Container**: Deployment内の実行コンテナ。名前、CPU/Memory requests、CPU/Memory limitsを持つ
- **HPA（HorizontalPodAutoscaler）**: Deployment のオートスケーリング設定。minReplicas、maxReplicas、targetメトリクス（CPU利用率等）、関連Deploymentを持つ
- **Metrics**: Prometheusから取得した時系列メトリクス。期間、パーセンタイル統計（P50/P95/P99）、対象コンテナまたはDeployment（Pod数）への紐付けを持つ
- **ResourceRecommendation**: リソース（requests/limits）の推奨値算出結果。対象Deployment/Container、現状requests、推奨requests、判定（過剰/過小/適正）、削減率、根拠（統計値、係数）を持つ
- **HPARecommendation**: HPA設定の推奨値算出結果。対象HPA/Deployment、現状minReplicas/maxReplicas、推奨minReplicas/maxReplicas、判定（過剰/過小/適正）、削減/増加効果、根拠（Pod数統計、余裕係数）を持つ
- **LLMProvider**: LLMプロバイダの抽象化。名前（openai, claude等）、API仕様、認証情報、エンドポイントを持つ
- **Report**: 最終出力物。Markdownレポート、YAMLパッチファイル群（Deployment resources、HPA設定）、JSONサマリを含む

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: ユーザーは10分以内にツールをインストールし、最初のレポートを生成できる（README Quickstartに従う）
- **SC-002**: システムは対象Namespace内の少なくとも1つのDeploymentに対して、根拠付き推奨値（統計値、係数、推奨値、削減率）を提示できる
- **SC-003**: 生成されたレポートには、推奨値の算出根拠（使用したパーセンタイル、セーフティファクター、丸め処理）が明記されている
- **SC-004**: 生成されたYAMLパッチファイルは、kubectl applyで直接適用可能な形式である
- **SC-005**: システムはPrometheusまたはKubernetes APIへのアクセスに失敗した場合、エラー内容と確認手順を具体的に提示する
- **SC-006**: システムは最小3つのCLIコマンド（scan/suggest/report）をサポートし、各コマンドは独立して実行できる
- **SC-007**: LLMオプションを無効化した場合でも、統計ベースの推奨値とレポートが正常に生成される
- **SC-008**: 推奨値算出ロジック（統計+ルール）はユニットテストでカバーされ、再現性が担保されている
- **SC-009**: HPA設定済みDeploymentに対して、過去のPod数傾向に基づくminReplicas/maxReplicas推奨値が提示される
- **SC-010**: HPA推奨値のレポートには、Pod数統計（P5/P99等）と余裕係数が明記されている
- **SC-011**: ユーザーは設定ファイルでLLMプロバイダ（openai、claude等）を選択でき、それぞれ正常に動作する
- **SC-012**: Claude APIを使用した場合、OpenAI APIと同等の説明品質でレポートが生成される
- **SC-013**: いずれかのLLMプロバイダでAPI接続に失敗した場合でも、ルールベース説明が生成されレポートは完成する
- **SC-014**: システムは設定ファイル、ログファイル、レポート出力のいずれにもAPI認証情報を含まない
- **SC-015**: 静的コード解析（golangci-lint with gosec）がCI/CDパイプラインで実行され、セキュリティ警告がゼロである
- **SC-016**: 依存関係の脆弱性スキャン（go mod vulnerabilities check）がCI/CDパイプラインで実行され、既知の高/重大脆弱性がゼロである
- **SC-017**: Kubernetes RBAC設定が最小権限の原則に従い、read-only権限のみで動作する

## Assumptions

本仕様では以下の前提を置いています：

- **Prometheusの前提**: クラスタ内にPrometheus（またはPrometheus互換メトリクスストア）が導入済みで、container_cpu_usage_seconds_total、container_memory_working_set_bytes、kube_deployment_status_replicasのメトリクスが取得可能
- **RBAC権限**: ツールを実行するユーザー/ServiceAccountは、対象NamespaceのDeployment/Pod/HPAに対する読み取り権限を持つ
- **ワークロード種別**: MVP段階ではDeploymentのみを対象とし、StatefulSet、DaemonSet、Job等は対象外
- **HPA設定**: HPA設定がない Deploymentも正常に処理され、その場合はHPA推奨値セクションはスキップされる
- **メトリクス期間**: デフォルト7日間のメトリクスで十分な統計精度が得られる
- **パーセンタイル選択**: requests/limitsはP95をデフォルト、HPAのminReplicasはP5、maxReplicasはP99をデフォルトとする
- **セーフティファクター**: requests/limitsは1.2倍、HPAのminReplicasは0.8倍、maxReplicasは1.3倍をデフォルトとする
- **LLMプロバイダ**: OpenAI（GPT-4等）とClaude（Claude 3.5 Sonnet等）の2つのプロバイダをサポートし、設定ファイルまたは環境変数で選択可能
- **LLM API認証**: LLMプロバイダのAPI認証情報（APIキー等）は環境変数で提供され、設定ファイルには含まれない
- **出力ディレクトリ**: デフォルト`./out`にレポート、パッチ、サマリを出力する
- **言語**: ドキュメント、レポートは日本語。コード、エラーメッセージは英語（国際化を考慮）
- **セキュリティ**: API認証情報は環境変数で管理し、設定ファイルに含めない。最小権限の原則に従い、Kubernetes APIへはread-only権限のみでアクセスする
- **脆弱性管理**: 依存関係は定期的に更新し、既知の脆弱性を含むバージョンの使用を避ける。CI/CDパイプラインで自動スキャンを実施する
- **入力検証**: 全てのユーザー入力（Namespace名、ラベルセレクタ、出力パス等）は適切にバリデーションし、injection攻撃を防止する
