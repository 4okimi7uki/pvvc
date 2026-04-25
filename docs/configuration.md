# Configuration Reference

設定値の読み込み優先度（高い順）:

1. 環境変数
2. プロジェクトルートの `.env` ファイル
3. `~/.config/pvvc/config.toml`

`pvvc init` を実行すると `config.toml` に対話形式で書き込めます。

---

## 設定項目一覧

### Vercel

| 環境変数 | config.toml キー | 必須 | 説明 |
| --- | --- | :---: | --- |
| `VERCEL_TOKEN` | `vercel.token` | ✅ | Vercel のアクセストークン |
| `TEAM_ID` | `vercel.team_id` | ⚠️ | チーム ID。未設定の場合は個人アカウントのみ対象 |
| `PROJECT_ID` | `vercel.project_id` | ⚠️ | プロジェクト ID。未設定だとコストが $0 になる |

### GA4

| 環境変数 | config.toml キー | 必須 | 説明 |
| --- | --- | :---: | --- |
| `PROPERTY_ID` | `ga4.property_id` | ✅ | GA4 プロパティ ID（数値文字列） |
| `GOOGLE_ANALYTICS_CREDENTIAL` | `ga4.credential` | ✅ | サービスアカウントの JSON を1行に圧縮した文字列 |

### AI

| 環境変数 | config.toml キー | 必須 | 説明 |
| --- | --- | :---: | --- |
| `GEMINI_API_KEY` | `ai.gemini_key` | - | Gemini API キー。未設定の場合は AI 分析をスキップ |

### Slack

| 環境変数 | config.toml キー | 必須 | 説明 |
| --- | --- | :---: | --- |
| `SLACK_WEBHOOK_URL` | `slack.webhook_url` | - | Incoming Webhook URL。`--notify` 使用時のみ必要 |

### Service

| 環境変数 | config.toml キー | 必須 | 説明 |
| --- | --- | :---: | --- |
| `TARGET_WEBSITE_NAME` | `service.name` | - | サービス名。レポートや Slack メッセージに表示される |

---

## config.toml の例

```toml
[vercel]
token = "your_vercel_token"
team_id = "team_xxxxxxxx"
project_id = "prj_xxxxxxxx"

[ga4]
property_id = "123456789"
credential = '{"type":"service_account","project_id":"..."}' # サービスアカウント JSON

[ai]
gemini_key = "AIza..."

[slack]
webhook_url = "https://hooks.slack.com/services/..."

[service]
name = "Your Site Name"
```

---

## .env の例

```env
VERCEL_TOKEN=your_vercel_token
TEAM_ID=team_xxxxxxxx
PROJECT_ID=prj_xxxxxxxx

PROPERTY_ID=123456789
GOOGLE_ANALYTICS_CREDENTIAL={"type":"service_account",...}

GEMINI_API_KEY=AIza...

SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...

TARGET_WEBSITE_NAME=Your Site Name
```

---

## GA4 サービスアカウントの取得

1. [Google Cloud Console](https://console.cloud.google.com/) でプロジェクトを選択
2. 「IAM と管理」→「サービスアカウント」でアカウントを作成
3. 「Google Analytics Data API」を有効化
4. サービスアカウントに **閲覧者** ロールを付与
5. GA4 管理画面でサービスアカウントのメールアドレスをプロパティに追加（「管理」→「プロパティのアクセス管理」）
6. キーを JSON 形式でダウンロードし、ファイル全体を1行に圧縮して `GOOGLE_ANALYTICS_CREDENTIAL` に設定

```bash
# JSON ファイルを1行に圧縮する例
cat service-account.json | tr -d '\n'
```

> 詳細は [Google Analytics Data API ドキュメント](https://developers.google.com/analytics/devguides/reporting/data/v1/quickstart-client-libraries) を参照してください。

---

## Vercel トークンの取得

1. Vercel ダッシュボード → 「Settings」→「Tokens」
2. 「Create Token」でトークンを発行

> スコープの設定については各自で最新の [Vercel ドキュメント](https://vercel.com/docs/rest-api/authentication) を確認してください。Billing API へのアクセスに必要なスコープは Vercel のプランや仕様変更によって異なる場合があります。

**Team ID / Project ID の確認方法:**

- Team ID: チームの「Settings」→「General」→ `Team ID`
- Project ID: プロジェクトの「Settings」→「General」→ `Project ID`
