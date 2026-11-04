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

# Black-box Test Plan

## EPIC 1 - User Management (manual)

### User story 1-1 Register an account

> As a new user, I want to register an account so that I can start using the platform.

**Input equivalence classes**

| ID | Input | Validation |
|---|---|---|
| 1 | Role is Customer | Valid |
| 2 | Role is Artist | Valid |
| 3 | Role is not selected | Invalid |
| 4 | Email is empty | Invalid |
| 5 | Email is not in a valid email format (e.g. `john.gmail.com`) | Invalid |
| 6 | Email is in a valid format and not yet registered | Valid |
| 7 | Email is in a valid format and already registered | Invalid |
| 8 | Username is empty | Invalid |
| 9 | Username contains only spaces | Invalid |
| 10 | Username has 1 to 2 characters | Invalid |
| 11 | Username has 3 to 20 characters | Valid |
| 12 | Username has more than 20 characters | Invalid |
| 13 | Password is empty | Invalid |
| 14 | Password has 1 to 7 characters | Invalid |
| 15 | Password has 8 to 16 characters | Valid |
| 16 | Password has more than 16 characters | Invalid |
| 17 | Confirm password is the same as Password | Valid |
| 18 | Confirm password is different from Password | Invalid |
| 19 | Bank is not selected | Invalid |
| 20 | Bank is one of the 10 listed banks | Valid |
| 21 | Account holder name is empty | Invalid |
| 22 | Account holder name contains only spaces | Invalid |
| 23 | Account holder name contains at least one non-space character | Valid |
| 24 | Account number is empty | Invalid |
| 25 | Account number contains a non-digit character | Invalid |
| 26 | Account number has 1 to 7 digits | Invalid |
| 27 | Account number has 8 to 16 digits | Valid |
| 28 | Account number has more than 16 digits | Invalid |
| 29 | Terms of Service and Privacy Policy are not accepted | Invalid |
| 30 | Terms of Service and Privacy Policy are accepted | Valid |

**Precondition**: the user is not logged in and has no account. The emails `john@gmail.com`, `mei@gmail.com` and `max@gmail.com` are not registered; `jane@artmission.test` is registered.

**Valid input set V** (used by every test case below, which then changes one input):

| Step | Field | Value |
|---|---|---|
| 1 | Role | Customer |
| 2 | Email | `john@gmail.com` |
| 2 | Username | `John` |
| 2 | Password | `john12345` |
| 2 | Confirm password | `john12345` |
| 3 | Bank | `ธนาคารกสิกรไทย (KBANK)` |
| 3 | Account holder name | `John Doe` |
| 3 | Account number | `1234567890` |
| 3 | Terms checkbox | checked |

**Test cases**

Action for every test case: enter the fields of each step and click **Next** (steps 1 and 2) and **Create account** (step 3), until the flow stops.

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC1-1-1 | V (unchanged) | Screen `Account created!` with text `Your account has been created successfully. Please log in to continue.` and button **Log in**; clicking **Log in** opens the Login page | Pending run | 1, 6, 11, 15, 17, 20, 23, 27, 30 |
| TC1-1-2 | Role: Artist<br>Email: `mei@gmail.com`<br>Username: `Mei` (3 characters)<br>Password: `mei12345` (8 characters)<br>Confirm password: `mei12345`<br>Bank: `ธนาคารไทยพาณิชย์ (SCB)`<br>Account holder name: `Mei Tan`<br>Account number: `12345678` (8 digits)<br>Terms checkbox: checked | Same as TC1-1-1 | Pending run | 2, 6, 11, 15, 17, 20, 23, 27, 30 |
| TC1-1-3 | Role: Customer<br>Email: `max@gmail.com`<br>Username: `maximumlengthuser123` (20 characters)<br>Password: `max1234567890123` (16 characters)<br>Confirm password: `max1234567890123`<br>Bank: `United Overseas Bank (UOB)`<br>Account holder name: `Max Maximum`<br>Account number: `1234567890123456` (16 digits)<br>Terms checkbox: checked | Same as TC1-1-1 | Pending run | 1, 6, 11, 15, 17, 20, 23, 27, 30 |
| TC1-1-4 | Role: *(none selected)* | Stays on step 1; text `Please select your role.` is shown under the role cards; no account is created | Pending run | 3 |
| TC1-1-5 | V, except Email: *(empty)* | Stays on step 2; text `Please enter your email address.` is shown under Email; no account is created | Pending run | 1, 4, 11, 15, 17 |
| TC1-1-6 | V, except Email: `john.gmail.com` | Stays on step 2; text `Please enter a valid email address.` is shown under Email; no account is created | Pending run | 1, 5, 11, 15, 17 |
| TC1-1-7 | V, except Email: `jane@artmission.test` | After **Create account**, the page returns to step 2; text `This email is already registered.` is shown under Email; no `Account created!` screen | Pending run | 1, 7, 11, 15, 17, 20, 23, 27, 30 |
| TC1-1-8 | V, except Username: *(empty)* | Stays on step 2; text `Please enter your username.` is shown under Username; no account is created | Pending run | 1, 6, 8, 15, 17 |
| TC1-1-9 | V, except Username: `   ` (3 spaces) | Same as TC1-1-8 | Pending run | 1, 6, 9, 15, 17 |
| TC1-1-10 | V, except Username: `Jo` (2 characters) | Stays on step 2; text `Username must be at least 3 characters.` is shown under Username; no account is created | Pending run | 1, 6, 10, 15, 17 |
| TC1-1-11 | V, except Username: `maximumlengthuser1234` (21 characters) | Stays on step 2; text `Username must be 20 characters or less.` is shown under Username; no account is created | Pending run | 1, 6, 12, 15, 17 |
| TC1-1-12 | V, except Password: *(empty)* and Confirm password: *(empty)* | Stays on step 2; text `Please enter a password.` is shown under Password; no account is created | Pending run | 1, 6, 11, 13, 17 |
| TC1-1-13 | V, except Password: `john123` (7 characters) and Confirm password: `john123` | Stays on step 2; text `Password must be at least 8 characters.` is shown under Password; no account is created | Pending run | 1, 6, 11, 14, 17 |
| TC1-1-14 | V, except Password: `john1234567890123` (17 characters) and Confirm password: `john1234567890123` | Stays on step 2; text `Password must be 16 characters or less.` is shown under Password; no account is created | Pending run | 1, 6, 11, 16, 17 |
| TC1-1-15 | V, except Confirm password: `john54321` | Stays on step 2; text `Passwords do not match.` is shown under Confirm password; no account is created | Pending run | 1, 6, 11, 15, 18 |
| TC1-1-16 | V, except Bank: *(not selected)* | Stays on step 3; text `Please select your bank.` is shown under Bank; no account is created | Pending run | 1, 6, 11, 15, 17, 19, 23, 27, 30 |
| TC1-1-17 | V, except Account holder name: *(empty)* | Stays on step 3; text `Please enter the account holder's name.` is shown under Account holder name; no account is created | Pending run | 1, 6, 11, 15, 17, 20, 21, 27, 30 |
| TC1-1-18 | V, except Account holder name: `   ` (3 spaces) | Same as TC1-1-17 | Pending run | 1, 6, 11, 15, 17, 20, 22, 27, 30 |
| TC1-1-19 | V, except Account number: *(empty)* | Stays on step 3; text `Please enter your bank account number.` is shown under Account number; no account is created | Pending run | 1, 6, 11, 15, 17, 20, 23, 24, 30 |
| TC1-1-20 | V, except Account number: `12345abc90` | Stays on step 3; text `Please enter a valid bank account number.` is shown under Account number; no account is created | Pending run | 1, 6, 11, 15, 17, 20, 23, 25, 30 |
| TC1-1-21 | V, except Account number: `1234567` (7 digits) | Same as TC1-1-20 | Pending run | 1, 6, 11, 15, 17, 20, 23, 26, 30 |
| TC1-1-22 | V, except Account number: `12345678901234567` (17 digits) | Same as TC1-1-20 | Pending run | 1, 6, 11, 15, 17, 20, 23, 28, 30 |
| TC1-1-23 | V, except Terms checkbox: not checked | Stays on step 3; text `Please agree to the Terms of Service and Privacy Policy to continue.` is shown under the checkbox; no account is created | Pending run | 1, 6, 11, 15, 17, 20, 23, 27, 29 |

---

### User story 1-2 Save bank account information

> As an authenticated user, I want to save my bank account information so that I can send or receive payments through the system.

**Input equivalence classes**

| ID | Input | Validation |
|---|---|---|
| 1 | Bank is one of the 10 listed banks | Valid |
| 2 | Account holder name contains at least one non-space character | Valid |
| 3 | Account holder name is empty | Invalid |
| 4 | Account holder name contains only spaces | Invalid |
| 5 | Account number is empty | Invalid |
| 6 | Account number contains a non-digit character | Invalid |
| 7 | Account number has 1 to 7 digits | Invalid |
| 8 | Account number has 8 to 16 digits | Valid |
| 9 | Account number has more than 16 digits | Invalid |
| 10 | User clicks **Done** | Valid |
| 11 | User clicks **Cancel** after editing | Valid |

**Precondition**: customer john@gmail.com is logged in, has opened the Profile page and clicked **Edit** on **User Bank Account Info**. Saved bank account before TC1-2-1: bank `ธนาคารกสิกรไทย (KBANK)`, holder `John Doe`, number `1234567890` (`••••7890`).

**Test cases**

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC1-2-1 | Name: `ธนาคารไทยพาณิชย์ (SCB)`<br>Account holder name: `John Smith`<br>Account holder number: `9876543210`<br>Action: click **Done** | Form closes; card shows Bank `ธนาคารไทยพาณิชย์ (SCB)`, Account holder name `John Smith`, Account holder number `••••3210` | Pending run | 1, 2, 8, 10 |
| TC1-2-2 | Name: `ธนาคารกสิกรไทย (KBANK)`<br>Account holder name: *(cleared)*<br>Account holder number: `9876543210`<br>Action: click **Done** | Form stays open; text `Please enter the account holder's name.` is shown under Account holder name; nothing is saved | Pending run | 1, 3, 8, 10 |
| TC1-2-3 | Name: `ธนาคารกสิกรไทย (KBANK)`<br>Account holder name: `   ` (3 spaces)<br>Account holder number: `9876543210`<br>Action: click **Done** | Same as TC1-2-2 | Pending run | 1, 4, 8, 10 |
| TC1-2-4 | Name: `ธนาคารกสิกรไทย (KBANK)`<br>Account holder name: `John Doe`<br>Account holder number: *(empty)*<br>Action: click **Done** | Form stays open; text `Please enter your bank account number.` is shown under Account holder number; nothing is saved | Pending run | 1, 2, 5, 10 |
| TC1-2-5 | Name: `ธนาคารกสิกรไทย (KBANK)`<br>Account holder name: `John Doe`<br>Account holder number: `12345abc90`<br>Action: click **Done** | Form stays open; text `Please enter a valid bank account number.` is shown under Account holder number; nothing is saved | Pending run | 1, 2, 6, 10 |
| TC1-2-6 | Name: `ธนาคารกสิกรไทย (KBANK)`<br>Account holder name: `John Doe`<br>Account holder number: `1234567` (7 digits)<br>Action: click **Done** | Same as TC1-2-5 | Pending run | 1, 2, 7, 10 |
| TC1-2-7 | Name: `ธนาคารกสิกรไทย (KBANK)`<br>Account holder name: `John Doe`<br>Account holder number: `12345678901234567` (17 digits)<br>Action: click **Done** | Same as TC1-2-5 | Pending run | 1, 2, 9, 10 |
| TC1-2-8 | Name: `ธนาคารไทยพาณิชย์ (SCB)`<br>Account holder name: `Changed Name`<br>Account holder number: `1111111111`<br>Action: click **Cancel** | Form closes; card shows the values saved before **Edit** was clicked, not the typed values | Pending run | 1, 2, 8, 11 |

---

### User story 1-4 Log in to an account

> As a registered user, I want to log in to my account so that I can access features associated with my account.

**Input equivalence classes**

| ID | Input | Validation |
|---|---|---|
| 1 | Email is registered and belongs to a customer account | Valid |
| 2 | Email is registered and belongs to an artist account | Valid |
| 3 | Email is in a valid format but not registered | Invalid |
| 4 | Email is not in a valid email format | Invalid |
| 5 | Email is empty | Invalid |
| 6 | Password is the correct password of the account with that Email | Valid |
| 7 | Password is not the correct password of the account with that Email | Invalid |
| 8 | Password is empty | Invalid |


**Precondition**: user is not logged in; the Login page is open; the accounts jane@artmission.test (customer) and ann@artmission.test (artist) exist, both with password `Password123!`.

**Test cases**

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC1-4-1 | Email: `jane@artmission.test`<br>Password: `Password123!`<br>Action: click **Log in** | Home page opens with heading `Browse artists`; the header navigation shows **Home**, **Order** and **Notification** | Pending run | 1, 6 |
| TC1-4-2 | Email: `ann@artmission.test`<br>Password: `Password123!`<br>Action: click **Log in** | The order history page opens (an artist is sent there after logging in); the header navigation shows **Order**, **Artist Profile** and **Notification** | Pending run | 2, 6 |
| TC1-4-3 | Email: `jane@artmission.test`<br>Password: `Password123` (wrong)<br>Action: click **Log in** | Text `Cannot login because the email or password is incorrect.` is shown; stays on the Login page | Pending run | 1, 7 |
| TC1-4-4 | Email: `nobody@artmission.test`<br>Password: `Password123!`<br>Action: click **Log in** | Same as TC1-4-3 | Pending run | 3 |
| TC1-4-5 | Email: `jane.artmission.test`<br>Password: `Password123!`<br>Action: click **Log in** | A browser validation message is shown at the Email field; stays on the Login page | Pending run | 4 |
| TC1-4-6 | Email: *(empty)*<br>Password: `Password123!` | **Log in** button is disabled, so the form cannot be submitted | Pending run | 5 |
| TC1-4-7 | Email: `jane@artmission.test`<br>Password: *(empty)* | **Log in** button is disabled, so the form cannot be submitted | Pending run | 1, 8 |

---

### User story 1-6 Delete an account

> As an authenticated user, I want to delete my account so that my account can no longer be used to access the platform.

**Input equivalence classes**

| ID | Input | Validation |
|---|---|---|
| 1 | Account has no orders | Valid |
| 2 | Account has orders, all in `SUCCESS` or `CANCEL` | Valid |
| 3 | Account has at least one order in `PENDING`, `NOT_PAID` or `IN_PROCESS` | Invalid |
| 4 | User clicks **Delete Account** and confirms in the popup | Valid |
| 5 | User clicks **Delete Account**, then **Cancel** in the popup | Valid |

**Precondition**: the account named in the test case is logged in and has opened the Profile page. `delete1@artmission.test` and `delete2@artmission.test` (password `Delete1234`, no orders) are registered beforehand.

**Test cases**

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC1-6-1 | Account: delete1@artmission.test (no orders)<br>Action: click **Delete Account**<br>Then: click **Delete Account** in the popup<br>Then: click **Close**<br>Then: on the Login page, log in with `delete1@artmission.test` / `Delete1234` | Popup `Account Deleted` with text `Your account has been successfully deleted.`; **Close** opens the Login page; the login attempt shows `Cannot login because the email or password is incorrect.` | Pending run | 1, 4 |
| TC1-6-2 | Account: cleo@artmission.test (only order is `CANCEL`)<br>Action: click **Delete Account**<br>Then: click **Delete Account** in the popup<br>Then: click **Close**<br>Then: on the Login page, log in with `cleo@artmission.test` / `Password123!` | Same as TC1-6-1 | Pending run | 2, 4 |
| TC1-6-3 | Account: jane@artmission.test (orders in `PENDING`, `NOT_PAID`, `IN_PROCESS`)<br>Action: click **Delete Account**<br>Then: click **Delete Account** in the popup<br>Then: click **Close** | Popup `Unable to Delete Account` with text `Your account cannot be deleted at the moment, as there are still ongoing orders.`; **Close** closes the popup and the Profile page is still shown | Pending run | 3, 4 |
| TC1-6-4 | Account: delete2@artmission.test (no orders)<br>Action: click **Delete Account**<br>Then: click **Cancel** in the popup | Popup closes; the Profile page is still shown; the account is not deleted | Pending run | 1, 5 |

---

## EPIC 6 - Order (automated with Playwright)

### User story 6-1 Customer places an order

> As a customer, I want to place an order so that the artist can accept the order.

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

**Precondition**: customer jane@artmission.test is logged in and has opened the order page of "Custom Portrait", so the minimum deadline M = 5 days. E is the earliest accepted deadline, today + M + 1 = today + 6 days (the date the Deadline Date field is pre-filled with).

**Test cases**

| Test Case ID | Inputs | Expected Output | Actual Output | Classes Covered |
|---|---|---|---|---|
| TC6-1-1 | Order Name: `Portrait for my cat`<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Order is created and the order detail page opens, showing the order name and status `PENDING` | Same as Expected Output. Screenshot: `TC6-1-1.png` | 1, 4, 7 |
| TC6-1-2 | Order Name: *(empty)*<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Text `Please enter the order name and request details.` is shown; stays on the order page; no order is created | Same as Expected Output. Screenshot: `TC6-1-2.png` | 2, 4, 7 |
| TC6-1-3 | Order Name: `   ` (3 spaces)<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Same as TC6-1-2 | Same as Expected Output. Screenshot: `TC6-1-3.png` | 3, 4, 7 |
| TC6-1-4 | Order Name: `Portrait for my cat`<br>Request Details: *(empty)*<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Same as TC6-1-2 | Same as Expected Output. Screenshot: `TC6-1-4.png` | 1, 5, 7 |
| TC6-1-5 | Order Name: `Portrait for my cat`<br>Request Details: `   ` (3 spaces)<br>Deadline Date: today + 6 days (E)<br>Action: click **Create Order** | Same as TC6-1-2 | Same as Expected Output. Screenshot: `TC6-1-5.png` | 1, 6, 7 |
| TC6-1-6 | Order Name: `Portrait for my cat`<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: today + 5 days (E − 1 day)<br>Action: click **Create Order** | Text `Deadline must be at least 5 days from today.` is shown; no order is created | Same as Expected Output. Screenshot: `TC6-1-6.png` | 1, 4, 8 |
| TC6-1-7 | Order Name: `Portrait for my cat`<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: yesterday (today − 1 day)<br>Action: click **Create Order** | Text `Deadline must be at least 5 days from today.` is shown; no order is created | Same as Expected Output. Screenshot: `TC6-1-7.png` | 1, 4, 9 |
| TC6-1-8 | Order Name: `Portrait for my cat`<br>Request Details: `Black and white cat, watercolor style`<br>Deadline Date: *(cleared)*<br>Action: click **Create Order** | Text `Deadline must be at least 5 days from today.` is shown; no order is created | Same as Expected Output. Screenshot: `TC6-1-8.png` | 1, 4, 10 |

---

### User story 6-2 Artist confirms an order

> As an artist, I want to confirm an order, so that I can get hired for a commission.

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
| TC6-2-1 | Order status: `PENDING`<br>Action: click **Accept**<br>Then: click **Accept** in popup `Accept this Order?` | Popup closes; status badge shows `NOT PAID`; Accept and Reject buttons are no longer shown | Same as Expected Output. Screenshot: `TC6-2-1.png` | 1, 3 |
| TC6-2-2 | Order status: `PENDING`<br>Action: click **Reject**<br>Then: click **Reject** in popup `Reject this Order?` | Popup closes; status badge shows `CANCEL`; Accept and Reject buttons are no longer shown | Same as Expected Output. Screenshot: `TC6-2-2.png` | 1, 4 |
| TC6-2-3 | Order status: `PENDING`<br>Action: click **Accept**<br>Then: click **Cancel** in popup `Accept this Order?` | Popup closes; status badge still `PENDING`; Accept and Reject buttons still shown | Same as Expected Output. Screenshot: `TC6-2-3.png` | 1, 5 |
| TC6-2-4 | Order status: `PENDING`<br>Action: click **Reject**<br>Then: click **Cancel** in popup `Reject this Order?` | Popup closes; status badge still `PENDING`; Accept and Reject buttons still shown | Same as Expected Output. Screenshot: `TC6-2-4.png` | 1, 5 |
| TC6-2-5 | Order status: `NOT_PAID`<br>Action: open the order | Accept and Reject buttons are not shown | Same as Expected Output. Screenshot: `TC6-2-5.png` | 2 |
| TC6-2-6 | Order status: `IN_PROCESS`<br>Action: open the order | Accept and Reject buttons are not shown | Same as Expected Output. Screenshot: `TC6-2-6.png` | 2 |
| TC6-2-7 | Order status: `SUCCESS`<br>Action: open the order | Accept and Reject buttons are not shown | Same as Expected Output. Screenshot: `TC6-2-7.png` | 2 |
| TC6-2-8 | Order status: `CANCEL`<br>Action: open the order | Accept and Reject buttons are not shown | Same as Expected Output. Screenshot: `TC6-2-8.png` | 2 |

---

### User story 6-6 Artist requests a deadline extension

> As an artist, I want to request a deadline extension for an order so that I can secure extra time to finish the work if I am running late.

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
| TC6-6-1 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: `I need more time for details`<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Popup closes; banner `Postpone Request Pending` is shown with text `You have requested to extend the deadline to <today + 21 days as dd Mon yyyy>. Waiting for the customer's response (Auto declines in 48 hours).`; **Postpone Order** button is disabled | Same as Expected Output. Screenshot: `TC6-6-1.png` | 1, 5, 7, 11 |
| TC6-6-2 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 14 days (D, same date as the current deadline)<br>Reason for Postponement: `I need more time for details`<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Request is not sent: popup stays open, a browser validation message is shown at the New Proposed Deadline field, and no banner appears | Same as Expected Output. Screenshot: `TC6-6-2.png` (the browser's validation message is not part of the page, so it is not visible in the image) | 2, 5, 7, 11 |
| TC6-6-3 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 13 days (D − 1 day)<br>Reason for Postponement: `I need more time for details`<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Request is not sent: popup stays open, a browser validation message is shown at the New Proposed Deadline field, and no banner appears | Same as Expected Output. Screenshot: `TC6-6-3.png` (the browser's validation message is not visible in the image) | 3, 5, 7, 11 |
| TC6-6-4 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: *(empty)*<br>Reason for Postponement: `I need more time for details`<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Request is not sent: popup stays open, a browser validation message is shown at the New Proposed Deadline field, and no banner appears | Same as Expected Output. Screenshot: `TC6-6-4.png` (the browser's validation message is not visible in the image) | 4, 5, 7, 11 |
| TC6-6-5 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: *(not selected)*<br>Explanation: `Need extra time to refine the shading.`<br>Then: click **Send Request** | Alert `กรุณาเลือกวันหลัง Deadline เดิม พร้อมเหตุผลและคำอธิบาย` is shown; popup stays open; no banner | Same as Expected Output. Screenshot: `TC6-6-5.png` | 1, 6, 7, 11 |
| TC6-6-6 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: `I need more time for details`<br>Explanation: *(empty)*<br>Then: click **Send Request** | Request is not sent: popup stays open, a browser validation message is shown at the Explanation field, and no banner appears | Same as Expected Output. Screenshot: `TC6-6-6.png` (the browser's validation message is not visible in the image) | 1, 5, 8, 11 |
| TC6-6-7 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: `I need more time for details`<br>Explanation: `   ` (3 spaces)<br>Then: click **Send Request** | Alert `กรุณาเลือกวันหลัง Deadline เดิม พร้อมเหตุผลและคำอธิบาย` is shown; popup stays open; no banner | Same as Expected Output. Screenshot: `TC6-6-7.png` | 1, 5, 9, 11 |
| TC6-6-8 | Order status: `IN_PROCESS`<br>Action: click **Postpone Order**<br>New Proposed Deadline: today + 21 days (D + 7 days)<br>Reason for Postponement: `I need more time for details`<br>Explanation: 301 characters (the letter `A` repeated 301 times) | Counter shows `300/300`; field contains the first 300 characters only | Same as Expected Output. Screenshot: `TC6-6-8.png` | 1, 5, 10, 11 |
| TC6-6-9 | Order status: `NOT_PAID`<br>Action: open the order | **Postpone Order** button is not shown | Same as Expected Output. Screenshot: `TC6-6-9.png` | 12 |
| TC6-6-10 | Order status: `IN_PROCESS`, with a postponement request awaiting the customer's decision<br>Action: open the order | Banner `Postpone Request Pending` is shown and the **Postpone Order** button is disabled | Same as Expected Output. Screenshot: `TC6-6-10.png` | 13 |

---

### User story 6-7 Artist submits the work for an order

> As an artist, I want to submit the work for an order so that the customer can verify and accept the artwork.

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
| TC6-7-1 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.jpg` (valid JPEG, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Popup closes; a `#1 Draft` card appears with badge `Waiting for Customer Review`; the **Submit Draft** section is no longer shown | Same as Expected Output. Screenshot: `TC6-7-1.png` | 1, 4, 9, 12 |
| TC6-7-2 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.png` (valid PNG, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Same as TC6-7-1 | Same as Expected Output. Screenshot: `TC6-7-2.png` | 1, 5, 9, 12 |
| TC6-7-3 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.webp` (valid WebP, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Same as TC6-7-1 | Same as Expected Output. Screenshot: `TC6-7-3.png` | 1, 6, 9, 12 |
| TC6-7-4 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.gif`<br>Action: choose the file | Alert `เลือกไฟล์ JPG, PNG หรือ WebP เท่านั้น` is shown; no preview and no **Submit** button | Same as Expected Output. Screenshot: `TC6-7-4.png` | 1, 7 |
| TC6-7-5 | Order status: `IN_PROCESS` (no drafts)<br>File: *(none chosen)*<br>Action: open the order | Upload area `Add your artwork` is shown; no **Submit** button | Same as Expected Output. Screenshot: `TC6-7-5.png` | 1, 8 |
| TC6-7-6 | Order status: `IN_PROCESS` (no drafts)<br>File: `fake.png` (a text file renamed to `.png`, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Popup stays open with an error alert; no draft card is created. Expected text `deliverable image must be a JPEG, PNG, or WebP image` | Same as Expected Output. Screenshot: `TC6-7-6.png` | 1, 5, 9, 11, 12 |
| TC6-7-7 | Order status: `IN_PROCESS` (no drafts)<br>File: `big.png` (valid PNG image of 21 MiB)<br>Action: click **Submit**<br>Then: click **Submit** in popup `Submit Draft to Customer?` | Popup stays open with an error alert; no draft card is created. Expected text `deliverable image exceeds the 20 MiB limit` | Same as Expected Output. Screenshot: `TC6-7-7.png` | 1, 5, 10, 12 |
| TC6-7-8 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.jpg` (valid JPEG, under 1 MiB)<br>Action: click **Submit**<br>Then: click **Cancel** in popup `Submit Draft to Customer?` | Popup closes; preview and **Submit** button still shown; no draft card is created | Same as Expected Output. Screenshot: `TC6-7-8.png` | 1, 4, 9, 13 |
| TC6-7-9 | Order status: `IN_PROCESS` (no drafts)<br>File: `draft.jpg` (valid JPEG, under 1 MiB)<br>Action: click ✕ on the preview | Preview disappears; upload area `Add your artwork` shown again; no **Submit** button | Same as Expected Output. Screenshot: `TC6-7-9.png` | 1, 4, 9, 14 |
| TC6-7-10 | Order status: `NOT_PAID`<br>Action: open the order | **Submit Draft** section is not shown | Same as Expected Output. Screenshot: `TC6-7-10.png` | 2 |
| TC6-7-11 | Order status: `IN_PROCESS`, latest draft awaiting customer review<br>Action: open the order | **Submit Draft** section is not shown; the draft still shows `Waiting for Customer Review` | Same as Expected Output. Screenshot: `TC6-7-11.png` | 3 |

---

# Test Coverage Matrix

| User Story ID | Test Cases ID | Status |
|---|---|---|
| 1-1 | TC1-1-1, TC1-1-2, TC1-1-3, TC1-1-4, TC1-1-5, TC1-1-6, TC1-1-7, TC1-1-8, TC1-1-9, TC1-1-10, TC1-1-11, TC1-1-12, TC1-1-13, TC1-1-14, TC1-1-15, TC1-1-16, TC1-1-17, TC1-1-18, TC1-1-19, TC1-1-20, TC1-1-21, TC1-1-22, TC1-1-23 | P, P, P, P, P, P, P, P, P, P, P, P, P, P, P, P, P, P, P, P, P, P, P |
| 1-2 | TC1-2-1, TC1-2-2, TC1-2-3, TC1-2-4, TC1-2-5, TC1-2-6, TC1-2-7, TC1-2-8 | P, P, P, P, P, P, P, P |
| 1-4 | TC1-4-1, TC1-4-2, TC1-4-3, TC1-4-4, TC1-4-5, TC1-4-6, TC1-4-7 | P, P, P, P, P, P, P |
| 1-6 | TC1-6-1, TC1-6-2, TC1-6-3, TC1-6-4 | P, P, P, P |
| 6-1 | TC6-1-1, TC6-1-2, TC6-1-3, TC6-1-4, TC6-1-5, TC6-1-6, TC6-1-7, TC6-1-8 | P, P, P, P, P, P, P, P |
| 6-2 | TC6-2-1, TC6-2-2, TC6-2-3, TC6-2-4, TC6-2-5, TC6-2-6, TC6-2-7, TC6-2-8 | P, P, P, P, P, P, P, P |
| 6-6 | TC6-6-1, TC6-6-2, TC6-6-3, TC6-6-4, TC6-6-5, TC6-6-6, TC6-6-7, TC6-6-8, TC6-6-9, TC6-6-10 | P, P, P, P, P, P, P, P, P, P |
| 6-7 | TC6-7-1, TC6-7-2, TC6-7-3, TC6-7-4, TC6-7-5, TC6-7-6, TC6-7-7, TC6-7-8, TC6-7-9, TC6-7-10, TC6-7-11 | P, P, P, P, P, P, P, P, P, P, P |
