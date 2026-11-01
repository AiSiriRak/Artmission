"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { getAccount } from "@/lib/api/users";
import { routes } from "@/lib/routes";
import { isApiError } from "@/lib/api/error";
import {
  getNotifications,
  markNotificationAsRead,
} from "@/lib/api/notification";
import { UserAccount, NotificationList, Notification } from "@/lib/api/types";
import { NOTI_TYPE } from "@/lib/types";

import { Loading } from "@/components/ui/Loading";
import MainLayout from "@/components/feature/main/MainLayout";
import { NotificationCard } from "@/components/feature/notification/NotificationCard";
import NotiTest from "@/components/feature/notification/NotiTest";

export default function NotificationPage() {
  const router = useRouter();
  const [user, setUser] = useState<UserAccount | null>(null);
  const [notification, setNotification] = useState<NotificationList>([]);

  const notiRouteMap: Record<string, string> = {
    [NOTI_TYPE.review]: routes.home,
    [NOTI_TYPE.order_update]: routes.order.history,
    [NOTI_TYPE.pospone]: routes.settings,
    [NOTI_TYPE.payment]: routes.order.history,
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
    async function loadUser() {
      try {
        const accountdata = await getAccount();
        setUser(accountdata);

        await fetchNotifications();
      } catch (error: unknown) {
        if (isApiError(error) && error.status === 401) {
          router.replace(routes.login);
        }
      }
    }
    loadUser();
  }, [router, fetchNotifications]);

  if (!user) {
    return <Loading />;
  }

  const handleClick = async (
    e: React.MouseEvent,
    notificaion: Notification,
  ) => {
    e.preventDefault();

    await markNotificationAsRead(notificaion.id);

    await fetchNotifications();

    router.replace(notiRouteMap[notificaion.type]);
  };

  return (
    <MainLayout usertype={user.role}>
      <p className="mx-20 my-10 text-h2">Recently Notifications</p>
      <div className="space-y-6 mx-20 my-10">
        <NotiTest />
        {notification.map((option) => (
          <NotificationCard
            key={option.id}
            notification={option}
            onClick={(e) => {
              handleClick(e, option);
            }}
          />
        ))}
      </div>
    </MainLayout>
  );
}
