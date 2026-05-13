---
phase: testing
title: Testing Strategy
description: Define testing approach, test cases, and quality assurance
---

# Testing Strategy

## Test Coverage Goals

- Unit test coverage target: all business logic in `internal/usecase/` and `internal/domain/`.
- Integration scope: HTTP handlers tested with `httptest` + interface mocks; DB layer not unit-tested (requires live Postgres).
- No end-to-end tests (Evolution API + Gemini are external services).
- Run with: `make test` or `go test ./... -v`
- Coverage report: `make test-coverage`

---

## Unit Tests — Current Status

### `internal/config`
- [x] `TestLoad_Success` — all required vars present
- [x] `TestLoad_DefaultPort` — PORT defaults to 8080
- [x] `TestLoad_DefaultEvolutionAPIURL` — defaults to `http://evolution:8082`
- [x] `TestLoad_InvalidPort` — non-numeric PORT returns error
- [x] `TestLoad_Missing*` — each required var missing individually
- [x] `TestLoad_MultipleErrors` — accumulated errors joined
- [x] `TestParseAllowedNumbers_*` — single, multiple, already-suffixed, empty

### `internal/domain`
- [x] `TestCategoryLabel` — all Category → Portuguese label mappings
- [x] `TestPaymentMethodLabel` — all PaymentMethod → Portuguese label mappings
- [x] `purchase_test.go` — `NewPurchase`, `NewIncome`, `NewTransfer` constructors; `Cancel()`
- [x] `payment_test.go` — `NewPayment`; `PaidAt` auto-set when status is PAID

### `internal/infra/gemini`
- [x] `TestParseResponse_Success` — valid JSON → `ExpenseAnalysis`
- [x] `TestParseResponse_AmountNull` — null amount handled
- [x] `TestParseResponse_NilResponse` / `TestParseResponse_EmptyCandidates` — empty responses

### `internal/infra/evolution`
- [x] `TestSendText_Success` — extracts message ID from response
- [x] `TestSendText_StripAtSign` — strips `@s.whatsapp.net` before sending
- [x] `TestSendText_HTTPError` — 4xx returns error

### `internal/infra/http` — middleware
- [x] `TestWebhookSourceMiddleware_AllowsKnownHost`
- [x] `TestWebhookSourceMiddleware_BlocksUnknownIP`
- [x] `TestWebhookSourceMiddleware_BlocksUnresolvableHost`
- [x] `TestWebhookSourceMiddleware_BlocksMalformedRemoteAddr`
- [x] `TestIPRateLimiter_AllowsUnderLimit` / `BlocksOverLimit` / `IsolatesIPs` / `ResetsAfterWindow`
- [x] `TestAdminRateLimitMiddleware_Returns429WhenLimitExceeded`

### `internal/infra/http` — handlers
- [x] `TestHandle_InvalidJSON` / `TestHandle_InvalidBody`
- [x] `TestHandle_BotMessageDedup` / `TestHandle_AlreadyProcessed`
- [x] `TestHandle_NotAllowedNumber`
- [x] `TestHandle_TextMessage_Success` / `TestHandle_TextMessage_AllowedNumber`
- [x] `TestHandle_ExtendedTextMessage`
- [x] `TestHandle_EmptyText_UnsupportedMessage`
- [x] `TestHandle_ImageMessage_WithBase64` / `FetchBase64` / `FetchBase64_Error` / `InvalidBase64`
- [x] `TestHandle_MessengerSendText_StoresSentID` / `MessengerSendText_Error`
- [x] `TestHandleExportCommand_SendsDocument` / `EmptyMonth_SendsText` / `ExporterError_Returns500` / `UsesMonthFromAnalyzer`
- [x] `TestQRCodeHandler_*` — auth, connected state, HTML with QR, 503 for empty base64
- [x] `TestHandleError_*` / `TestNotifyError_*` / `TestDecodeBase64Image_*`

### `internal/usecase`
- [x] `TestExecuteText_Success` / `AnalyzerError` / `AmountNil` / `RepoSaveError` / `DescriptionFallback*`
- [x] `TestExecuteImage_Success` / `AnalyzerError` / `RawInputFormat`
- [x] `TestInferPaymentMethod` / `TestParsePaymentMethod` / `TestParseCategory` / `TestResolvePaymentMethod`
- [x] `TestProcessAnalysis_InvalidPaymentMethod` / `InvalidAmount`
- [x] `TestExecuteText_Installment_*` — success, nil amount, no installment info, repo error, amount calc, due dates
- [x] `TestExecuteText_Recurring_*` — success, first payment generated, nil amount, default day of month
- [x] `TestExecuteText_CancelRecurring_*` — success, not found, no description, description fallback
- [x] `TestGenerateRecurringExpenses_*` — skip wrong day, generate on target day, skip already paid month
- [x] `TestLastValidDay_*` — normal day, day exceeds month, leap year, day 31 in 30-day month
- [x] `TestProcessQuery_*` — current month, specific month/year, empty result, repo error, no query info
- [x] `TestExportCSV_*` — empty month, repo error, filename, BOM, header, single/installment/recurring/transfer rows, totals, nil description
- [x] `TestBuildExportCaption_*` — with transfers, without transfers
- [x] `TestMonthlyReport_*` — sends document, empty month skips, exporter error, messenger error, correct phone

### Missing test coverage (gaps to address)
- [ ] `processIncome` — success, nil amount, repo error — **no test file**
- [ ] `processIncomeRecurring` — success, day-of-month default, first payment reference month — **no test file**
- [ ] `processTransfer` — success, nil amount, invalid direction, repo error — **no test file**
- [ ] `ExecuteDocument` — auto-saved transactions, duplicate → pending, Gemini error — **no test file**
- [ ] `SavePendingTransaction` — EXPENSE / INCOME / TRANSFER kinds — **no test file**
- [ ] `saveStatementTransaction` — each kind — **no test file**
- [ ] `tryHandlePendingConfirmation` — yes confirmation, no skip, session expiry, non-confirmation text passes through — **no HTTP test**
- [ ] `handleDocumentImport` — base64 present, base64 fetch, decode error — **no HTTP test**

---

## Integration Tests

No automated integration tests exist. The following paths require a live Postgres + Evolution instance and are validated manually:

- [ ] Full message → classify → save → reply cycle
- [ ] Bank statement PDF import + confirmation flow
- [ ] Monthly report scheduling (1st of month)
- [ ] WhatsApp reconnection via QR code endpoint

---

## Test Data

- Mocks are defined in package-local `mocks_test.go` files (`mockPurchaseRepo`, `mockAnalyzer`, `mockMessenger`, `mockQRProvider`, `mockCSVExporter`).
- No external fixtures or seed data required for unit tests.

---

## Test Reporting & Coverage

Run coverage:
```bash
make test-coverage
# or
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Known gaps** (files with no or partial test coverage):
- `internal/usecase/process_income.go`
- `internal/usecase/process_transfer.go`
- `internal/usecase/process_statement.go`
- `internal/infra/http/webhook_handler.go` — document import and pending confirmation paths

