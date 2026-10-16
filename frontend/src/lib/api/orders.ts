import { apiFetch } from "./client";
import type {
  CancelOrderInput,
  CancelOrderResult,
  ConfirmDelayRequestInput,
  ConfirmDelayRequestResult,
  ConfirmDeliverableInput,
  ConfirmDeliverableResult,
  CreateOrderInput,
  Order,
  OrderDetail,
  OrderHistory,
  OrderStatus,
} from "./types";

export type { OrderDeliverable, OrderDetail } from "./types";

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
export async function getOrderById(orderId: string): Promise<OrderDetail> {
  return apiFetch<OrderDetail>(`/orders/${orderId}`);
}

/**
 * ยกเลิกออเดอร์ (backend ต้องการ reason และ explanation)
 * PUT /orders/{order_id}/cancel
 */
export async function cancelOrder(
  orderId: string,
  input: CancelOrderInput = {
    reason: "NO_LONGER_NEEDED",
    explanation: "Cancelled by customer",
  },
): Promise<CancelOrderResult> {
  return apiFetch<CancelOrderResult>(`/orders/${orderId}/cancel`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

/**
 * ลูกค้าอนุมัติหรือปฏิเสธงานที่ส่งมอบ (ปฏิเสธต้องมี comment)
 * PUT /orders/{order_id}/deliverables/{deliverable_id}/confirm
 */
export async function confirmDeliverable(
  orderId: string,
  deliverableId: string,
  input: ConfirmDeliverableInput,
): Promise<ConfirmDeliverableResult> {
  return apiFetch<ConfirmDeliverableResult>(
    `/orders/${orderId}/deliverables/${deliverableId}/confirm`,
    { method: "PUT", body: JSON.stringify(input) },
  );
}

/**
 * ลูกค้าอนุมัติหรือปฏิเสธคำขอเลื่อนเดดไลน์
 * PUT /orders/{order_id}/delay-requests/{delay_request_id}/confirm
 */
export async function confirmDelayRequest(
  orderId: string,
  delayRequestId: string,
  input: ConfirmDelayRequestInput,
): Promise<ConfirmDelayRequestResult> {
  return apiFetch<ConfirmDelayRequestResult>(
    `/orders/${orderId}/delay-requests/${delayRequestId}/confirm`,
    { method: "PUT", body: JSON.stringify(input) },
  );
}
