# Kubernetes × FinOps × Generative AI OSS: Development Process & MVP Design (Detailed Version)

This document serves as an implementation-ready plan for developing an OSS focused on **optimizing Kubernetes cluster resource allocation (requests/limits)**, with emphasis on **actionable suggestions** rather than visualization (like OpenCost/Kubecost).

- Target product (working name): `k8s-finops-advisor`
- Format: CLI (future GitHub Action / GitOps PR integration)
- Core value: **Seamlessly delivers "Why wasteful?" → "How to fix? (concrete patches)" → "Expected impact"**

---

## 1. Purpose & Background

### 1.1 Purpose (Goal)
One of the most common challenges in Kubernetes operations is **excessive resource requests/limits configuration**.

- Excessive requests lead to reduced scheduling efficiency (node waste)
- Insufficient requests risk latency degradation, OOM, and restarts
- Determining "what" and "how much" to fix requires metrics + contextual understanding

This OSS collects information from metrics (e.g., Prometheus) and manifests (K8s API), then outputs **concrete improvement proposals (recommended values, applicable patches, explanations)**.

### 1.2 Non-goals (What MVP Won't Do)
To clarify MVP scope, the following are defined as "won't do":

- **Complete cost calculation / cost attribution** (OpenCost/Kubecost domain)
- Automatic application (automatic `kubectl apply` execution)
  - MVP stops at **proposal/patch generation**. Application is human decision
- Complete coverage of all optimization areas (node optimization, spot optimization, SLO optimization, etc.)
- Full support for complex workload types (start with Deployment first)

---

## 2. Differentiation (Competitive Landscape)

### 2.1 OpenCost / Kubecost
- Strengths: Cost visualization, allocation, reporting
- Weaknesses: **Specific "what to change and how" proposal generation** is not a core feature

### 2.2 Generative AI Debugging Tools (e.g., K8sGPT, etc.)
- Strengths: Troubleshooting support, state analysis
- Weaknesses: Integration of FinOps/resource optimization **quantitative recommendations** and **applicable patches** is limited

### 2.3 Position of This OSS
> Focused on **improvement proposals** rather than **visualization**,
> providing *recommended values (quantitative) + rationale (statistical) + patches (applicable) + explanations (natural language)*.

---

## 3. Target Users & Use Cases

### 3.1 Target Users
- SRE / Platform Engineers operating Kubernetes
- Team members managing cluster costs (FinOps)
- Teams practicing GitOps operations (Argo CD / Flux, etc.)

### 3.2 Primary Use Case (MVP: Focused Approach)
**Detect excessive/insufficient requests for Deployments in a namespace, and output recommended values + patches + report.**

---

## 4. Product Overview (MVP)

### 4.1 MVP One-Liner
**"CLI that detects excessive/insufficient requests/limits and outputs recommended values, rationale, and YAML patches"**

### 4.2 Inputs
- kubeconfig / in-cluster config (K8s API)
- Prometheus endpoint (or in-cluster)
- Target specification (namespace / label selector / deployment name)
- Aggregation period (e.g., 7d) and statistics (P50/P95/P99)
- Optional: Exclusion rules (system namespaces, job-type workloads)

### 4.3 Outputs
- `report.md`: Human-readable summary (top improvement proposals, rationale, notes)
- `patches/`: Candidate patches (per Deployment)
  - `patches/<namespace>/<deployment>.yaml`
- `summary.json`: Machine-readable (future GitHub Action / PR creation)

---

## 5. Specific MVP Specifications (Commands, Configuration, Algorithms)

### 5.1 CLI Command Design (Minimum 3 Commands)
MVP provides 3 commands, internally splitting the same execution pipeline:

1) `scan`: Target collection (manifest info from K8s)
2) `suggest`: Recommendation generation (statistics & rules)
3) `report`: Report generation (Markdown / JSON, optional AI summary)

Example:
```bash
k8s-finops scan --namespace prod
k8s-finops suggest --namespace prod --window 7d --cpu p95 --mem p95
k8s-finops report --namespace prod --llm on
```

A `run` command (scan→suggest→report) could be added later, but the MVP phase recommends easy-to-debug separation.

### 5.2 Configuration File (Example: `config.yaml`)
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

> **Note**: Model names are examples. Implementation should allow environment variable overrides for operational flexibility.

### 5.3 Metrics to Retrieve (Prometheus)
MVP focuses on CPU/Memory utilization:

- CPU usage (example): `rate(container_cpu_usage_seconds_total[5m])`
- Memory usage (example): `container_memory_working_set_bytes`

**Important**: Queries vary depending on node-exporter / cAdvisor / kube-state-metrics deployment.
MVP assumes "values can be retrieved from Prometheus", so docs must clearly state prerequisites.

### 5.4 Recommendation Calculation (Don't Leave to AI: Reproducibility and Safety)
Recommendations are determined by **statistics + rules**, with AI limited to explanation/prioritization.

#### 5.4.1 Basic Formula for Recommended Requests
- `recommended_cpu_request_m = max(Pxx_cpu_m * safetyFactor, minCpuMilli)`
- `recommended_mem_request_mi = max(Pxx_mem_mi * safetyFactor, minMemMi)`

Example (CPU):
- P95 = 120m, safetyFactor=1.2 → recommended = 144m (rounded to 150m)

#### 5.4.2 Judgment Rules (Example)
- **Overprovisioned**: `current_request > Pxx * 2.0`
- **Underprovisioned**: `current_request < Pxx * 1.1`
- **Risky limit** (optional): `limit >> request` (e.g., exceeds 10x)

#### 5.4.3 Expected Impact (Estimate)
MVP presents **reduction rate** (proportion of request reduction) rather than absolute cost:

- `saving_cpu_request_ratio = (current - recommended) / current` (when recommendation is lower)
- Same for memory

> Integration with OpenCost could enable cost conversion, but MVP prioritizes reduction rate and rationale.

### 5.5 Patch Generation (GitOps-oriented)
Output patches should minimally update Deployment `resources.requests/limits`:

- Recommended: **Strategic Merge Patch** (human-readable)
- Alternative: JSONPatch (mechanically safe but harder to read)

Example (Strategic Merge Patch):
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

## 6. Using Generative AI (Limited Use for MVP Value)

### 6.1 AI Role (MVP)
- **Prioritization** of proposals (by impact/risk)
- Human-readable **explanations** (why, how to fix, notes)
- Verbalizing risks (peak times, seasonality, batch workload mixing, etc.)

### 6.2 Input to AI (Structured Data is Key)
Passing structured **key points** to the LLM rather than full logs ensures stability:

- Current requests/limits
- P50/P95/P99 utilization (CPU/Memory)
- Recommended values (pre-calculated)
- Over/under-provisioning judgment
- Past events (OOMKilled, restart count) *if available
- Expected reduction rate (CPU/Memory)

### 6.3 Example: LLM Prompt Approach (Overview)
- System: Propose based on facts from FinOps + K8s SRE perspective
- User: Structured data above (JSON)
- Output: Markdown report section (concise, concrete)

> In implementation, safely insert LLM response into template rather than pasting raw output.

---

## 7. Development Process (OSS Sustainable Workflow)

### 7.1 Phase 0: Solidify Design (Prevent Drift)
- Clearly state "Purpose", "Non-goals", "MVP" at top of README
- Diagram main I/O (input→processing→output) (ASCII is fine)
- Make differentiation from competitors writable in 3 lines

Deliverables:
- Initial `README.md`
- `docs/design.md` (this document)
- `LICENSE`

### 7.2 Phase 1: Repository Skeleton (Ease of Contribution)
Minimally prepare the following to facilitate external participation:

- `README.md`
- `LICENSE`
- `CONTRIBUTING.md`
- `.github/ISSUE_TEMPLATE/*`
- `.github/pull_request_template.md`
- CI (lint / test / build)

Recommended directory structure (adjust by language):
```
.
├── cmd/                    # CLI entry (for Go)
├── internal/
│   ├── k8s/                # K8s API access
│   ├── metrics/            # Prometheus queries
│   ├── engine/             # Recommendation calculation
│   ├── patch/              # Patch generation
│   ├── report/             # md/json output
│   └── llm/                # LLM I/F (optional)
├── docs/
├── examples/
└── out/                    # Output (gitignore)
```

### 7.3 Phase 2: Complete Small MVP (Milestones)
Design to accumulate **"small completions"**.

#### Milestone 0: Skeleton
- CLI template (subcommands, config loading)
- Generate `report.md` with dummy data

#### Milestone 1: Data Acquisition
- Retrieve Deployment/Container requests/limits from K8s API
- Target only Deployments (for now)

#### Milestone 2: Suggestion Engine (Statistics & Rules)
- Retrieve CPU/Memory utilization from Prometheus
- Calculate recommendations based on statistics like P95
- Calculate judgment (over/under) and reduction rate

#### Milestone 3: Patch / Report Output
- Generate `patches/`
- Generate `report.md` from template
- Output `summary.json`

#### Milestone 4: LLM Narrative (Optional for MVP)
- LLM limited to explanation/notes generation
- Fallback to rule-based text on failure

#### Milestone 5: GitHub Action (MVP+)
- Generate report on cron and output artifact
- Future: PR creation (next phase)

---

## 8. Quality & Safety (Conditions for Trusted OSS)

### 8.1 Safety Design
- Default is `--dry-run`
- **No automatic application** (MVP)
- Always show rationale for recommendations (P95, coefficient, rounding)

### 8.2 Testing Approach
- Engine (recommendation calculation) requires unit tests
- Ensure reproducibility with fixtures for Prometheus/K8s responses
- Minimum: `lint + test + build` in GitHub Actions

### 8.3 Exception & Error Handling
- When Prometheus data unavailable:
  - Error message should present "required metrics" and "verification steps"
- K8s permission insufficient:
  - Document required RBAC (read-only) in docs

---

## 9. Initial Issue List (Initial Backlog Example)

### 9.1 Repo/CI
- [ ] Initial README (purpose, non-goals, quickstart, demo)
- [ ] Add license
- [ ] CI (lint/test/build)
- [ ] Issue/PR templates

### 9.2 Features
- [ ] `scan`: Retrieve Deployment/Container resources
- [ ] `metrics`: Implement Prometheus queries (CPU/Memory)
- [ ] `suggest`: Calculate recommendations (P95 + safety)
- [ ] `patch`: Generate Strategic Merge Patch
- [ ] `report`: Implement Markdown template
- [ ] `summary.json` generation (future extension)
- [ ] LLM narrative (optional, with fallback)

### 9.3 Docs
- [ ] Prerequisites (Prometheus/kube-state-metrics)
- [ ] RBAC example (read-only)
- [ ] FAQ (common failures: metrics unavailable, etc.)

---

## 10. Future Roadmap (Post-MVP)

### 10.1 Phase 2 (After MVP)
- OpenCost integration for cost conversion (savings estimate)
- GitHub Action (weekly reports)
- GitOps PR creation (automated PR for proposal patches)

### 10.2 Phase 3 (Advanced)
- HPA/VPA proposals
- Workload type expansion (StatefulSet/Job)
- Staged change application (safe rollout plan)
- Integration with SLO/error rates (consider reliability, not just cost)

---

## Appendix A: Report (report.md) Template Proposal

```md
# k8s-finops-advisor report

Target: namespace=prod / window=7d / percentile=P95 / safety=1.2

## Top findings
1. Deployment `api` container `app`
   - Current request(cpu)=500m / P95=120m → recommended=150m (-70%)
   - Rationale: Low utilization during normal operation, sufficient margin even at P95
   - Note: If monthly batch or other peaks exist, apply in stages

## Suggested patches
- patches/prod/api.yaml

## Method
- CPU: P95(rate(container_cpu_usage_seconds_total[5m])) × safety
- Memory: P95(container_memory_working_set_bytes) × safety
```

---

## Appendix B: MVP "Definition of Done"

- Deployable within 10 minutes (README Quickstart works)
- `scan → suggest → report` generates `report.md`
- At least 1 Deployment shows "recommendation with rationale"
- `patches/` generated in applicable format
- CI green (lint/test/build)
- `--dry-run` is default for safety

---

End.
