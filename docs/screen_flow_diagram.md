```mermaid
flowchart LR
    AUTH[🔐 ログイン\nGoogle OAuth]

    AUTH --> HOME

    subgraph NAV[ボトムナビ]
        HOME[🏠 ホーム\n検索バー・フィルター・商品一覧]
        LISTING[📦 出品タブ\n出品中・取引中セクション]
        MYPAGE[👤 マイページ]
    end

    subgraph PRODUCT[商品]
        PRODUCT_DETAIL[🎧 商品詳細\n3Dビューア・傷ピン・Q&A]
        PRODUCT_EDIT[✏️ 商品詳細\n編集・削除]
    end

    subgraph PURCHASE[購入フロー]
        PURCHASE_CONFIRM[💳 購入確認\n商品名・価格確認]
        PURCHASE_COMPLETE[✅ 購入完了]
    end

    subgraph LISTING_FLOW[出品フロー]
        SHOOTING[📸 ガイド付き撮影\n5方向固定]
        PRODUCT_INFO[📝 商品情報入力\n傷検出バックグラウンド実行中]
        LISTING_CONFIRM[✔️ 出品確認\n傷検出完了で出品ボタン活性化]
    end

    subgraph TRADE_SELLER[取引・出品者]
        TRADE_DETAIL_SELLER[💬 取引詳細\nDM・WebSocket]
    end

    subgraph TRADE_BUYER[取引・購入者]
        TRADE_DETAIL_BUYER[💬 取引詳細\nDM・WebSocket]
        DAMAGE_REPORT[🔍 傷報告\n3Dモデル上でタップ]
    end

    subgraph MYPAGE_DETAIL[マイページ詳細]
        LIKED[❤️ いいね一覧]
        HISTORY[👁️ 閲覧履歴]
        PURCHASED[🛍️ 購入した商品]
    end

    HOME -->|商品タップ| PRODUCT_DETAIL

    PRODUCT_DETAIL -->|購入するボタン| PURCHASE_CONFIRM
    PURCHASE_CONFIRM -->|購入を確定する| PURCHASE_COMPLETE

    LISTING -->|出品するボタン| SHOOTING
    LISTING -->|出品中セクションタップ| PRODUCT_EDIT
    LISTING -->|取引中セクションタップ| TRADE_DETAIL_SELLER
    PRODUCT_EDIT -->|保存| LISTING

    SHOOTING -->|5方向撮影完了| PRODUCT_INFO
    PRODUCT_INFO -->|次へ| LISTING_CONFIRM
    LISTING_CONFIRM -->|出品する| LISTING

    MYPAGE --> PURCHASED
    MYPAGE --> LIKED
    MYPAGE --> HISTORY
    PURCHASED -->|タップ| TRADE_DETAIL_BUYER
    TRADE_DETAIL_BUYER -->|傷を報告するボタン| DAMAGE_REPORT
    LIKED -->|タップ| PRODUCT_DETAIL
    HISTORY -->|タップ| PRODUCT_DETAIL

    style AUTH fill:#6366f1,color:#fff
    style HOME fill:#0ea5e9,color:#fff
    style LISTING fill:#0ea5e9,color:#fff
    style MYPAGE fill:#0ea5e9,color:#fff
    style PRODUCT_DETAIL fill:#8b5cf6,color:#fff
    style PRODUCT_EDIT fill:#8b5cf6,color:#fff
    style PURCHASE_CONFIRM fill:#f59e0b,color:#fff
    style PURCHASE_COMPLETE fill:#10b981,color:#fff
    style SHOOTING fill:#ec4899,color:#fff
    style PRODUCT_INFO fill:#ec4899,color:#fff
    style LISTING_CONFIRM fill:#ec4899,color:#fff
    style TRADE_DETAIL_SELLER fill:#f97316,color:#fff
    style TRADE_DETAIL_BUYER fill:#f97316,color:#fff
    style DAMAGE_REPORT fill:#ef4444,color:#fff
    style PURCHASED fill:#0ea5e9,color:#fff
    style LIKED fill:#0ea5e9,color:#fff
    style HISTORY fill:#0ea5e9,color:#fff
```
