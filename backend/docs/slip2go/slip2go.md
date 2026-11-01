# Slip2Go Integration Notes

> Purpose: Give AI coding agents reliable project-specific context for
> integrating Slip2Go with an existing Go REST API and PostgreSQL
> backend.
>
> Documentation checked: 2026-10-31.
>
> Official documentation: - PromptPay QR payload:
> https://slip2go.com/guide/rest-api/promptpay-qr - PromptPay QR image
> URL: https://slip2go.com/guide/rest-api/promptpay-qr-url - Verify slip
> by QR payload: https://slip2go.com/guide/rest-api/qr-code - Verify
> slip by image: https://slip2go.com/guide/rest-api/image -
> Authentication: https://slip2go.com/guide/authentication - Response:
> https://slip2go.com/guide/response - Account information:
> https://slip2go.com/guide/rest-api/account-info

## 1. Project context

The project is an existing Artist-Customer Commission Platform.

-   Backend language: Go
-   API style: REST
-   Database: PostgreSQL
-   Payment provider: Slip2Go
-   Existing Order system: Must be inspected and reused.
-   This document describes provider integration only. It does not
    define the entire application architecture or replace the existing
    Order model.

### Business payment rules

Use the actual order amount from PostgreSQL. Never trust an amount
supplied by the frontend.

Settlement rules:

  Order outcome          Artist amount   Customer refund
  -------------------- --------------- -----------------
  Order completed                 100%                0%
  Deadline expired                  0%              100%
  Artist cancelled                  0%              100%
  Customer cancelled               50%               50%

All internal monetary calculations must use integer satang (for example,
THB 100.00 = 10000 satang). Convert to the exact decimal string format
required by Slip2Go only at the provider boundary. Do not use
floating-point arithmetic for money.

**Important:** Slip2Go QR generation and slip verification do not, by
themselves, provide bank payout/refund execution. A separate bank
transfer/disbursement integration is required for real transfers to
Artists or Customers. Until such a provider exists, settlement records
must not be represented as completed real transfers.

## 2. Authentication and configuration

Slip2Go REST API examples use:

``` http
Authorization: Bearer <SLIP2GO_SECRET_KEY>
Content-Type: application/json
```

The secret must be held by the Go backend only.

Suggested environment variables (confirm actual API base URL and account
details from the Slip2Go dashboard/account documentation):

``` env
SLIP2GO_API_BASE_URL=
SLIP2GO_SECRET_KEY=
SLIP2GO_PROMPTPAY_CODE=
SLIP2GO_PROMPTPAY_TYPE=phone_number
SLIP2GO_ACCOUNT_NAME=
```

Rules:

-   Load configuration from environment variables using the project's
    existing configuration approach.
-   Never hardcode the secret.
-   Never return the secret to the frontend.
-   Never log Authorization headers, API secrets, or full sensitive bank
    details.
-   Do not commit `.env`; ensure it is ignored by Git.
-   Commit only a `.env.example` with blank or placeholder values.
-   Use HTTPS for provider requests.
-   Set explicit HTTP client timeouts.
-   Handle non-2xx responses, malformed JSON, provider error codes, and
    network timeouts.
-   Do not assume GitHub repository secrets automatically become
    available to developers' local machines. GitHub Actions secrets are
    available only to authorized workflows; a deployed backend must
    receive its runtime secret through its deployment environment/secret
    manager.

The official Authentication guide documents the `Authorization` header
with a Bearer token and describes optional IP whitelisting (up to 10 IP
addresses; the guide says the initial setting is `*`, which accepts all
IPs).

## 3. Generate PromptPay QR

### 3.1 Endpoint

``` http
POST /api/qr-payment/generate-qr-code
```

Authentication is required.

### 3.2 Request body

The official guide documents these fields:

  -----------------------------------------------------------------------
  Field                   Type                    Meaning
  ----------------------- ----------------------- -----------------------
  `promptPayCode`         string                  PromptPay identifier of
                                                  the receiving account

  `promptPayType`         string                  `phone_number`,
                                                  `citizen_id`, or
                                                  `e_wallet`

  `accountName`           string                  Recipient account name

  `amount`                string                  Amount to encode in the
                                                  QR, e.g. `"88.88"`
  -----------------------------------------------------------------------

Example (illustrative only; do not copy the sample recipient into the
project):

``` json
{
  "promptPayCode": "<PLATFORM_PROMPTPAY_CODE>",
  "promptPayType": "phone_number",
  "accountName": "<PLATFORM_ACCOUNT_NAME>",
  "amount": "100.00"
}
```

### 3.3 Response

The QR-payload endpoint returns:

``` json
{
  "code": "200",
  "message": "Success",
  "data": {
    "qrCode": "<QR_PAYLOAD>",
    "accountName": "<ACCOUNT_NAME>",
    "amount": "100.00"
  }
}
```

`data.qrCode` is the QR payload string. Persist it with the Payment and
render it in the frontend; do not treat it as an image URL or store a
generated image.

### 3.4 Image-link alternative (not used by this backend)

``` http
POST /api/qr-payment/generate-qr-image-link
```

Request fields are the same as QR payload generation. The documented
response contains:

``` json
{
  "code": "200",
  "message": "Success",
  "data": {
    "qrImageLink": "<QR_IMAGE_URL>",
    "accountName": "<ACCOUNT_NAME>",
    "amount": "100.00"
  }
}
```

This endpoint returns an image URL and is not used by the current
payment lifecycle.

### 3.5 Application flow

1.  The artist accepts a pending Order. In one database transaction, the
    backend changes its status to `NOT_PAID`, captures the transition
    timestamp, creates its Payment using the stored Order amount, and
    stores `qr_expire = transition timestamp + 24 hours`.
2.  The customer calls `POST /api/v1/orders/{order_id}/payment/qr`. The
    backend checks ownership, `NOT_PAID` status, and expiry; if a payload
    is already stored it is returned without calling Slip2Go. Otherwise
    it takes a PostgreSQL advisory lock per Payment (a concurrent call
    gets `409`), re-checks under the lock, and calls Slip2Go outside any
    database transaction or row lock.
3.  The backend validates the provider response and conditionally stores
    the QR payload only while the Payment is pending, the Order remains
    `NOT_PAID`, and the persisted expiry is still in the future.
4.  The customer's read-only
    `GET /api/v1/orders/{order_id}/payment/qr` returns the stored payload
    and expiry, or `404` ("payment QR has not been generated") if none
    exists. It never calls Slip2Go.
5.  The worker clears expired payloads, marks Payments `EXPIRED`, and
    cancels still-unpaid Orders with a cancellation record. Slip
    verification repeats the status and expiry checks while holding the
    Order and Payment locks before recording the payment.

QR generation is never automatic and never retried in a loop.

Do not accept PromptPay destination details from an untrusted frontend
request. Do not create multiple active payments for the same Order
unless the existing business rules explicitly allow it.

Provider errors leave the QR payload `NULL`, so no usable QR is exposed;
the customer may call POST again. Advisory locking and the stored-payload
check prevent duplicate generation, and conditional persistence prevents
duplicate payload storage. Slip2Go's documented API does not provide an
idempotency key for QR generation; after an ambiguous provider timeout, a
manual retry can result in another remote request even though only one
payload can be stored.
Slip2Go configuration is supplied through `SLIP2GO_API_BASE_URL`,
`SLIP2GO_SECRET_KEY`, `SLIP2GO_PROMPTPAY_CODE`,
`SLIP2GO_PROMPTPAY_TYPE`, and `SLIP2GO_ACCOUNT_NAME`.

## 4. Verify slip by QR payload

### 4.1 Endpoint

``` http
POST /api/verify-slip/qr-code/info
```

Authentication is required.

This endpoint accepts a QR payload decoded from the customer's transfer
slip. It does not accept an arbitrary uploaded image in the `qrCode`
field.

### 4.2 Request body

``` json
{
  "payload": {
    "qrCode": "<DECODED_SLIP_QR_PAYLOAD>",
    "checkCondition": {
      "checkDuplicate": true,
      "checkReceiver": [
        {
          "accountType": "<RECEIVER_ACCOUNT_TYPE>",
          "accountNumber": "<RECEIVER_ACCOUNT_NUMBER_OR_PROMPTPAY_ID>"
        }
      ],
      "checkAmount": {
        "type": "eq",
        "amount": 10000
      }
    }
  }
}
```

The exact receiver fields must match the receiving account details and
account type documented by Slip2Go. Do not guess or hardcode sample
bank/account values.

Documented verification conditions include:

-   `checkDuplicate`: boolean; enable it to ask Slip2Go to check whether
    the slip has already been used.
-   `checkReceiver`: array of acceptable recipient-account conditions.
    The documentation says a slip is considered valid if at least one
    receiver condition matches.
-   `checkAmount`: amount comparison with `eq`, `gte`, or `lte`. The
    documentation's example uses integer amount `10000` for THB 100.00
    (satang).
-   `checkDate`: date comparison with `eq`, `gte`, or `lte`; date is
    specified in GMT/UTC format.

For a payment system, prefer exact amount matching (`eq`) and an
explicit receiver check. Enable duplicate checking. Validate the
returned transaction data in the application as an additional safeguard.

### 4.3 Response data to inspect

The official guide documents fields including:

-   top-level `code` and `message`
-   `data.referenceId`
-   `data.decode`
-   `data.transRef`
-   `data.dateTime`
-   `data.amount`
-   `data.receiver`
-   `data.sender`

The provider's sample success response uses `code: "200000"` and
`message: "Slip found"`. Do not assume every HTTP 200 means a valid
slip. Interpret the provider's documented response code and data, and
treat unknown codes as unverified until confirmed against the current
Response/Error Codes guide.

Persist a provider transaction reference such as `transRef` when
present. Add a database uniqueness constraint for the relevant provider
transaction reference so the same bank transaction cannot be applied to
multiple Payments even if concurrent requests occur. Also use Slip2Go
duplicate checking; provider checking does not replace local database
uniqueness and idempotency.

### 4.4 Verify an uploaded slip image directly

If the customer uploads an image, Slip2Go documents this endpoint:

``` http
POST /api/verify-slip/qr-image/info
Content-Type: multipart/form-data
```

The documented multipart fields are:

-   `file`: the uploaded image file.
-   `payload`: a JSON string containing optional verification
    conditions.

Example payload value:

``` json
{
  "checkDuplicate": true,
  "checkReceiver": [
    {
      "accountType": "<RECEIVER_ACCOUNT_TYPE>",
      "accountNumber": "<RECEIVER_ACCOUNT_NUMBER_OR_PROMPTPAY_ID>"
    }
  ],
  "checkAmount": {
    "type": "eq",
    "amount": "200.31"
  }
}
```

The official example sends the file as multipart field `file` and the
JSON string as multipart field `payload`, with the same Bearer
Authorization header. This endpoint can perform image verification
directly, so local QR decoding is not required for this approach.

**Important amount-format discrepancy to verify:** the QR-payload
endpoint's `checkAmount` example documents an integer such as `10000`,
while the image-upload endpoint example documents a decimal string such
as `"200.31"`. Do not assume these formats are interchangeable. Follow
the selected endpoint's current documentation and confirm with a
controlled test using a known amount before implementing production
checks.

The documentation page does not establish all upload limits in the
material reviewed here. The application must still enforce a
conservative maximum upload size, validate the actual file type, and
reject malformed files before forwarding them.

Official image-verification documentation:
https://slip2go.com/guide/rest-api/image

### 4.4.1 Decision: file upload vs. image link

Slip2Go also offers `POST /api/verify-slip/qr-image-link/info` with
`payload.imageUrl`. This backend keeps the multipart file upload because
the link variant needs the slip stored at a publicly reachable URL first
(extra upload step, public storage of payment slips, and a customer-chosen
URL that could be swapped or abused), while the upload sends the bytes
straight to Slip2Go and nothing is stored. Revisit the link endpoint only
if slips are already uploaded to object storage by another feature.

### 4.4.2 Application response of `POST /orders/{id}/payment/verify`

Slip outcomes return HTTP 200 with `success` and a readable `message`:

``` json
{ "success": false, "reason": "AMOUNT_MISMATCH",
  "message": "The slip amount (50.00 THB) does not match the order amount (100.00 THB).",
  "order_id": "..." }
```

Reasons: `SLIP_ALREADY_USED`, `WRONG_RECEIVER`, `AMOUNT_MISMATCH`,
`SLIP_NOT_FOUND`, `SLIP_FRAUD`, `INVALID_IMAGE`, `SLIP_REJECTED` (unknown
provider code, included in the message). On success: `success: true`,
`payment_id`, and `status`. Order or QR state problems (not `NOT_PAID`,
QR expired, not found) and provider outages still use HTTP error statuses.

### 4.5 Documented response codes

The official Response guide lists these important verification results:

  ------------------------------------------------------------------------------
  Code                    Meaning                 Application handling
  ----------------------- ----------------------- ------------------------------
  `200000`                Slip found              Candidate success; validate
                                                  amount, receiver, transaction
                                                  reference and duplicate status

  `200200`                Slip is valid           Success only if required data
                                                  and local checks also pass

  `200202`                Successfully queued     Asynchronous/pending; do not
                                                  mark PAID yet

  `200401`                Recipient account does  Reject verification
                          not match               

  `200402`                Transfer amount does    Reject verification
                          not match               

  `200403`                Transfer date does not  Reject verification
                          match                   

  `200404`                Slip not found          Reject verification

  `200500`                Slip is fraud           Reject verification

  `200501`                Slip is duplicated      Reject verification

  `200502`                Bank error; try again   Temporary/unknown result; do
                                                  not mark PAID

  `400001`                QR code missing         Request/input error

  `400002`                File incorrect          Request/input error

  `400400`                Request object invalid  Request/input error

  `400409`                Request conflicted      Investigate conflict; do not
                                                  mark PAID

  `401001`                Authentication token    Configuration/authentication
                          mismatch                error

  `401004`                Package expired         Provider account issue

  `401005`                Insufficient token      Provider quota/billing issue

  `401006`                Insufficient credit     Provider quota/billing issue

  `401007`                IP address not allowed  Provider IP whitelist
                                                  configuration issue
  ------------------------------------------------------------------------------

The list is not necessarily exhaustive. Check the current official
Response guide for other codes and changes.

**Do not treat every HTTP 200 as payment success.** Slip2Go documents
business result codes inside the response body. Parse and interpret the
provider `code` explicitly.

## 5. Payment status and safe state transitions

Suggested application-level statuses (adapt to existing project naming):

-   `PENDING`: QR generated; payment not yet verified.
-   `VERIFYING`: verification request is in progress.
-   `PAID`: verified successfully and all required checks passed.
-   `VERIFICATION_FAILED`: provider confirms invalid/failed
    verification.
-   `EXPIRED`: payment window expired.
-   `REFUND_PENDING`, `PAYOUT_PENDING`: settlement requested but not
    confirmed by a real transfer provider.
-   `REFUNDED`, `PAYOUT_COMPLETED`: only use when a real transfer
    provider confirms completion.

Do not mark a Payment as `PAID` because: - the QR was generated, - an
image was uploaded, - the client says payment succeeded, - Slip2Go timed
out, - the provider returned an unknown response, - the backend received
HTTP 200 without validating the provider response body.

If the provider request times out, the result may be unknown. Do not
blindly retry in a way that could double-apply a transaction. Reconcile
using the provider's documented lookup/history facilities if available,
or leave the payment in a reviewable state.

## 6. Duplicate protection and concurrency

Implement multiple layers of protection:

1.  Send `checkDuplicate: true` to Slip2Go.
2.  Store the provider transaction reference (`transRef`) when
    available.
3.  Add a PostgreSQL unique constraint/index for the provider
    transaction reference, scoped to the provider if appropriate.
4.  Ensure only one successful Payment can be associated with the same
    Order according to the business rules.
5.  Use a database transaction and row locking or conditional updates to
    prevent concurrent verification requests from confirming the same
    Order twice.
6.  Make verification and settlement operations idempotent.
7.  Never create a second refund or payout for the same settlement
    event.

Do not rely only on a frontend button being disabled.

## 7. Error handling and observability

-   Distinguish provider rejection from network timeout and internal
    database failure.
-   Keep provider HTTP status, provider response code, and a sanitized
    error message for troubleshooting.
-   Do not log secrets, Authorization headers, full slip payloads, or
    unnecessary personal/bank information.
-   Return stable application error responses to the frontend; do not
    expose raw provider internals.
-   Record timestamps and correlation/payment IDs for troubleshooting.
-   Add tests for success, invalid slip, duplicate slip, wrong amount,
    wrong receiver, timeout, malformed provider response, and concurrent
    verification.

## 8. Settlement is separate from Slip2Go

Slip2Go's QR generation and slip verification endpoints establish
payment QR data and verify incoming transfer slips. They do not, based
on the documented endpoints above, establish a bank payout/refund API.

The application's settlement calculation must be independent:

-   Completed: Artist 100%, Customer 0%.
-   Deadline expired: Artist 0%, Customer 100%.
-   Artist cancelled: Artist 0%, Customer 100%.
-   Customer cancelled: Artist 50%, Customer 50%.

Use integer satang. Record settlement instructions and statuses in
PostgreSQL. Until a separate real bank-transfer provider is integrated,
do not mark a refund or payout as actually transferred.

## 9. Implementation instructions for AI agents

Before coding:

1.  Read the root `AGENTS.md`.
2.  Inspect the existing Go architecture, routes, models, repositories,
    services, migrations, authentication, and Order status flow.
3.  Read this `docs/slip2go.md`.
4.  Re-check the official Slip2Go documentation links in this file,
    especially the image verification endpoint and response/error codes.
5.  Report the current project structure and identify files to reuse or
    change.
6.  Identify any missing account configuration or provider details.
7.  Propose a minimal implementation plan and wait for approval.

During implementation:

-   Implement one feature at a time: QR generation, slip verification,
    then settlement calculation.
-   Follow existing project conventions; do not introduce a parallel
    architecture unnecessarily.
-   Explain each important struct, function, API flow, and database
    change in beginner-friendly language.
-   Do not invent undocumented API endpoints, request fields, response
    fields, or provider behavior.
-   Do not make real payout/refund calls unless a separately approved
    transfer provider exists.
-   Add migrations and tests using the project's existing tools.
-   Run formatting, tests, and relevant checks after changes.
-   Report exactly what was changed, what passed, and what remains
    unimplemented.

## 10. Official references

-   Generate PromptPay QR payload:
    https://slip2go.com/guide/rest-api/promptpay-qr
-   Generate PromptPay QR image URL:
    https://slip2go.com/guide/rest-api/promptpay-qr-url
-   Verify slip by QR code: https://slip2go.com/guide/rest-api/qr-code
-   Verify slip by image: https://slip2go.com/guide/rest-api/image
-   Authentication: https://slip2go.com/guide/authentication
-   Response and error codes: https://slip2go.com/guide/response
-   Account information: https://slip2go.com/guide/rest-api/account-info

Always treat the official documentation as the source of truth if it
changes. If this file conflicts with the current official docs, stop and
report the discrepancy before coding.
