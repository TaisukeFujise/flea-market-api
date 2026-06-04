# flea-market-api

3DスキャンとAIで商品の傷を検出・可視化するフリマアプリの**バックエンドAPI**。

**-> フロントエンドリポジトリ:** [flea-market-web](https://github.com/TaisukeFujise/flea-market-web) 

---

## 概要

商品状態の認識ズレによるトラブルを解消するため、AIが商品の傷・汚れを客観的に検出・評価するフリマプラットフォームのAPIサーバー。

## 技術スタック

| 領域 | 技術 |
| --- | --- |
| バックエンド | Go (Cloud Run) |
| データベース | Cloud SQL (PostgreSQL) + pgvector |
| ストレージ | Cloud Storage |
| 認証 | Firebase Authentication |
| AI: 傷検出・状態判定 | Gemini Vision API (structured output) |
| AI: フィードバックEmbedding | Vertex AI Multimodal Embedding |
| 3Dモデル生成 | Meshy API |
| リアルタイム通信 | WebSocket |
| インフラ管理 | Terraform |

## 主な機能

- ユーザー登録・認証（Firebase Authentication）
- 商品CRUD・購入フロー
- Gemini Visionによる傷検出（bbox）・商品状態の自動判定・説明文生成
- 購入者の傷報告 → Vertex AI EmbeddingでフィードバックをGeminiのfew-shotに反映
- Meshy APIを使った写真 → 3Dモデル（GLB）生成（3Dフェーズ）
- WebSocketを使ったDM・リアルタイム通知
- 商品検索（キーワード・カテゴリ・価格帯・写真検索）

## アーキテクチャ

```
React (Vercel)
  ↓ REST / WebSocket
Go API (Cloud Run)                    ← このリポジトリ
  ├─ Gemini Vision API（傷検出・状態判定）
  ├─ Vertex AI Multimodal Embedding（フィードバック）
  ├─ Meshy API（3Dモデル生成）
  ├─ Cloud Storage
  └─ Cloud SQL (PostgreSQL + pgvector)
```

アーキテクチャは Clean Architecture に基づいて設計。

