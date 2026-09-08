import { apiFetch } from "./client";
import {
  BankAccount,
  UpdateAccountInput,
  UpdateBankAccountInput,
  UserAccount,
} from "./types";

export async function getAccount(token: string): Promise<UserAccount> {
  return apiFetch<UserAccount>("/users/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

export async function updateAccount(
  token: string,
  data: UpdateAccountInput,
): Promise<UserAccount> {
  return apiFetch<UserAccount>("/users/me", {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });
}

// Not implement yet.
export async function deleteAccount(token: string): Promise<UserAccount> {
  return apiFetch<UserAccount>("/users/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

// Not implement yet.
export async function getBankAccount(token: string): Promise<UserAccount> {
  return apiFetch<UserAccount>("/users/me/bank-account", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

export async function updateBankAccount(
  token: string,
  data: BankAccount,
): Promise<BankAccount> {
  return apiFetch<BankAccount>("/users/me/bank-account", {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });
}
