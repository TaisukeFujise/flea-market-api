# 傷検出仕様書

## 1. 方針
- 傷検出エンジン：Gemini Vision API（structured output）に一本化。YOLOv8 は使用しない
- 画像解像度：Cloud Storage 保存前に 1024×1024 にリサイズして統一（座標変換を安定させるため）
- 3Dフェーズは Week6-7 で追加。2Dフェーズと独立して動作し、スキーマ変更なしで対応可能

---

## 2. 2Dフェーズ フロー（Week3-4）

```
[1] ガイド付き撮影（フロント）
    5方向の画像をアップロード
    → 1024×1024 にリサイズして Cloud Storage に保存
    → product_images INSERT（angle付き）

[2] 傷検出 + 状態サマリー生成（Go → Gemini Vision）
    product_images の画像5枚 + プロンプトを Gemini に送信
    1回の API 呼び出しで傷リスト・condition・condition_note をまとめて取得

    ※ pgvector フィードバックがある場合
    → 撮影画像を Vertex AI Multimodal Embedding でベクトル化
    → feedback_embeddings で類似検索（同 category_id・上位3件）
    → フィードバック画像3枚をプロンプトに追加（参照専用・index 5-7）

    Gemini レスポンス（JSON）
    → damage_detection_summaries INSERT（condition・condition_note・user_id）
    → UPDATE product_images SET summary_id WHERE id IN (image_ids)
    → damages INSERT（bbox座標・damage_type・description・image_id）

[3] フロントに反映
    → 2D画像上に bbox マーカー表示（bbox_x1/y1/x2/y2 をそのまま使用）
    → condition_note を商品詳細に表示
```

---

## 3. Gemini リクエスト設計

### 3-1. 画像インデックス

画像を送る順番を固定し、インデックスで管理する。

| index | angle |
|-------|-------|
| 0 | front |
| 1 | back |
| 2 | right |
| 3 | left |
| 4 | top |

フィードバック画像がある場合は index 5-7 に追加（参照専用）。

### 3-2. プロンプト

**フィードバックなし：**
```
画像0が正面、画像1が背面、画像2が右側面、画像3が左側面、画像4が上面です。
各画像で検出した傷を指定のJSON形式で返してください。
```

**フィードバックあり（画像5-7追加時）：**
```
画像0が正面、画像1が背面、画像2が右側面、画像3が左側面、画像4が上面です。
各画像で検出した傷を指定のJSON形式で返してください。

## 傷の参考例（別商品です）
画像5〜7は過去の同カテゴリ商品で報告された傷の例です。
傷の検出パターンの参考にしてください。damagesには含めないでください。
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
                    "image_index": {
                        Type: genai.TypeInteger,
                        Enum: []string{"0", "1", "2", "3", "4"},
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
                    "image_index", "damage_type",
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
      "image_index": 2,
      "damage_type": "scratch",
      "bbox_x1": 100,
      "bbox_y1": 320,
      "bbox_x2": 140,
      "bbox_y2": 360,
      "description": "右側面に約2cmの線状の傷"
    },
    {
      "image_index": 0,
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

## 5. image_index → image_id 変換（Go）

```go
angleMap := map[int]string{
    0: "front",
    1: "back",
    2: "right",
    3: "left",
    4: "top",
}

// DBから取得済みのマップ: angle → ProductImage
imageByAngle := map[string]ProductImage{ ... }

for _, d := range geminiResp.Damages {
    angle := angleMap[d.ImageIndex]
    img := imageByAngle[angle]
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
- **response_schema**：`image_index` の enum を 0-4 に制約し、スキーマレベルで5以上の値を返せなくする（3-3. 参照）

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
