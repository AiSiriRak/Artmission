import { apiFetch } from "./client";
import { Order, OrderHistory, OrderStatus } from "./types";

export interface CreateOrderInput {
  artwork_id: string;
  name: string;
  customer_description: string;
  /** ISO 8601 date-time */
  deadline_at: string;
}

/**
 * สร้างออเดอร์ใหม่ของลูกค้า
 * POST /orders
 */
export async function createOrder(input: CreateOrderInput): Promise<Order> {
  return apiFetch<Order>("/orders", {
    method: "POST",
    body: JSON.stringify(input),
  });
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

  return apiFetch<OrderHistory>(`/orders${query ? `?${query}` : ""}`);
}
