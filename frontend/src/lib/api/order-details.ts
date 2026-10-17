import { apiFetch } from "./client";
import type { OrderStatus } from "./types";

// The running Swagger includes these endpoints; v1.d.ts is not yet regenerated.
export interface OrderDeliverable {
  id: string;
  version: number;
  decision: "WAIT" | "APPROVED" | "REJECTED";
  comment: string | null;
  preview_image_url: string;
  created_at: string;
  updated_at: string;
}

export type PostponeReason =
  | "NEED_MORE_TIME"
  | "WAITING_FOR_FEEDBACK"
  | "PERSONAL_REASON"
  | "HIGH_WORKLOAD"
  | "TECHNICAL_ISSUE"
  | "OTHER";

export interface OrderDelayRequest {
  id: string;
  order_id: string;
  original_deadline_at: string;
  deadline_request_at: string;
  reason: PostponeReason;
  explanation: string;
  customer_decision: "WAIT" | "APPROVED" | "REJECTED";
  created_at: string;
  updated_at: string;
}

export interface OrderDetail {
  id: string;
  customer_id: string;
  artist_id: string;
  artwork_id?: string;
  name: string;
  artwork_snapshot: { artwork_name: string; category: string; styles: string[] | null };
  price_satang: number;
  customer_description: string;
  deadline_at: string;
  status: OrderStatus;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  other_party: { id: string; name: string; email: string; artist_review_score?: number };
  deliverables: OrderDeliverable[] | null;
  active_delay_request?: OrderDelayRequest | null;
}

export function requestOrderPostpone(orderId: string, input: {
  deadline_request_at: string;
  reason: PostponeReason;
  explanation: string;
}) {
  return apiFetch<OrderDelayRequest>(`/orders/${encodeURIComponent(orderId)}/delay-requests`, {
    method: "POST", body: JSON.stringify(input),
  });
}

export function getOrderDetail(orderId: string, options: RequestInit = {}) {
  return apiFetch<OrderDetail>(`/orders/${encodeURIComponent(orderId)}`, options);
}

export function confirmOrder(orderId: string, accept: boolean) {
  return apiFetch<{ message: string; status: OrderStatus }>(
    `/orders/${encodeURIComponent(orderId)}/confirm`,
    { method: "PUT", body: JSON.stringify({ accept }) },
  );
}

export type ArtistCancelReason =
  | "UNABLE_TO_COMPLETE"
  | "UNABLE_TO_MEET_DEADLINE"
  | "CUSTOMER_CHANGED_REQUIREMENTS"
  | "CUSTOMER_UNRESPONSIVE"
  | "MUTUAL_AGREEMENT"
  | "PERSONAL_OR_TECHNICAL_ISSUE"
  | "OTHER";

export function cancelOrder(orderId: string, input: { reason: ArtistCancelReason; explanation: string }) {
  return apiFetch<{ order_id: string; reason: string; explanation: string }>(
    `/orders/${encodeURIComponent(orderId)}/cancel`,
    { method: "PUT", body: JSON.stringify(input) },
  );
}

export function submitOrderDraft(orderId: string, file: File) {
  const body = new FormData();
  body.append("deliverable_image", file);
  return apiFetch<OrderDeliverable>(`/orders/${encodeURIComponent(orderId)}/deliverables`, {
    method: "POST", body,
  });
}
