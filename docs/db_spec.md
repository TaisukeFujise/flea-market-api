DB設計書
次世代フリマアプリ


# 1. テーブル一覧

| テーブル名 | カテゴリ | 役割 |
| --- | --- | --- |
| users | 認証・ユーザー | ユーザー情報管理（firebase_uidをPKとして使用） |
| products | 商品 | 出品商品の情報管理 |
| product_images | 商品 | 商品写真の管理 |
| product_models | 商品 | 3DモデルGLBの管理 |
| categories | 商品 | 商品カテゴリの管理 |
| damages | 傷検出 | AIが検出した傷の情報 |
| damage_reports | 傷検出 | 購入者が報告した傷の情報 |
| orders | 取引 | 購入・取引の管理 |
| message_rooms | コミュニケーション | DMルームの管理 |
| messages | コミュニケーション | メッセージの管理 |
| comments | コミュニケーション | 商品へのコメント・質問管理 |
| likes | その他 | いいねの管理 |
| viewing_history | その他 | 閲覧履歴の管理 |
| damage_detection_summaries | 傷検出 | 傷検出の結果（condition/condition_note）を一時保存 |
| feedback_embeddings | その他 | 傷報告のEmbeddingを保存・few-shot検索に使用 |




# 2. テーブル詳細

## 2-1. users

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | VARCHAR(255) | PRIMARY KEY | Firebase AuthenticationのユーザーID |
| display_name | VARCHAR(255) | NOT NULL | 表示名 |
| avatar_url | VARCHAR(500) | NULL許容 | プロフィール画像URL（Cloud Storage） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-2. products

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 商品ID |
| user_id | VARCHAR(255) | NOT NULL FK→users | 出品者ID |
| category_id | UUID | NOT NULL FK→categories | カテゴリID |
| title | VARCHAR(255) | NOT NULL | 商品タイトル |
| description | TEXT | NULL許容 | 商品説明 |
| price | INT | NOT NULL | 出品価格（円） |
| condition | ENUM | NOT NULL | 商品状態（good / fair / poor） |
| condition_note | TEXT | NULL許容 | AIが生成した状態サマリー文 |
| status | ENUM | NOT NULL | 出品状態（draft / on_sale / sold / deleted） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-3. product_images
angleカラムはRaycasterでのカメラ位置再現に使用

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 画像ID |
| product_id | UUID | NULL許容 FK→products | 商品ID（POST /api/images時はNULL・product作成時に紐づけ） |
| summary_id | UUID | NULL許容 FK→damage_detection_summaries | 傷検出サマリーID（傷検出完了時に紐づけ） |
| url | VARCHAR(500) | NOT NULL | 画像URL（Cloud Storage） |
| angle | ENUM | NULL許容 | 撮影角度（front / back / right / left / top） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-4. product_models

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | モデルID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| glb_url | VARCHAR(500) | NULL許容 | GLBファイルURL（Cloud Storage） |
| job_id | VARCHAR(255) | NULL許容 | MeshyのタスクID（Webhook用） |
| status | ENUM | NOT NULL | 生成状態（pending / processing / done / failed） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-5. categories

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | カテゴリID |
| parent_id | UUID | NULL許容 FK→categories | 親カテゴリID（NULLなら親カテゴリ） |
| name | VARCHAR(255) | NOT NULL | カテゴリ名 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |



## 2-6. damages
bbox座標はGemini Visionの検出結果（1024×1024正規化済み）。model座標はRaycasterで変換後の3D座標（3Dフェーズ・保留）

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 傷ID |
| image_id | UUID | NOT NULL FK→product_images | 検出元の写真ID |
| damage_type | ENUM | NOT NULL | 傷の種類（scratch / dirt / wear） |
| bbox_x1 | INT | NULL許容 | バウンディングボックス左上X座標 |
| bbox_y1 | INT | NULL許容 | バウンディングボックス左上Y座標 |
| bbox_x2 | INT | NULL許容 | バウンディングボックス右下X座標 |
| bbox_y2 | INT | NULL許容 | バウンディングボックス右下Y座標 |
| model_x | FLOAT | NULL許容 | GLBモデル上のX座標（3Dフェーズ・保留） |
| model_y | FLOAT | NULL許容 | GLBモデル上のY座標（3Dフェーズ・保留） |
| model_z | FLOAT | NULL許容 | GLBモデル上のZ座標（3Dフェーズ・保留） |
| description | TEXT | NULL許容 | 個別の傷の説明文 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-7. damage_reports
購入者が受け取り後に商品画像上（2Dフェーズ）または3Dモデル上（3Dフェーズ・保留）でタップして報告。フィードバックEmbeddingの元データ。返金・補償の根拠になる

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 報告ID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| user_id | VARCHAR(255) | NOT NULL FK→users | 報告者（購入者）ID |
| image_id | UUID | NULL許容 FK→product_images | 報告対象の画像ID（2Dフェーズ） |
| damage_type | ENUM | NOT NULL | 傷の種類（scratch / dirt / wear） |
| bbox_x1 | INT | NULL許容 | 報告範囲の左上X座標（2Dフェーズ） |
| bbox_y1 | INT | NULL許容 | 報告範囲の左上Y座標（2Dフェーズ） |
| bbox_x2 | INT | NULL許容 | 報告範囲の右下X座標（2Dフェーズ） |
| bbox_y2 | INT | NULL許容 | 報告範囲の右下Y座標（2Dフェーズ） |
| model_x | FLOAT | NULL許容 | 3Dモデル上のX座標（3Dフェーズ・保留） |
| model_y | FLOAT | NULL許容 | 3Dモデル上のY座標（3Dフェーズ・保留） |
| model_z | FLOAT | NULL許容 | 3Dモデル上のZ座標（3Dフェーズ・保留） |
| description | TEXT | NULL許容 | 報告内容の説明 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-8. orders
priceは購入時点のproducts.priceをコピーして保存。出品者が後から価格変更しても影響しない

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 注文ID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| buyer_id | VARCHAR(255) | NOT NULL FK→users | 購入者ID |
| price | INT | NOT NULL | 購入時の確定価格（円） |
| status | ENUM | NOT NULL | 取引状態（pending / completed / cancelled） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |



## 2-9. message_rooms
DMは購入後の取引連絡専用。購入前の質問はcommentsテーブルで管理

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | ルームID |
| order_id | UUID | NOT NULL FK→orders | 注文ID（購入後のみ作成） |
| buyer_id | VARCHAR(255) | NOT NULL FK→users | 購入者ID |
| seller_id | VARCHAR(255) | NOT NULL FK→users | 出品者ID |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-10. messages
既読管理はスコープ外

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | メッセージID |
| room_id | UUID | NOT NULL FK→message_rooms | ルームID |
| sender_id | VARCHAR(255) | NOT NULL FK→users | 送信者ID |
| content | TEXT | NOT NULL | メッセージ内容 |
| created_at | TIMESTAMP | NOT NULL | 送信日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-11. comments
商品詳細ページのQ&Aセクション。全ユーザーに公開される

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | コメントID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| user_id | VARCHAR(255) | NOT NULL FK→users | コメント投稿者ID |
| content | TEXT | NOT NULL | コメント内容 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |



## 2-12. likes

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | いいねID |
| user_id | VARCHAR(255) | NOT NULL FK→users | ユーザーID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| UNIQUE | (user_id, product_id) |  | 同じ商品への二重いいね防止 |



## 2-13. viewing_history

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 閲覧履歴ID |
| user_id | VARCHAR(255) | NOT NULL FK→users | ユーザーID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| viewed_at | TIMESTAMP | NOT NULL | 最終閲覧日時 |
| UNIQUE | (user_id, product_id) |  | 同じ商品の重複なし・viewed_atを更新 |



## 2-14. damage_detection_summaries
傷検出完了時にcondition/condition_noteを保存。product作成時にサーバーが参照してproductsに適用する

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | サマリーID |
| user_id | VARCHAR(255) | NOT NULL FK→users | アップロードしたユーザーID |
| condition | ENUM | NOT NULL | 商品状態（good / fair / poor） |
| condition_note | TEXT | NOT NULL | AIが生成した状態サマリー文 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |



## 2-15. feedback_embeddings
damage_reportsから生成したEmbeddingを保存。次回の傷検出でfew-shot参照に使用（pgvector）

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | EmbeddingID |
| damage_report_id | UUID | NOT NULL FK→damage_reports | 元の傷報告ID |
| category_id | UUID | NOT NULL FK→categories | 商品カテゴリID（類似検索のフィルタに使用） |
| embedding | vector(1408) | NOT NULL | Vertex AI Multimodal Embeddingのベクトル |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |




# 3. 設計方針
- ・テーブルIDはUUID（推測されにくい・分散環境に強い）、usersテーブルのみid（VARCHAR）にFirebase UIDを格納
- ・論理削除：deleted_atカラムで管理（取引履歴を残すため）
- ・金額：INT型（円単位で管理）
- ・3Dモデルのジョブ管理：product_modelsのstatusカラムで管理
- ・damages：Gemini VisionがbboxでJSON出力（bbox_x1/y1/x2/y2）。画像は1024×1024に正規化済み。model_x/y/zは3Dフェーズ（保留）
- ・damage_reports：2Dフェーズはbbox_x1/y1/x2/y2 + image_idで報告（フロントで外接bboxに変換して送信）。model_x/y/zは3Dフェーズ（保留）
- ・feedback_embeddings：pgvector（vector(1408)）でカテゴリ内類似検索。Geminiへのfew-shot参照に使用
- ・認証：Firebase Authentication（firebase_uidをusers PKとして使用）
