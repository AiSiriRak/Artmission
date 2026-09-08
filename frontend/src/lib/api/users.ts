import { apiFetch } from "./client";

import {
  BankAccount,
  UpdateAccountInput,
  UpdateBankAccountInput,
  UserAccount,
} from "./types";

export async function getAccount(): Promise<UserAccount> {
  return apiFetch<UserAccount>("/users/me");
}

export async function updateAccount(
  data: UpdateAccountInput,
): Promise<UserAccount> {
  return apiFetch<UserAccount>("/users/me", {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

// Not implement yet.
export async function deleteAccount(): Promise<UserAccount> {
  return apiFetch<UserAccount>("/users/me");
}

export async function getBankAccount(): Promise<BankAccount> {
  return apiFetch<BankAccount>("/users/me/bank-account");
}

export async function updateBankAccount(
  data: UpdateBankAccountInput,
): Promise<BankAccount> {
  return apiFetch<BankAccount>("/users/me/bank-account", {
    method: "PUT",
    body: JSON.stringify(data),
  });
}
