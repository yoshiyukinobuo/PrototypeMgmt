# Prototype Management System (試作管理)

Tracks prototype requests -> lots -> serialised units -> test results -> issues, with an ERP adapter.
Stack: Go (Gin) + MySQL 8 + Vue 3 (Vite). Containerised for GCP.

## Run locally
    cp .env.example .env      # set JWT_SECRET and ADMIN_PASS
    docker compose up --build
    # UI: http://localhost:8081   login: admin / <ADMIN_PASS>

With `ERP_URL` empty the backend uses a mock ERP, so every screen works without an ERP.

## Dev without Docker
    docker compose up db
    cd backend && go mod tidy && DB_DSN='app:app@tcp(localhost:3306)/prototype?parseTime=true' JWT_SECRET=x ADMIN_PASS=x go run .
    cd frontend && npm install && npm run dev      # proxies /api to :8080

## What is in this first slice
- Requests with approval status; lots (one serial number per unit, auto-created)
- Stage tracker per lot; test results with spec limits (pass/fail computed server-side)
- Issues with countermeasure; dashboard; login (JWT)
- ERP: item master sync (ERP -> system), work order creation per lot (system -> ERP), sync log

Not built yet: BOM revisions, shipments/sample tracking, file upload (a test result only stores a file URL),
user roles, CSV import of measurement data.

## ERP integration (ERPNext)
Set `ERP_URL`, `ERP_API_KEY`, `ERP_API_SECRET`, `ERP_COMPANY`. API key/secret come from an ERPNext integration user.
`backend/erp.go` hides the ERP behind the `ERPClient` interface; add another implementation to target a different ERP.
Calls used: `GET /api/resource/Item`, `POST /api/resource/Work Order`.
Work order creation sends only item, qty, company and a description; your ERPNext may require more fields
(BOM, warehouses) depending on its manufacturing settings. Extend `CreateWorkOrder` to match.

## Deploy on GCP
- Cloud SQL (MySQL 8): create the DB, run `db/schema.sql`. Set `DB_DSN` to the Unix-socket form
  `user:pass@unix(/cloudsql/PROJECT:REGION:INSTANCE)/prototype?parseTime=true`.
- Backend -> Cloud Run (listens on `$PORT`). Put `JWT_SECRET`, `ADMIN_PASS`, `ERP_API_*` in Secret Manager.
- Frontend -> second Cloud Run service; set `BACKEND_URL` to the backend service URL. Restrict access with IAP.

## Security notes
- The single admin account comes from env vars; replace with real users/roles before wider use.
- All SQL is parameterised. The login has no rate limit yet.


## 履歴に関する注意

このGit履歴は、ZIPスナップショットとマニフェストを基にレビュー用として新規作成したものです。各コミット日時・差分は元のGit履歴から復元されたものではなく、当時の実作業を証明する記録ではありません。実際の作業記録・バックアップ・チケット等と照合してから利用してください。

### 作業メモ
- 2020-08-03：試作依頼 — 依頼登録画面の入力条件を整理（レビュー用メモ 001）
- 2020-08-12：依頼ステータス — 依頼ステータス更新処理の仕様を整理（レビュー用メモ 002）
- 2020-08-24：ロット管理 — ロット一覧と詳細表示の構成を整理（レビュー用メモ 003）
- 2020-08-31：個体管理 — 個体単位の管理処理を確認（レビュー用メモ 004）
- 2020-09-09：シリアル番号 — シリアル番号の採番規則を整理（レビュー用メモ 005）
- 2020-09-18：工程管理 — 工程ステータス表示の構成を整理（レビュー用メモ 006）
- 2020-09-28：試験結果 — 試験結果登録時の入力項目を整理（レビュー用メモ 007）
- 2020-10-07：合否判定 — 測定値に基づく合否判定条件を確認（レビュー用メモ 008）
- 2020-10-16：不具合管理 — 不具合一覧と登録フォームを整理（レビュー用メモ 009）
- 2020-10-26：ダッシュボード — ダッシュボードの集計項目を整理（レビュー用メモ 010）
- 2020-11-04：ERP連携 — ERP接続設定とクライアント構成を確認（レビュー用メモ 011）
- 2020-11-13：品目同期 — 品目マスター同期処理の仕様を整理（レビュー用メモ 012）
- 2020-11-23：作業指示 — ERP作業指示作成時の送信項目を整理（レビュー用メモ 013）
- 2020-12-02：同期ログ — ERP同期ログの表示項目を整理（レビュー用メモ 014）
- 2020-12-11：認証 — ログインとトークン発行処理を確認（レビュー用メモ 015）
- 2020-12-21：API通信 — フロントエンド共通API処理を整理（レビュー用メモ 016）
- 2020-12-30：データベース — データベース構造と関連制約を確認（レビュー用メモ 017）
- 2021-01-08：画面共通化 — 画面構成とナビゲーションを整理（レビュー用メモ 018）
- 2021-01-18：運用設定 — ローカル実行環境の設定を整理（レビュー用メモ 019）
- 2021-01-26：開発環境 — 開発環境のセットアップ手順を整理（レビュー用メモ 020）
- 2021-02-05：試作依頼 — 依頼登録画面の入力条件を整理（レビュー用メモ 021）
- 2021-02-15：依頼ステータス — 依頼ステータス更新処理の仕様を整理（レビュー用メモ 022）
- 2021-02-23：ロット管理 — ロット一覧と詳細表示の構成を整理（レビュー用メモ 023）
- 2021-03-04：個体管理 — 個体単位の管理処理を確認（レビュー用メモ 024）
- 2021-03-15：シリアル番号 — シリアル番号の採番規則を整理（レビュー用メモ 025）
- 2021-03-23：工程管理 — 工程ステータス表示の構成を整理（レビュー用メモ 026）
