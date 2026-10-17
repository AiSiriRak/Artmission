import { apiFetch } from "./client";
import { OrderHistory, OrderStatus } from "./types";
import type { CancelOrderInput, CreateOrderInput, Order } from "./types";
export { getOrderDetail as getOrderById } from "./order-details";
export type { OrderDetail, OrderDeliverable } from "./order-details";

export function createOrder(input: CreateOrderInput) {
  return apiFetch<Order>("/orders", { method: "POST", body: JSON.stringify(input) });
}

export function cancelOrder(orderId: string, input: CancelOrderInput) {
  return apiFetch<{ order_id: string; reason: string; explanation: string }>(
    `/orders/${encodeURIComponent(orderId)}/cancel`,
    { method: "PUT", body: JSON.stringify(input) },
  );
}

export function confirmDeliverable(orderId: string, deliverableId: string, input: { accept: boolean; comment?: string }) {
  return apiFetch<{ message: string; status: string }>(
    `/orders/${encodeURIComponent(orderId)}/deliverables/${encodeURIComponent(deliverableId)}/confirm`,
    { method: "PUT", body: JSON.stringify(input) },
  );
}

export function confirmDelayRequest(orderId: string, requestId: string, input: { accept: boolean }) {
  return apiFetch<{ message: string; customer_decision: string; deadline_at: string }>(
    `/orders/${encodeURIComponent(orderId)}/delay-requests/${encodeURIComponent(requestId)}/confirm`,
    { method: "PUT", body: JSON.stringify(input) },
  );
}

interface GetOrderHistoryParams {
  status?: OrderStatus[];
  sort?: "deadline" | "price" | "updated_at";
  order?: "asc" | "desc";
  limit?: number;
  offset?: number;
}

export async function getOrderHistory(
  params?: GetOrderHistoryParams,
  options: RequestInit = {},
): Promise<OrderHistory> {
  const searchParams = new URLSearchParams();

  if (params?.status) {
    params.status.forEach((status) => {
      searchParams.append("status", status);
    });
  }

  if (params?.sort) {
    searchParams.set("sort", params.sort);
  }

  if (params?.order) {
    searchParams.set("order", params.order);
  }

  if (params?.limit !== undefined) {
    searchParams.set("limit", params.limit.toString());
  }

  if (params?.offset !== undefined) {
    searchParams.set("offset", params.offset.toString());
  }

  const query = searchParams.toString();

  return apiFetch<OrderHistory>(`/orders${query ? `?${query}` : ""}`, options);
}
