# Test plan working notes (NOT part of the PDF)

`test_plan.md` is exactly what gets exported to the PDF for the course. Everything that helps the team but is not in the course examples lives here: conventions, test data, and the description of each page under test.

## Conventions

- **Test Case ID**: `TC<epic>-<story>-<n>`, e.g. `TC6-1-3` is the 3rd test case of user story 6-1.
- **Equivalence classes** are numbered per user story, starting at 1.
- Each test case changes **one invalid input at a time**; the valid-path case covers all valid classes at once. **Classes Covered** lists every class the test case's inputs belong to, including the valid classes of the inputs that are not the one under test (same as the black-box example in the course material).
- **Actual Output** is filled in after the run (EPIC 6: result of the automated test with the same Test Case ID; EPIC 1: what the tester observed).
- Dates are relative to the run date (`today + N days`), so the test cases stay valid over time.
- Some invalid inputs are stopped by the browser's built-in form validation, which shows a message next to the field. The wording of that message is set by the browser, so Expected Output only says that a message is shown at the field, and the exact text seen in the test browser (Chromium) is recorded in Actual Output.
- EPIC 1 registration (1-1): the flow stops at the first step with an error, so the fields of later steps are not entered in those test cases and **Classes Covered** lists only the fields that were entered.

## Test data (EPIC 1)

| Account | Email | Password | Role | Used in |
|---|---|---|---|---|
| Jane | jane@artmission.test | `Password123!` | Customer | 1-1 (email already registered), 1-4, 1-6. Has orders in `PENDING`, `NOT_PAID`, `IN_PROCESS`, `SUCCESS` and `CANCEL` |
| Ann | ann@artmission.test | `Password123!` | Artist | 1-4 |
| Cleo | cleo@artmission.test | `Password123!` | Artist | 1-6. Her only order is `CANCEL` |
| John | john@gmail.com | `john12345` | Customer | 1-2. Registered with bank `ธนาคารกสิกรไทย (KBANK)`, holder `John Doe`, account number `1234567890` (this is the account created by TC1-1-1) |
| Delete1, Delete2 | delete1@artmission.test, delete2@artmission.test | `Delete1234` | Customer | 1-6. Registered through the registration page before the run; no orders |

Every account exists, in the state described in the table (orders, bank account), before the run.

## Test data (EPIC 6)

| Role | Account | Notes |
|---|---|---|
| Customer | jane@artmission.test | places orders |
| Artist | ann@artmission.test | owns the artwork "Custom Portrait" (price 1,500 THB, minimum deadline 5 days) |
| Artist | bo@artmission.test | |

All passwords are `Password123!`. Order statuses are `PENDING`, `NOT_PAID`, `IN_PROCESS`, `SUCCESS`, `CANCEL`. Test cases that need an order in a given status get their own order in that status, so no test case depends on the result of another.

## Pages under test

### 1-1 Register an account

Registration page (`/register`, reached from **Create Account** on the Login page). Registration has 3 steps; the fields of one step are validated when **Next** (steps 1 and 2) or **Create account** (step 3) is clicked, and the flow stops at the first step that has an error.

- Step 1 **Choose your role**: **Artist** or **Customer**.
- Step 2 **Account information**: **Email**, **Username**, **Password**, **Confirm password**.
- Step 3 **Bank account**: **Bank** (10 banks, e.g. `ธนาคารกสิกรไทย (KBANK)`), **Account holder name**, **Account number**, checkbox `I agree to the Terms of Service and Privacy Policy`.

The 10 banks are treated as one class.

### 1-2 Save bank account information

Profile page (header menu → **Profile**), card **User Bank Account Info**. It shows **Bank**, **Account holder name** and **Account holder number** (the account number, shown as `••••` plus its last 4 digits). **Edit** opens the form: **Name** (the bank, a dropdown with the 10 banks and no empty choice, pre-selected with the saved bank), **Account holder name** (pre-filled with the saved name), **Account holder number** (always starts empty; the saved number is never shown), and buttons **Cancel** and **Done**.

### 1-4 Log in to an account

Login page (`/login`) with fields **Email** and **Password** and button **Log in**. The button is disabled while Email or Password is empty. The Email field is a browser email field, so a text that is not an email address is stopped by the browser's validation message when **Log in** is clicked.

### 1-6 Delete an account

Profile page, card **Delete Account** with button **Delete Account**. The button opens a popup `Delete Account` with **Cancel** and **Delete Account**. After confirming, the popup shows either `Account Deleted` (button **Close**, which opens the Login page) or `Unable to Delete Account` (button **Close**, which closes the popup).

An account cannot be deleted while it has an order in `PENDING`, `NOT_PAID` or `IN_PROCESS`. Orders in `SUCCESS` or `CANCEL` do not block deletion.

### 6-1 Customer places an order

Order page of an artwork, with fields **Order Name**, **Request Details**, **Deadline Date** and button **Create Order**. The page shows `Minimum M days required`, where **M** is the artwork's minimum deadline (M = 5 for "Custom Portrait"). **E** is the earliest accepted deadline: the date the Deadline Date field is pre-filled with, today + M + 1 days (today + 6 days for "Custom Portrait").

### 6-2 Artist confirms an order

Order detail page as the artist. Buttons **Reject** / **Accept** are shown for `PENDING` orders; each opens a confirmation popup with **Cancel** and a confirm button.

### 6-6 Artist requests a deadline extension

Order detail page as the artist. Button **Postpone Order** (shown for `IN_PROCESS` orders) opens popup `Request to Postpone Order?` with **New Proposed Deadline** (date), **Reason for Postponement** (select), **Please explain your reason** (text, up to 300 characters) and button **Send Request**.

### 6-7 Artist submits the work for an order

Order detail page as the artist. Section **Submit Draft** where an image is chosen. After a file is chosen a preview and a **Submit** button appear; **Submit** opens popup `Submit Draft to Customer?`. The ✕ button on the preview removes the chosen file.
