import { expect, type APIRequestContext } from "@playwright/test";
// Type-only imports: the request/response shapes come from the OpenAPI types the app
// uses, but no app code runs here, so a bug in src/lib/api cannot break the fixtures.
import type {
  AuthResult,
  CreateOrderInput,
  LoginInput,
  Order,
  OrderHistory,
  SearchArtworksOutput,
} from "../../src/lib/api/types";
import { accounts, type AccountKey } from "./accounts";

const API_BASE_URL = process.env.E2E_API_BASE_URL ?? "http://localhost:8080/api/v1";

async function call<T>(
  request: APIRequestContext,
  method: "GET" | "POST" | "PUT",
  path: string,
  token: string | undefined,
  data?: object,
): Promise<T> {
  const response = await request.fetch(`${API_BASE_URL}${path}`, {
    method,
    data,
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });
  expect(response.ok(), `${method} ${path} -> ${response.status()} ${await response.text()}`).toBe(true);
  return (await response.json()) as T;
}

export async function apiLogin(request: APIRequestContext, account: AccountKey): Promise<string> {
  const credentials: LoginInput = accounts[account];
  const body = await call<AuthResult>(request, "POST", "/auth/login", undefined, credentials);
  return body.access_token;
}

export async function findArtworkId(
  request: APIRequestContext,
  account: AccountKey,
  name: string,
): Promise<string> {
  const token = await apiLogin(request, account);
  const body = await call<SearchArtworksOutput>(request, "GET", "/artworks?limit=100", token);
  const artwork = (body.artworks ?? []).find((item) => item.name === name);
  expect(artwork, `artwork "${name}" is not seeded`).toBeDefined();
  return artwork!.id;
}

// Places a real order through the API as the customer, so the artist-side tests
// get a fresh PENDING order each time instead of depending on seed state.
export async function placeOrder(
  request: APIRequestContext,
  options: { artworkName?: string; name: string },
): Promise<Order> {
  const customer = "customerJane";
  const token = await apiLogin(request, customer);
  const order: CreateOrderInput = {
    artwork_id: await findArtworkId(request, customer, options.artworkName ?? "Custom Portrait"),
    name: options.name,
    customer_description: "Created by an automated test",
    deadline_at: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
  };
  return call<Order>(request, "POST", "/orders", token, order);
}

// Returns the order id of a seeded order, looked up by its name through the API.
export async function findOrderId(
  request: APIRequestContext,
  account: AccountKey,
  name: string,
): Promise<string> {
  const token = await apiLogin(request, account);
  const body = await call<OrderHistory>(request, "GET", "/orders?limit=100", token);
  const order = (body.orders ?? []).find((item) => item.name === name);
  expect(order, `order "${name}" is not visible to ${account}; reseed the database`).toBeDefined();
  return order!.id;
}
