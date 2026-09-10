import { getAccessToken } from "@/lib/token";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;
const API_TOKEN_TEST = process.env.NEXT_PUBLIC_API_TOKEN_TEST;

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  if (!API_BASE_URL) {
    throw new Error("NEXT_PUBLIC_API_BASE_URL is not defined");
  }

  const headers = new Headers(options.headers);

  headers.set("Content-Type", "application/json");

  if (!headers.has("Authorization")) {
    const accessToken = getAccessToken();

    if (accessToken) {
      headers.set("Authorization", `Bearer ${accessToken}`);
    }
  }

  if (API_TOKEN_TEST) {
    headers.set("Authorization", `Bearer ${API_TOKEN_TEST}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  });
  if (!response.ok) {
    const text = await response.text();

    let message = `API request failed: ${response.status}`;

    if (text) {
      try {
        const errorData = JSON.parse(text);
        message =
          errorData.detail || errorData.message || errorData.title || message;
      } catch {
        // Response is not JSON
        message = text;
      }
    }

    throw new Error(message);
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
