import { apiFetch } from "./client";
import {
  MarkNotificationReadOutput,
  NotificationList,
  NotificationUnreadCount,
} from "./types";

export async function getNotifications(): Promise<NotificationList> {
  return apiFetch<NotificationList>("/notifications");
}

export async function getUnreadCount(): Promise<NotificationUnreadCount> {
  return apiFetch<NotificationUnreadCount>("/notifications/unread-count");
}

export async function markNotificationAsRead(
  id: string,
): Promise<MarkNotificationReadOutput> {
  return apiFetch<MarkNotificationReadOutput>(`/notifications/${id}/read`, {
    method: "PATCH",
  });
}

export type NotificationType =
  | "GENERAL"
  | "ORDER_CREATED"
  | "ORDER_ACCEPTED"
  | "PAYMENT_SECURED"
  | "WALLET_CREDITED"
  | "REFUND_CREDITED"
  | "ORDER_DECLINED"
  | "ORDER_CANCELED"
  | "DRAFT_SUBMITTED"
  | "REVISION_REQUESTED"
  | "ORDER_COMPLETED"
  | "FINAL_ARTWORK_READY"
  | "DEADLINE_EXTENSION_REQUESTED"
  | "DEADLINE_EXTENSION_APPROVED"
  | "DEADLINE_EXTENSION_REJECTED";
