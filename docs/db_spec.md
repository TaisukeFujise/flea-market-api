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
| damage_reports | 購入者がフィードバック画面から報告。`feedback_embeddings` の元データになる。AIの精度向上目的のみで出品者には通知されない |
| ratings | 購入者が取引完了後に出品者を5段階評価。`order_id` に対して1レコード（重複不可）。`ratee_id` が被評価者（出品者）、`rater_id` が評価者（購入者） |
| orders | `price` は購入時点の `products.price` のスナップショット（後から出品者が値段変更しても影響しない） |
| message_rooms | 購入後の取引連絡専用。購入前の質問は `comments` テーブルで管理 |
| messages | 既読管理はスコープ外 |
| comments | 商品詳細の Q&A。全ユーザーに公開 |
| likes | `(user_id, product_id)` に UNIQUE 制約で二重いいね防止 |
| viewing_history | `(user_id, product_id)` に UNIQUE 制約。同じ商品を再閲覧したら `viewed_at` を UPDATE |
| damage_detection_summaries | 傷検出の進捗と結果を保持。`status` は `processing`（検出中）/ `done`（完了）/ `failed`（失敗）の3値。`condition` / `condition_note` は `done` 時のみ設定され、商品出品時にサーバーが参照して `products` に適用する |
| feedback_embeddings | `vector(3072)`（`gemini-embedding-2-preview` の出力次元数）。傷報告の bbox クロップ画像を埋め込んで保存。`cropped_image_url` にクロップ画像の GCS URL を持ち、検索時の再クロップを不要にする。`category_id` でフィルタした pgvector 類似検索で傷検出プロンプトの few-shot 参照例を取得する |

---

## 設計方針

- テーブル PK は UUID（`users` のみ Firebase UID の VARCHAR）
- 論理削除は `deleted_at TIMESTAMP` で管理（取引履歴を残すため）
- 金額は INT 型（円単位）
- 傷の bbox 座標は画像左上 (0,0)・右下 (1000,1000) とした正規化座標（整数）
- `products.status` は `on_sale` / `sold_out` の2値（`draft` / `sold` / `deleted` は廃止）。購入確定時に `sold_out` に更新、キャンセル時に `on_sale` に戻す
- `feedback_embeddings.embedding` は `vector(3072)`（`gemini-embedding-2-preview` の出力次元数）。`category_id` と組み合わせたカテゴリ内 pgvector 類似検索で Gemini への few-shot 参照に使用
- `ratings` はフィードバック送信時に作成。`order_id` に UNIQUE 制約で二重評価防止。ユーザーの平均評価スコアは `GET /api/me` および `GET /api/products/:id` の seller フィールドで返す
