# flea-market-api

3DスキャンとAIで商品の傷を検出・可視化するフリマアプリの**バックエンドAPI**。

**-> フロントエンドリポジトリ:** [flea-market-web](https://github.com/TaisukeFujise/flea-market-web) 

---

## 概要

商品状態の認識ズレによるトラブルを解消するため、AIが商品の傷・汚れを客観的に検出・評価するフリマプラットフォームのAPIサーバー。

## 技術スタック

| 領域 | 技術 |
| --- | --- |
| バックエンド | Go (Cloud Run) |
| データベース | Cloud SQL (PostgreSQL) + pgvector |
| ストレージ | Cloud Storage |
| 認証 | Firebase Authentication |
| AI: 傷検出・状態判定 | Gemini Vision API (structured output) |
| AI: フィードバックEmbedding | Vertex AI Multimodal Embedding |
| 3Dモデル生成 | Meshy API |
| リアルタイム通信 | WebSocket |
| インフラ管理 | Terraform |

## 主な機能

- ユーザー登録・認証（Firebase Authentication）
- 商品CRUD・購入フロー
- Gemini Visionによる傷検出（bbox）・商品状態の自動判定・説明文生成
- 購入者の傷報告 → Vertex AI EmbeddingでフィードバックをGeminiのfew-shotに反映
- Meshy APIを使った写真 → 3Dモデル（GLB）生成（3Dフェーズ）
- WebSocketを使ったDM・リアルタイム通知
- 商品検索（キーワード・カテゴリ・価格帯・写真検索）

## アーキテクチャ

```
React (Vercel)
  ↓ REST / WebSocket
Go API (Cloud Run)                    ← このリポジトリ
  ├─ Gemini Vision API（傷検出・状態判定）
  ├─ Vertex AI Multimodal Embedding（フィードバック）
  ├─ Meshy API（3Dモデル生成）
  ├─ Cloud Storage
  └─ Cloud SQL (PostgreSQL + pgvector)
```

アーキテクチャは Clean Architecture に基づいて設計。

---

## ローカル開発

### 前提条件

- [Docker](https://docs.docker.com/get-docker/) / Docker Compose
- [Go 1.25.5+](https://go.dev/dl/)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [gcloud CLI](https://cloud.google.com/sdk/docs/install)（Gemini / Vertex AI 利用時）

### セットアップ

```bash
# 1. 環境変数ファイルを作成
cp .env.example .env
# .env を編集して各値を設定

# 2. Google ADC（Gemini / Vertex AI 用）
gcloud auth application-default login

# 3. コンテナ起動（アプリ + PostgreSQL）
make up

# 4. マイグレーション実行（コンテナ起動後）
make migrate-local

# 5. 開発用サンプルデータ投入（任意）
make seed-local
```

### 主要 make コマンド

| コマンド | 説明 |
|---|---|
| `make up` | Docker Compose でアプリ + DB を起動 |
| `make rebuild` | キャッシュなしでビルドして起動 |
| `make down` | コンテナ停止 |
| `make down-v` | コンテナ停止 + volume 削除 |
| `make dev` | ローカルで直接サーバーを起動（要 DB 起動済み） |
| `make migrate-local` | ローカル DB にマイグレーション適用 |
| `make migrate-local-down` | ローカル DB のマイグレーションをロールバック |
| `make seed-local` | ローカル DB に開発用サンプルデータを投入 |

カテゴリマスターは本番でも必要な初期データのため、マイグレーションで管理します。`db/seeds` 配下はローカル開発用のユーザー・商品・注文データです。

### 環境変数

| 変数名 | 説明 |
|---|---|
| `DATABASE_URL` | PostgreSQL 接続 URL（例: `postgres://postgres:pass@localhost:5432/flea?sslmode=disable`） |

---

## デプロイ

`main` ブランチへの push で GitHub Actions が自動実行：

1. Cloud SQL Auth Proxy 経由でマイグレーション
2. Docker イメージをビルドして Artifact Registry に push
3. Cloud Run にデプロイ

GCP 認証は Workload Identity Federation を使用（サービスアカウントキー不要）。

### Terraform（インフラ初期構築）

```bash
cd terraform

# 設定ファイルを作成
cp terraform.tfvars.example terraform.tfvars
# terraform.tfvars を編集して frontend_origin を設定

terraform init
terraform apply
```

| 変数名 | 説明 |
|---|---|
| `frontend_origin` | フロントエンドの Vercel URL（本番 CORS 許可オリジン） |

`terraform.tfvars` は `.gitignore` に含まれているためコミットされません。
