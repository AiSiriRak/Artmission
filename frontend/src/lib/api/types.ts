import type { components } from "./v1";

// Auth
export type AuthResult = components["schemas"]["AuthResultBody"];
export type LoginInput = components["schemas"]["LoginInputBody"];
export type RegisterInput = components["schemas"]["RegisterInputBody"];

// User
export type UserAccount = components["schemas"]["AccountView"];
export type UpdateAccountInput =
  components["schemas"]["UpdateAccountInputBody"];

// Bank Account
export type BankAccount = components["schemas"]["BankAccountView"];
export type UpdateBankAccountInput =
  components["schemas"]["UpdateBankAccountInputBody"];

// Order
export type OrderHistory = components["schemas"]["ViewHiringHistoryOutputBody"];
export type Order = components["schemas"]["OrderView"];
