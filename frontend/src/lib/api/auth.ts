import { clearAccessToken } from "@/lib/token";
import { ApiError, apiFetch } from "./client";
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

export async function logout(accessToken?: string): Promise<void> {
  try {
    await apiFetch<void>("/auth/logout", {
      method: "POST",
      credentials: "include",
      headers: authHeaders(accessToken),
    });
  } catch (error) {
    // apiFetch already tried a refresh before surfacing this 401, so the
    // session is gone server-side (expired, revoked, or the DB was reset).
    // That is the state logout is after, so don't block signing out locally.
    if (!(error instanceof ApiError && error.status === 401)) {
      throw error;
    }
  }

  clearAccessToken();
}
