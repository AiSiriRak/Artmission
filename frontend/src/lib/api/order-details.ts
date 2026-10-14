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

// Current cancel API has no request body. Reasons require a Backend extension.
export function cancelOrder(orderId: string) {
  return apiFetch<{ message: string; status: OrderStatus }>(
    `/orders/${encodeURIComponent(orderId)}/cancel`, { method: "PUT" },
  );
}

export function submitOrderDraft(orderId: string, file: File) {
  const body = new FormData();
  body.append("deliverable_image", file);
  return apiFetch<OrderDeliverable>(`/orders/${encodeURIComponent(orderId)}/deliverables`, {
    method: "POST", body,
  });
}
