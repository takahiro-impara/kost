# Tasks: kost (Kubernetes Optimization & Sizing Tool) MVPコア機能

**Input**: Design documents from `/specs/001-mvp-core/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Repository root**: `cmd/kost/`, `internal/`, `tests/`, `examples/`
- Follows Go standard project layout
- Paths shown below follow plan.md structure

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create project directory structure: cmd/kost/, internal/{config,k8s,metrics,engine,patch,report,llm}/, tests/fixtures/{k8s,prometheus}/, examples/, docs/
- [X] T002 Initialize Go modules with go.mod in repository root
- [X] T003 [P] Create Makefile with build, test, lint, clean targets
- [X] T004 [P] Create .gitignore for Go project (out/, bin/, *.test, vendor/)
- [X] T005 [P] Create .golangci.yml configuration file for golangci-lint
- [X] T006 [P] Create GitHub Actions workflow file .github/workflows/ci.yml with lint, test, build jobs
- [X] T007 [P] Create examples/config.yaml with default configuration values
- [X] T008 [P] Create examples/rbac.yaml with Kubernetes RBAC definitions
- [X] T009 [P] Create README.md with project overview and quickstart guide
- [X] T010 [P] Create LICENSE file (Apache 2.0 or MIT)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T011 Create Config struct and loader in internal/config/config.go with viper integration
- [X] T012 [P] Create base types for Deployment in internal/k8s/types.go
- [X] T013 [P] Create base types for Container in internal/k8s/types.go
- [X] T014 [P] Create Kubernetes client initialization in internal/k8s/client.go
- [X] T015 [P] Create Prometheus client initialization in internal/metrics/client.go
- [X] T016 [P] Create base CLI structure with cobra in cmd/kost/main.go
- [X] T017 Create root command with config loading in cmd/kost/root.go
- [X] T018 [P] Create test fixtures: tests/fixtures/k8s/deployment.json
- [X] T019 [P] Create test fixtures: tests/fixtures/prometheus/cpu_metrics.json
- [X] T020 [P] Create test fixtures: tests/fixtures/prometheus/memory_metrics.json
- [X] T021 [P] Add go dependencies: go get github.com/spf13/cobra github.com/spf13/viper k8s.io/client-go github.com/prometheus/client_golang github.com/montanaflynn/stats sigs.k8s.io/yaml
- [X] T022 [P] Add test dependencies: go get github.com/stretchr/testify

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Deploymentリソース最適化の分析と推奨値提示 (Priority: P1) 🎯 MVP

**Goal**: DeploymentのCPU/Memory requests/limitsの過剰・過小判定と推奨値算出

**Independent Test**: 対象NamespaceにDeploymentが1つ以上存在し、Prometheusからメトリクスが取得できる状態で、`scan`→`suggest`→`report`コマンドを順次実行し、レポートファイルに推奨値と根拠が記載されていることを確認する

### Implementation for User Story 1

#### Scan Command (K8s Data Acquisition)

- [ ] T023 [P] [US1] Implement Deployment listing in internal/k8s/deployment.go with client-go
- [ ] T024 [P] [US1] Implement Container resource extraction in internal/k8s/deployment.go
- [ ] T025 [US1] Create scan command in cmd/kost/scan.go with namespace flag
- [ ] T026 [US1] Add Deployment filtering logic (namespace exclusion, label selector) in internal/k8s/deployment.go
- [ ] T027 [US1] Add error handling for K8s API access failures with actionable messages

#### Suggest Command (Metrics & Recommendation)

- [ ] T028 [P] [US1] Create Metrics struct in internal/metrics/types.go
- [ ] T029 [P] [US1] Implement CPU metrics query in internal/metrics/query.go with PromQL
- [ ] T030 [P] [US1] Implement Memory metrics query in internal/metrics/query.go with PromQL
- [ ] T031 [US1] Implement percentile calculation (P50/P95/P99) in internal/metrics/stats.go using montanaflynn/stats
- [ ] T032 [P] [US1] Create ResourceRecommendation struct in internal/engine/types.go
- [ ] T033 [US1] Implement CPU recommendation calculator in internal/engine/resource.go (P95 * safety factor)
- [ ] T034 [US1] Implement Memory recommendation calculator in internal/engine/resource.go (P95 * safety factor)
- [ ] T035 [US1] Implement judgement logic (overprovisioned/underprovisioned/appropriate) in internal/engine/resource.go
- [ ] T036 [US1] Implement saving ratio calculation in internal/engine/resource.go
- [ ] T037 [US1] Create suggest command in cmd/kost/suggest.go
- [ ] T038 [US1] Add error handling for Prometheus connection failures with troubleshooting steps

#### Report Command (Markdown/JSON/Patch Output)

- [ ] T039 [P] [US1] Create Report struct in internal/report/types.go
- [ ] T040 [P] [US1] Implement Markdown template for report in internal/report/markdown.go
- [ ] T041 [P] [US1] Implement JSON summary generation in internal/report/json.go
- [ ] T042 [P] [US1] Implement resource patch generation (Strategic Merge Patch) in internal/patch/resource.go
- [ ] T043 [US1] Create report command in cmd/kost/report.go
- [ ] T044 [US1] Add output directory creation and file writing logic in internal/report/writer.go
- [ ] T045 [US1] Add Top Findings sorting logic (by saving ratio) in internal/report/markdown.go

#### Unit Tests for User Story 1

- [ ] T046 [P] [US1] Unit test for Config loading in internal/config/config_test.go
- [ ] T047 [P] [US1] Unit test for percentile calculation in internal/metrics/stats_test.go
- [ ] T048 [P] [US1] Unit test for CPU recommendation calculator in internal/engine/resource_test.go
- [ ] T049 [P] [US1] Unit test for Memory recommendation calculator in internal/engine/resource_test.go
- [ ] T050 [P] [US1] Unit test for judgement logic in internal/engine/resource_test.go
- [ ] T051 [P] [US1] Unit test for saving ratio calculation in internal/engine/resource_test.go
- [ ] T052 [P] [US1] Unit test for Markdown report generation in internal/report/markdown_test.go
- [ ] T053 [P] [US1] Unit test for patch generation in internal/patch/resource_test.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - 適用可能なYAMLパッチの生成 (Priority: P2)

**Goal**: 推奨値を反映したStrategic Merge Patch形式のYAMLファイル生成

**Independent Test**: User Story 1の完了後、`patches/`ディレクトリ配下にNamespaceとDeployment名ごとのYAMLファイルが生成され、その内容が推奨値を反映したStrategic Merge Patch形式になっていることを確認する

### Implementation for User Story 2

- [ ] T054 [P] [US2] Enhance patch generation to include namespace directory structure in internal/patch/resource.go
- [ ] T055 [US2] Implement patch file naming convention (<deployment>.yaml) in internal/patch/resource.go
- [ ] T056 [US2] Add patch validation logic (kubectl dry-run compatible) in internal/patch/resource.go
- [ ] T057 [US2] Add patch writing to report command in cmd/kost/report.go
- [ ] T058 [P] [US2] Unit test for patch file structure in internal/patch/resource_test.go
- [ ] T059 [P] [US2] Unit test for patch content validation in internal/patch/resource_test.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - HPA設定の最適化推奨 (Priority: P3)

**Goal**: HPA minReplicas/maxReplicasの推奨値算出とレポート出力

**Independent Test**: 対象NamespaceにHPA設定済みDeploymentが1つ以上存在し、過去7日間のPod数メトリクス（kube_deployment_status_replicas）が取得できる状態で、`scan`→`suggest`→`report`コマンドを実行し、レポートにHPA推奨値（minReplicas/maxReplicas）と根拠が記載されていることを確認する

### Implementation for User Story 3

#### HPA Data Acquisition

- [ ] T060 [P] [US3] Create HPA struct in internal/k8s/types.go
- [ ] T061 [P] [US3] Implement HPA listing in internal/k8s/hpa.go with client-go
- [ ] T062 [US3] Add HPA association to Deployment in internal/k8s/deployment.go
- [ ] T063 [US3] Enhance scan command to include HPA detection in cmd/kost/scan.go

#### HPA Metrics & Recommendation

- [ ] T064 [P] [US3] Implement Replicas metrics query (kube_deployment_status_replicas) in internal/metrics/query.go
- [ ] T065 [P] [US3] Create HPARecommendation struct in internal/engine/types.go
- [ ] T066 [US3] Implement HPA minReplicas calculator (P5 * factor) in internal/engine/hpa.go
- [ ] T067 [US3] Implement HPA maxReplicas calculator (P99 * factor) in internal/engine/hpa.go
- [ ] T068 [US3] Implement HPA judgement logic (too_high/too_low/appropriate) in internal/engine/hpa.go
- [ ] T069 [US3] Enhance suggest command to include HPA recommendations in cmd/kost/suggest.go

#### HPA Report & Patch

- [ ] T070 [P] [US3] Add HPA recommendations section to Markdown template in internal/report/markdown.go
- [ ] T071 [P] [US3] Implement HPA patch generation in internal/patch/hpa.go
- [ ] T072 [US3] Add HPA recommendations to JSON summary in internal/report/json.go
- [ ] T073 [US3] Enhance report command to output HPA patches (<deployment>-hpa.yaml) in cmd/kost/report.go

#### Unit Tests for User Story 3

- [ ] T074 [P] [US3] Unit test for HPA listing in internal/k8s/hpa_test.go
- [ ] T075 [P] [US3] Unit test for Replicas metrics query in internal/metrics/query_test.go
- [ ] T076 [P] [US3] Unit test for HPA minReplicas calculator in internal/engine/hpa_test.go
- [ ] T077 [P] [US3] Unit test for HPA maxReplicas calculator in internal/engine/hpa_test.go
- [ ] T078 [P] [US3] Unit test for HPA judgement logic in internal/engine/hpa_test.go
- [ ] T079 [P] [US3] Unit test for HPA patch generation in internal/patch/hpa_test.go

**Checkpoint**: User Stories 1, 2, and 3 should now be independently functional

---

## Phase 6: User Story 4 - 複数LLMプロバイダ対応（Claude Code含む） (Priority: P4)

**Goal**: OpenAIとClaudeの両方のLLMプロバイダをサポート

**Independent Test**: 設定ファイルでLLMプロバイダを「openai」または「claude」に設定し、それぞれでレポート生成を実行し、適切なAPI呼び出しが行われ、説明が生成されることを確認する

### Implementation for User Story 4

#### LLM Provider Interface

- [ ] T080 [P] [US4] Create LLMProvider interface in internal/llm/provider.go
- [ ] T081 [P] [US4] Create StructuredInput type for LLM prompts in internal/llm/types.go
- [ ] T082 [P] [US4] Create LLM config loading (provider, model, API key) in internal/config/config.go

#### OpenAI Provider

- [ ] T083 [P] [US4] Add OpenAI SDK dependency: go get github.com/sashabaranov/go-openai
- [ ] T084 [US4] Implement OpenAI provider in internal/llm/openai.go
- [ ] T085 [US4] Implement OpenAI Chat Completions API call in internal/llm/openai.go
- [ ] T086 [P] [US4] Unit test for OpenAI provider in internal/llm/openai_test.go

#### Claude Provider

- [ ] T087 [P] [US4] Add Anthropic SDK dependency: go get github.com/anthropics/anthropic-sdk-go
- [ ] T088 [US4] Implement Claude provider in internal/llm/claude.go
- [ ] T089 [US4] Implement Claude Messages API call in internal/llm/claude.go
- [ ] T090 [P] [US4] Unit test for Claude provider in internal/llm/claude_test.go

#### Fallback & Integration

- [ ] T091 [P] [US4] Implement rule-based fallback explanation in internal/llm/fallback.go
- [ ] T092 [US4] Add LLM provider factory in internal/llm/provider.go (creates OpenAI or Claude based on config)
- [ ] T093 [US4] Add error handling for LLM connection failures with fallback to rule-based in internal/llm/provider.go
- [ ] T094 [US4] Update report command to optionally call LLM provider in cmd/kost/report.go
- [ ] T095 [P] [US4] Unit test for LLM fallback in internal/llm/fallback_test.go

**Checkpoint**: User Stories 1-4 should now be independently functional

---

## Phase 7: User Story 5 - 生成AIによる説明と優先順位付け (Priority: P5)

**Goal**: LLMによる自然言語説明と優先順位付けをレポートに含める

**Independent Test**: User Story 1, 2の完了後、LLMオプション有効でレポート生成を実行し、レポートに優先順位付けされた改善案と自然言語での説明（理由、注意点）が含まれていることを確認する。LLM接続に失敗した場合、フォールバックとしてルールベースの説明が出力されることを確認する

### Implementation for User Story 5

- [ ] T096 [P] [US5] Create narrative generation prompt template in internal/llm/prompt.go
- [ ] T097 [US5] Implement structured input builder (current/recommended/statistics) in internal/llm/prompt.go
- [ ] T098 [US5] Implement priority ranking logic (by saving ratio, risk) in internal/report/markdown.go
- [ ] T099 [US5] Add narrative field to Report struct in internal/report/types.go
- [ ] T100 [US5] Enhance Markdown template to include LLM-generated narrative in internal/report/markdown.go
- [ ] T101 [US5] Add --llm flag to report command in cmd/kost/report.go
- [ ] T102 [P] [US5] Unit test for narrative generation in internal/llm/provider_test.go
- [ ] T103 [P] [US5] Unit test for priority ranking in internal/report/markdown_test.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T104 [P] Add logging framework (logrus or zap) to all commands
- [ ] T105 [P] Add progress indicators for long-running operations (Prometheus queries, LLM calls)
- [ ] T106 [P] Create CONTRIBUTING.md with development guidelines
- [ ] T107 [P] Create architecture diagram in docs/architecture.md
- [ ] T108 [P] Add version command in cmd/kost/version.go
- [ ] T109 [P] Create GitHub issue templates in .github/ISSUE_TEMPLATE/
- [ ] T110 [P] Create pull request template in .github/pull_request_template.md
- [ ] T111 [P] Add code coverage reporting to CI workflow in .github/workflows/ci.yml
- [ ] T112 [P] Create Docker build target in Makefile
- [ ] T113 [P] Create Dockerfile for container execution
- [ ] T114 [P] Add goreleaser configuration for multi-platform builds in .goreleaser.yml
- [ ] T115 Run quickstart.md validation with real Kubernetes cluster
- [ ] T116 [P] Add golangci-lint to CI and fix all linter warnings
- [ ] T117 [P] Add error message standardization across all commands
- [ ] T118 Code cleanup and gofmt formatting
- [ ] T119 Update README.md with complete usage examples
- [ ] T120 Add performance benchmarks for recommendation engine in internal/engine/benchmark_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4 → P5)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Extends User Story 1 patch functionality
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent HPA functionality
- **User Story 4 (P4)**: Can start after Foundational (Phase 2) - Independent LLM provider support
- **User Story 5 (P5)**: Depends on User Story 4 completion - Uses LLM provider interface

### Within Each User Story

- Scan command before suggest command
- Suggest command before report command
- Models/types before services
- Services before CLI commands
- Core implementation before unit tests (test-after approach as spec doesn't request TDD)

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- Models within a story marked [P] can run in parallel
- Unit tests within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all parallel implementation tasks for User Story 1:
Task T023: "Implement Deployment listing in internal/k8s/deployment.go"
Task T024: "Implement Container resource extraction in internal/k8s/deployment.go"
Task T028: "Create Metrics struct in internal/metrics/types.go"
Task T029: "Implement CPU metrics query in internal/metrics/query.go"
Task T030: "Implement Memory metrics query in internal/metrics/query.go"
Task T032: "Create ResourceRecommendation struct in internal/engine/types.go"

# Launch all parallel unit tests for User Story 1:
Task T046: "Unit test for Config loading"
Task T047: "Unit test for percentile calculation"
Task T048: "Unit test for CPU recommendation calculator"
Task T049: "Unit test for Memory recommendation calculator"
Task T050: "Unit test for judgement logic"
Task T051: "Unit test for saving ratio calculation"
Task T052: "Unit test for Markdown report generation"
Task T053: "Unit test for patch generation"
```

---

## Parallel Example: Across User Stories

```bash
# After Foundational phase completes, launch multiple stories in parallel:
# Team Member 1 works on User Story 1 (P1)
# Team Member 2 works on User Story 3 (P3 - HPA)
# Team Member 3 works on User Story 4 (P4 - LLM providers)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Add User Story 4 → Test independently → Deploy/Demo
6. Add User Story 5 → Test independently → Deploy/Demo
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (P1) - Core optimization
   - Developer B: User Story 3 (P3) - HPA optimization
   - Developer C: User Story 4 (P4) - LLM providers
3. Stories complete and integrate independently
4. Developer D: User Story 2 (P2) - Enhance patches (depends on US1)
5. Developer E: User Story 5 (P5) - LLM narratives (depends on US4)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Follow Go best practices: gofmt, golint, Effective Go
- All code comments and error messages in English
- All documentation (README, reports) in Japanese
- Use table-driven tests (Go convention)
- Use testify for assertions
- Use fixtures for K8s and Prometheus API responses
