import { apiFetch } from "./client";
import { OrderHistory } from "./types";

export async function getOrderHistory(): Promise<OrderHistory> {
  return apiFetch<OrderHistory>("/orders/history");
}
