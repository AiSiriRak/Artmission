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
  if (API_TOKEN_TEST) {
    headers.set("Authorization", `Bearer ${API_TOKEN_TEST}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  return response.json();
}
