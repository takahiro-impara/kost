# Feature Specification: kost (Kubernetes Optimization & Sizing Tool) MVP Core Features

**Feature Branch**: `001-mvp-core`
**Created**: 2026-01-14
**Updated**: 2026-01-15
**Status**: Implemented (P1-P2 complete, P3-P5 next phase)
**Input**: User description: "kost (Kubernetes Optimization & Sizing Tool) MVP core feature implementation (including HPA optimization, Claude Code LLM support)"

**Implementation Status Summary**:
- ✅ User Story 1 (P1): Deployment resource optimization - Implementation complete, tested
- ✅ User Story 2 (P2): YAML patch generation - Implementation complete, tested
- ⚠️ User Story 3 (P3): HPA optimization - Next phase
- ⚠️ User Story 4 (P4): Multiple LLM providers - Partial implementation (configuration only, testing incomplete)
- ⚠️ User Story 5 (P5): AI explanation generation - Partial implementation (testing incomplete)
- ✅ Security enhancement: gosec 0 issues, input validation 97.7% coverage
- ✅ Unit tests: 63.8% coverage (60% target achieved)
- ✅ E2E tests: Main scenarios passing

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Deployment Resource Optimization Analysis and Recommendations (Priority: P1)

SRE/Platform Engineers want to verify whether Deployment resource settings (requests/limits) in production Kubernetes clusters are excessive or insufficient. Excessive settings lead to cost waste, while insufficient settings risk performance degradation and OOM.

Users execute a scan targeting a specific namespace and receive a report comparing current Deployment status with recommended values based on metric-driven statistical analysis.

**Why this priority**: This is the core MVP feature and essential for realizing the differentiating factor of "concrete improvement proposals". This feature alone provides value.

**Independent Test**: With at least one Deployment in the target namespace and metrics retrievable from Prometheus, execute `scan`→`suggest`→`report` commands sequentially and verify that the report file contains recommended values and rationale.

**Acceptance Scenarios**:

1. **Given** multiple Deployments running in production namespace, **When** user executes scan specifying namespace, **Then** resource configuration information for all Deployments and containers is collected
2. **Given** scan is complete, **When** recommendation generation is executed with P95 statistics based on 7 days of metrics, **Then** recommended CPU/Memory request values are calculated for each Deployment, and over/under-provisioning judgments are made
3. **Given** recommendations are generated, **When** report generation is executed, **Then** Markdown report contains top improvement proposals, rationale (statistical values, coefficients), and reduction rates

---

### User Story 2 - Applicable YAML Patch Generation (Priority: P2)

After reviewing recommendations, SREs want to obtain patch files in a form applicable to actual Kubernetes manifests. Rather than manually copying and pasting values, it's desirable to provide them in a format that can be integrated into GitOps operations.

Users obtain YAML files in Strategic Merge Patch format reflecting recommended values during report generation.

**Why this priority**: Providing not just recommended values but also applicable patches increases the feasibility of implementing improvement proposals. However, since patch application itself is left to human judgment, priority is lower than P1.

**Independent Test**: After completing User Story 1, verify that YAML files are generated under the `patches/` directory per namespace and Deployment name, and their contents are in Strategic Merge Patch format reflecting recommended values.

**Acceptance Scenarios**:

1. **Given** recommendations are generated, **When** report generation is executed, **Then** patch files are generated in `patches/<namespace>/<deployment>.yaml` format
2. **Given** patch files are generated, **When** file contents are reviewed, **Then** Deployment name, namespace, container name, and recommended requests values are correctly listed
3. **Given** patch files are generated, **When** patch is applied with kubectl apply, **Then** Deployment's resources.requests are updated to recommended values (manual test)

---

### User Story 3 - HPA Configuration Optimization Recommendations (Priority: P3)

When HPA (Horizontal Pod Autoscaler) is configured for a Deployment, SREs want to know appropriate minReplicas/maxReplicas values based on historical resource usage trends. If current settings are excessive (maxReplicas too large) or insufficient (minReplicas too small to handle spikes), issues arise in both cost and performance aspects.

Users detect Deployments with HPA configured during scanning and receive recommended minReplicas/maxReplicas values in the report based on historical Pod count fluctuation trends (min/max/P95, etc.).

**Why this priority**: Compared to requests/limits optimization (P1), HPA optimization provides additional value. However, it's an important optimization point in autoscaling environments, so it's included as P3.

**Independent Test**: With at least one HPA-configured Deployment in the target namespace and Pod count metrics (kube_deployment_status_replicas) for the past 7 days retrievable, execute `scan`→`suggest`→`report` commands and verify that HPA recommendations (minReplicas/maxReplicas) and rationale are included in the report.

**Acceptance Scenarios**:

1. **Given** HPA is configured for Deployment, **When** scan is executed, **Then** current HPA configuration (minReplicas/maxReplicas/targetCPU, etc.) is collected
2. **Given** HPA configuration is collected, **When** recommendation generation is executed based on 7 days of Pod count metrics, **Then** recommended minReplicas (P5 statistics + margin) and recommended maxReplicas (P99 statistics + margin) are calculated
3. **Given** HPA recommendations are generated, **When** report generation is executed, **Then** report contains current values, recommended values, rationale (statistical values, margin coefficients), and reduction/increase effects

---

### User Story 4 - Multiple LLM Provider Support (Including Claude Code) (Priority: P4)

SREs want to be able to select multiple LLM providers according to organizational policy and budget. In particular, there are cases where they want to use Claude Code (Anthropic Claude API).

Users select an LLM provider (OpenAI, Claude Code, etc.) in the configuration file and receive explanation generation according to each API specification.

**Why this priority**: LLM functionality itself is added value (P1, P2 work without LLM), and expanding provider options is further added value, so it's P4. However, implementation value is high from a flexibility perspective.

**Independent Test**: Set LLM provider to "openai" or "claude" in configuration file, execute report generation with each, and verify that appropriate API calls are made and explanations are generated.

**Acceptance Scenarios**:

1. **Given** LLM provider is set to "openai" in configuration file, **When** report generation is executed, **Then** OpenAI API is called and explanation is generated
2. **Given** LLM provider is set to "claude" in configuration file, **When** report generation is executed, **Then** Claude API (Anthropic) is called and explanation is generated
3. **Given** either LLM provider is configured, **When** API connection fails, **Then** rule-based fallback explanation is generated

---

### User Story 5 - AI-Powered Explanations and Prioritization (Priority: P5)

When there are many Deployments, SREs want to receive prioritization of where to improve first, and natural language explanations of why those recommendations are made. Having explanations with context, not just statistical values, makes judgment easier.

Users enable the LLM option during report generation to include prioritization of recommendations and natural language explanations in the report.

**Why this priority**: AI-powered explanations are added value, but since recommended values themselves are calculated by statistics and rules, they're not essential. In the MVP stage, positioned as a "nice to have" feature. Since User Story 3, 4 have been reassigned to HPA optimization and LLM provider selection, priority changed to P5.

**Independent Test**: After completing User Story 1, 2, execute report generation with LLM option enabled and verify that the report contains prioritized improvement proposals and natural language explanations (reasons, notes). When LLM connection fails, verify that rule-based explanations are output as fallback.

**Acceptance Scenarios**:

1. **Given** LLM option is enabled, **When** report generation is executed, **Then** report contains prioritization such as "High Impact", "High Risk"
2. **Given** LLM option is enabled, **When** report generation is executed, **Then** each recommendation has natural language explanations for "Why wasteful", "How to fix", "Notes"
3. **Given** LLM connection failed, **When** report generation is executed, **Then** rule-based fallback explanation is included in report

---

### Edge Cases

- **When Prometheus metrics cannot be retrieved**: System displays error message and presents required metrics (container_cpu_usage_seconds_total, container_memory_working_set_bytes, kube_deployment_status_replicas) and verification procedures (Prometheus connection destination, query validation method)
- **When no Deployments exist in target namespace**: Scan results are empty and report indicates "No target resources"
- **When current requests are unset (null)**: System treats as "unset" and presents only recommended values (no over/under-provisioning judgment)
- **When P95 statistical value is extremely small**: Minimum values (minCpuMilli=20m, minMemMi=64Mi) are applied, ensuring recommended values don't fall below these
- **When K8s API cannot be accessed due to insufficient RBAC permissions**: System displays error message and documents required permissions (read-only for Deployment/Pod/HPA/Namespaces)
- **When Deployment has no HPA configured**: HPA recommendation section is skipped, only requests/limits recommendations are presented
- **When Pod count metrics for HPA are insufficient**: HPA recommendation is treated as "insufficient data", only current values are listed
- **When LLM provider API credentials are unset**: LLM functionality is disabled and report is generated with rule-based explanations only
- **When unsupported LLM provider is specified**: Error message is displayed listing supported providers (openai, claude)
- **When malicious input (PromQL injection attempt) is detected**: System rejects input and executes only safe queries
- **When output directory path contains path traversal**: System returns error and rejects paths containing relative paths (../, etc.)
- **When TLS certificate verification fails**: System rejects connection and provides option to disable certificate verification (--insecure-skip-tls-verify)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System must be able to connect to Kubernetes API and retrieve information for all Deployment resources in specified namespace
- **FR-002**: System must be able to extract CPU/Memory requests/limits configured for all containers in each Deployment
- **FR-003**: System must be able to connect to Prometheus endpoint and retrieve CPU usage rate and Memory usage metrics for specified period (default 7 days)
- **FR-004**: System must be able to calculate P50/P95/P99 percentile statistics from retrieved metrics
- **FR-005**: System must be able to calculate recommended CPU/Memory requests values using statistical values and safety factor (default 1.2)
- **FR-006**: System must be able to compare current requests values with recommended values and judge over-provisioning (current > Pxx * 2.0) or under-provisioning (current < Pxx * 1.1)
- **FR-007**: System must be able to calculate CPU/Memory reduction rate (%) from difference between recommended values and current values
- **FR-008**: System must be able to generate Markdown format report file (report.md) listing top improvement proposals, reasons, reduction rates, and notes
- **FR-009**: System must be able to generate YAML files in Strategic Merge Patch format in `patches/<namespace>/<deployment>.yaml` structure
- **FR-010**: System must be able to generate machine-readable JSON format summary file (summary.json) for future automation integration
- **FR-011**: System must be able to read from configuration file (config.yaml): Kubernetes context, Prometheus endpoint, analysis period, percentile, safety factor, minimum values, excluded namespaces, output format
- **FR-012**: Users must be able to independently execute CLI commands for scan, recommendation generation (suggest), and report generation (report)
- **FR-013**: System must be able to connect to Kubernetes API and retrieve HPA (HorizontalPodAutoscaler) resource information in specified namespace
- **FR-014**: System must be able to extract current minReplicas, maxReplicas, targetCPUUtilizationPercentage (or other metrics) for each HPA
- **FR-015**: System must be able to connect to Prometheus endpoint and retrieve actual Pod count (kube_deployment_status_replicas) metrics for specified period (default 7 days)
- **FR-016**: System must be able to calculate P5/P50/P99 percentile statistics from retrieved Pod count metrics
- **FR-017**: System must be able to calculate recommended minReplicas/maxReplicas using Pod count statistics and margin coefficients (e.g., min 0.8x, max 1.3x)
- **FR-018**: System must be able to compare current HPA configuration with recommended values and judge over-provisioning (max too large) or under-provisioning (min too small)
- **FR-019**: System must be able to output HPA recommendations in report (report.md) and patch files (patches/<namespace>/<deployment>-hpa.yaml)
- **FR-020**: System must be able to read LLM provider (openai, claude, etc.) from configuration file or environment variables
- **FR-021**: System must be able to send structured data according to appropriate API specifications (OpenAI Chat Completions, Anthropic Messages, etc.) based on selected LLM provider
- **FR-022**: When LLM provider is OpenAI, system must be able to call OpenAI Chat Completions API (GPT-4, etc.) and retrieve prioritization and natural language explanations
- **FR-023**: When LLM provider is Claude, system must be able to call Anthropic Messages API (Claude 3.5 Sonnet, etc.) and retrieve prioritization and natural language explanations
- **FR-024**: When LLM connection fails (network error, authentication failure, API limits, etc.), system must generate rule-based fallback explanations
- **FR-025**: System must default to dry-run mode and not automatically modify Kubernetes resources

### Security Requirements

- **FR-026**: System must read all API credentials (LLM API keys, Kubernetes tokens, etc.) from environment variables and must not include them in configuration files, logs, or report outputs
- **FR-027**: When executing Prometheus queries, system must properly escape user input to prevent PromQL injection
- **FR-028**: When outputting files, system must validate paths to prevent path traversal attacks and prohibit writing outside specified output directory
- **FR-029**: When connecting to Kubernetes/Prometheus APIs, system must verify TLS certificates (can be disabled in configuration) to mitigate man-in-the-middle attack risks
- **FR-030**: Following principle of least privilege, system must request only read-only (get, list) permissions for Kubernetes API access and must not require write permissions (create, update, delete)
- **FR-031**: System must periodically scan dependency vulnerabilities and avoid using libraries with known vulnerabilities
- **FR-032**: System must not include sensitive information such as API keys, tokens, passwords in error messages or logs

### Key Entities

- **Deployment**: Kubernetes workload execution unit. Has namespace, name, container list, resources configuration for each container, associated HPA (optional)
- **Container**: Execution container within Deployment. Has name, CPU/Memory requests, CPU/Memory limits
- **HPA (HorizontalPodAutoscaler)**: Deployment autoscaling configuration. Has minReplicas, maxReplicas, target metrics (CPU utilization, etc.), associated Deployment
- **Metrics**: Time-series metrics retrieved from Prometheus. Has period, percentile statistics (P50/P95/P99), association to target container or Deployment (Pod count)
- **ResourceRecommendation**: Resource (requests/limits) recommendation calculation result. Has target Deployment/Container, current requests, recommended requests, judgment (over/under/appropriate), reduction rate, rationale (statistical values, coefficients)
- **HPARecommendation**: HPA configuration recommendation calculation result. Has target HPA/Deployment, current minReplicas/maxReplicas, recommended minReplicas/maxReplicas, judgment (over/under/appropriate), reduction/increase effect, rationale (Pod count statistics, margin coefficients)
- **LLMProvider**: LLM provider abstraction. Has name (openai, claude, etc.), API specification, credentials, endpoint
- **Report**: Final output. Includes Markdown report, YAML patch file collection (Deployment resources, HPA configuration), JSON summary

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can install tool and generate first report within 10 minutes (following README Quickstart)
- **SC-002**: System can present recommended values with rationale (statistical values, coefficients, recommended values, reduction rates) for at least one Deployment in target namespace
- **SC-003**: Generated reports clearly state calculation rationale for recommended values (percentile used, safety factor, rounding)
- **SC-004**: Generated YAML patch files are in format directly applicable with kubectl apply
- **SC-005**: When access to Prometheus or Kubernetes API fails, system specifically presents error content and verification procedures
- **SC-006**: System supports at least 3 CLI commands (scan/suggest/report), each executable independently
- **SC-007**: When LLM option is disabled, statistics-based recommended values and reports are generated normally
- **SC-008**: Recommendation calculation logic (statistics + rules) is covered by unit tests with reproducibility guaranteed
- **SC-009**: For HPA-configured Deployments, minReplicas/maxReplicas recommendations based on historical Pod count trends are presented
- **SC-010**: HPA recommendation reports clearly state Pod count statistics (P5/P99, etc.) and margin coefficients
- **SC-011**: Users can select LLM provider (openai, claude, etc.) in configuration file, and each works normally
- **SC-012**: When using Claude API, reports are generated with explanation quality equivalent to OpenAI API
- **SC-013**: Even when API connection fails with any LLM provider, rule-based explanations are generated and report is completed
- **SC-014**: System does not include API credentials in configuration files, log files, or report outputs
- **SC-015**: Static code analysis (golangci-lint with gosec) is executed in CI/CD pipeline with zero security warnings
- **SC-016**: Dependency vulnerability scanning (go mod vulnerabilities check) is executed in CI/CD pipeline with zero known high/critical vulnerabilities
- **SC-017**: Kubernetes RBAC configuration follows principle of least privilege and operates with read-only permissions only

## Assumptions

This specification makes the following assumptions:

- **Prometheus prerequisite**: Prometheus (or Prometheus-compatible metrics store) is deployed in cluster with container_cpu_usage_seconds_total, container_memory_working_set_bytes, kube_deployment_status_replicas metrics retrievable
- **RBAC permissions**: User/ServiceAccount executing tool has read permissions for Deployment/Pod/HPA in target namespace
- **Workload types**: MVP stage targets only Deployments, excluding StatefulSet, DaemonSet, Job, etc.
- **HPA configuration**: Deployments without HPA configuration are processed normally, HPA recommendation section is skipped in such cases
- **Metrics period**: Default 7 days of metrics provides sufficient statistical accuracy
- **Percentile selection**: Default P95 for requests/limits, P5 for HPA minReplicas, P99 for maxReplicas
- **Safety factors**: Default 1.2x for requests/limits, 0.8x for HPA minReplicas, 1.3x for maxReplicas
- **LLM providers**: Support two providers OpenAI (GPT-4, etc.) and Claude (Claude 3.5 Sonnet, etc.), selectable via configuration file or environment variables
- **LLM API authentication**: LLM provider API credentials (API keys, etc.) are provided via environment variables and not included in configuration files
- **Output directory**: Default output to `./out` for reports, patches, summary
- **Language**: Documentation and reports in Japanese. Code and error messages in English (considering internationalization)
- **Security**: Manage API credentials via environment variables, not in configuration files. Follow principle of least privilege, access Kubernetes API with read-only permissions only
- **Vulnerability management**: Update dependencies regularly, avoid using versions with known vulnerabilities. Implement automatic scanning in CI/CD pipeline
- **Input validation**: Properly validate all user inputs (namespace names, label selectors, output paths, etc.) to prevent injection attacks

---

## Security Implementation Details

### Implemented Security Features

#### 1. Input Validation (internal/security/validation.go)

**Implementation**: Validate all user inputs to prevent injection attacks

- **ValidateNamespace**: Namespace validation compliant with Kubernetes naming conventions
  - Reject empty strings
  - Reject uppercase and special characters
  - Length limit (63 characters)
  - Regex: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`

- **ValidateLabelSelector**: Label selector injection prevention
  - Reject quotation marks (' / ")
  - Reject semicolons (;)
  - Reject other special characters

- **ValidatePrometheusURL**: URL validation
  - Allow only http/https schemes
  - Reject special characters
  - Reject empty strings

- **ValidateOutputPath**: Path traversal attack prevention
  - Detect and reject `..` patterns
  - Prevent access to sensitive directories (/etc, /root, /sys, /proc, /dev)

- **SanitizePromQLQuery**: PromQL injection prevention
  - Detect SQL injection-like patterns (;, --, /*, */)
  - Query sanitization

- **MaskSensitiveValue**: API key masking for log output
  - 8 characters or less: Full mask (`***`)
  - 9 characters or more: Show only first 4 and last 4 characters (`sk-1...cdef`)

**Test Coverage**: 97.7% (internal/security/validation_test.go)

#### 2. File Permissions (gosec compliant)

**Implementation**: Secure file and directory permission settings

- **Directory creation**: 0750 (owner: rwx, group: r-x, other: ---)
- **File creation**: 0600 (owner: rw-, group: ---, other: ---)
- **Applied in**: internal/report/writer.go

#### 3. API Credential Management

**Implementation**: Manage API credentials via environment variables only

- **Kubernetes**: kubeconfig file or in-cluster authentication
- **OpenAI**: `OPENAI_API_KEY` environment variable
- **Claude**: `ANTHROPIC_API_KEY` environment variable
- **Validation**: Check environment variable existence when loading configuration file (internal/config/config.go)
- **Prohibited**: Including credentials in configuration files

#### 4. RBAC Least Privilege

**Implementation**: Use read-only permissions only

```yaml
# examples/rbac.yaml
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list"]
```

**Prohibited operations**: create, update, delete, patch

### Security Test Results

#### gosec (Static Security Analysis)
- **Files scanned**: 21 files
- **Lines of code scanned**: 2,226 lines
- **Issues detected**: 0 ✅

#### govulncheck (Vulnerability Scan)
- **Go dependencies**: No vulnerabilities ✅
- **Go standard library**: 2 detected ⚠️
  - GO-2025-4175: Improper application of DNS constraints in crypto/x509
  - GO-2025-4155: Resource consumption during crypto/x509 certificate verification
  - **Mitigation**: Upgrade to Go 1.25.5 or later recommended
  - **Risk**: Low (connects only to trusted Prometheus URLs)

### .gitignore Configuration

**Excluding sensitive information**:
```gitignore
# Configuration files with potential secrets
config.yaml
config-*.yaml
kubeconfig
*.kubeconfig

# Security keys and certificates
*.key
*.pem
*.crt
*.p12
*.pfx

# Security scan reports
gosec-report*.json
trivy-report*.txt
```

### Recommended Security Practices

#### Production Use
1. **Go version**: Use 1.25.5 or later (crypto/x509 vulnerability mitigation)
2. **RBAC**: Use read-only ServiceAccount
3. **API keys**: Manage via environment variables, never include in configuration files
4. **TLS**: Enable TLS certificate verification for Prometheus connections (default)
5. **Logs**: Verify sensitive information is not output to logs in production

#### Development Environment
1. **Security scanning**: Run gosec before commits
2. **Dependency checking**: Run govulncheck regularly
3. **Test coverage**: Target 90%+ for security-related code

#### CI/CD Integration
```yaml
# GitHub Actions example
- name: Security Scan
  run: |
    gosec ./...
    govulncheck ./...
```

### Known Limitations

1. **minikube environment**: Insufficient container labels (addressed, using id=~".*/.*" filter)
2. **Go standard library vulnerabilities**: Resolved by upgrading to Go 1.25.5
3. **TLS disabling**: Possible via configuration but not recommended for production

### Security Documentation

- [SECURITY_TEST_PLAN.md](../../SECURITY_TEST_PLAN.md) - Comprehensive security checklist
- [TEST_RESULTS.md](../../TEST_RESULTS.md) - Detailed security scan results
- [E2E_TEST_PLAN.md](../../E2E_TEST_PLAN.md) - Security-related E2E tests

---

**Last Updated**: 2026-01-15
**Security Review**: Complete ✅
**OSS Publication Readiness**: Complete ✅
