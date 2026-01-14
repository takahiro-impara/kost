# セキュリティテスト計画

## 概要
OSS公開前のセキュリティチェックリスト。全項目を検証し、脆弱性がないことを確認する。

## 1. 依存関係の脆弱性スキャン

### 1.1 govulncheck（Go公式ツール）
```bash
# インストール
go install golang.org/x/vuln/cmd/govulncheck@latest

# 実行
govulncheck ./...
```

**期待結果**: 既知の脆弱性が検出されないこと

### 1.2 Trivy（コンテナ/ファイルシステムスキャン）
```bash
# インストール（macOS）
brew install aquasecurity/trivy/trivy

# ファイルシステムスキャン
trivy fs .

# 深刻度HIGH以上のみ表示
trivy fs --severity HIGH,CRITICAL .
```

**期待結果**: HIGH/CRITICAL脆弱性が検出されないこと

## 2. 静的セキュリティ解析

### 2.1 gosec（静的解析ツール）
```bash
# インストール
go install github.com/securego/gosec/v2/cmd/gosec@latest

# 実行
gosec -fmt=json -out=gosec-report.json ./...
gosec ./...
```

**チェック項目**:
- SQL injection
- Command injection
- Path traversal
- Hardcoded credentials
- Weak crypto
- Race conditions
- Integer overflow

**期待結果**: セキュリティ上の問題が検出されないこと

### 2.2 golangci-lint（包括的linter）
```bash
# 実行
golangci-lint run --enable-all ./...
```

## 3. 入力検証のセキュリティテスト

### 3.1 Namespace検証
- ✅ 空文字列の拒否
- ✅ 大文字の拒否
- ✅ 特殊文字の拒否
- ✅ 長さ制限（63文字）
- ✅ Kubernetes命名規則準拠

**テストファイル**: `internal/security/validation_test.go`

### 3.2 Label Selector検証
- ✅ インジェクション攻撃の防止（引用符、セミコロン）
- ✅ 空文字列の許可
- ✅ 正しいセレクタ構文の許可

### 3.3 Prometheus URL検証
- ✅ http/httpsスキームのみ許可
- ✅ 特殊文字の拒否
- ✅ 空文字列の拒否

### 3.4 PromQL Query検証
- ✅ SQLインジェクション的なパターンの拒否（セミコロン、コメント）
- ✅ サニタイゼーション実施

### 3.5 出力パス検証
- ✅ パストラバーサル攻撃の防止（`..`）
- ✅ センシティブディレクトリへのアクセス防止（/etc, /root, /sys, /proc, /dev）

**カバレッジ**: 97.7%（internal/security）

## 4. API認証情報の管理

### 4.1 環境変数のみ使用
- ✅ `OPENAI_API_KEY`（環境変数）
- ✅ `ANTHROPIC_API_KEY`（環境変数）
- ✅ kubeconfigファイルまたはin-cluster認証

### 4.2 設定ファイルに含めない
```yaml
# ❌ 絶対にやってはいけない
llm:
  apiKey: "sk-xxxxx"  # これは含めない！

# ✅ 正しい方法
llm:
  enabled: true
  provider: "openai"
  # API keyは環境変数から読み取る
```

### 4.3 ログ出力時のマスキング
- ✅ `MaskSensitiveValue()` 関数実装済み
- ✅ API keyを`sk-1...cdef`形式でマスク
- ✅ エラーメッセージに認証情報を含めない

**検証方法**:
```bash
# ソースコード全体をgrep
grep -r "apiKey" --include="*.go" .
grep -r "OPENAI_API_KEY" --include="*.go" .
grep -r "ANTHROPIC_API_KEY" --include="*.go" .
```

### 4.4 .gitignoreの確認
```
# 以下が含まれていることを確認
*.env
.env.*
config.yaml  # 個人設定を含む可能性
*.key
*.pem
kubeconfig
```

## 5. RBAC権限の最小化

### 5.1 必要な権限のみ
```yaml
# examples/rbac.yaml
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list"]  # read-onlyのみ
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["get", "list"]
```

**禁止事項**:
- ❌ `create`, `update`, `delete`, `patch`権限
- ❌ cluster-admin権限
- ❌ secrets, configmapsへのアクセス

### 5.2 RBAC設定のテスト
```bash
# examples/rbac.yamlを適用してテスト
kubectl apply -f examples/rbac.yaml
kubectl auth can-i list deployments --as=system:serviceaccount:default:kost-sa
kubectl auth can-i delete deployments --as=system:serviceaccount:default:kost-sa  # 拒否されるべき
```

## 6. TLS証明書検証

### 6.1 デフォルトで有効
```go
// internal/metrics/client.go
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: false,  // デフォルト
        },
    },
}
```

### 6.2 設定での無効化（非推奨）
```yaml
prometheus:
  url: "https://prometheus.example.com"
  insecureSkipVerify: true  # 本番環境では非推奨
```

**警告**: ドキュメントに非推奨であることを明記

## 7. エラーメッセージの安全性

### 7.1 センシティブ情報を含めない
```go
// ❌ 悪い例
return fmt.Errorf("failed to connect to %s with token %s", url, token)

// ✅ 良い例
return fmt.Errorf("failed to connect to prometheus")
```

### 7.2 スタックトレースの制御
- 本番環境ではスタックトレースを表示しない
- デバッグモードでのみ詳細情報を表示

## 8. コードレビューチェックリスト

### 8.1 コマンド実行
- [ ] `exec.Command()`使用箇所の確認
- [ ] ユーザー入力をコマンドに渡していないか

### 8.2 ファイル操作
- [ ] ファイルパスのサニタイゼーション
- [ ] パーミッション設定（0600推奨）
- [ ] シンボリックリンクのチェック

### 8.3 ネットワーク通信
- [ ] タイムアウト設定
- [ ] TLS証明書検証
- [ ] リダイレクト回数制限

## 9. CI/CDパイプラインへの統合

### 9.1 GitHub Actions設定例
```yaml
name: Security Scan
on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Run gosec
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec ./...

      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          severity: 'HIGH,CRITICAL'
```

## 10. ペネトレーションテストシナリオ

### 10.1 インジェクション攻撃
```bash
# Namespace injection
./bin/kost scan -n "prod; rm -rf /"
./bin/kost scan -n "../../../etc/passwd"

# Label selector injection
./bin/kost scan -n prod --label-selector "app=nginx';DROP TABLE--"

# 期待結果: すべて拒否されること
```

### 10.2 パストラバーサル
```bash
# config.yamlで試行
output:
  dir: "../../etc"  # 拒否されるべき
  dir: "/root/test"  # 拒否されるべき
```

### 10.3 大量リクエスト（DoS耐性）
```bash
# タイムアウト設定の確認
prometheus:
  timeoutSeconds: 30  # 適切なタイムアウト

# コンテキストキャンセルの確認
ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
defer cancel()
```

## チェックリスト完了基準

- [ ] govulncheck: 脆弱性0件
- [ ] Trivy: HIGH/CRITICAL脆弱性0件
- [ ] gosec: セキュリティ問題0件
- [ ] 入力検証テスト: 全項目パス
- [ ] API認証情報: 環境変数のみ使用確認
- [ ] RBAC: read-only権限のみ確認
- [ ] ログ: センシティブ情報マスキング確認
- [ ] ペネトレーションテスト: 全項目で攻撃防御成功
- [ ] ドキュメント: セキュリティガイドライン記載

## 参考資料

- OWASP Top 10: https://owasp.org/www-project-top-ten/
- CWE Top 25: https://cwe.mitre.org/top25/
- Go Security Best Practices: https://go.dev/doc/security/
- Kubernetes Security Best Practices: https://kubernetes.io/docs/concepts/security/
