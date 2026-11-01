import { apiFetch } from "./client";
import {
  NotificationList,
  NotificationUnreadCount,
  Notification,
} from "./types";

let mockNotifications: NotificationList = [];

export async function getNotifications(): Promise<NotificationList> {
  return mockNotifications;

  //   return apiFetch<NotificationList>("/notifications");
}

export async function getUnreadCount(): Promise<NotificationUnreadCount> {
  return {
    count: mockNotifications.filter((n) => !n.is_read).length,
  } as NotificationUnreadCount;
  //   return apiFetch<NotificationUnreadCount>("/notifications/unread-count");
}

export async function markNotificationAsRead(id: string) {
  mockNotifications = mockNotifications.map((n) =>
    n.id === id ? { ...n, is_read: true } : n,
  );
  //   return apiFetch<void>(`/notifications/${id}/read`, {
  //     method: "PUT",
  //   });
}

// TEMP FOR TESTING
export async function createMockNotification(
  notification: Omit<Notification, "id" | "is_read" | "created_at" | "user_id">,
) {
  const newNotification: Notification = {
    ...notification,
    id: crypto.randomUUID(),
    user_id: "mock-user",
    is_read: false,
    created_at: new Date().toISOString(),
  };

  mockNotifications.unshift(newNotification);

  return newNotification;
}
