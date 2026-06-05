# 残タスク一覧

## 実装状況サマリー

| 層 | 状況 |
|---|---|
| インフラ（postgres / firebase） | 完了 |
| 認証ミドルウェア | 完了 |
| apperror | 完了 |
| handler / service / repository | **すべてスタブ（中身なし）** |
| ルーティング登録（router.go） | **未登録** |

---

## Issue 一覧

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
- DELETE は DB 論理削除のみ（Firebase Auth は残す）
- Register は 409 CONFLICT を返す

---

### #2 いいね一覧・閲覧履歴 API

**エンドポイント**
- `GET /api/me/likes`
- `GET /api/me/viewing-history`

**実装内容**
- `internal/repository/user_repository.go` — GetLikesByUserID / GetViewingHistoryByUserID（ページネーション付き）
- `internal/service/user_service.go` — GetMyLikes / GetMyViewingHistory
- `internal/handler/user_handler.go` — 2ハンドラ追加

**備考**
- 共通ページネーション（limit / offset）対応

---

### #3 カテゴリ一覧 API

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

### #4 商品一覧・詳細 API

**エンドポイント**
- `GET /api/products`
- `GET /api/products/:id`

**実装内容**
- `internal/repository/product_repository.go` — List（フィルタ・ソート・ページネーション）/ FindByID
- `internal/repository/viewing_history_repository.go` — Upsert（閲覧履歴の自動記録）
- `internal/service/product_service.go` — ListProducts / GetProduct
- `internal/handler/product_handler.go` — 2ハンドラ・レスポンス構造体

**備考**
- 一覧のクエリパラメータ: `q`, `category_id`, `min_price`, `max_price`, `condition`, `sort`
- 詳細: 認証済みの場合のみ `liked` フラグを返す・閲覧履歴を記録
- `model` フィールドは product_models を JOIN して返す

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

**備考**
- 出品時: product INSERT → product_images の product_id 更新 → damage_detection_summaries から condition / condition_note を取得して products に反映
- PATCH / DELETE は出品者本人のみ（403 FORBIDDEN）

---

### #6 画像アップロード・傷検出 API

**エンドポイント**
- `POST /api/images` (multipart/form-data)

**実装内容**
- GCS への画像アップロード
- `internal/repository/product_image_repository.go` — Create（5枚分）
- `internal/repository/damage_detection_summary_repository.go` — Create
- Vertex AI Multimodal / Gemini を呼び出して非同期で傷検出
- 検出結果を `damages` テーブルに INSERT、`damage_detection_summaries` を更新
- WebSocket でフロントに完了通知（Issue #13 と連動）

**備考**
- ファイル形式: JPEG / PNG、10MB 以下
- 非同期処理（ goroutine + チャネル or キュー）
- Google ADC 必須（`gcloud auth application-default login`）

---

### #7 傷情報 API

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

---

### #8 コメント API

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

### #9 いいね API

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

### #10 注文 API

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

### #11 傷報告 API

**エンドポイント**
- `POST /api/orders/:id/damage-reports`

**実装内容**
- `internal/repository/damage_report_repository.go` — Create
- `internal/service/damage_report_service.go` — 権限チェック（buyer かつ order.status = completed）
- `internal/handler/damage_report_handler.go`

**備考**
- buyer_id 一致 かつ orders.status = `completed` のみ可
- feedback_embeddings への Vertex AI Embedding 保存は将来フェーズ（スコープ外でも可）

---

### #12 メッセージ API

**エンドポイント**
- `GET /api/message-rooms/:id/messages`
- `POST /api/message-rooms/:id/messages`

**実装内容**
- `internal/repository/message_repository.go` — ListByRoomID / Create
- `internal/service/message_service.go` — 参加者チェック（buyer or seller）
- `internal/handler/message_handler.go`

**備考**
- 参加者以外は 403 FORBIDDEN
- POST 成功後に WebSocket で相手に通知（Issue #13 と連動）

---

### #13 WebSocket リアルタイム通信

**エンドポイント**
- `WS /ws?token=<Firebase ID Token>`

**実装内容**
- `internal/handler/ws_handler.go` — 接続管理・ルーム管理・ブロードキャスト
- クライアント接続時に Firebase トークン検証
- 送信イベント:
  - `new_message` — メッセージ送信時（Issue #12 と連動）
  - `damage_detection_complete` — 傷検出完了時（Issue #6 と連動）
  - `model_generation_complete` — 3Dモデル生成完了時（3D フェーズ）

**備考**
- gorilla/websocket など WebSocket ライブラリの導入が必要
- Hub パターン（接続を中央管理）推奨

---

## 依存関係

```
#1 (User CRUD)
  └─ #2 (いいね・閲覧履歴)
  └─ #10 (Order)
       └─ #11 (傷報告)
       └─ #12 (Message)
            └─ #13 (WebSocket)

#6 (画像アップロード)
  └─ #5 (商品出品) — damage_detection_summaries を参照
  └─ #13 (WebSocket) — 検出完了通知

#4 (商品一覧・詳細)
  └─ #7 (傷情報)
  └─ #8 (コメント)
  └─ #9 (いいね)

#3 (Category) — 独立
```

## 推奨実装順序

1. **#1** User CRUD（最も他に依存されるため最初）
2. **#3** Category（独立・シンプル）
3. **#6** 画像アップロード・傷検出（出品の前提）
4. **#4 → #5** 商品一覧/詳細 → 商品出品/編集/削除
5. **#7 #8 #9 #2** 傷・コメント・いいね・閲覧履歴（並行可）
6. **#10 → #11** 注文 → 傷報告
7. **#12 → #13** メッセージ → WebSocket
