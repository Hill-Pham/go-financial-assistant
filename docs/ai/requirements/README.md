---
phase: requirements
title: Requirements & Problem Understanding
description: Clarify the problem space, gather requirements, and define success criteria
---

# Requirements & Problem Understanding

## Problem Statement

Managing personal finances manually is tedious and error-prone. The owner wants a frictionless way to log every transaction — expenses, income, and transfers — directly from WhatsApp without switching to a separate app or spreadsheet.

**Who is affected**: A single owner (and optionally a small allowlist of trusted numbers) who actively uses WhatsApp as their primary communication channel.

**Current situation**: Manual entry in spreadsheets or banking apps after the fact, leading to missed records and poor visibility into spending patterns.

---

## Goals & Objectives

### Primary goals
- Accept natural-language text messages describing transactions and classify them automatically via AI.
- Accept photos of receipts and PDF bank statements and extract transactions from them.
- Persist all transactions in a structured PostgreSQL database.
- Reply on WhatsApp with a confirmation of what was recorded.
- Generate a monthly summary report (CSV) and send it automatically on the 1st of each month.

### Secondary goals
- Allow on-demand financial queries ("How much did I spend in March?").
- Allow on-demand CSV export for any month.
- Support recurring expenses and incomes (e.g., Netflix every month on the 15th).
- Support transfers between own accounts (e.g., investments) that are excluded from expense/income totals.
- Allow cancellation of recurring entries.

### Non-goals
- Multi-user SaaS product — the system is designed for a single owner.
- Web or mobile UI — WhatsApp is the only interface.
- Automated payment execution — the assistant only records, never moves money.
- Full bank synchronisation — import is triggered manually by sending a PDF statement.

---

## User Stories & Use Cases

| Story | Example message |
|---|---|
| Record a single expense | "I spent 45 usd on lunch at Pix." |
| Record an instalment purchase | "I bought sneakers for 300 USD in 3 installments on credit." |
| Record a recurring expense | "Netflix is ​​55 USD every month, due on the 15th." |
| Cancel a recurring expense | "cancel Netflix" |
| Record an income | "I received BRL 6000 in salary." |
| Record a recurring income | "salary 6000 every month" |
| Record a transfer out (investment) | "I put 2000 in the piggy bank." |
| Record a transfer in (redemption) | "I retrieved 500 from the piggy bank." |
| Query monthly expenses | "How much did I spend in March?" |
| Export monthly CSV | "export my March expenses" |
| Import bank statement | send a PDF of the bank statement |
| Photo of a receipt | send a photo of a receipt |

**Edge cases**:
- Duplicate transactions in a bank statement (same amount + same day) — surface as "pending" for manual confirmation.
- WhatsApp reconnection — QR code available at `/admin/qrcode`.
- Recurring expense day falls on a month with fewer days (e.g., day 31 in February) — clamp to last valid day of the month.

---

## Success Criteria

- All transaction types (SINGLE, INSTALLMENT, RECURRING, INCOME, TRANSFER) are correctly classified and persisted.
- AI classification confidence is surfaced in replies and logs.
- A monthly CSV is sent automatically on the 1st of every month.
- The system starts within seconds and recovers gracefully from a temporary Evolution API outage.
- Only owner and explicitly allowlisted numbers can interact with the bot.
- The `/admin/qrcode` endpoint is rate-limited and requires a secret token.

---

## Constraints & Assumptions

- **Single owner**: `OWNER_PHONE` is always in the allowlist; the bot always replies to `ownerPhone`.
- **Evolution API**: WhatsApp connectivity depends on a self-hosted Evolution API instance. Outages are tolerated with a retry loop.
- **Gemini AI**: Classification accuracy depends on the `gemini-2.5-flash-lite` model. Low-confidence results are handled by including confidence in error replies.
- **In-memory state**: Deduplication maps and pending import sessions are in-memory (`sync.Map`). They are lost on restart; a brief window of duplicate processing is possible.
- **No authentication on `/webhook`**: The webhook is protected by source-IP validation (DNS lookup of Evolution host) rather than a shared secret.
- **English language**: All user-facing messages, categories, and payment method labels are in Brazilian English.

---

## Questions & Open Items

- Should duplicate detection in statement import use description as an additional key (currently only amount + date)?
- Should `pendingImports` sessions be scoped per sender JID to support multiple allowlisted users importing simultaneously?
- Should `GenerateRecurringExpenses` be exposed via the `ExpenseAnalyzer` interface to improve testability?

