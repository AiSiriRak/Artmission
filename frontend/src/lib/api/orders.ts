import { apiFetch } from "./client";
import { Order } from "./types";

export async function getAccount(): Promise<Order> {
  return apiFetch<Order>("/orders/history");
}
