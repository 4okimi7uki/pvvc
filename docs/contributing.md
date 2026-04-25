# Contributing

---

## Branch Naming

| プレフィックス | 用途 |
| --- | --- |
| `feat/` | 新機能 |
| `fix/` | バグ修正 |
| `chore/` | 設定・依存関係・CI など機能に影響しない変更 |
| `docs/` | ドキュメントのみの変更 |
| `refactor/` | 動作を変えないリファクタリング |

例: `feat/custom-prompt`, `fix/rate-limit-handling`

---

## Commit Message

[Conventional Commits](https://www.conventionalcommits.org/) に準じて**英語**で書きます。

```
<type>: <subject>

[optional body]
```

**type の例:**

| type | 用途 |
| --- | --- |
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `chore` | ビルド・CI・設定変更 |
| `docs` | ドキュメント |
| `refactor` | リファクタリング |
| `test` | テスト追加・修正 |

**例:**

```
feat: add --prompt flag to support custom template path

Allow users to pass a local file path or HTTPS URL via --prompt/-p.
Falls back to prompts/analyze.tmpl when not specified.
```

---

## Pull Request

- `main` ブランチへ直接プッシュせず、必ずブランチを切って PR を作成してください。
- PR タイトルもコミットメッセージと同様の形式（英語・Conventional Commits）で書きます。
- CI（Lint）が通ることを確認してからマージしてください。

---

## Lint

PR 作成前にローカルで確認することを推奨します。

```bash
golangci-lint run ./...
```

lefthook を導入していれば `git commit` 時に自動で実行されます。  
セットアップは [getting-started.md](./getting-started.md) を参照してください。

---

## Release

リリースは **セマンティックバージョニング** に従います。

| 変更の種類 | バージョン |
| --- | --- |
| 後方互換な新機能・フラグの追加 | `minor` (`v1.x.0`) |
| バグ修正・内部リファクタリング | `patch` (`v1.0.x`) |
| 後方互換のない変更 | `major` (`vx.0.0`) |

### リリース手順

1. `main` ブランチで最終確認
2. バージョンタグを打つ

```bash
git tag v1.2.0
git push origin v1.2.0
```

3. GitHub Actions の [Release ワークフロー](../.github/workflows/release.yml) が自動で起動し、全プラットフォーム向けバイナリをビルドして GitHub Release を作成します。

> タグは `v` から始まる形式（`v1.2.0`）にしてください。ワークフローのトリガーが `v*` パターンになっています。
