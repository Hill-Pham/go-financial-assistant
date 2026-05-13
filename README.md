# Go Financial Assistant

Personal finance assistant via WhatsApp. Send text messages, receipt photos, or bank statement PDFs to automatically record expenses, income, and transfers. The assistant uses AI (Google Gemini) to interpret transactions and store them in a PostgreSQL database.

## How It Works

You send a message to **yourself** on WhatsApp - a text describing a purchase, a receipt photo, or a bank statement PDF - and the assistant records the transactions automatically.

**Message examples:**
- `"spent BRL 45 on lunch via PIX"`
- `"netflix BRL 55 every month on day 15"`
- `"cancel netflix"`
- `"received BRL 6000 as salary"`
- `"put BRL 2000 in the piggy bank"`
- `"how much did I spend in March?"`
- `"export my March expenses"`
- Receipt or invoice photo
- Bank statement PDF (Itau and others)

The assistant classifies each transaction into three categories:

| Type | Description | Example |
|------|-------------|---------|
| **Expense** | Real money spent | Lunch, Netflix, electricity bill |
| **Income** | Money received | Salary, freelance, reimbursement |
| **Transfer** | Movement between your own accounts | Savings deposit, CDB redemption |

Transfers are excluded from expense and income totals, so savings deposits do not distort your financial summary.

## Technologies

- **Go** - main application
- **Google Gemini** - message analysis and interpretation
- **Evolution API** - WhatsApp integration
- **PostgreSQL** - expense storage
- **Redis** - Evolution API cache
- **Docker / Docker Compose** - infrastructure

## Prerequisites

- [Docker](https://www.docker.com/) with Docker Compose
- [Google Gemini](https://aistudio.google.com/app/apikey) API key
- WhatsApp account

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/MarcosAAlbanoJunior/go-financial-assistant.git
cd go-financial-assistant
```

### 2. Configure environment variables

Copy the example file and fill it in with your own values:

```bash
cp .env.example .env
```

Edit `.env`:

```env
PORT=3000
DATABASE_URL=postgres://finassist:finassist@localhost:5432/finassist?sslmode=disable
GEMINI_API_KEY=your-gemini-key

EVOLUTION_API_KEY=any-key-to-protect-the-api
EVOLUTION_API_URL=http://localhost:8082
EVOLUTION_INSTANCE=Financial Assistant
OWNER_PHONE=5511999999999

# Optional - add this if Evolution API returns your number in a different format (known bug)
ALLOWED_NUMBERS=

# Password for the /admin/qrcode endpoint (required for production use)
ADMIN_SECRET=your-secure-password
```

| Variable | Description |
| --- | --- |
| `GEMINI_API_KEY` | Google Gemini API key - get it from [aistudio.google.com](https://aistudio.google.com/app/apikey) |
| `EVOLUTION_API_KEY` | Key used to protect your Evolution API instance - any value is fine |
| `EVOLUTION_INSTANCE` | Instance name in Evolution API |
| `OWNER_PHONE` | Your WhatsApp number with country code and area code, without `+` or spaces (example: `5511999999999`) |
| `ALLOWED_NUMBERS` | Optional alternate number if Evolution API returns your number in a different format |
| `ADMIN_SECRET` | Password for the `/admin/qrcode` endpoint - set a strong value in production |

### 3. Start the project

```bash
docker compose up -d --build && docker compose logs -f app
```

On the first run, the assistant will automatically create the instance in Evolution API and display a **QR code in the terminal**.

Basically, the QR code comes from the Go app logs.

### 4. Connect WhatsApp

Scan the QR code shown in the terminal with your WhatsApp app:

> WhatsApp -> **Linked devices** -> **Link a device** -> scan the QR code

After scanning, the assistant will be ready to use.

> On future runs, if WhatsApp is already connected, the QR code will not be shown.

## Usage

With the container running and WhatsApp connected, send messages to **yourself** on WhatsApp.

### Register a simple expense
```
spent BRL 45 on lunch via PIX
```

### Register an installment purchase
```
bought a pair of shoes for BRL 300 in 3 installments by card
```

### Register a recurring expense
```
netflix BRL 55 every month on day 15
```

### Cancel a recurring entry
```
cancel netflix
```

### Register income (salary, freelance)
```
received BRL 6000 as salary
received BRL 500 from freelance via PIX
```

### Register a transfer between your own accounts
```
put BRL 2000 in the piggy bank
redeemed 500 from CDB
```
Transfers do not affect expenses or income - they are only used to track movements between your own accounts.

### Query the monthly summary
```
how much did I spend this month?
how much did I spend in February?
```

The summary shows:
- **Expenses** by category
- **Total income** (if any)
- **Monthly result** (income - expenses)
- **Investments for the month**: how much was applied and redeemed
- **In account**: result after subtracting the amount that stayed invested

### Import a bank statement (PDF)

Send your bank statement PDF directly in WhatsApp. The assistant processes all transactions automatically:

- Expenses, income, and transfers are classified by AI
- Piggy bank deposits and CDB redemptions are detected as **Transfer** - they do not inflate expenses
- Transactions that already exist in the database are flagged for individual confirmation
- Supports Itau statement format (and other formats with DD/MM/YYYY or YYYY-MM-DD dates)

### Export a CSV spreadsheet

Ask the assistant to export a month's expenses and it will send a `.csv` file directly in WhatsApp - ready to open in Excel or Google Sheets:

```
export my March expenses
send me the CSV for February 2024
I want the January spreadsheet
export
```

- If no month is specified, it exports the **current month**.
- If there are no entries in the period, the assistant responds with a text message.
- The file includes **UTF-8 BOM** for Excel compatibility.
- Columns: Date, Description, Category, Payment Method, Type, Installment, Amount (R$).
- Total rows at the end: **TOTAL EXPENSES**, **TOTAL INCOME** (if any), **BALANCE**, **TOTAL APPLIED** / **TOTAL REDEEMED** (if transfers exist).
- The accompanying message already includes the financial summary: expenses, income, result, applied/redeemed, and in-account amount.

> Gemini interprets the export intent, so natural phrases like _"I want to see my expenses in a spreadsheet"_ or _"generate a CSV for me"_ also work.

### Automatic monthly report

On the first day of each month, the assistant automatically sends the CSV spreadsheet with all expenses from the previous month - no request needed.

### Send a receipt or invoice
Take a photo or forward the receipt image directly in WhatsApp.

## Security and remote management

### Generate a QR code to connect WhatsApp

Open this in your browser, replacing it with your VPS IP or `localhost` if running locally:

```bash
http://<VPS-IP-OR-LOCALHOST>:3000/admin/qrcode?token=your-password
```

If WhatsApp is already connected, it shows a confirmation message. Otherwise, it shows the QR code to scan - the page refreshes automatically every 30 seconds.

**Implemented protections:**
- Requires `ADMIN_SECRET` to be configured (returns `503` if empty)
- Rate limit of 10 requests per minute per IP
- `/webhook` only accepts connections from the Evolution API container (IP verification via Docker internal DNS)

## Useful commands

```bash
# Start and follow application logs
docker compose up -d --build && docker compose logs -f app

# View logs in real time
docker compose logs -f app

# Stop the containers
docker compose down

# Stop and delete all data (database, volumes)
docker compose down -v

# Access the database
docker compose exec postgres psql -U finassist -d finassist
```

## Project structure

```
cmd/                                        application entrypoint
internal/
    config/                               environment variable loading
    domain/                               business entities and rules
    usecase/                              use cases (analysis, recurring, queries, export)
    infra/
        db/                               PostgreSQL repository
        evolution/                        Evolution API (WhatsApp) client
        gemini/                           Google Gemini client
        http/                             HTTP server and webhook handler
migrations/                               database creation SQL scripts
```

## License

MIT
