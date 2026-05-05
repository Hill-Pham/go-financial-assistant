---
phase: planning
title: Project Planning & Task Breakdown
description: Break down work into actionable tasks and estimate timeline
---

# Project Planning & Task Breakdown

## Milestones

- [x] Milestone 1: Core expense recording — SINGLE, INSTALLMENT, RECURRING via text/image
- [x] Milestone 2: Income & transfer support — INCOME, INCOME_RECURRING, TRANSFER kinds
- [x] Milestone 3: Bank statement import — PDF analysis, duplicate detection, confirmation flow
- [x] Milestone 4: Reporting — monthly query, on-demand CSV export, automated monthly report
- [x] Milestone 5: Operations — graceful shutdown, daily scheduler, WhatsApp QR flow, admin endpoint
- [ ] Milestone 6: Quality & observability — missing test coverage, DNS cache, design docs

---

## Task Breakdown

### Phase 1: Foundation (completed)
- [x] Domain entities: `Purchase`, `Payment`, `PurchaseKind`, `TransferDirection`
- [x] Domain ports: `PurchaseRepository`, `AIAnalyzer`, `Messenger`
- [x] PostgreSQL schema (migrations 001–004)
- [x] `db.NewPostgres` + `db.RunMigrations` (goose embedded FS)
- [x] `config.Load` with fail-fast validation
- [x] `cmd/main.go` startup sequence + graceful shutdown

### Phase 2: Core Features (completed)
- [x] `gemini.Client` — `AnalyzeText`, `AnalyzeImage`, `AnalyzeDocument`
- [x] System prompt for expense classification (Portuguese, all types)
- [x] `usecase.AnalyzeExpense` — routes by `ExpenseType`
- [x] `processAnalysis` (SINGLE), `processInstallment`, `processRecurring`, `processCancel`
- [x] `processIncome`, `processIncomeRecurring`
- [x] `processTransfer`
- [x] `processQuery` — monthly expense + income + transfer summary
- [x] `processExportCSV` — resolves export month
- [x] `usecase.ExportCSV` — CSV with BOM, totals footer, `BuildExportCaption`
- [x] `usecase.MonthlyReport` — sends CSV on 1st of month

### Phase 3: Integration & Polish (completed)
- [x] `evolution.Client` — `SendText`, `SendDocument`, `FetchImageBase64`, `EnsureInstance`
- [x] `webhookHandler.Handle` — deduplication, allowlist, text/image/document routing
- [x] Statement import — `ExecuteDocument`, pending confirmation flow (`tryHandlePendingConfirmation`)
- [x] `qrcodeHandler` — HTML page with auto-refresh + bearer auth
- [x] Middleware — source IP validation, admin rate limiting
- [x] Daily scheduler goroutine + startup `GenerateRecurringExpenses`

### Phase 4: Quality (in progress)
- [ ] Add tests for `processIncome` / `processIncomeRecurring`
- [ ] Add tests for `processTransfer`
- [ ] Add tests for `ExecuteDocument` / `SavePendingTransaction` / statement duplicate detection
- [ ] Add tests for `tryHandlePendingConfirmation` and document import HTTP flow
- [ ] Cache DNS resolution in `webhookSourceMiddleware` (resolve once at startup or TTL cache)
- [ ] Expose `GenerateRecurringExpenses` via `ExpenseAnalyzer` interface (or separate `RecurringScheduler`)
- [ ] Fix startup log typo: `"eerror"` → `"error"` in `cmd/main.go`

---

## Dependencies

- Phase 4 test tasks are independent of each other.
- DNS cache improvement requires no schema or interface changes.
- Interface extension for `GenerateRecurringExpenses` requires updating `cmd/main.go` and any mock implementations.

---

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| Evolution API unavailable at startup | 5-second retry loop; `ctx.Done()` exits cleanly |
| In-memory session lost on restart | Pending import TTL is 30 min; user can re-send the document |
| Gemini misclassification | Confidence score surfaced in reply; user can re-send with clearer text |
| Statement false-positive duplicate detection (same amount + day) | Marked as "pending" for user confirmation rather than auto-skipped |
| Month-end day rollover for recurring (e.g., day 31 in February) | `lastValidDay()` helper clamps to last valid day via `time.Date` normalisation |

