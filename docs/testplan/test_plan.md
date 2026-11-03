# Test plan Artmission

## แมวชุบแป้งทอด (Group xx)

| Name | Student ID | Role |
|---|---|---|
| Navaporn Homjandee | 6730265821 | TODO |
| Napas Ruttanapunyagorn | 6730262921 | TODO |
| Kochakorn Srichay | 6731301921 | TODO |
| Pumwaree Pipithsuksunt | 6731338121 | TODO |
| Siripudsorn Raksutakan | 6731353521 | TODO |
| Thirada Thomnam | 6732015621 | TODO |
| Peeravas Piboolvorakul | 6732033921 | TODO |
| Ashira Aungsumal | 6732041921 | TODO |

TODO: indicate one leader and each member's contribution to this document.

2110423 Software Engineering, 17 November 2026
Department of Computer Engineering, Faculty of Engineering, Chulalongkorn University

---

## Conventions

- **Test Case ID**: `TC<epic>-<story>-<n>`, e.g. `TC6-1-3` is the 3rd test case of user story 6-1.
- **Equivalence classes** are numbered per user story, starting at 1.
- Each test case changes **one invalid input at a time**; the valid-path case covers all valid classes at once. **Classes Covered** lists every class the test case's inputs belong to, including the valid classes of the inputs that are not the one under test (same as the black-box example in the course material).
- **Actual Output** is filled in after the run (EPIC 6: result of the automated test with the same Test Case ID; EPIC 1: what the tester observed).
- Dates are relative to the run date (`today + N days`), so the test cases stay valid over time.
- Some invalid inputs are stopped by the browser's built-in form validation, which shows a message next to the field. The wording of that message is set by the browser, so Expected Output only says that a message is shown at the field, and the exact text seen in the test browser (Chromium) is recorded in Actual Output.

### Test data (EPIC 6)

| Role | Account | Notes |
|---|---|---|
| Customer | jane@artmission.test | places orders |
| Artist | ann@artmission.test | owns the artwork "Custom Portrait" (price 1,500 THB, minimum deadline 5 days) |
| Artist | bo@artmission.test | |

All passwords are `Password123!`. Order statuses are `PENDING`, `NOT_PAID`, `IN_PROCESS`, `SUCCESS`, `CANCEL`. Test cases that need an order in a given status get their own order in that status, so no test case depends on the result of another.

---

# Black-box Test Plan

## EPIC 6 - Order (automated with Playwright)

### User story 6-1 Customer places an order

> As a customer, I want to place an order so that the artist can accept the order.

Page: order page of an artwork, with fields **Order Name**, **Request Details**, **Deadline Date** and button **Create Order**. The page shows `Minimum M days required`, where **M** is the artwork's minimum deadline (M = 5 for "Custom Portrait"). **E** is the earliest accepted deadline: the date the Deadline Date field is pre-filled with, today + M + 1 days (today + 6 days for "Custom Portrait").

**Input equivalence classes**

| ID | Input | Validation |
|---|---|---|
| 1 | Order Name contains at least one non-space character | Valid |
| 2 | Order Name is empty | Invalid |
| 3 | Order Name contains only spaces | Invalid |
| 4 | Request Details contains at least one non-space character | Valid |
| 5 | Request Details is empty | Invalid |
| 6 | Request Details contains only spaces | Invalid |
| 7 | Deadline is E or later | Valid |
| 8 | Deadline is today or later but before E | Invalid |
| 9 | Deadline is in the past | Invalid |
| 10 | Deadline is empty | Invalid |

**Precondition**: customer jane@artmission.test is logged in and has opened the order page of "Custom Portrait", so M = 5 and E = today + 6 days.

**Test cases**

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC6-1-1 | Order Name: `Portrait for my cat`<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Order is created and the order detail page opens, showing the order name and status `PENDING` | Pending run | 1, 4, 7 |
| TC6-1-2 | Order Name: *(empty)*<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Text `Please enter the order name and request details.` is shown; stays on the order page; no order is created | Pending run | 2, 4, 7 |
| TC6-1-3 | Order Name: `   ` (3 spaces)<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Same message as TC6-1-2; no order is created | Pending run | 3, 4, 7 |
| TC6-1-4 | Order Name: `Portrait for my cat`<br>Request Details: *(empty)*<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Same message as TC6-1-2; no order is created | Pending run | 1, 5, 7 |
| TC6-1-5 | Order Name: `Portrait for my cat`<br>Request Details: `   ` (3 spaces)<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Same message as TC6-1-2; no order is created | Pending run | 1, 6, 7 |
| TC6-1-6 | Order Name: `Portrait for my cat`<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: today + 5 days (E − 1 day)<br>Action: click **Create Order** | Text `Deadline must be at least 5 days from today.` is shown; no order is created | Pending run | 1, 4, 8 |
| TC6-1-7 | Order Name: `Portrait for my cat`<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: yesterday (today − 1 day)<br>Action: click **Create Order** | Text `Deadline must be at least 5 days from today.` is shown; no order is created | Pending run | 1, 4, 9 |
| TC6-1-8 | Order Name: `Portrait for my cat`<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: *(cleared)*<br>Action: click **Create Order** | Text `Deadline must be at least 5 days from today.` is shown; no order is created | Pending run | 1, 4, 10 |

---

### User story 6-2 Artist confirms an order

> As an artist, I want to confirm an order, so that I can get hired for a commission.

Page: order detail page as the artist. Buttons **Reject** / **Accept** are shown for `PENDING` orders; each opens a confirmation popup with **Cancel** and a confirm button.

**Input equivalence classes**

| ID | Input | Validation |
|---|---|---|
| 1 | Order status is `PENDING` | Valid |
| 2 | Order status is not `PENDING` (`NOT_PAID`, `IN_PROCESS`, `SUCCESS` or `CANCEL`) | Invalid |
| 3 | Artist presses Accept and confirms in the popup | Valid |
| 4 | Artist presses Reject and confirms in the popup | Valid |
| 5 | Artist presses Accept or Reject, then presses Cancel in the popup | Valid |

**Precondition**: the artist who owns the order is logged in.

**Test cases**

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC6-2-1 | Order status: `PENDING`<br>Action: click **Accept**<br>Then: click **Accept** in popup `Accept this Order?` | Popup closes; status badge shows `NOT PAID`; Accept and Reject buttons are no longer shown | Pending run | 1, 3 |
| TC6-2-2 | Order status: `PENDING`<br>Action: click **Reject**<br>Then: click **Reject** in popup `Reject this Order?` | Popup closes; status badge shows `CANCEL`; Accept and Reject buttons are no longer shown | Pending run | 1, 4 |
| TC6-2-3 | Order status: `PENDING`<br>Action: click **Accept**<br>Then: click **Cancel** in popup `Accept this Order?` | Popup closes; status badge still `PENDING`; Accept and Reject buttons still shown | Pending run | 1, 5 |
| TC6-2-4 | Order status: `PENDING`<br>Action: click **Reject**<br>Then: click **Cancel** in popup `Reject this Order?` | Popup closes; status badge still `PENDING`; Accept and Reject buttons still shown | Pending run | 1, 5 |
| TC6-2-5 | Order status: `NOT_PAID`<br>Action: open the order | Accept and Reject buttons are not shown | Pending run | 2 |
| TC6-2-6 | Order status: `IN_PROCESS`<br>Action: open the order | Accept and Reject buttons are not shown | Pending run | 2 |
| TC6-2-7 | Order status: `SUCCESS`<br>Action: open the order | Accept and Reject buttons are not shown | Pending run | 2 |
| TC6-2-8 | Order status: `CANCEL`<br>Action: open the order | Accept and Reject buttons are not shown | Pending run | 2 |

---

### User story 6-6 Artist requests a deadline extension

> As an artist, I want to request a deadline extension for an order so that I can secure extra time to finish the work if I am running late.

Page: order detail page as the artist. Button **Postpone Order** (shown for `IN_PROCESS` orders) opens popup `Request to Postpone Order?` with **New Proposed Deadline** (date), **Reason for Postponement** (select), **Please explain your reason** (text, up to 300 characters) and button **Send Request**.

**Input equivalence classes**

| ID | Input | Validation |
|---|---|---|
| 1 | New deadline is later than the current deadline | Valid |
| 2 | New deadline is the same date as the current deadline | Invalid |
| 3 | New deadline is earlier than the current deadline | Invalid |
| 4 | New deadline is empty | Invalid |
| 5 | Reason is one of the 5 options (`I need more time for details`, `I am waiting for additional customer feedback`, `I have a personal emergency / illness`, `I am experiencing an unexpected high workload`, `Other`) | Valid |
| 6 | Reason is not selected | Invalid |
| 7 | Explanation has 1 to 300 characters (after trimming) | Valid |
| 8 | Explanation is empty | Invalid |
| 9 | Explanation contains only spaces | Invalid |
| 10 | Explanation entered is longer than 300 characters (field keeps the first 300) | Valid |
| 11 | Order status is `IN_PROCESS` and has no postponement request awaiting the customer | Valid |
| 12 | Order status is not `IN_PROCESS` | Invalid |
| 13 | Order is `IN_PROCESS` but already has a postponement request awaiting the customer's decision | Invalid |

**Precondition**: the artist who owns the order is logged in. For TC6-6-1 to TC6-6-8 the order's current deadline is today + 14 days (D).

**Test cases**

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC6-6-1 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: `I need more time for details`<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Popup closes; banner `Postpone Request Pending` is shown with text `You have requested to extend the deadline to <today + 21 days as dd Mon yyyy>. Waiting for the customer's response (Auto declines in 48 hours).`; **Postpone Order** button is disabled | Pending run | 1, 5, 7, 11 |
| TC6-6-2 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 14 days (D, same date as the current deadline)<br>Reason for Postponement: `I need more time for details`<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Request is not sent: popup stays open, a validation message is shown at the New Proposed Deadline field, and no banner appears | Pending run | 2, 5, 7, 11 |
| TC6-6-3 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 13 days (D − 1 day)<br>Reason for Postponement: `I need more time for details`<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Request is not sent: popup stays open, a validation message is shown at the New Proposed Deadline field, and no banner appears | Pending run | 3, 5, 7, 11 |
| TC6-6-4 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: *(empty)*<br>Reason for Postponement: `I need more time for details`<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Request is not sent: popup stays open, a validation message is shown at the New Proposed Deadline field, and no banner appears | Pending run | 4, 5, 7, 11 |
| TC6-6-5 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: *(not selected)*<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Alert `กรุณาเลือกวันหลัง Deadline เดิม พร้อมเหตุผลและคำอธิบาย` is shown; popup stays open; no banner | Pending run | 1, 6, 7, 11 |
| TC6-6-6 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: `I need more time for details`<br>Explanation: *(empty)*<br>Then: click **Send Request** | Request is not sent: popup stays open, a validation message is shown at the Explanation field, and no banner appears | Pending run | 1, 5, 8, 11 |
| TC6-6-7 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: `I need more time for details`<br>Explanation: `   ` (3 spaces)<br>Then: click **Send Request** | Alert `กรุณาเลือกวันหลัง Deadline เดิม พร้อมเหตุผลและคำอธิบาย` is shown; popup stays open; no banner | Pending run | 1, 5, 9, 11 |
| TC6-6-8 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: `I need more time for details`<br>Explanation: 301 characters (the letter `A` repeated 301 times) | Counter shows `300/300`; field contains the first 300 characters only | Pending run | 1, 5, 10, 11 |
| TC6-6-9 | Order status: `NOT_PAID`<br>Action: open the order | **Postpone Order** button is not shown | Pending run | 12 |
| TC6-6-10 | Order status: `IN_PROCESS`, with a postponement request awaiting the customer's decision<br>Action: open the order | Banner `Postpone Request Pending` is shown and the **Postpone Order** button is disabled | Pending run | 13 |

---

### User story 6-7 Artist submits the work for an order

> As an artist, I want to submit the work for an order so that the customer can verify and accept the artwork.

Page: order detail page as the artist. Section **Submit Draft** where an image is chosen. After a file is chosen a preview and a **Submit** button appear; **Submit** opens popup `Submit Draft to Customer?`. The ✕ button on the preview removes the chosen file.

**Input equivalence classes**

| ID | Input | Validation |
|---|---|---|
| 1 | Order is `IN_PROCESS` and its latest draft is not awaiting customer review | Valid |
| 2 | Order is not `IN_PROCESS` | Invalid |
| 3 | Order is `IN_PROCESS` but its latest draft is awaiting customer review | Invalid |
| 4 | File type is JPEG (`.jpg`) | Valid |
| 5 | File type is PNG (`.png`) | Valid |
| 6 | File type is WebP (`.webp`) | Valid |
| 7 | File type is not JPEG/PNG/WebP (e.g. `.gif`, `.pdf`) | Invalid |
| 8 | No file is chosen | Invalid |
| 9 | Image file is 20 MiB or smaller | Valid |
| 10 | Image file is larger than 20 MiB | Invalid |
| 11 | File has an image extension but its content is not a valid image | Invalid |
| 12 | Artist presses Submit and confirms in the popup | Valid |
| 13 | Artist presses Submit, then presses Cancel in the popup | Valid |
| 14 | Artist removes the chosen file with ✕ before submitting | Valid |

**Precondition**: the artist who owns the order is logged in.

**Test cases**

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC6-7-1 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.jpg` (valid JPEG, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Popup closes; a `#1 Draft` card appears with badge `Waiting for Customer Review`; the **Submit Draft** section is no longer shown | Pending run | 1, 4, 9, 12 |
| TC6-7-2 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.png` (valid PNG, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Same as TC6-7-1 | Pending run | 1, 5, 9, 12 |
| TC6-7-3 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.webp` (valid WebP, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Same as TC6-7-1 | Pending run | 1, 6, 9, 12 |
| TC6-7-4 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.gif`<br>Action: choose the file | Alert `เลือกไฟล์ JPG, PNG หรือ WebP เท่านั้น` is shown; no preview and no **Submit** button | Pending run | 1, 7 |
| TC6-7-5 | Order status: `IN_PROCESS` (no drafts)<br>File: *(none chosen)*<br>Action: open the order | Upload area `Add your artwork` is shown; no **Submit** button | Pending run | 1, 8 |
| TC6-7-6 | Order status: `IN_PROCESS` (no drafts)<br>File: `fake.png` (a text file renamed to `.png`, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Popup stays open with an error alert; no draft card is created. Expected text `deliverable image must be a JPEG, PNG, or WebP image` | Pending run | 1, 5, 9, 11, 12 |
| TC6-7-7 | Order status: `IN_PROCESS` (no drafts)<br>File: `big.png` (valid PNG image of 21 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Popup stays open with an error alert; no draft card is created. Expected text `deliverable image exceeds the 20 MiB limit` | Pending run | 1, 5, 10, 12 |
| TC6-7-8 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.jpg` (valid JPEG, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Cancel** in popup `Submit Draft to Customer?` | Popup closes; preview and **Submit** button still shown; no draft card is created | Pending run | 1, 4, 9, 13 |
| TC6-7-9 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.jpg` (valid JPEG, under 1 MiB)<br>Action: click ✕ on the preview | Preview disappears; upload area `Add your artwork` shown again; no **Submit** button | Pending run | 1, 4, 9, 14 |
| TC6-7-10 | Order status: `NOT_PAID`<br>Action: open the order | **Submit Draft** section is not shown | Pending run | 2 |
| TC6-7-11 | Order status: `IN_PROCESS`, latest draft awaiting customer review<br>Action: open the order | **Submit Draft** section is not shown; the draft still shows `Waiting for Customer Review` | Pending run | 3 |

---

## EPIC 1 - User Management (manual)

Design pending: 1-1 Register an account, 1-2 Save bank account information, 1-4 Log in to an account, 1-6 Delete an account.

---

# Test Coverage Matrix

| User Story ID | Test Cases ID | Status |
|---|---|---|
| 6-1 | TC6-1-1, TC6-1-2, TC6-1-3, TC6-1-4, TC6-1-5, TC6-1-6, TC6-1-7, TC6-1-8 | Not run |
| 6-2 | TC6-2-1, TC6-2-2, TC6-2-3, TC6-2-4, TC6-2-5, TC6-2-6, TC6-2-7, TC6-2-8 | Not run |
| 6-6 | TC6-6-1, TC6-6-2, TC6-6-3, TC6-6-4, TC6-6-5, TC6-6-6, TC6-6-7, TC6-6-8, TC6-6-9, TC6-6-10 | Not run |
| 6-7 | TC6-7-1, TC6-7-2, TC6-7-3, TC6-7-4, TC6-7-5, TC6-7-6, TC6-7-7, TC6-7-8, TC6-7-9, TC6-7-10, TC6-7-11 | Not run |
| 1-1, 1-2, 1-4, 1-6 | TBD | Not designed |
