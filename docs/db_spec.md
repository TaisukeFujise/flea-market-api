# DB設計書

スキーマの正は `migrations/000001_init_schema.up.sql`。このファイルには SQL に書けない設計意図を記録する。

---

## テーブルの役割と設計上の注意点

| テーブル | 役割・注意点 |
|---|---|
| users | Firebase UID を VARCHAR(255) の PK として使用 |
| products | 出品商品。`condition` / `condition_note` は傷検出完了後に `damage_detection_summaries` から適用 |
| product_images | `product_id` は `POST /api/images` 時点では NULL。`POST /api/products` で紐づける |
| product_models | Meshy の job_id を保存し Webhook で status 更新。`product_id` に対して1レコード |
| categories | 親子2階層。`parent_id` が NULL なら親カテゴリ |
| damages | Gemini Vision の検出結果（1024×1024 正規化済み bbox）。`model_x/y/z` は 3D フェーズまで NULL |
| damage_reports | 購入者が受け取り後に報告。`feedback_embeddings` の元データになる |
| orders | `price` は購入時点の `products.price` のスナップショット（後から出品者が値段変更しても影響しない） |
| message_rooms | 購入後の取引連絡専用。購入前の質問は `comments` テーブルで管理 |
| messages | 既読管理はスコープ外 |
| comments | 商品詳細の Q&A。全ユーザーに公開 |
| likes | `(user_id, product_id)` に UNIQUE 制約で二重いいね防止 |
| viewing_history | `(user_id, product_id)` に UNIQUE 制約。同じ商品を再閲覧したら `viewed_at` を UPDATE |
| damage_detection_summaries | 傷検出完了時に一時保存。商品出品時にサーバーが参照して `products` に適用後は参照のみ |
| feedback_embeddings | `vector(1408)`（Vertex AI Multimodal Embedding の次元数）。`category_id` でフィルタして pgvector 類似検索 |

---

## 設計方針

- テーブル PK は UUID（`users` のみ Firebase UID の VARCHAR）
- 論理削除は `deleted_at TIMESTAMP` で管理（取引履歴を残すため）
- 金額は INT 型（円単位）
- 傷の bbox 座標は 1024×1024 正規化済み画像の絶対ピクセル値
- `feedback_embeddings.embedding` は `vector(1408)`。`category_id` と組み合わせたカテゴリ内類似検索で Gemini への few-shot 参照に使用
