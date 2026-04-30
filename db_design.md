# DB設計書

次世代フリマアプリ

---

## 1. テーブル一覧

| テーブル名 | カテゴリ | 役割 |
| --- | --- | --- |
| users | 認証・ユーザー | ユーザー情報管理 |
| products | 商品 | 出品商品の情報管理 |
| product_images | 商品 | 商品写真の管理 |
| product_models | 商品 | 3DモデルGLBの管理 |
| categories | 商品 | 商品カテゴリの管理 |
| damages | 傷検出 | AIが検出した傷の情報 |
| damage_reports | 傷検出 | 購入者が報告した傷の情報 |
| orders | 取引 | 購入・取引の管理 |
| message_rooms | コミュニケーション | DMルームの管理 |
| messages | コミュニケーション | メッセージの管理 |
| likes | その他 | いいねの管理 |
| viewing_history | その他 | 閲覧履歴の管理 |

---

## 2. テーブル詳細

### 2-1. users

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | ユーザーID |
| google_id | VARCHAR(255) | UNIQUE NOT NULL | Google OAuthのユーザーID |
| email | VARCHAR(255) | UNIQUE NOT NULL | メールアドレス |
| display_name | VARCHAR(255) | NOT NULL | 表示名 |
| avatar_url | VARCHAR(500) | NULL許容 | プロフィール画像URL（Cloud Storage） |
| role | ENUM | DEFAULT 'user' | ロール（user / admin） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |

### 2-2. products

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 商品ID |
| user_id | UUID | NOT NULL FK→users | 出品者ID |
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

### 2-3. product_images

> angleカラムはRaycasterでのカメラ位置再現に使用

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 画像ID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| url | VARCHAR(500) | NOT NULL | 画像URL（Cloud Storage） |
| angle | ENUM | NULL許容 | 撮影角度（front / back / right / left / top） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |

### 2-4. product_models

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | モデルID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| glb_url | VARCHAR(500) | NULL許容 | GLBファイルURL（Cloud Storage） |
| job_id | VARCHAR(255) | NULL許容 | MeshyのタスクID（Webhook用） |
| status | ENUM | NOT NULL | 生成状態（pending / processing / done / failed） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### 2-5. categories

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | カテゴリID |
| parent_id | UUID | NULL許容 FK→categories | 親カテゴリID（NULLなら親カテゴリ） |
| name | VARCHAR(255) | NOT NULL | カテゴリ名 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### 2-6. damages

> pixel座標はYOLO/Gemini Visionの検出結果。model座標はRaycaster/Geminiで変換後の3D座標。両方NULL許容で実装フェーズで判断

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 傷ID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| image_id | UUID | NULL許容 FK→product_images | 検出元の写真ID |
| damage_type | ENUM | NOT NULL | 傷の種類（scratch / dirt / wear） |
| pixel_x | INT | NULL許容 | 元写真上のピクセルX座標（2D） |
| pixel_y | INT | NULL許容 | 元写真上のピクセルY座標（2D） |
| model_x | FLOAT | NULL許容 | GLBモデル上のX座標（3D） |
| model_y | FLOAT | NULL許容 | GLBモデル上のY座標（3D） |
| model_z | FLOAT | NULL許容 | GLBモデル上のZ座標（3D） |
| description | TEXT | NULL許容 | 個別の傷の説明文 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |

### 2-7. damage_reports

> 購入者が受け取り後に3Dモデル上でタップして報告。YOLOの再学習データとして活用。返金・補償の根拠になる

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 報告ID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| user_id | UUID | NOT NULL FK→users | 報告者（購入者）ID |
| damage_type | ENUM | NOT NULL | 傷の種類（scratch / dirt / wear） |
| model_x | FLOAT | NULL許容 | 3Dモデル上のX座標（購入者がタップした箇所） |
| model_y | FLOAT | NULL許容 | 3Dモデル上のY座標 |
| model_z | FLOAT | NULL許容 | 3Dモデル上のZ座標 |
| description | TEXT | NULL許容 | 報告内容の説明 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |

### 2-8. orders

> priceは購入時点のproducts.priceをコピーして保存。出品者が後から価格変更しても影響しない

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 注文ID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| buyer_id | UUID | NOT NULL FK→users | 購入者ID |
| price | INT | NOT NULL | 購入時の確定価格（円） |
| status | ENUM | NOT NULL | 取引状態（pending / completed / cancelled） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### 2-9. message_rooms

> 購入前（order_id=NULL）と購入後（order_id有）の両方に対応

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | ルームID |
| order_id | UUID | NULL許容 FK→orders | 注文ID（購入前はNULL） |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| buyer_id | UUID | NOT NULL FK→users | 購入者（質問者）ID |
| seller_id | UUID | NOT NULL FK→users | 出品者ID |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |
| UNIQUE | (product_id, buyer_id) | | 同じユーザーが同じ商品に複数ルームを作れない |

### 2-10. messages

> 既読管理はスコープ外

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | メッセージID |
| room_id | UUID | NOT NULL FK→message_rooms | ルームID |
| sender_id | UUID | NOT NULL FK→users | 送信者ID |
| content | TEXT | NOT NULL | メッセージ内容 |
| created_at | TIMESTAMP | NOT NULL | 送信日時 |
| deleted_at | TIMESTAMP | NULL許容 | 論理削除日時 |

### 2-11. likes

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | いいねID |
| user_id | UUID | NOT NULL FK→users | ユーザーID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| UNIQUE | (user_id, product_id) | | 同じ商品への二重いいね防止 |

### 2-12. viewing_history

| カラム名 | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| id | UUID | PRIMARY KEY | 閲覧履歴ID |
| user_id | UUID | NOT NULL FK→users | ユーザーID |
| product_id | UUID | NOT NULL FK→products | 商品ID |
| viewed_at | TIMESTAMP | NOT NULL | 最終閲覧日時 |
| UNIQUE | (user_id, product_id) | | 同じ商品の重複なし・viewed_atを更新 |

---

## 3. 設計方針

- 全テーブルのIDはUUID（推測されにくい・分散環境に強い）
- 論理削除：deleted_atカラムで管理（取引履歴を残すため）
- 金額：INT型（円単位で管理）
- 3Dモデルのジョブ管理：product_modelsのstatusカラムで管理
- damages/damage_reportsの座標：pixel座標・3D座標ともNULL許容（実装フェーズで判断）
- 認証：Google OAuth + JWT（アクセストークンのみ・有効期限7日）
