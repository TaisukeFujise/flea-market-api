# flea-market-api

3DスキャンとAIで商品の傷を検出・可視化するフリマアプリのバックエンドAPI。

**フロントエンドリポジトリ:** [flea-market-web](https://github.com/TaisukeFujise/flea-market-web) 

---

## 概要

商品状態の認識ズレによるトラブルを解消するため、AIが商品の傷・汚れを客観的に検出・評価するフリマプラットフォームのAPIサーバー。

## 技術スタック

| 領域 | 技術 |
| --- | --- |
| バックエンド | Go (Cloud Run) |
| データベース | Cloud SQL (PostgreSQL) |
| ストレージ | Cloud Storage |
| AI | Gemini Vision API |
| 3Dモデル生成 | Meshy API |
| 傷検出 | YOLOv8 (Python / Cloud Run) |
| 内部通信 | gRPC (Go → Python) |
| リアルタイム通信 | WebSocket |

## 主な機能

- ユーザー登録・認証（Google OAuth + JWT）
- 商品CRUD・購入フロー
- Gemini Visionによる商品状態の自動判定・説明文生成
- Meshy APIを使った写真 → 3Dモデル（GLB）生成
- YOLOv8による傷検出・3D座標へのピン留め
- WebSocketを使ったDM・リアルタイム通知
- 商品検索（キーワード・カテゴリ・価格帯・写真検索）

## アーキテクチャ

```
React (Vercel)
  ↓ REST / WebSocket
Go API (Cloud Run)          ← このリポジトリ
  ├─ Gemini Vision API
  ├─ Meshy API
  ├─ Cloud Storage
  └─ Cloud SQL
  ↓ gRPC
Python Service (Cloud Run)  ← YOLOv8 傷検出
  └─ Cloud SQL
```

アーキテクチャは Clean Architecture に基づいて設計。

