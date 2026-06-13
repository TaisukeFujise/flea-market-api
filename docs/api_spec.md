# API仕様書

## エンドポイント一覧

| メソッド | パス | 認証 | 概要 |
|---|---|---|---|
| POST | /api/users/register | 必要 | ユーザー登録 |
| GET | /api/me | 必要 | 自分のプロフィール取得 |
| PATCH | /api/me | 必要 | 自分のプロフィール更新 |
| DELETE | /api/me | 必要 | アカウント削除 |
| POST | /api/me/avatar | 必要 | アバター画像アップロード |
| GET | /api/me/likes | 必要 | いいね一覧 |
| GET | /api/me/viewing-history | 必要 | 閲覧履歴 |
| GET | /api/categories | 不要 | カテゴリ一覧 |
| GET | /api/products | 不要 | 商品一覧 |
| POST | /api/images | 必要 | 画像アップロード・傷検出開始 |
| POST | /api/products | 必要 | 商品出品 |
| GET | /api/products/:id | 不要 | 商品詳細 |
| PATCH | /api/products/:id | 必要（出品者のみ） | 商品情報更新 |
| DELETE | /api/products/:id | 必要（出品者のみ） | 商品削除 |
| GET | /api/products/:id/damages | 不要 | 傷情報一覧 |
| GET | /api/products/:id/comments | 不要 | コメント一覧 |
| POST | /api/products/:id/comments | 必要 | コメント投稿 |
| POST | /api/products/:id/likes | 必要 | いいね |
| DELETE | /api/products/:id/likes | 必要 | いいね解除 |
| POST | /api/products/:id/orders | 必要（自分の商品は不可） | 商品購入 |
| PATCH | /api/damages/:id | 必要 | 傷3D座標更新（3Dフェーズ） |
| DELETE | /api/comments/:id | 必要（投稿者のみ） | コメント削除 |
| GET | /api/orders | 必要 | 注文一覧 |
| GET | /api/orders/:id | 必要（関係者のみ） | 注文詳細 |
| PATCH | /api/orders/:id | 必要（関係者のみ・操作内容による） | 注文ステータス更新 |
| POST | /api/orders/:id/feedback | 必要（購入者のみ・completed状態） | フィードバック送信（評価） |
| POST | /api/orders/:id/damage-reports | 必要（購入者のみ・completed状態） | 傷報告 |
| GET | /api/message-rooms/:id/messages | 必要（参加者のみ） | メッセージ一覧 |
| POST | /api/message-rooms/:id/messages | 必要（参加者のみ） | メッセージ送信 |
| WS | /ws | 必要 | リアルタイム通信 |

---

## 共通仕様

### ベースURL
```
本番: https://api.flea-market.example.com
ローカル: http://localhost:8080
```

全エンドポイントに `/api` プレフィックスを付与。

### 認証

Firebase IDトークンをAuthorizationヘッダーに付与。

```
Authorization: Bearer <Firebase ID Token>
```

認証が必要なエンドポイントで未付与または無効なトークンの場合は `401 UNAUTHORIZED` を返す。

### ページネーション

一覧系エンドポイントは共通のクエリパラメータでページネーション。

| パラメータ | 型 | デフォルト | 説明 |
|---|---|---|---|
| limit | int | 20 | 取得件数（最大100） |
| offset | int | 0 | 取得開始位置 |

レスポンス形式：
```json
{
  "items": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

### Content-Type

| 種別 | Content-Type |
|---|---|
| JSONリクエスト | `application/json` |
| ファイルアップロード | `multipart/form-data` |
| JSONレスポンス | `application/json` |

### エラーレスポンス

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "human readable message"
  }
}
```

---

## Users

### POST /api/users/register
ユーザー登録（Firebase Auth後の初回のみ）

**認証：必要**

**Request Body**
```json
{
  "display_name": "山田太郎",
  "avatar_url": "https://storage.example.com/avatars/xxx.jpg"
}
```

**Response 201** No Content

**Errors**
| Status | code | 条件 |
|---|---|---|
| 409 | CONFLICT | 同一ユーザーが既に登録済み |

---

### GET /api/me
自分のプロフィール取得

**認証：必要**

**Response 200**
```json
{
  "id": "firebase_uid",
  "display_name": "山田太郎",
  "avatar_url": "https://storage.example.com/avatars/xxx.jpg",
  "created_at": "2026-06-01T00:00:00Z",
  "rating_avg": 4.5,
  "rating_count": 12,
  "updated_at": "2026-06-01T00:00:00Z"
}
```

---

### PATCH /api/me
自分のプロフィール更新

**認証：必要**

**Request Body**（更新するフィールドのみ）
```json
{
  "display_name": "山田太郎",
  "avatar_url": "https://storage.example.com/avatars/xxx.jpg"
}
```

**Response 204** No Content

---

### DELETE /api/me
アカウント削除

**認証：必要**

DBの論理削除（deleted_at更新）のみ実施。Firebase Authのアカウントは残るが、以降のリクエストは認証ミドルウェアのdeleted_atチェックで弾く。

**Response 204** No Content

---

### POST /api/me/avatar
アバター画像をアップロードし、自分の `avatar_url` を更新する。

**認証：必要**

**Request** `multipart/form-data`

| フィールド | 型 | 説明 |
|---|---|---|
| avatar | file | 画像ファイル（JPEG/PNG・10MB以下） |

**Response 204** No Content

**Errors**
| Status | code | 条件 |
|---|---|---|
| 400 | VALIDATION_ERROR | ファイル形式不正・サイズ超過 |

---

### GET /api/me/likes
自分のいいね一覧

**認証：必要**

**Response 200**
```json
{
  "items": [
    {
      "product": {
        "id": "uuid",
        "title": "Sony WH-1000XM5",
        "price": 25000,
        "thumbnail_url": "https://...",
        "status": "on_sale"
      },
      "created_at": "2026-06-01T00:00:00Z"
    }
  ],
  "total": 10,
  "limit": 20,
  "offset": 0
}
```

---

### GET /api/me/viewing-history
自分の閲覧履歴

**認証：必要**

**Response 200**
```json
{
  "items": [
    {
      "product": {
        "id": "uuid",
        "title": "Sony WH-1000XM5",
        "price": 25000,
        "thumbnail_url": "https://...",
        "status": "on_sale"
      },
      "viewed_at": "2026-06-01T00:00:00Z"
    }
  ],
  "total": 30,
  "limit": 20,
  "offset": 0
}
```

---

## Categories

### GET /api/categories
カテゴリ一覧取得

**認証：不要**

**Response 200**
```json
{
  "categories": [
    {
      "id": "uuid",
      "parent_id": null,
      "name": "家電",
      "children": [
        {
          "id": "uuid",
          "parent_id": "uuid",
          "name": "スマートフォン",
          "children": []
        }
      ]
    }
  ]
}
```

---

## Products

### GET /api/products
商品一覧取得

**認証：不要**

**Query Parameters**
| パラメータ | 型 | 説明 |
|---|---|---|
| q | string | キーワード検索 |
| category_id | uuid | カテゴリフィルター |
| min_price | int | 最低価格（円） |
| max_price | int | 最高価格（円） |
| condition | string | 商品状態（good / fair / poor） |
| sort | string | ソート（created_at_desc / price_asc / price_desc）デフォルト: created_at_desc |
| limit | int | - |
| offset | int | - |

**Response 200**
```json
{
  "items": [
    {
      "id": "uuid",
      "category_id": "uuid",
      "title": "Sony WH-1000XM5",
      "price": 25000,
      "condition": "fair",
      "status": "on_sale",
      "thumbnail_url": "https://storage.example.com/images/xxx.jpg",
      "model": {
        "status": "done",
        "glb_url": "https://storage.example.com/models/xxx.glb"
      },
      "created_at": "2026-06-01T00:00:00Z"
    }
  ],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

---

### POST /api/products
商品出品。サーバー側の処理：
1. products INSERT
2. UPDATE product_images SET product_id WHERE id IN (image_ids)
3. product_images[0].summary_id → damage_detection_summaries を参照してcondition/condition_noteをproductsに適用

**認証：必要**

**Request Body**
```json
{
  "image_ids": ["uuid", "uuid", "uuid", "uuid", "uuid"],
  "category_id": "uuid",
  "title": "Sony WH-1000XM5",
  "description": "1年使用しました。",
  "price": 25000
}
```

**Response 201**
```json
{
  "id": "uuid",
  "status": "on_sale",
  "created_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 400 | VALIDATION_ERROR | 必須項目不足・価格が0以下等 |
| 404 | NOT_FOUND | image_idが存在しない |

---

### GET /api/products/:id
商品詳細取得

認証済みユーザーの場合は閲覧履歴を自動記録。

**認証：不要**

**Response 200**
```json
{
  "id": "uuid",
  "seller": {
    "id": "firebase_uid",
    "display_name": "山田太郎",
    "avatar_url": "https://...",
    "rating_avg": 4.5,
    "rating_count": 12
  },
  "category_id": "uuid",
  "title": "Sony WH-1000XM5",
  "description": "1年使用しました。",
  "price": 25000,
  "condition": "fair",
  "condition_note": "全体的に使用感があり、右側面に目立つ傷があります。",
  "status": "on_sale",
  "damage_count": 3,
  "images": [
    {
      "id": "uuid",
      "url": "https://storage.example.com/images/xxx.jpg",
      "angle": "front"
    }
  ],
  "model": {
    "status": "done",
    "glb_url": "https://storage.example.com/models/xxx.glb"
  },
  "liked": true,
  "created_at": "2026-06-01T00:00:00Z",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

※ `liked` は認証済みの場合のみ返す。未認証は `null`。
※ `model` は3Dモデルが未生成の場合 `null`。生成中は `{"status": "processing", "glb_url": null}`。

**Errors**
| Status | code | 条件 |
|---|---|---|
| 404 | NOT_FOUND | 商品が存在しない |

---

### PATCH /api/products/:id
商品情報更新

**認証：必要 / 出品者のみ**

**Request Body**（更新するフィールドのみ）
```json
{
  "title": "Sony WH-1000XM5",
  "description": "1年使用しました。",
  "price": 23000,
  "status": "on_sale"
}
```

**Response 200**
```json
{
  "id": "uuid",
  "title": "Sony WH-1000XM5",
  "price": 23000,
  "status": "on_sale",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 403 | FORBIDDEN | 出品者以外が更新しようとした |
| 404 | NOT_FOUND | 商品が存在しない |

---

### DELETE /api/products/:id
商品削除（論理削除）

**認証：必要 / 出品者のみ**

**Response 204** No Content

**Errors**
| Status | code | 条件 |
|---|---|---|
| 403 | FORBIDDEN | 出品者以外が削除しようとした |
| 404 | NOT_FOUND | 商品が存在しない |

---

### POST /api/images
画像アップロード・傷検出開始

5方向の画像をアップロードし、傷検出を非同期で開始。image_idsを返すので、出品時にPOST /api/productsのリクエストに含める。画像プレビューはフロントエンド側でアップロード済みファイルを使って表示する。

**認証：必要**

**Request** `multipart/form-data`

| フィールド | 型 | 説明 |
|---|---|---|
| front | file | 正面画像（JPEG/PNG・10MB以下） |
| back | file | 背面画像 |
| right | file | 右側面画像 |
| left | file | 左側面画像 |
| top | file | 上面画像 |

**Response 201**
```json
{
  "image_ids": ["uuid", "uuid", "uuid", "uuid", "uuid"],
  "damage_detection": "processing"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 400 | VALIDATION_ERROR | ファイル形式不正・サイズ超過 |

---

## Damages

### GET /api/products/:id/damages
商品の傷情報一覧取得

**認証：不要**

**Response 200**
```json
{
  "damages": [
    {
      "id": "uuid",
      "image_id": "uuid",
      "damage_type": "scratch",
      "bbox_x1": 100,
      "bbox_y1": 320,
      "bbox_x2": 140,
      "bbox_y2": 360,
      "description": "右側面に約2cmの線状の傷",
      "model_x": null,
      "model_y": null,
      "model_z": null
    }
  ]
}
```

---

### PATCH /api/damages/:id
傷の3D座標更新（Raycaster変換後にフロントから送信）

※ 3Dフェーズで追加

**認証：必要**

**Request Body**
```json
{
  "model_x": 0.12,
  "model_y": -0.34,
  "model_z": 0.56
}
```

**Response 200**
```json
{
  "id": "uuid",
  "model_x": 0.12,
  "model_y": -0.34,
  "model_z": 0.56
}
```

---

## Comments

### GET /api/products/:id/comments
コメント一覧取得

**認証：不要**

**Response 200**
```json
{
  "items": [
    {
      "id": "uuid",
      "user": {
        "id": "firebase_uid",
        "display_name": "田中花子",
        "avatar_url": "https://..."
      },
      "content": "状態について教えてください",
      "created_at": "2026-06-01T00:00:00Z"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

---

### POST /api/products/:id/comments
コメント投稿

**認証：必要**

**Request Body**
```json
{
  "content": "状態について教えてください"
}
```

**Response 201**
```json
{
  "id": "uuid",
  "content": "状態について教えてください",
  "created_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 400 | VALIDATION_ERROR | contentが空 |
| 404 | NOT_FOUND | 商品が存在しない |

---

### DELETE /api/comments/:id
コメント削除（論理削除）

**認証：必要 / 投稿者のみ**

**Response 204** No Content

**Errors**
| Status | code | 条件 |
|---|---|---|
| 403 | FORBIDDEN | 投稿者以外が削除しようとした |
| 404 | NOT_FOUND | コメントが存在しない |

---

## Likes

### POST /api/products/:id/likes
いいね

**認証：必要**

**Response 201**
```json
{
  "product_id": "uuid",
  "created_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 404 | NOT_FOUND | 商品が存在しない |
| 409 | CONFLICT | 既にいいね済み |

---

### DELETE /api/products/:id/likes
いいね解除

**認証：必要**

**Response 204** No Content

**Errors**
| Status | code | 条件 |
|---|---|---|
| 404 | NOT_FOUND | いいねが存在しない |

---

## Orders

### POST /api/products/:id/orders
商品購入

購入と同時にmessage_roomを作成。

**認証：必要**

**Response 201**
```json
{
  "id": "uuid",
  "product_id": "uuid",
  "price": 25000,
  "status": "pending",
  "message_room_id": "uuid",
  "created_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 403 | FORBIDDEN | 自分の出品商品を購入しようとした |
| 409 | CONFLICT | 既に購入済みの商品 |

---

### GET /api/orders
自分の注文一覧（購入・出品どちらも）

**認証：必要**

**Query Parameters**
| パラメータ | 型 | 説明 |
|---|---|---|
| role | string | buyer / seller（省略時は両方） |

**Response 200**
```json
{
  "items": [
    {
      "id": "uuid",
      "product": {
        "id": "uuid",
        "title": "Sony WH-1000XM5",
        "thumbnail_url": "https://..."
      },
      "price": 25000,
      "status": "pending",
      "role": "buyer",
      "created_at": "2026-06-01T00:00:00Z"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

---

### GET /api/orders/:id
注文詳細取得

**認証：必要 / 関係者のみ（buyer or seller）**

**Response 200**
```json
{
  "id": "uuid",
  "product_id": "uuid",
  "buyer_id": "firebase_uid",
  "price": 25000,
  "status": "pending",
  "message_room_id": "uuid",
  "created_at": "2026-06-01T00:00:00Z",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 403 | FORBIDDEN | 関係者以外のアクセス |
| 404 | NOT_FOUND | 注文が存在しない |

---

### PATCH /api/orders/:id
注文ステータス更新

**認証：必要 / 関係者のみ**

設定可能なステータス遷移：

| 操作 | 実行者 | orders.status変更 |
|---|---|---|
| 受け取り完了 | buyer | pending → completed |
| キャンセル | buyer or seller | pending → cancelled |

※ 購入確定時（`POST /api/products/:id/orders`）に `products.status` を `on_sale` → `sold_out` に更新。キャンセル時は `sold_out` → `on_sale` に戻す。

**Request Body**
```json
{
  "status": "completed"
}
```

**Response 200**
```json
{
  "id": "uuid",
  "status": "completed",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 400 | BAD_REQUEST | 無効なステータス遷移 |
| 403 | FORBIDDEN | 関係者以外・または権限のないステータス変更 |
| 404 | NOT_FOUND | 注文が存在しない |

---

### POST /api/orders/:id/feedback
フィードバック送信（評価）

**認証：必要 / buyer_idが一致 かつ orders.status = 'completed' のみ**

1注文につき1回のみ送信可能。送信後は再送不可。

**Request Body**
```json
{
  "score": 5
}
```

**Response 201**
```json
{
  "id": "uuid",
  "order_id": "uuid",
  "score": 5,
  "created_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 403 | FORBIDDEN | 購入者以外・または注文がcompleted状態でない |
| 404 | NOT_FOUND | 注文が存在しない |
| 409 | CONFLICT | 既にフィードバック送信済み |

---

### POST /api/orders/:id/damage-reports
傷報告

**認証：必要 / buyer_idが一致 かつ orders.status = 'completed' のみ**

**Request Body**
```json
{
  "image_id": "uuid",
  "damage_type": "scratch",
  "bbox_x1": 100,
  "bbox_y1": 200,
  "bbox_x2": 150,
  "bbox_y2": 250,
  "description": "右側面に傷がありました"
}
```

**Response 201**
```json
{
  "id": "uuid",
  "product_id": "uuid",
  "damage_type": "scratch",
  "created_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 403 | FORBIDDEN | 購入者以外・または注文がcompleted状態でない |
| 404 | NOT_FOUND | 注文が存在しない |

---

## Message Rooms

### GET /api/message-rooms/:id/messages
メッセージ一覧取得

**認証：必要 / 参加者のみ**

**Response 200**
```json
{
  "items": [
    {
      "id": "uuid",
      "sender": {
        "id": "firebase_uid",
        "display_name": "山田太郎",
        "avatar_url": "https://..."
      },
      "content": "よろしくお願いします",
      "created_at": "2026-06-01T00:00:00Z"
    }
  ],
  "total": 10,
  "limit": 20,
  "offset": 0
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 403 | FORBIDDEN | 参加者以外のアクセス |
| 404 | NOT_FOUND | ルームが存在しない |

---

### POST /api/message-rooms/:id/messages
メッセージ送信

**認証：必要 / 参加者のみ**

**Request Body**
```json
{
  "content": "よろしくお願いします"
}
```

**Response 201**
```json
{
  "id": "uuid",
  "content": "よろしくお願いします",
  "created_at": "2026-06-01T00:00:00Z"
}
```

**Errors**
| Status | code | 条件 |
|---|---|---|
| 400 | VALIDATION_ERROR | contentが空 |
| 403 | FORBIDDEN | 参加者以外のアクセス |
| 404 | NOT_FOUND | ルームが存在しない |

---

## WebSocket

### WS /ws
リアルタイム通信

**認証：必要**（接続時にクエリパラメータでトークンを渡す）

```
ws://localhost:8080/ws?token=<Firebase ID Token>
```

**サーバーから送信されるイベント**

```json
// 新着メッセージ
{
  "type": "new_message",
  "payload": {
    "room_id": "uuid",
    "message": {
      "id": "uuid",
      "sender_id": "firebase_uid",
      "content": "よろしくお願いします",
      "created_at": "2026-06-01T00:00:00Z"
    }
  }
}

// 傷検出完了
{
  "type": "damage_detection_complete",
  "payload": {
    "condition": "fair",
    "condition_note": "全体的に使用感があり、右側面に目立つ傷があります。",
    "damages": [
      {
        "image_id": "uuid",
        "damage_type": "scratch",
        "bbox_x1": 100,
        "bbox_y1": 320,
        "bbox_x2": 140,
        "bbox_y2": 360,
        "description": "右側面に約2cmの線状の傷"
      }
    ]
  }
}

// 3Dモデル生成完了（3Dフェーズ）
{
  "type": "model_generation_complete",
  "payload": {
    "product_id": "uuid",
    "glb_url": "https://storage.example.com/models/xxx.glb"
  }
}
```
