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

export async function deleteAccount(): Promise<void> {
  await apiFetch<void>("/users/me", {
    method: "DELETE",
  });
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
