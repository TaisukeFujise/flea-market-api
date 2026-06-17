# 残タスク一覧

## 実装状況サマリー

| # | 内容 | 状況 |
|---|------|------|
| #1 ユーザー CRUD | ✅ 完了 |
| #2 カテゴリ一覧 | ✅ 完了 |
| #3 画像アップロード | ✅ 完了 |
| #4 商品一覧・詳細 | ✅ 完了 |
| #5 商品出品・編集・削除 | ✅ 完了 |
| #6 コメント | ✅ 完了 |
| #7 いいね | ✅ 完了 |
| #8 いいね一覧・閲覧履歴 | ✅ 完了 |
| #9 注文 | ✅ 完了 |
| #10 傷報告 | ✅ 完了 |
| #11 メッセージ | ✅ 完了 |
| #12 WebSocket（new_message） | ✅ 完了 |
| #13 傷検出AI連携 | ✅ 完了 |
| #14 傷情報API | **未実装** |
| #15 3Dモデル生成AI連携 | **未実装** |
| #16 WebSocket AI通知 | 🔶 部分完了（damage_detection_complete / damage_detection_failed 実装済み。model_generation_complete は #15 待ち） |

---

## 残タスク詳細

### #14 傷情報API

**エンドポイント**
- `GET /api/products/:id/damages`
- `PATCH /api/damages/:id`

**実装内容**
- `internal/repository/damage_repository.go` — ListByProductID / UpdateCoordinates（CreateAll は実装済み）
- `internal/service/damage_service.go`
- `internal/handler/damage_handler.go`
- `router.go` の `notImplemented` を差し替え

**備考**
- GET は認証不要
- PATCH は 3D フェーズ用（Raycaster で算出した model_x / y / z を保存）
- #13 の完了が前提（damages テーブルにデータが存在する）→ 前提は満たされている

---

### #15 3Dモデル生成AI連携

**概要**
Meshy などの外部サービスで3Dモデルを非同期生成する。

**実装内容**
- `internal/repository/product_model_repository.go` — Create / UpdateStatus / UpdateGlbURL
- 外部サービス（Meshy）への生成リクエスト送信
- Webhook でステータス更新（`pending → processing → done / failed`）
- 完了後に #16 の `model_generation_complete` 通知を呼ぶ

**差し替え箇所**
- `internal/service/product_service.go` の CreateProduct — 出品時に `product_models` を `pending` で作成する処理を追加

**備考**
- `product_models.job_id` に Meshy のジョブIDを保存する
- `GET /api/products/:id` の `model` フィールドは追加実装なしで自動的に実データを返すようになる（product_repository が product_models を JOIN 済みのため）

---

### #16 model_generation_complete 通知（#15 完了後）

**実装内容**
- `internal/handler/hub.go` に `NotifyModelGenerationComplete` を追加
- `router.go` の imageService と同様に productModelService から hub を呼ぶ

**備考**
- `damage_detection_complete` / `damage_detection_failed` は実装済み
- `model_generation_complete` のみ未実装

---

## 依存関係

```
#14 傷情報API（独立して着手可能）

#15 3Dモデル生成AI連携
  └─ #16 model_generation_complete 通知
```
