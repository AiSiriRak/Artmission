import { clearAccessToken, getAccessToken, setAccessToken } from "@/lib/token";
import type { AuthResult } from "./types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly body: unknown,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

const NO_REFRESH_RETRY_PATHS: Record<string, true> = {
  "/auth/login": true,
  "/auth/register": true,
  "/auth/refresh": true,
};

function messageFromErrorBody(body: unknown, fallback: string) {
  if (body && typeof body === "object") {
    const errorData = body as Record<string, unknown>;
    const message = errorData.detail || errorData.message || errorData.title;

    if (typeof message === "string" && message.trim()) {
      return message;
    }
  }

  if (typeof body === "string" && body.trim()) {
    return body;
  }

  return fallback;
}

async function rawFetch<T>(path: string, options: RequestInit): Promise<T> {
  if (!API_BASE_URL) {
    throw new Error("NEXT_PUBLIC_API_BASE_URL is not defined");
  }

  const headers = new Headers(options.headers);

  if (options.body instanceof FormData) {
    headers.delete("Content-Type");
  } else if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  if (!headers.has("Authorization")) {
    const accessToken = getAccessToken();

    if (accessToken) {
      headers.set("Authorization", `Bearer ${accessToken}`);
    }
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const text = await response.text();
    let body: unknown = text;

    if (text) {
      try {
        body = JSON.parse(text) as unknown;
      } catch {
        // Response is not JSON; keep the raw text body.
      }
    }

    const message = messageFromErrorBody(
      body,
      `API request failed: ${response.status}`,
    );

    throw new ApiError(message, response.status, body);
  }

  // 204 No Content
  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();

  // Response with no Body
  if (!text) {
    return undefined as T;
  }

  return JSON.parse(text);
}

let refreshPromise: Promise<boolean> | null = null;

function refreshAccessToken(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = rawFetch<AuthResult>("/auth/refresh", {
    method: "POST",
    credentials: "include",
  })
    .then((result) => {
      setAccessToken(result.access_token);
      return true;
    })
    .catch(() => {
      clearAccessToken();
      return false;
    })
    .finally(() => {
      refreshPromise = null;
    });

  return refreshPromise;
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
  _retrying = false,
): Promise<T> {
  try {
    return await rawFetch<T>(path, options);
  } catch (error) {
    const shouldRetry =
      error instanceof ApiError &&
      error.status === 401 &&
      !_retrying &&
      !NO_REFRESH_RETRY_PATHS[path];

    if (shouldRetry && (await refreshAccessToken())) {
      return apiFetch<T>(path, options, true);
    }

    throw error;
  }
}
