# Loupe — 中古品の状態確認を変えるフリマアプリ

AIによる商品の傷検出と3Dモデル上への可視化を備えたフリマWebアプリのバックエンドAPIです。
写真が出品者ごとにバラつき比較しにくい問題と、主観的な状態評価によるトラブルを解消します。

**関連リポジトリ：** [フロントエンド](https://github.com/TaisukeFujise/flea-market-web) 

## デモ

<a href="https://youtu.be/xpiIk5NCf9U">
  <img src="https://img.youtube.com/vi/xpiIk5NCf9U/maxresdefault.jpg" width="600">
  <br>▶ デモを見る
</a>

---

## 主な機能

- ユーザー登録・認証（Firebase Authentication）
- アバター画像のアップロード
- 商品CRUD・購入フロー
- 画像アップロード（5方向: front/back/right/left/top）→ Gemini 2.5 Flashによる傷検出（bbox付き）・商品状態の自動判定・説明文生成（非同期）
- Meshy APIを使った写真 → 3Dモデル（GLB）生成（非同期・WebSocket通知）
- 傷の3Dモデル上座標マッピング
- 購入者による傷レポート（取引完了後）
- 売買双方による評価（取引完了後）
- コメント機能
- いいね機能
- 閲覧履歴
- WebSocketを使ったDMメッセージ・リアルタイム通知（傷検出完了・3Dモデル生成完了）
- 商品検索（キーワード・カテゴリ・価格帯・コンディション・並び順）

## 技術スタック

| 領域 | 技術 |
| --- | --- |
| バックエンド | Go (Cloud Run) |
| データベース | Cloud SQL (PostgreSQL) + pgvector |
| ストレージ | Cloud Storage |
| 認証 | Firebase Authentication |
| AI: 傷検出・状態判定 | Gemini 2.5 Flash via Vertex AI (structured output) |
| 3Dモデル生成 | Meshy API |
| リアルタイム通信 | WebSocket |
| インフラ管理 | Terraform |

## アーキテクチャ

```
React (Vercel)
  ↓ REST / WebSocket
Go API (Cloud Run)                    ← このリポジトリ
  ├─ Gemini 2.5 Flash via Vertex AI（傷検出・状態判定）
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
| `GOOGLE_CLOUD_PROJECT` | GCP プロジェクト ID（Vertex AI 用） |
| `VERTEX_AI_LOCATION` | Vertex AI リージョン（例: `us-central1`） |
| `GCS_BUCKET_NAME` | Cloud Storage バケット名（画像・3Dモデル保存先） |
| `MESHY_API_KEY` | Meshy API キー（3Dモデル生成用） |
| `FRONTEND_ORIGIN` | フロントエンドの URL（CORS・WebSocket Origin 許可用） |

`DATABASE_URL` のみローカル開発に必須。Vertex AI / Meshy を使う機能はそれぞれの変数が必要。

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
