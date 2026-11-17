# 試作管理システム (Prototype Management System)

試作依頼からロット、個体（シリアル番号付き）、試験結果、課題に至るまでを追跡・管理するシステムです。ERP連携用アダプターを備えています。
技術スタック: Go (Gin) + MySQL 8 + Vue 3 (Vite)。GCP（Google Cloud Platform）向けにコンテナ化されています。

## ローカルでの実行方法
cp .env.example .env      # JWT_SECRET と ADMIN_PASS を設定してください
docker compose up --build

UI: http://localhost:8081   
ログイン: admin / <ADMIN_PASS>

`ERP_URL` を空のままにすると、バックエンドはモック（擬似）ERPを使用するため、ERP環境がなくてもすべての画面が動作します。

## Dockerを使用しない開発環境
docker compose up db
cd backend && go mod tidy && DB_DSN='app:app@tcp(localhost:3306)/prototype?parseTime=true' JWT_SECRET=x ADMIN_PASS=x go run .
cd frontend && npm install && npm run dev      # /api へのリクエストは :8080 へプロキシされます

## 今回のリリース範囲（初期実装）
- 承認ステータス付きの試作依頼、ロット管理（個体ごとにシリアル番号を自動生成）
- ロットごとの工程追跡、試験結果の記録（仕様限界に基づき、合否判定をサーバー側で自動計算）
- 課題管理（対策内容を含む）、ダッシュボード、ログイン機能（JWT）
- ERP連携: 品目マスターの同期（ERP → 本システム）、ロットごとの製造指図（Work Order）作成（本システム → ERP）、同期ログ

未実装項目: BOM（部品表）リビジョン管理、出荷・サンプル追跡、ファイルアップロード（試験結果にはファイルURLのみを保存）、ユーザーロール（権限管理）、測定データのCSVインポート。

## ERP連携 (ERPNext)
`ERP_URL`、`ERP_API_KEY`、`ERP_API_SECRET`、`ERP_COMPANY` を設定してください。APIキーおよびシークレットは、ERPNext側で連携用ユーザーを作成して取得します。
`backend/erp.go` では、ERPとのやり取りを `ERPClient` インターフェースで抽象化しています。別のERPと連携させる場合は、このインターフェースを実装した新しいクラス（構造体）を追加してください。
使用するAPI: `GET /api/resource/Item`、`POST /api/resource/Work Order`。
製造指図の作成時には、品目、数量、会社、説明のみが送信されます。 ERPNextの製造設定によっては、追加のフィールド（BOM、倉庫など）が必要になる場合があります。これに合わせて `CreateWorkOrder` を拡張してください。

## GCPへのデプロイ
- Cloud SQL (MySQL 8): データベースを作成し、`db/schema.sql` を実行します。`DB_DSN` は、Unixソケット形式（`user:pass@unix(/cloudsql/PROJECT:REGION:INSTANCE)/prototype?parseTime=true`）で設定してください。
- バックエンド: Cloud Run（`$PORT` でリッスン）。`JWT_SECRET`、`ADMIN_PASS`、`ERP_API_*` などの変数は Secret Manager に格納します。
- フロントエンド: 別の Cloud Run サービスとしてデプロイし、`BACKEND_URL` にバックエンドサービスのURLを設定します。IAP（Identity-Aware Proxy）を使用してアクセスを制限してください。

## セキュリティに関する注意点
- 現在、管理者アカウントは環境変数から読み込まれる単一のものとなっています。本格的な運用に際しては、適切なユーザー管理やロールベースのアクセス制御（RBAC）に置き換えてください。
- すべてのSQLクエリはパラメータ化されています。なお、ログイン機能には現時点でレート制限（試行回数制限）は実装されていません。