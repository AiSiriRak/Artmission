import { apiFetch } from "./client";
import { Order } from "./types";

export async function getAccount(token: string): Promise<Order> {
  return apiFetch<Order>("/orders/history", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}
