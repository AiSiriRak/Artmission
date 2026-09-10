import { apiFetch } from "./client";
import type { AuthResult, LoginInput, RegisterInput } from "./types";

function authHeaders(accessToken?: string): HeadersInit | undefined {
  if (!accessToken) {
    return undefined;
  }

  return {
    Authorization: `Bearer ${accessToken}`,
  };
}

export async function login(data: LoginInput): Promise<AuthResult> {
  return apiFetch<AuthResult>("/auth/login", {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(data),
  });
}

export async function register(data: RegisterInput): Promise<void> {
  return apiFetch<void>("/auth/register", {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(data),
  });
}

export async function refreshAccessToken(): Promise<AuthResult> {
  return apiFetch<AuthResult>("/auth/refresh", {
    method: "POST",
    credentials: "include",
  });
}

export async function logout(accessToken?: string): Promise<void> {
  return apiFetch<void>("/auth/logout", {
    method: "POST",
    credentials: "include",
    headers: authHeaders(accessToken),
  });
}
