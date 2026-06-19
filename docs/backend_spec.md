# バックエンド詳細仕様書

概要・アーキテクチャ・開発コマンドは README / CLAUDE.md を参照。このファイルにはコードから自明でない設計判断を記録する。

**関連ドキュメント**
- API 仕様：[api_spec.md](api_spec.md)
- DB 設計：[db_spec.md](db_spec.md)
- 傷検出仕様：[damage_detection_spec.md](damage_detection_spec.md)

---

## Echo ルートグループ構成

```go
api    := e.Group("/api")
public := api.Group("")        // 認証ミドルウェアなし
authed := api.Group("")
authed.Use(authMW.AuthRequired)
```

`GET /api/products/:id` のように「未認証でも使えるが認証状態で挙動が変わる」エンドポイントは `public` に置き、ハンドラー内で `c.Get("firebase_uid") == nil` を確認して分岐する。

---

## 認可方針（ミドルウェアではなくサービス層で行う理由）

「出品者のみ」「投稿者のみ」「参加者のみ」などのリソース単位の認可はサービス/ハンドラー層でチェックする。

- **責務の分離**：ミドルウェアは認証・ロギング・CORS など横断的関心事を担う場所。「このユーザーがこのリソースを操作できるか」はビジネスロジック。
- **二重クエリの回避**：ミドルウェアで認可チェックするにはリソースを DB から取得する必要があり、ハンドラーでも同じリソースを取得するため 2 回クエリが走る。サービス層でまとめれば 1 回で済む。

---

## 環境変数一覧

| 変数名 | 説明 |
|---|---|
| `DATABASE_URL` | PostgreSQL 接続文字列 |
| `FRONTEND_ORIGIN` | CORS 許可オリジン（1件のみ）。未設定時は CORS ミドルウェアを適用しない（全拒否）。本番は `https://loupe-market.vercel.app`、ローカルは `http://localhost:5173` を指定 |
| `FIREBASE_PROJECT_ID` | Firebase プロジェクト ID |
| `GOOGLE_CLOUD_PROJECT` | GCP プロジェクト ID（Vertex AI 用） |
| `VERTEX_AI_LOCATION` | Vertex AI リージョン（例: `us-central1`） |
| `MESHY_API_KEY` | Meshy API キー（Secret Manager 経由） |
| `MESHY_WEBHOOK_SECRET` | Meshy Webhook 署名検証キー（Secret Manager 経由） |
| `GCS_BUCKET_NAME` | Cloud Storage バケット ID |

Google Cloud 系サービス（Firebase / Vertex AI / Cloud Storage）は ADC で認証。ローカルは `gcloud auth application-default login`、Cloud Run はアタッチされたサービスアカウントの ADC を使用。

---

## 外部サービス連携

| サービス | Go パッケージ | タイムアウト | エラー時レスポンス | 備考 |
|---|---|---|---|---|
| Vertex AI Gemini（傷検出） | `google.golang.org/genai` | 60 秒（goroutine 内） | WebSocket `damage_detection_failed` | — |
| Vertex AI Multimodal Embedding（`gemini-embedding-2-preview`） | `google.golang.org/genai` | 15 秒 | 503 SERVICE_UNAVAILABLE | 出力次元 3072（固定）。傷報告時は bbox クロップ画像をインライン送信して `feedback_embeddings` に保存。傷検出時は全5方向の画像それぞれでクエリして pgvector 検索し、スコアマージ・重複排除した上位3件のクロップ画像 + damage_type + description を Gemini プロンプトに few-shot として追加 |
| Meshy API（初回リクエスト） | — | 10 秒 | 503 SERVICE_UNAVAILABLE | — |
| Meshy（生成完了） | — | Webhook 待ち（最大 5 分） | — | — |
| Firebase Auth | `firebase.google.com/go/v4` | — | — | — |
| Cloud Storage | `cloud.google.com/go/storage` | — | — | — |

Meshy フロー：`POST` → `job_id` 取得 → Webhook で完了通知 → GLB を Cloud Storage に保存 → WebSocket でフロントに通知。

---

## 使用パッケージ

| パッケージ | 用途 |
|---|---|
| `github.com/labstack/echo/v5` | HTTP フレームワーク |
| `database/sql` + `github.com/lib/pq` | DB 操作（PostgreSQL ドライバー） |
| `github.com/go-playground/validator/v10` | リクエストボディのバリデーション |
| `github.com/google/uuid` | UUID 生成 |
| `google.golang.org/genai` | Vertex AI Gemini クライアント（傷検出・Embedding） |
| `image` / `image/jpeg` / `image/png` | bbox クロップ・オーバーレイ描画（Go 標準ライブラリ） |

---

## セキュリティ固有の実装要件

- **画像アップロード**：MIME type 検証（画像のみ許可）・サイズ上限 10MB
- **Meshy Webhook**：署名検証（`MESHY_WEBHOOK_SECRET`）を必ず実施
- **アカウント削除**：DB の論理削除（`deleted_at` 更新）のみ。Firebase Auth 側は残すが、`AuthRequired` ミドルウェアの `deleted_at` チェックで以降のリクエストをすべて弾く
