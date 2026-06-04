# バックエンド詳細仕様書

## 1. 概要

Go製のRESTful APIサーバー。Firebase Authenticationで認証し、Gemini Vision APIによる傷検出・Vertex AI Embeddingによるフィードバックループ・Meshy APIによる3Dモデル生成を組み合わせたフリマプラットフォームのバックエンド。

関連ドキュメント：
- DB設計：`de_spec.md`
- 傷検出仕様：`damage_detection_spec.md`
- API仕様：`api_spec.md`

---

## 2. 環境構成

2環境構成。

| 環境 | 用途 | インフラ |
|---|---|---|
| local | ローカル開発 | Docker Compose（PostgreSQL） |
| production | 本番 | Cloud Run + Cloud SQL + Cloud Storage |

stagingは設けない。本番確認はローカルで完結させる。

---

## 3. 開発・デプロイフロー

### ローカル開発

ローカル開発はMakefileで管理。

```bash
make up       # PostgreSQL起動（docker compose up -d）
make migrate  # マイグレーション実行
make dev      # サーバー起動
make down     # PostgreSQL停止
```

Google Cloud系サービス（Firebase Admin SDK / Vertex AI / Cloud Storage）の認証はADC（Application Default Credentials）で行う。事前に一度だけ実行しておく。

```bash
gcloud auth application-default login
```

### インフラ管理（Terraform）

全インフラ（Cloud Run / Cloud SQL / Cloud Storage）をTerraformで管理。

```bash
cd terraform/
terraform init
terraform plan
terraform apply
```

### CI/CD（Cloud Build + Workload Identity Federation）

`main` ブランチへのpushで自動ビルド・Cloud Runへのデプロイが走る。GitHubとCloud BuildはWorkload Identity Federationで連携し、サービスアカウントキーを使用しない。

---

## 4. アーキテクチャ

### フレームワーク：Echo v5

`github.com/labstack/echo/v5` を使用。

- `/api` プレフィックスのルートグループ化が簡単
- 認証ミドルウェアをグループ単位で適用できる
- CORS・Logger・Recoverが標準ミドルウェアとして揃っている
- WebSocketサポートあり

```go
e := echo.New()

api := e.Group("/api")

// 認証不要グループ
api.GET("/products", handler.ListProducts)

// 認証必須グループ
auth := api.Group("")
auth.Use(middleware.AuthRequired)
auth.GET("/me", handler.GetMe)
```

### レイヤー構成

```
internal/
├── apperror/            # アプリケーションエラー定義（ErrCode, AppError）
├── domain/              # エンティティ定義（Product, User, Damage...）
├── handler/             # HTTPハンドラー（リクエスト/レスポンス変換・ErrorHandler）
├── infra/
│   ├── fbapp/           # Firebase Admin SDK初期化
│   └── postgres/        # PostgreSQL接続（database/sql）
├── middleware/          # 認証等
├── repository/          # リポジトリ実装（database/sql + PostgreSQL）
└── service/             # ビジネスロジック（インターフェース定義含む）
```

### ディレクトリ構造

```
.
├── cmd/
│   └── server/          # サーバーエントリーポイント（main.go, router.go）
├── internal/            # アプリケーションコア（上記参照）
├── migrations/          # SQLマイグレーションファイル
├── terraform/           # インフラ定義
├── docker-compose.yaml
└── Makefile
```

---

## 5. 認証・認可方針

### Firebase IDトークン検証フロー

```
1. クライアントがFirebase AuthでログインしてIDトークンを取得
2. リクエストヘッダーに Bearer トークンを付与
   Authorization: Bearer <Firebase ID Token>
3. Goミドルウェアでトークンを検証（Firebase Admin SDK）
4. DBのusers.deleted_atを確認し、削除済みユーザーは401を返す
5. 検証済みのfirebase_uid（users.id）をContextに格納
6. ハンドラーでContextからユーザーIDを取得
```

### Echoグループ構成

グループは2つ。`GET /api/products/:id` のように未ログインでも使えるがログイン状態によって挙動が変わるエンドポイントは `authed` に入れず、ハンドラー内で `c.Get("firebase_uid")` が nil かどうかを確認して分岐する。

```go
api    := e.Group("/api")
public := api.Group("")   // ミドルウェアなし
authed := api.Group("")   // AuthRequired
authed.Use(authMW.AuthRequired)
```

### AuthRequired ミドルウェア

```
1. AuthorizationヘッダーからBearerトークン取得 → なければ401
2. Firebase Admin SDKでトークン検証 → 無効なら401
3. DBでusers.deleted_atチェック → 削除済みなら401
4. Context("firebase_uid")にfirebase_uidをセット
```

### リソース単位の認可はサービス層で行う

「出品者のみ」「投稿者のみ」「参加者のみ」などの認可はミドルウェアではなくサービス/ハンドラー層でチェックする。

理由は2つ：

**責務の分離**：ミドルウェアは認証・ロギング・CORSなど横断的関心事を担う場所。「このユーザーがこのリソースを操作できるか」はビジネスロジックなのでサービス層が適切。ミドルウェアに入れると専用ミドルウェアが増えて管理しにくくなる。

**二重クエリの回避**：ミドルウェアで認可チェックするにはリソースをDBから取得する必要があり、ハンドラーでも同じリソースを取得するため2回クエリが走る。サービス層でまとめて行えば1回で済む。

```
ミドルウェア → 「誰か」の確認（認証）← 全エンドポイント共通
サービス → 「その人がこのリソースに触れるか」← エンドポイントごとのビジネスロジック
```

### アカウント削除方針

DBの論理削除（deleted_at更新）のみ実施。Firebase Authのアカウントは残すが、AuthRequiredミドルウェアのステップ3のdeleted_atチェックで以降のリクエストをすべて弾く。Firebase Auth側の削除は優先度低のため省略。

### 認証要否

| エンドポイント | 認証 | 認可（追加条件） |
|---|---|---|
| GET /api/categories | 不要 | - |
| GET /api/products | 不要 | - |
| GET /api/products/:id | 不要 | - |
| GET /api/products/:id/damages | 不要 | - |
| GET /api/products/:id/comments | 不要 | - |
| POST /api/users/register | 必要 | - |
| GET /api/me | 必要 | - |
| PATCH /api/me | 必要 | - |
| DELETE /api/me | 必要 | - |
| GET /api/me/likes | 必要 | - |
| GET /api/me/viewing-history | 必要 | - |
| POST /api/products | 必要 | - |
| PATCH /api/products/:id | 必要 | 出品者のみ |
| DELETE /api/products/:id | 必要 | 出品者のみ |
| POST /api/images | 必要 | - |
| PATCH /api/damages/:id | 必要 | - |
| POST /api/products/:id/orders | 必要 | - |
| GET /api/orders | 必要 | - |
| GET /api/orders/:id | 必要 | 関係者のみ（buyer or seller） |
| PATCH /api/orders/:id | 必要 | 関係者のみ・操作内容による |
| POST /api/orders/:id/damage-reports | 必要 | buyer_id一致 かつ status=completed |
| POST /api/products/:id/comments | 必要 | - |
| DELETE /api/comments/:id | 必要 | 投稿者のみ |
| POST /api/products/:id/likes | 必要 | - |
| DELETE /api/products/:id/likes | 必要 | - |
| GET /api/message-rooms | 必要 | - |
| GET /api/message-rooms/:id/messages | 必要 | 参加者のみ |
| POST /api/message-rooms/:id/messages | 必要 | 参加者のみ |
| WS /ws | 必要 | - |

---

## 6. エラーレスポンス定義

### 6-1. フォーマット

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "human readable message"
  }
}
```

### 6-2. HTTPステータスとエラーコード対応表

| HTTP | code | 説明 |
|---|---|---|
| 400 | BAD_REQUEST | リクエストパラメータ不正 |
| 400 | VALIDATION_ERROR | バリデーションエラー |
| 401 | UNAUTHORIZED | 認証トークンなし・無効・有効期限切れ |
| 403 | FORBIDDEN | 認証済みだが権限なし（他ユーザーのリソースへのアクセス等） |
| 404 | NOT_FOUND | リソースが存在しない |
| 409 | CONFLICT | 重複登録（同一ユーザーID・いいね済み等） |
| 429 | TOO_MANY_REQUESTS | レートリミット超過（外部AIエンドポイント等） |
| 500 | INTERNAL_SERVER_ERROR | サーバー内部エラー |
| 503 | SERVICE_UNAVAILABLE | 外部API（Gemini / Vertex AI / Meshy等）の一時障害 |

---

## 7. セキュリティ

| 項目 | 方針 |
|---|---|
| CORS | フロントエンドのオリジン（Vercel URL）のみ許可 |
| 入力バリデーション | ハンドラー層でリクエストボディを検証。不正な値は400を返す |
| SQLインジェクション | GORMのプレースホルダーを使用。生SQLは書かない |
| ファイルアップロード | 画像のみ許可（MIME type検証）・サイズ上限10MB |
| Webhookの検証 | Meshy Webhookの署名検証を行う |
| シークレット管理 | 外部APIキー・DB認証情報はSecret Managerで管理しCloud Run環境変数にマウント。Google Cloud系サービスはADCで認証。コードにハードコードしない |

---

## 8. 外部サービス連携仕様

### Gemini Vision API
- 用途：傷検出・商品状態判定・説明文生成
- 詳細：`damage_detection_spec.md` 参照
- タイムアウト：30秒
- エラー時：503 SERVICE_UNAVAILABLE を返す

### Vertex AI Multimodal Embedding
- 用途：傷報告画像のEmbedding生成（フィードバックfew-shot）
- モデル：multimodalembedding@001（1408次元）
- 詳細：`damage_detection_spec.md` 参照
- タイムアウト：15秒

### Meshy API
- 用途：商品画像5枚 → GLB形式3Dモデル生成
- フロー：リクエスト → job_id取得 → Webhookで完了通知 → GLBをCloud Storageに保存
- Webhook受信後にWebSocketでフロントに完了通知
- タイムアウト：初回リクエストは10秒。生成はWebhook待ち（最大5分）

### Firebase Admin SDK
- 用途：IDトークン検証・ユーザー情報取得
- 初期化：ADCで自動認証（ローカルは`gcloud auth application-default login`、Cloud RunはサービスアカウントのADCを使用）

---

## 9. 環境変数一覧

| 変数名 | 説明 | 例 |
|---|---|---|
| PORT | サーバーポート | 8080 |
| ENV | 実行環境 | local / production |
| DATABASE_URL | PostgreSQL接続文字列 | postgres://user:pass@host:5432/dbname?sslmode=disable |
| FIREBASE_PROJECT_ID | FirebaseプロジェクトID | my-project |
| GEMINI_API_KEY | Gemini APIキー | Secret Manager経由 |
| VERTEX_AI_PROJECT_ID | Vertex AIプロジェクトID | my-project |
| VERTEX_AI_LOCATION | Vertex AIリージョン | us-central1 |
| MESHY_API_KEY | Meshy APIキー | Secret Manager経由 |
| MESHY_WEBHOOK_SECRET | Meshy Webhook署名検証キー | Secret Manager経由 |
| GCS_BUCKET_NAME | Cloud StorageバケットID | flea-market-assets |
| FRONTEND_ORIGIN | CORS許可オリジン | https://flea-market.vercel.app |

**認証情報の管理方針：**
- Google Cloud系サービス（Firebase / Vertex AI / Cloud Storage）：ADCで自動解決。ローカルは`gcloud auth application-default login`、Cloud RunはアタッチされたサービスアカウントのADCを使用
- 外部APIキー・DB認証情報（Gemini / Meshy / DB Password）：Secret Managerに保存し、Cloud Runの環境変数としてマウント。コードは環境変数を読むだけ

---

## 10. 使用外部パッケージ

### フレームワーク

| パッケージ | 用途 |
|---|---|
| `github.com/labstack/echo/v5` | HTTPフレームワーク |
| `github.com/gorilla/websocket` | WebSocket（Echo経由） |

### DB

| パッケージ | 用途 |
|---|---|
| `database/sql`（標準ライブラリ） | DB操作 |
| `github.com/lib/pq` | PostgreSQLドライバー |

### 認証

| パッケージ | 用途 |
|---|---|
| `firebase.google.com/go/v4` | Firebase Admin SDK（IDトークン検証） |

### Google Cloud / AI

| パッケージ | 用途 |
|---|---|
| `github.com/google/generative-ai-go/genai` | Gemini Vision API（傷検出・状態判定） |
| `cloud.google.com/go/aiplatform/apiv1beta1` | Vertex AI Multimodal Embedding |
| `cloud.google.com/go/storage` | Cloud Storage（画像・GLBファイルのアップロード） |

### ユーティリティ

| パッケージ | 用途 |
|---|---|
| `github.com/google/uuid` | UUID生成 |
| `github.com/go-playground/validator/v10` | リクエストボディのバリデーション |
| `github.com/disintegration/imaging` | 画像リサイズ（1024×1024正規化） |
| `github.com/joho/godotenv` | ローカル開発用の.envファイル読み込み |
