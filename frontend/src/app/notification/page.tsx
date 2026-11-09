"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";

import { getAccount } from "@/lib/api/users";
import { routes } from "@/lib/routes";
import { isApiError } from "@/lib/api/error";
import { UserAccount, NotificationList, Notification } from "@/lib/api/types";
import {
  getNotifications,
  markNotificationAsRead,
  NotificationType,
} from "@/lib/api/notification";

import { Loading } from "@/components/ui/Loading";
import MainLayout from "@/components/feature/main/MainLayout";
import { NotificationCard } from "@/components/feature/notification/NotificationCard";

export default function NotificationPage() {
  const router = useRouter();
  const [user, setUser] = useState<UserAccount | null>(null);
  const [notification, setNotification] = useState<NotificationList>();

  const notiRouteMap = (type: NotificationType, order?: string): string => {
    if (type === "WALLET_CREDITED" || type === "REFUND_CREDITED") {
      return routes.settings;
    }
    if (type === "GENERAL" || !order) {
      return routes.order.history;
    }
    return routes.order.detail(order);
  };

  const fetchNotifications = async () => {
    try {
      const notificationdata = await getNotifications();
      setNotification(notificationdata);
    } catch (error) {
      console.error("Failed to fetch notifications:", error);
    }
  };

  useEffect(() => {
    async function loadUserAndData() {
      try {
        const accountdata = await getAccount();
        setUser(accountdata);

        const notificationdata = await getNotifications();
        setNotification(notificationdata);
      } catch (error: unknown) {
        if (isApiError(error) && error.status === 401) {
          router.replace(routes.login);
        }
      }
    }

    loadUserAndData();

    const handleNotificationUpdate = () => {
      fetchNotifications();
    };

    window.addEventListener("notification-updated", handleNotificationUpdate);

    return () => {
      window.removeEventListener(
        "notification-updated",
        handleNotificationUpdate,
      );
    };
  }, [router]);

  if (!user) {
    return <Loading />;
  }

  const handleClick = async (
    e: React.MouseEvent,
    notificaion: Notification,
    orderId: string,
  ) => {
    e.preventDefault();

    await markNotificationAsRead(notificaion.id);

    await fetchNotifications();

    router.replace(notiRouteMap(notificaion.type as NotificationType, orderId));
  };

  return (
    <MainLayout usertype={user.role}>
      <div className="w-full max-w-[80%] flex-1 justify-self-center items-center">
        <p className="mx-20 my-10 text-h2">Recently Notifications</p>
        <div className="space-y-6 mx-20 my-10">
          {notification?.items?.map((option) => (
            <NotificationCard
              key={option.id}
              notification={option}
              orderId={option.order_id}
              user={user}
              onClick={(e) => {
                handleClick(e, option, option.order_id);
              }}
            />
          ))}
          {(!notification ||
            !notification.items ||
            notification?.items.length == 0) && (
            <div className="w-full relative flex flex-col items-center justify-center gap-6 my-40">
              <Image
                src="/icons/notification/bell-off.svg"
                alt={""}
                width={96}
                height={96}
              />
              <p className="text-h1 text-neutral">No notifications yet</p>
              <p className="text-body text-neutral">
                Updates about your commissions will appear here.
              </p>
            </div>
          )}
        </div>
      </div>
    </MainLayout>
  );
}
