# Getting Started

ローカルで pvvc を開発・実行するための環境構築手順です。

---

## Prerequisites

| ツール | バージョン | 用途 |
| --- | --- | --- |
| [Go](https://go.dev/dl/) | 1.25.4 以上 | ビルド・実行 |
| [golangci-lint](https://golangci-lint.run/welcome/install/) | v2.6.2 以上 | Linter（lefthook 経由で自動実行） |
| [lefthook](https://github.com/evilmartians/lefthook) | 最新 | Git フック管理 |

---

## Setup

### 1. リポジトリのクローン

```bash
git clone https://github.com/4okimi7uki/pvvc.git
cd pvvc
```

### 2. 依存パッケージのインストール

```bash
go mod download
```

### 3. lefthook のインストール

lefthook は Git フックを管理するツールです。コミット前に自動で Lint が走ります。

```bash
# Homebrew (macOS)
brew install lefthook

# Go でインストールする場合
go install github.com/evilmartians/lefthook@latest
```

### 4. Git フックの登録

```bash
lefthook install
```

これで `pre-commit` 時に `golangci-lint run ./...` が自動実行されます。  
設定内容は [`lefthook.yml`](../lefthook.yml) を参照してください。

---

## Build

```bash
# 現在の OS 向けにビルド（dist/pvvc に出力）
make build

# バージョンを指定してビルド
make build VERSION=v1.2.0

# 全プラットフォーム向けにビルド
make build-all VERSION=v1.2.0
```

---

## Run

ビルドせずにソースから直接実行することもできます。

```bash
go run . report
go run . analyze
```

---

## Configuration

開発時は `.env` ファイルをプロジェクトルートに置くと便利です。  
設定項目の詳細は [configuration.md](./configuration.md) を参照してください。

```bash
cp .env.example .env  # .env.example がある場合
# なければ手動で作成して設定値を記入
```

---

## Lint

```bash
golangci-lint run ./...
```

lefthook を導入していれば `git commit` のタイミングで自動実行されます。  
コミット前に手動で確認したい場合は上記コマンドを実行してください。
