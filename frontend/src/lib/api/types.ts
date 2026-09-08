import type { components } from "./v1";

// User
export type UserAccount = components["schemas"]["AccountView"];
export type UpdateAccountInput =
  components["schemas"]["UpdateAccountInputBody"];

// Bank Account
export type BankAccount = components["schemas"]["BankAccountView"];
export type UpdateBankAccountInput =
  components["schemas"]["UpdateBankAccountInputBody"];

// Order
export type Order = components["schemas"]["OrderView"];
