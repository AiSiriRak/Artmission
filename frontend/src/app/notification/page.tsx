"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { UserAccount } from "@/lib/api/types";
import { getAccount } from "@/lib/api/users";
import { routes } from "@/lib/routes";
import { isApiError } from "@/lib/api/error";

import { Loading } from "@/components/ui/Loading";
import MainLayout from "@/components/feature/main/MainLayout";
import {
  NotificationCard,
  notiType,
  Notification,
} from "@/components/feature/notification/NotificationCard";

// TEMP
let mockNoti = [
  {
    id: "1",
    user_id: "1",
    title: "Payment noti",
    message:
      "a very long messageeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    type: notiType.payment,
    is_read: false,
    created_at: "2026-10-30T20:32:21.129Z",
    payment: {
      amount: 50,
    },
  },
  {
    id: "2",
    user_id: "2",
    title: "Order noti",
    message:
      "a very long messageeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    type: notiType.orderUpdate,
    is_read: false,
    created_at: "2026-09-30T20:32:21.129Z",
    order: {
      status: "PENDING",
    },
  },
  {
    id: "3",
    user_id: "3",
    title: "Review noti",
    message:
      "a very long messageeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    type: notiType.review,
    is_read: true,
    created_at: "2025-10-30T20:32:21.129Z",
  },
  {
    id: "4",
    user_id: "4",
    title: "Pospone noti",
    message:
      "a very long messageeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    type: notiType.pospone,
    is_read: true,
    created_at: "2026-10-05T20:32:21.129Z",
  },
];
export default function NotificationPage() {
  const router = useRouter();
  const [user, setUser] = useState<UserAccount | null>(null);

  const notiRouteMap: Record<string, string> = {
    [notiType.review]: routes.order.history,
    [notiType.orderUpdate]: routes.order.history,
    [notiType.pospone]: routes.order.history,
    [notiType.payment]: routes.order.history,
  };

  useEffect(() => {
    async function loadUser() {
      try {
        const accountdata = await getAccount();
        setUser(accountdata);
      } catch (error: unknown) {
        if (isApiError(error) && error.status === 401) {
          router.replace(routes.login);
        }
      }
    }
    loadUser();
  }, [router]);

  if (!user) {
    return <Loading />;
  }

  const handleClick = (e: React.MouseEvent, notificaion: Notification) => {
    e.preventDefault();
    router.replace(notiRouteMap[notificaion.type]);
  };

  return (
    <MainLayout usertype={user.role}>
      <p className="mx-20 my-10 text-h2">Recently Notifications</p>
      <div className="space-y-6 mx-20 my-10">
        {mockNoti.map((option) => (
          <NotificationCard
            key={option.id}
            notification={option}
            onClick={(e) => {
              handleClick(e, option);
              option.is_read = true;
            }}
          />
        ))}
      </div>
    </MainLayout>
  );
}
