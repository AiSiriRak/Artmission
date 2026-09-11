import type { components, paths } from "./v1";

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
export type OrderHistory = components["schemas"]["ViewOrdersOutputBody"];
export type Order = components["schemas"]["OrderSummaryView"];
export type OrderQuery = NonNullable<
  paths["/orders"]["get"]["parameters"]["query"]
>;
export type OrderStatus = NonNullable<OrderQuery["status"]>[number];
export type OrderSortField = OrderQuery["sort"];
export type OrderSortOrder = OrderQuery["order"];
