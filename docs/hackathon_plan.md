次世代フリマアプリ
傷スキャンができるフリマアプリ
ハッカソン企画ドキュメント


# 1. コンセプト
一言で言うと：「傷スキャンができるフリマアプリ」
出品体験と購入体験の両方をAIで革新する次世代のフリマプラットフォーム。
デモ商品：ヘッドフォン（Sony WH-1000XM5想定）


# 2. 問題提起

## 2-1. 現状の課題
フリマアプリにおけるトラブルの多くは、「商品状態の認識のズレ」から発生している。特に中古品では「思っていたより傷が多い」「説明と状態が違う」といった問題が頻発する。
この問題の本質は、商品状態の評価が出品者の主観に依存している点にある。現在のフリマでは「良い」「やや傷あり」といった状態はすべて出品者の自己申告で決定されている。

## 2-2. 他サービスとの比較
セカンドストリートやRAGTAGのようなリユース店舗では、専門スタッフが実物を確認し、統一基準で状態を評価することで客観性と信頼性が担保されている。
フリマには「第三者による客観評価が存在しない」という構造的な欠陥がある。

## 2-3. 既存ECとの違い
AmazonやGoogleはすでに3D商品表示を導入している。しかしそれは新品向けであり、「傷」という概念が存在しない。中古品を扱うフリマ特有の問題は「商品状態の信頼性」だ。


# 3. ソリューション

## 3-1. AIによる客観的な商品状態判定
- スマートフォンで商品を3Dスキャン
- 全方向の形状・質感をデータ化
- AIが傷・汚れ・使用感を検出
- 客観的な商品状態を自動生成


## 3-2. なぜ3Dか

| 比較 | 内容 |
| --- | --- |
| 2D写真の問題 | 角度によって傷が隠れる。出品者が意図的・無意識に隠せる。 |
| 3Dモデルの強み | 全方向スキャン済み。AIが傷の位置をピン留め。買い手が自分で回転させて確認できる。 |



## 3-3. 持続可能な改善サイクル
購入者が商品を受け取った後、商品画像上で傷箇所を囲って報告できる仕組みを導入する。この報告データをVertex AI Multimodal Embeddingでベクトル化してfeedbackとして蓄積し、次回の傷検出でGeminiへのfew-shotとして活用することでAIが継続的に精度向上する。
- 報告が「自分を守るための行動」に直結（傷の報告が返金・補償の根拠になる）
- プラットフォームが成長するほど信頼性が上がる構造


# 4. 発表ストーリーの骨格
- 問題提起：フリマのトラブルの多くは商品状態の認識ズレから生まれる
- 原因：商品状態の評価が出品者の主観に依存している
- 比較：リユースショップは専門スタッフが客観評価する。フリマにはその店員がいない
- 解決策：だからAIが代わりになる
- 既存技術との違い：AmazonやGoogleの3Dは新品向け。本提案は「状態の信頼性」という価値を追加
- 結論：AIにより「誰が見ても同じ評価になる状態記述」を実現する

「メルカリで商品状態を決めているのは出品者です。私たちはそれをAIに変えました。」


# 5. 機能一覧

## 5-1. 必須機能

| 機能 | 内容 |
| --- | --- |
| ユーザー登録・認証 | 基本的なユーザー管理機能 |
| 商品出品 | 商品情報入力・ガイド付き撮影UI |
| 商品購入 | 購入フロー（決済API連携は不要） |
| DM機能 | ユーザー間でのメッセージ交換（WebSocket） |
| Gemini API連携 | Gemini Visionで商品状態を自動判定＋説明の生成。常駐AIアシスタント |
| デプロイ | バックエンド: CloudRun / フロントエンド: Vercel / DB: CloudSQL |



## 5-2. 追加機能

| 機能 | 内容 |
| --- | --- |
| いいね機能 | 商品へのいいね実装 |
| 商品検索 | キーワード検索・カテゴリフィルター・価格帯フィルター・写真検索 |
| 認証・認可強化 | Firebase Authentication |
| 3Dスキャン×傷検出 | ガイド付き撮影UI→Gemini Vision傷検出（bbox）→3D上ピン留め表示（3Dフェーズ） |
| 傷報告フロー | 購入者が商品画像上で傷箇所を囲って報告→Embeddingとして蓄積しGeminiのfew-shotに反映 |
| テスト・CI/CD | 単体テスト＋CI/CDパイプライン整備 |
| 多様な通信方式 | REST / WebSocket の使い分け |
| グラスモーフィズムUI | 3D周辺のみliquid glass等のグラスモーフィズムを採用 |




# 6. 技術スタック

| 領域 | 技術 | 役割 |
| --- | --- | --- |
| フロントエンド | React (Vercel) | UI実装・Three.js 3Dビューア |
| 認証 | Firebase Authentication | Firebase Admin SDKでIDトークン検証 |
| バックエンド | Go (CloudRun) | REST/WebSocket・Gemini Vision・Meshy API・Vertex AI呼び出し |
| データベース | CloudSQL (PostgreSQL + pgvector) | メインDB・Embeddingベクトル検索 |
| AI: 傷検出・状態判定 | Gemini Vision API (structured output) | 傷bbox検出・condition判定・説明文生成 |
| AI: フィードバックEmbedding | Vertex AI Multimodal Embedding | 傷報告画像のベクトル化・few-shot検索 |
| AI: 3D生成 | Meshy API | 写真→3Dモデル自動生成（3Dフェーズ） |
| 3D表示 | Three.js (@react-three/fiber) | 3Dビューア・傷ピン留め表示（3Dフェーズ） |
| 通信 | REST / WebSocket | 用途に応じた使い分け |
| ストレージ | Cloud Storage | 商品写真・3DモデルGLBファイルの保存・配信 |
| インフラ管理 | Terraform | Cloud Run・Cloud SQL・Cloud Storage等の全インフラをコードで管理 |
| CI/CD | Cloud Build | GitHubリポジトリ・ブランチ指定→自動ビルド→CloudRunへデプロイ |
| シークレット管理 | Secret Manager | Gemini/Meshy APIキー・DB接続情報などの機密情報管理 |



## 通信方式の使い分け
- REST：商品CRUD・認証・購入フロー（基本的なHTTPリクエスト）
- WebSocket：メッセージ機能・傷検出完了通知・3Dモデル生成完了通知


# 7. コア機能詳細：3Dスキャン×傷検出

## 7-1. 実装パイプライン（確定部分）
- ガイド付き撮影UI（5方向固定）
- 本人確認UIのように枠に商品を合わせて、指定角度で順番に撮影
- 正面・背面・右・左・上の5方向をスマホブラウザのカメラAPIで撮影
- 撮影角度が既知になることで3D座標変換が可能になる
- product_imagesのangleカラムで撮影角度を管理
- 3Dモデル生成（Meshy API）
- Meshy APIのProプラン（約$20/月）で複数枚の写真からGLBを生成
- GoがMeshyにリクエスト → job_idを取得してproduct_modelsに保存
- Meshyの処理完了時にWebhookでGoのエンドポイントを叩く
- GoがGLBをCloud Storageに保存 → WebSocketでフロントに完了通知
- 約30秒で完了・product_models.statusで状態管理（pending/processing/done/failed）
- React Three Fiberでの3Dモデル表示・ピン留め
- @react-three/fiberでGLBモデルをブラウザ上に表示
- damages.model_x/y/zの3D座標にピンマーカーを配置
- グラスモーフィズムのパネルで「右側面に約2cmの傷があります。状態：良い」と表示
- 商品状態の説明文生成（Gemini Vision）
- Gemini Visionに写真を渡して全体サマリーを生成（1回のAPI呼び出しで傷検出・状態判定・説明文生成を同時実施）
- products.condition_noteに全体サマリー文を保存
- damages.descriptionに個別の傷の説明文を保存


## 7-2. 確定した実装方針

① 傷検出：**Gemini Vision API (structured output) に確定**
- YOLOv8・Pythonサービス・gRPCは使用しない
- Gemini に5方向画像を一括送信し、bbox座標・damage_type・condition・condition_noteをJSONで取得
- 画像は1024×1024に正規化してCloud Storageに保存

② 2D→3D座標変換：**Raycaster（3Dフェーズ・保留）**
- Three.jsで撮影角度を再現したカメラからbbox中心座標にRayを飛ばしGLBと交差する3D座標を取得
- フロント処理。変換後にPATCH APIでdamages.model_x/y/zを更新



## 7-3. 持続可能な改善サイクル
- 購入者が商品受け取り後、商品画像上で傷箇所を囲って報告
- 報告は「返金・補償の根拠」になるため自分を守るための自然な動機が生まれる
- 報告画像をVertex AI Multimodal Embeddingでベクトル化→feedback_embeddingsに蓄積
- 同カテゴリの新商品出品時にpgvectorで類似検索→Geminiへのfewーshotとして活用→精度が継続的に向上


# 8. システムアーキテクチャ

## 8-1. 全体構成
React (Vercel)
↓ REST / WebSocket
Go API (CloudRun) ← メインバックエンド
├─ Gemini Vision API（傷検出・状態判定）
├─ Vertex AI Multimodal Embedding（フィードバック）
├─ Meshy API（3Dモデル生成）
├─ Cloud Storage（写真・3Dモデルの保存）
└─ CloudSQL (PostgreSQL + pgvector)


## 8-2. 設計方針
- アーキテクチャ：Clean Architecture（拡張性・保守性を重視）
- GoからGemini Vision・Vertex AI・Meshy APIを直接呼び出す
- Pythonサービス・gRPCは使用しない（Goで完結）
- 2Dフェーズ→フィードバックフロー→3Dフェーズの段階的な開発


# 9. 開発スケジュール（2ヶ月）

| 期間 | フェーズ | 主なタスク |
| --- | --- | --- |
| Week1-2 | 設計＋インフラ疎通 | ワイヤーフレーム・DB設計・API設計・CloudRun+Vercel+CloudSQL疎通確認 |
| Week3-4 | 必須機能バックエンド | 認証・商品CRUD・購入フロー・DM機能のAPI実装 |
| Week5 | フロントエンド実装 | 必須機能のUI・API繋ぎ込み |
| Week6 | 3Dモデル生成 | ガイド付き撮影UI・Meshy API組み込み・Three.js 3Dビューア |
| Week7 | 3Dフェーズ・仕上げ | Meshy API・Three.js 3Dビューア・Raycaster座標変換・傷ピン留め表示 |
| Week8 | 仕上げ・発表準備 | デモデータ投入・発表スライド作成・デモリハーサル |


最重要マイルストーン：Week2終わりに「Hello WorldがCloudRunで動いている」状態を必ず作る


# 10. 設計フェーズ（次のステップ）

## 優先順位
- インフラ疎通確認（最優先）
- CloudRun / Vercel / CloudSQL の接続確認
- Hello Worldを早めにデプロイ
- CI/CDパイプラインの雛形作成
- DB設計
- テーブル・カラム・リレーション定義
- 主要テーブル：users / products / product_images / product_models / damages / damage_detection_summaries / damage_reports / orders / messages / likes / categories / feedback_embeddings
- API設計
- OpenAPIでエンドポイント・リクエスト・レスポンス定義
- ワイヤーフレーム
- 主要画面の遷移図（Figma等）


# 11. 審査基準との対応

| 審査軸 | 評価 | 対応内容 |
| --- | --- | --- |
| AI活用 | ◎ | Gemini Vision（状態判定・写真検索・AIアシスタント）をコアに活用 |
| 新規ユーザー体験 | ◎ | 「誰が見ても同じ評価」という既存フリマにない体験 |
| 独自性・創造性 | ◎ | 傷スキャン×フリマという組み合わせは既存サービスにない |
| 内部設計 | ◎ | Clean Architecture / REST+WebSocketの使い分け・pgvectorフィードバックループ |
| 機能実装 | ○ | 必須機能＋AI機能を2ヶ月で実装 |
| デザイン | ○ | グラスモーフィズム（3D周辺のみ）でメリハリをつける |
| 発表・デモ | ◎ | 「撮影→3D生成→傷ピン留め」をその場で見せる |
