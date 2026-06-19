# 傷検出仕様書

## 1. 方針
- 傷検出エンジン：Vertex AI Gemini（structured output）に一本化。YOLOv8 は使用しない
- 使用モデル：`gemini-2.5-flash`（画像入力・構造化出力のバランスが最良）
- bbox 座標：画像左上を (0,0)・右下を (1000,1000) とした正規化座標（整数）で返す
- 3Dフェーズは Week6-7 で追加。2Dフェーズと独立して動作し、スキーマ変更なしで対応可能

---

## 2. 2Dフェーズ フロー（Week3-4）

**操作の順序：画像アップロード → AI検出 → 商品作成**

フロントは WebSocket で `damage_detection_complete` を受け取ってから `POST /api/products` を呼ぶ。
AI 検出が走る時点では商品（products レコード）がまだ存在しないため、`damages` テーブルは `product_id` を持たず `image_id` のみで紐付ける。商品作成後は `product_images.product_id` を経由して間接的に products と繋がる。

```
[1] ガイド付き撮影（フロント）
    5方向の画像をアップロード
    → 1024×1024 にリサイズして Cloud Storage に保存
    → damage_detection_summaries INSERT（status: processing）
    → product_images INSERT（angle・summary_id付き）
    ← image_ids を返す（商品作成にはまだ使わない）

[2] 傷検出 + 状態サマリー生成（Go goroutine → Vertex AI Gemini）
    ※ リクエストと独立した goroutine で非同期実行（タイムアウト 60 秒）
    product_images の画像5枚 + プロンプトを Vertex AI Gemini に送信
    1回の API 呼び出しで傷リスト・condition・condition_note をまとめて取得

    ※ pgvector フィードバックがある場合
    → 撮影画像を Vertex AI Multimodal Embedding でベクトル化
    → feedback_embeddings で類似検索（同 category_id・上位3件）
    → フィードバック画像3枚をプロンプトに追加（参照専用）

    成功時：
    → damages INSERT（bbox座標・damage_type・description・image_id）トランザクションで一括
    → damage_detection_summaries UPDATE（condition・condition_note・status: done）
    → WebSocket で damage_detection_complete 通知

    失敗時（API エラー・DB エラー等）：
    → damage_detection_summaries UPDATE（status: failed）
    → WebSocket で damage_detection_failed 通知

[3] 商品作成（フロントが WebSocket 通知を受け取ってから呼ぶ）
    POST /api/products に image_ids を含めて送信
    → damage_detection_summaries の status = 'done' を確認
      - status = 'processing'（検出中）または 'failed'（検出失敗）の場合は 400
    → condition・condition_note を summaries から取得して products に保存
    → product_images の所有権を全件確認してから product_id を紐づける

[4] フロントに反映
    → 2D画像上に bbox マーカー表示（bbox_x1/y1/x2/y2 をそのまま使用）
    → condition_note を商品詳細に表示
```

### WebSocket イベント一覧（傷検出関連）

| イベント | 送信タイミング | payload |
|---|---|---|
| `damage_detection_complete` | 検出・DB保存が正常完了 | `condition`, `condition_note`, `damages[]` |
| `damage_detection_failed` | 検出または DB 保存でエラー | なし |

### フロント側のタイムアウト処理（必須実装）

WebSocket 通知はベストエフォートであり、以下のケースで通知が届かないことがある:

- デプロイや Cloud Run スケールダウンのタイミングで goroutine が SIGKILL された（`status='processing'` のまま固着）
- ユーザーがアップロード後にブラウザを閉じた・ネットワーク瞬断

**フロントはアップロード完了から 90 秒以内に `damage_detection_complete` / `damage_detection_failed` のいずれも届かなかった場合、タイムアウトとして扱い「検出に失敗しました。再アップロードしてください」を表示すること。**

`status='processing'` 固着は上記ケースで発生しうるが、再アップロードすることで新しい `summary_id` で仕切り直せるため許容する。固着した古いレコードはユーザー操作に影響しない（`POST /api/products` に渡す `image_ids` が変わるため）。

> **WaitGroup によるシャットダウン待機を採用しない理由**  
> goroutine の最大実行時間（60 秒）が Cloud Run の termination grace period を超えるため、WaitGroup で待っても結局 SIGKILL される。根本解決は Cloud Tasks 等のジョブキューへの切り出しだが、現フェーズでは上記フロント側タイムアウトで許容する。

---

## 3. Gemini リクエスト設計

### 3-1. 撮影方向管理

各画像の直前に撮影方向を示すテキストを挿入して送信する。
angle は `front` / `back` / `right` / `left` / `top` の5値。
フィードバック画像がある場合は末尾に追加（参照専用）。

### 3-2. プロンプト

**System Instruction（共通）：**
`internal/infra/ai/client.go` の `systemInstruction` 定数を参照。
- condition の基準（good / fair / poor）
- damage_type の基準（scratch / dirt / wear）
- bbox 指定ルール（0-1000 正規化座標）
- condition_note は日本語 1〜2文

**User メッセージ構造（フィードバックなし）：**
```
「以下の商品画像を査定してください」
+ 繰り返し：「撮影方向: {angle}」テキスト + 画像（GCS URI）
```

**User メッセージ構造（フィードバックあり）：**
```
上記に加え末尾に追加：
「## 傷の参考例（別商品です）
過去の同カテゴリ商品で報告された傷の例です。
傷の検出パターンの参考にしてください。damagesには含めないでください。」
+ フィードバック画像（GCS URI、参照専用）
```

### 3-3. response_schema（Go）

```go
schema := &genai.Schema{
    Type: genai.TypeObject,
    Properties: map[string]*genai.Schema{
        "damages": {
            Type: genai.TypeArray,
            Items: &genai.Schema{
                Type: genai.TypeObject,
                Properties: map[string]*genai.Schema{
                    "image_angle": {
                        Type: genai.TypeString,
                        Enum: []string{"front", "back", "right", "left", "top"},
                    },
                    "damage_type": {
                        Type: genai.TypeString,
                        Enum: []string{"scratch", "dirt", "wear"},
                    },
                    "bbox_x1":     {Type: genai.TypeInteger},
                    "bbox_y1":     {Type: genai.TypeInteger},
                    "bbox_x2":     {Type: genai.TypeInteger},
                    "bbox_y2":     {Type: genai.TypeInteger},
                    "description": {Type: genai.TypeString},
                },
                Required: []string{
                    "image_angle", "damage_type",
                    "bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2",
                    "description",
                },
            },
        },
        "condition": {
            Type: genai.TypeString,
            Enum: []string{"good", "fair", "poor"},
        },
        "condition_note": {Type: genai.TypeString},
    },
    Required: []string{"damages", "condition", "condition_note"},
}
```

---

## 4. レスポンス JSON サンプル

```json
{
  "damages": [
    {
      "image_angle": "right",
      "damage_type": "scratch",
      "bbox_x1": 100,
      "bbox_y1": 320,
      "bbox_x2": 140,
      "bbox_y2": 360,
      "description": "右側面に約2cmの線状の傷"
    },
    {
      "image_angle": "front",
      "damage_type": "dirt",
      "bbox_x1": 60,
      "bbox_y1": 180,
      "bbox_x2": 100,
      "bbox_y2": 220,
      "description": "正面下部に汚れ"
    }
  ],
  "condition": "fair",
  "condition_note": "全体的に使用感があり、右側面に目立つ傷があります。"
}
```

---

## 5. image_angle → image_id 変換（Go）

```go
// DBから取得済みのマップ: angle → ProductImage
imageByAngle := map[string]ProductImage{ ... }

for _, d := range geminiResp.Damages {
    img := imageByAngle[d.ImageAngle]
    db.Create(&Damage{
        ImageID:     img.ID,
        DamageType:  DamageType(d.DamageType),
        BboxX1:      d.BboxX1,
        BboxY1:      d.BboxY1,
        BboxX2:      d.BboxX2,
        BboxY2:      d.BboxY2,
        Description: d.Description,
    })
}
```

---

## 6. ENUM 定義

### PostgreSQL
```sql
CREATE TYPE damage_type AS ENUM ('scratch', 'dirt', 'wear');
CREATE TYPE condition_type AS ENUM ('good', 'fair', 'poor');
```
GORM の AutoMigrate では生成されないため、マイグレーションファイルに明示的に記述する。

### Gemini response_schema
`damage_type` と `condition` に Enum フィールドで制約（3-3. 参照）。

---

## 7. フィードバック反映フロー

```
[1] 購入者が傷報告（フロント）
    前提条件：orders.buyer_id = 自分 かつ orders.status = 'completed'（受け取り済み）
    2D画像上で傷箇所を囲う（2Dフェーズ）または3Dモデル上でタップ（3Dフェーズ・保留）
    → image_id + bbox_x1/y1/x2/y2（2D・フロントで外接bboxに変換）または model_x/y/z（3D）+ damage_type + description を送信

[2] damage_reports INSERT（Go）
    前提条件をサーバー側でも検証（orders.buyer_id・status確認）
    product_id, user_id, image_id, damage_type, bbox_x1/y1/x2/y2, description を保存

[3] フィードバック画像の特定（Go）
    damage_reports.image_id → product_images から画像URL取得

[4] Embedding 生成（Go → Vertex AI）
    画像を Vertex AI Multimodal Embedding（1408次元）に送信
    → feedback_embeddings INSERT
        - damage_report_id
        - category_id（商品のカテゴリ）
        - embedding vector(1408)

[5] 次回の傷検出に反映
    同カテゴリの新商品が出品されたとき
    → feedback_embeddings で pgvector 類似検索（同 category_id・上位3件）
    → フィードバック画像として Gemini に追加（index 5-7）
```

### フィードバック画像のインデックス制御

2つの対策を組み合わせる：
- **プロンプト**：役割を構造的に分離し「damagesに含めない」と明示（3-2. 参照）
- **response_schema**：`image_angle` の enum を front/back/right/left/top に制約し、スキーマレベルでフィードバック画像の角度を返せなくする（3-3. 参照）

---

## 8. 3Dフェーズ（Week6-7・保留）

```
[4] 2D→3D座標変換（Raycaster・フロント処理）
    damages から bbox_x1/y1/x2/y2 + image_id（angle）を取得
    → Three.js で撮影角度を再現したカメラを設定
    → bbox 中心座標から Ray を飛ばし GLB と交差する 3D 座標を取得
    → PATCH /damages/:id で model_x/y/z を UPDATE

[5] フロントに反映
    2D bbox マーカー表示 → 3D ピン表示に切り替え
    → damages.model_x/y/z にピンマーカーを配置
```

2Dフェーズのスキーマをそのまま使用。damages の model_x/y/z は 2Dフェーズ中 NULL のまま。
