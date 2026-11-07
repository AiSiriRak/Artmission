# Wallet and incoming transactions

Order payment continues to use the existing [QR/slip payment APIs](payment-api.md). When a paid order reaches a terminal state through the order application, the backend automatically records each recipient, credits their internal wallet and completes settlement in the same database transaction. No bank transfer is initiated, and the frontend does not execute or acknowledge credits.

## Frontend APIs

Both endpoints are available under `/api/v1`, require `Authorization: Bearer <access-token>`, and accept customer/artist accounts. The owner comes from the token; client-supplied user IDs cannot change the scope. These read routes are available in both development and production. All amounts use integer `int64` satang, with 100 satang = 1 THB.

### GET /wallet/me

```json
{
  "user_id": "<authenticated-user-id>",
  "currency": "THB",
  "balance_satang": 100000,
  "updated_at": "2026-11-04T10:00:00+07:00"
}
```

If the user has never received a credit, return HTTP 200, `balance_satang: 0`, and `updated_at: null`. GET does not create or update the wallet. Wallet balances are internal accounting credits, not evidence of a bank payout.

### GET /transaction/me?limit=20&offset=0

```json
{
  "transactions": [
    {
      "id": "<payment-transaction-id>",
      "order_id": "<order-id>",
      "order_name": "วาดภาพ Portrait",
      "type": "RELEASE_TO_ARTIST",
      "amount_satang": 100000,
      "settlement_outcome": "COMPLETED",
      "created_at": "2026-11-04T10:00:00+07:00"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

- `type`: `RELEASE_TO_ARTIST` for artist income; `REFUND_TO_CUSTOMER` for customer refunds.
- `amount_satang`: the positive amount credited to this user's wallet.
- `settlement_outcome`: `COMPLETED`, `ARTIST_CANCELLED`, `CUSTOMER_CANCELLED`, or `DEADLINE_EXPIRED`. It can be null on historical records without outcome metadata; the API does not invent missing history.
- `created_at`: time the wallet credit was recorded, not the original charge time.
- `total`: all matching incoming credits, independently of the requested page. Empty/beyond-end pages return `transactions: []` while retaining the correct total.
- `limit`: default 20, minimum 1, maximum 100. `offset`: default 0, minimum 0. Invalid pagination returns HTTP 422.

History contains only completed wallet credits belonging to the caller. Charges, legacy top-ups, pending allocations and other users' entries are excluded. Ordering is credit time descending with transaction UUID as a deterministic tie-breaker. Responses and OpenAPI schemas are defined by the REST handler; Bruno examples are under `bruno/wallet`.

## Automatic credit rules

| Order event | Customer wallet | Artist wallet |
|---|---:|---:|
| Completed | 0 | Full held amount |
| Deadline expiration | Full held amount | 0 |
| Artist cancels | Full held amount | 0 |
| Customer cancels | Half, rounded down | Remainder |

Example: customer cancels a 10,001 satang payment → customer gets 5,000 and artist gets 5,001. Platform fee is currently zero. Zero-value allocations are omitted.

Credits run on approved deliverables, worker auto-approval, final-version automatic approval, participant cancellation, and deadline expiration. Only `HELD` payments are eligible. Explicit cancellation closes unpaid `PENDING` payments as `CANCELLED`; time-based expiry closes them as `EXPIRED`. Neither creates credits. Orders awaiting artist acceptance normally have no payment row. Existing terminal guards and QR cleanup remain intact. Historical HELD payments without CHARGE can settle; no historical charges are fabricated.

For each successful terminal transition, the same database transaction:

1. Conditionally changes payment `HELD` → `SETTLED` and records outcome/timestamp.
2. Inserts non-zero payout/refund rows in `payment_transactions`, including `recipient_user_id`.
3. Inserts matching `wallet_entries` journal credits.
4. Atomically increments `wallets.balance_satang` and updates its timestamp.

The credit transaction has `transfer_status: SUCCESS` and a `wallet-credit-` reference because its internal wallet credit commits with it. This does not represent a bank transfer. Errors roll back order, payment, transaction, journal and balance together. Conditional transitions, unique `(payment_id,type)` and unique wallet transaction IDs prevent duplicate credits. Atomic wallet increments protect shared recipients from lost updates; batch credits acquire recipient wallets in a consistent order.

Existing order list/detail payment blocks continue to show persisted settlement and transfer fields. Notifications remain TODOs at the application events. The previous experimental manual mock-wallet top-up/payment/execute APIs have been removed; order funding stays in the existing payment module.

Wallet/ledger migrations refuse rollback when any wallet journal entry or non-zero balance exists. Guards run before destructive statements and hold exclusive table locks through the migration transaction, preventing credits from racing the check. Empty development schemas can still Down/Up. A rejected wallet migration preserves its version, tables, recipient metadata, balances and history.

Goose commits each migration separately: an earlier migration may already have rolled back before the wallet guard stops a multi-step command. Reapply pending migrations before restarting the application. Downgrading a populated financial schema requires a separately reviewed preservation/recovery procedure; no automatic replay of settled payments is used to reconstruct deleted wallet credits.
