# 残タスク一覧

## 実装状況サマリー

| 層 | 状況 |
|---|---|
| インフラ（postgres / firebase） | 完了 |
| 認証ミドルウェア | 完了 |
| apperror | 完了 |
| #1 ユーザー CRUD | 完了 |
| #2 カテゴリ一覧 | 完了 |
| #3 画像アップロード（stub） | 完了 |
| #4 商品一覧・詳細 | 完了 |
| #5〜#12 | **未実装** |

---

## 実装フェーズ

```
フリマ本体フェーズ: #1 〜 #12
AIフェーズ:       #13 〜 #16（後回し）
```

AIフェーズでは #3・#5・#12 の stub を実装に差し替える。

---

## シードデータについて

API で作成できないマスターデータは `db/seeds/` に SQL ファイルとして管理する。

```
db/seeds/
  001_categories.sql   # カテゴリマスター（必須）
```

- `make seed-local` で流し込めるよう Makefile にターゲットを追加済み
- シードは `make migrate-local` の後に実行する

**シードデータが必要な Issue**
- **#2** カテゴリ一覧 API — `001_categories.sql` がないと空レスポンスになる
- **#5** 商品出品 API — `category_id` が外部キーのため、カテゴリが存在しないと出品できない

---

## フリマ本体フェーズ（#1〜#12）

### #1 ユーザー登録・プロフィールCRUD API

**エンドポイント**
- `POST /api/users/register`
- `GET /api/me`
- `PATCH /api/me`
- `DELETE /api/me`

**実装内容**
- `internal/repository/user_repository.go` — Create / FindByID / Update / SoftDelete
- `internal/service/user_service.go` — Register（重複チェック）/ GetMe / UpdateMe / DeleteMe
- `internal/handler/user_handler.go` — 各エンドポイントのハンドラ・レスポンス構造体
- `cmd/server/router.go` — 4ルート登録・DI完成

**備考**
- DELETE は DB 論理削除 + Cloud Scheduler による Firebase Auth の遅延削除
  - API 側: `deleted_at` を更新するのみ（即時応答）
  - Cloud Scheduler: 定期的に `deleted_at IS NOT NULL` のユーザーを Firebase Auth から削除
  - `terraform/` に Cloud Scheduler ジョブの定義を追加する
- Register は 409 CONFLICT を返す

---

### #2 カテゴリ一覧 API

**エンドポイント**
- `GET /api/categories`

**実装内容**
- `internal/repository/category_repository.go` — 全カテゴリ取得・親子ツリー構築
- `internal/service/category_service.go`
- `internal/handler/category_handler.go`
- `cmd/server/router.go` — ルート登録（認証不要）

**備考**
- 認証不要
- レスポンスは `children` を再帰的にネストしたツリー形式

---

### #3 画像アップロード API（GCSのみ・傷検出は stub）

**エンドポイント**
- `POST /api/images` (multipart/form-data)

**実装内容**
- GCS への画像アップロード
- `internal/repository/product_image_repository.go` — Create（5枚分）
- `internal/repository/damage_detection_summary_repository.go` — Create

**stub（AIフェーズ #13 で差し替える）**
- 傷検出AI呼び出しは行わない
- `damage_detection_summaries` にデフォルト値でレコードを INSERT する
  ```sql
  condition      = 'good'
  condition_note = ''
  ```
- `damages` テーブルへの INSERT は行わない
- `product_models` レコードは作成しない（#15 で追加）
- レスポンスの `"damage_detection"` は `"processing"` を返すが、完了通知（WebSocket）は送らない

**備考**
- ファイル形式: JPEG / PNG、10MB 以下
- Google ADC 必須（`gcloud auth application-default login`）

---

### #4 商品一覧・詳細 API

**エンドポイント**
- `GET /api/products`
- `GET /api/products/:id`

**実装内容**
- `internal/repository/product_repository.go` — List（フィルタ・ソート・ページネーション）/ FindByID
- `internal/repository/viewing_history_repository.go` — Upsert（閲覧履歴の自動記録）
- `internal/service/product_service.go` — ListProducts / GetProduct
- `internal/handler/product_handler.go` — 2ハンドラ・レスポンス構造体

**stub（AIフェーズ完了後に自動解消）**
- `damage_count` は常に `0`（damages テーブルが空のため）
- `model` フィールドは常に `null`（product_models レコードが存在しないため）
- #13・#15 が完了すれば追加実装なしで実データが返るようになる

**備考**
- 一覧のクエリパラメータ: `q`, `category_id`, `min_price`, `max_price`, `condition`, `sort`
- 詳細: 認証済みの場合のみ `liked` フラグを返す・閲覧履歴を記録

---

### #5 商品出品・編集・削除 API

**エンドポイント**
- `POST /api/products`
- `PATCH /api/products/:id`
- `DELETE /api/products/:id`

**実装内容**
- `internal/repository/product_repository.go` — Create / Update / SoftDelete
- `internal/repository/product_image_repository.go` — UpdateProductID（image_ids を product に紐付け）
- `internal/service/product_service.go` — CreateProduct / UpdateProduct / DeleteProduct
- `internal/handler/product_handler.go` — 3ハンドラ追加

**stub（AIフェーズ #13・#15 で差し替える）**
- 出品時に `damage_detection_summaries` から `condition` / `condition_note` を取得するが、#3 の stub が挿入したデフォルト値（`'good'` / `''`）が入る。動作は問題ない
- `product_models` レコードは作成しない（#15 で出品処理を修正して `pending` で作成する）

**備考**
- PATCH / DELETE は出品者本人のみ（403 FORBIDDEN）
- #3 の完了が前提（product_images・damage_detection_summaries が存在する）

---

### #6 コメント API

**エンドポイント**
- `GET /api/products/:id/comments`
- `POST /api/products/:id/comments`
- `DELETE /api/comments/:id`

**実装内容**
- `internal/repository/comment_repository.go` — ListByProductID / Create / SoftDelete
- `internal/service/comment_service.go`
- `internal/handler/comment_handler.go`

**備考**
- DELETE は投稿者本人のみ（403 FORBIDDEN）
- 一覧・取得は認証不要

---

### #7 いいね API

**エンドポイント**
- `POST /api/products/:id/likes`
- `DELETE /api/products/:id/likes`

**実装内容**
- `internal/repository/like_repository.go` — Create / Delete / ExistsByUserAndProduct
- `internal/service/like_service.go`
- `internal/handler/like_handler.go`

**備考**
- 重複いいねは 409 CONFLICT
- いいね解除対象がない場合は 404 NOT_FOUND

---

### #8 いいね一覧・閲覧履歴 API

**エンドポイント**
- `GET /api/me/likes`
- `GET /api/me/viewing-history`

**実装内容**
- `internal/repository/user_repository.go` — GetLikesByUserID / GetViewingHistoryByUserID（ページネーション付き）
- `internal/service/user_service.go` — GetMyLikes / GetMyViewingHistory
- `internal/handler/user_handler.go` — 2ハンドラ追加

**備考**
- 共通ページネーション（limit / offset）対応
- #4・#7 の完了が前提

---

### #9 注文 API

**エンドポイント**
- `POST /api/products/:id/orders`
- `GET /api/orders`
- `GET /api/orders/:id`
- `PATCH /api/orders/:id`

**実装内容**
- `internal/repository/order_repository.go` — Create / ListByUserID / FindByID / UpdateStatus
- `internal/repository/message_room_repository.go` — Create（注文と同時作成）
- `internal/service/order_service.go` — BuyProduct / ListOrders / GetOrder / UpdateOrderStatus
- `internal/handler/order_handler.go`

**備考**
- 購入時に message_rooms を同時作成（トランザクション）
- 自分の出品商品は購入不可（403）
- ステータス遷移: `pending → completed`（buyer）/ `pending → cancelled`（buyer or seller）
- `GET /api/orders` は `role` クエリパラメータ（buyer / seller）対応

---

### #10 傷報告 API

**エンドポイント**
- `POST /api/orders/:id/damage-reports`

**実装内容**
- `internal/repository/damage_report_repository.go` — Create
- `internal/service/damage_report_service.go` — 権限チェック（buyer かつ order.status = completed）
- `internal/handler/damage_report_handler.go`

**備考**
- buyer_id 一致 かつ orders.status = `completed` のみ可
- feedback_embeddings への Vertex AI Embedding 保存は将来フェーズ（スコープ外）
- #9 の完了が前提

---

### #11 メッセージ API

**エンドポイント**
- `GET /api/message-rooms/:id/messages`
- `POST /api/message-rooms/:id/messages`

**実装内容**
- `internal/repository/message_repository.go` — ListByRoomID / Create
- `internal/service/message_service.go` — 参加者チェック（buyer or seller）
- `internal/handler/message_handler.go`

**備考**
- 参加者以外は 403 FORBIDDEN
- POST 成功後に WebSocket で相手に通知（#12 と連動）
- #9 の完了が前提（message_rooms が存在する）

---

### #12 WebSocket（メッセージのリアルタイム通知のみ）

**エンドポイント**
- `WS /ws?token=<Firebase ID Token>`

**実装内容**
- `internal/handler/ws_handler.go` — 接続管理・Hub パターン・ブロードキャスト
- クライアント接続時に Firebase トークン検証
- `new_message` イベントの送信（#11 と連動）

**stub（AIフェーズ #16 で追加する）**
- `damage_detection_complete` イベントは未実装（#16 で追加）
- `model_generation_complete` イベントは未実装（#16 で追加）

**備考**
- gorilla/websocket など WebSocket ライブラリの導入が必要
- #11 の完了が前提

---

## AIフェーズ（#13〜#16）

> **前提**: フリマ本体フェーズ（#1〜#12）の完了後に着手する。

---

### #13 傷検出AI連携

**概要**
#3 の stub（ダミー summary INSERT）を本実装に差し替える。

**実装内容**
- Gemini / Vertex AI Multimodal を呼び出して傷検出
- 検出結果を `damages` テーブルに INSERT
- `damage_detection_summaries` の `condition` / `condition_note` を実データで更新
- 非同期処理（goroutine）で実行し、完了後に #16 の通知を呼ぶ

**差し替え箇所**
- `internal/repository/damage_detection_summary_repository.go` の Create — ダミー値を削除し、AI結果で更新
- `internal/service/` に傷検出サービスを追加

**備考**
- Google ADC 必須（`gcloud auth application-default login`）

---

### #14 傷情報 API

**エンドポイント**
- `GET /api/products/:id/damages`
- `PATCH /api/damages/:id`

**実装内容**
- `internal/repository/damage_repository.go` — ListByProductID / UpdateCoordinates
- `internal/service/damage_service.go`
- `internal/handler/damage_handler.go`

**備考**
- PATCH は 3D フェーズ用（Raycaster で算出した model_x / y / z を保存）
- 一覧は認証不要
- #13 の完了が前提（damages テーブルにデータが存在する）

---

### #15 3Dモデル生成AI連携

**概要**
Meshy などの外部サービスで3Dモデルを非同期生成する。

**実装内容**
- `internal/repository/product_model_repository.go` — Create / UpdateStatus / UpdateGlbURL
- 外部サービス（Meshy 等）への生成リクエスト送信
- ポーリング or Webhook でステータス更新（`pending → processing → done / failed`）
- 完了後に #16 の通知を呼ぶ

**差し替え箇所**
- `internal/service/product_service.go` の CreateProduct — 出品時に `product_models` を `pending` で作成する処理を追加（#5 の修正）

**備考**
- `product_models.job_id` に外部サービスのジョブIDを保存する
- `GET /api/products/:id` の `model` フィールドは追加実装なしで自動的に実データを返すようになる（#4 が product_models を JOIN 済みのため）

---

### #16 WebSocket AI通知

**概要**
#12 の stub に AI完了イベントを追加する。

**実装内容**
- `damage_detection_complete` イベント送信（#13 完了時に呼ぶ）
- `model_generation_complete` イベント送信（#15 完了時に呼ぶ）
- `internal/handler/ws_handler.go` にイベント種別を追加

**備考**
- #12 の Hub 実装が前提
- #13・#15 の完了が前提

---

## 依存関係

```
フリマ本体フェーズ
─────────────────
#1 User CRUD
  └─ #8 いいね一覧・閲覧履歴

#2 Category（独立）

#3 画像アップロード（stub）
  └─ #5 商品出品
       └─ #9 注文
            └─ #10 傷報告
            └─ #11 メッセージ
                 └─ #12 WebSocket（new_messageのみ）

#4 商品一覧・詳細
  └─ #5
  └─ #6 コメント
  └─ #7 いいね
  └─ #8

AIフェーズ
──────────
#13 傷検出AI（#3 stub差し替え）
  └─ #14 傷情報API

#15 3Dモデル生成AI（#5 修正）

#13 + #15
  └─ #16 WebSocket AI通知（#12 stub追加）
```
