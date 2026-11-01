"use client";

import { formatDistanceToNowStrict } from "date-fns";
import Image from "next/image";
import { ButtonHTMLAttributes } from "react";

import { Notification } from "@/lib/api/types";
import { NOTI_TYPE } from "@/lib/types";

interface NotificationCard extends ButtonHTMLAttributes<HTMLButtonElement> {
  notification: Notification;
}

export function NotificationCard({ notification, ...props }: NotificationCard) {
  const bgColor =
    notification.type === NOTI_TYPE.payment
      ? "bg-secondary-300"
      : "bg-secondary-200";

  const textColor = notification.is_read ? "text-neutral" : "text-primary-500";

  const notiIconMap: Record<string, string> = {
    [NOTI_TYPE.review]: "/icons/notification/pencil.svg",
    [NOTI_TYPE.order_update]: "/icons/notification/order.svg",
    [NOTI_TYPE.pospone]: "/icons/notification/hourglass.svg",
    [NOTI_TYPE.payment]: "/icons/notification/wallet.svg",
  };

  const notiIcon =
    notiIconMap[notification.type] ?? "/icons/notification/order.svg";

  const statusColor: Record<string, string> = {
    PENDING: "bg-status-pending-pale",
    NOT_PAID: "bg-status-notpaid-pale",
    IN_PROCESS: "bg-status-inprocess-pale",
    SUCCESS: "bg-status-success-pale",
    CANCEL: "bg-status-cancel-pale",
  };

  return (
    <button
      {...props}
      className={`${bgColor} ${notification.is_read ? "border-0" : "border-1"} relative flex w-full rounded-3xl py-6 gap-10 px-10 border-secondary-400 hover:brightness-90`}
    >
      {/* A red dot to indicate whether a notification has been read */}
      {!notification.is_read && (
        <div className="absolute -top-2 right-3 w-6 h-6 rounded-full bg-red-400 border-2 border-white" />
      )}
      <div className="flex flex-1 min-w-0 items-center gap-10">
        {/* Noti Icon */}
        <Image
          src={notiIcon}
          alt="Search"
          width={80}
          height={80}
          className={notification.is_read ? "opacity-50" : "opacity-100"}
        />
        <div className="flex-1 min-w-0 space-y-4">
          <div className="flex min-w-0 space-x-5">
            {/* If this is order noti, show order's status */}
            {notification.type === NOTI_TYPE.order_update &&
              notification.order && (
                <div
                  className={`${statusColor[notification.order.status]} opacity-${notification.is_read ? "50" : "100"} flex w-50 p-1.5 border rounded-full justify-center`}
                >
                  <p className="text-caption text-primary-500">
                    {notification.order.status}
                  </p>
                </div>
              )}
            {/* Noti title */}
            <p className={`${textColor} text-left text-h3 truncate`}>
              {notification.title}
            </p>
            {notification.type === NOTI_TYPE.payment &&
              notification.payment && (
                <p
                  className={`${notification.is_read ? "text-accent-300" : "text-accent-500"} text-left text-h3 truncate`}
                >
                  [{notification.payment.amount < 0 ? "-" : "+"}
                  {notification.payment.amount} THB]
                </p>
              )}
          </div>
          {/* Noti Message */}
          <p className={`${textColor} text-left text-body truncate`}>
            {notification.message}
          </p>
        </div>
      </div>
      {/* Noti Time */}
      <p className="shrink-0 whitespace-nowrap text-small text-secondary-600">
        {formatDistanceToNowStrict(new Date(notification.created_at), {
          addSuffix: true,
        })}
      </p>
    </button>
  );
}
