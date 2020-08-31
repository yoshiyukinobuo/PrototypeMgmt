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
