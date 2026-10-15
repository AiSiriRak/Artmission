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

/**
 * ดึงรายละเอียดออเดอร์ตาม ID
 * GET /orders/{orderId}
 */
export interface OrderDeliverable {
  id: string;
  version: number;
  decision: "WAIT" | "APPROVED" | "REJECTED";
  comment: string | null;
  preview_image_url: string;
  created_at: string;
  updated_at: string;
}

export interface OrderDetail {
  id: string;
  customer_id: string;
  artist_id: string;
  name: string;
  artwork_id?: string;
  artwork_snapshot: {
    artwork_name: string;
    category: string;
    styles: string[] | null;
  };
  price_satang: number;
  customer_description: string;
  deadline_at: string;
  status: OrderStatus;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  other_party: {
    id: string;
    name: string;
    email: string;
    artist_review_score?: number;
  };
  deliverables: OrderDeliverable[] | null;
}

export async function getOrderById(orderId: string): Promise<OrderDetail> {
  return apiFetch<OrderDetail>(`/orders/${orderId}`);
}

/**
 * ยกเลิกออเดอร์
 * PUT /orders/{order_id}/cancel
 */
export async function cancelOrder(
  orderId: string,
): Promise<{ message: string; status: string }> {
  return apiFetch(`/orders/${orderId}/cancel`, { method: "PUT" });
}


