"use client";

import { formatDistanceToNowStrict } from "date-fns";
import Image from "next/image";
import { ButtonHTMLAttributes, useEffect, useState } from "react";

import { Notification, OrderDetail, UserAccount } from "@/lib/api/types";

import { NotificationType } from "@/lib/api/notification";
import { getOrderById } from "@/lib/api/orders";

interface NotificationCardProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  notification: Notification;
  user: UserAccount;
  orderId?: string;
}

export function NotificationCard({
  notification,
  user,
  orderId,
  ...props
}: NotificationCardProps) {
  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [payment, setPayment] = useState<number | null>(null);

  const isWalletOrRefund =
    notification.type === "WALLET_CREDITED" ||
    notification.type === "REFUND_CREDITED";

  const bgColor = isWalletOrRefund ? "bg-secondary-300" : "bg-secondary-200";

  const textColor = notification.is_read ? "text-neutral" : "text-primary-500";

  const notiIconMap = (type: NotificationType): string => {
    if (
      type === "WALLET_CREDITED" ||
      type === "REFUND_CREDITED" ||
      type === "PAYMENT_SECURED"
    ) {
      return "/icons/notification/wallet.svg";
    }
    if (
      type === "DEADLINE_EXTENSION_REQUESTED" ||
      type === "DEADLINE_EXTENSION_APPROVED" ||
      type === "DEADLINE_EXTENSION_REJECTED"
    ) {
      return "/icons/notification/hourglass.svg";
    }
    if (type === "REVISION_REQUESTED" || type === "DRAFT_SUBMITTED") {
      return "/icons/notification/pencil.svg";
    }
    return "/icons/notification/order.svg";
  };

  const notiIcon = notiIconMap(notification.type as NotificationType);

  const statusColor: Record<string, string> = {
    PENDING: "bg-status-pending-pale",
    NOT_PAID: "bg-status-notpaid-pale",
    IN_PROCESS: "bg-status-inprocess-pale",
    SUCCESS: "bg-status-success-pale",
    CANCEL: "bg-status-cancel-pale",
  };

  useEffect(() => {
    if (!orderId) return;

    let isMounted = true;

    async function loadOrder() {
      try {
        const orderData = await getOrderById(orderId!);
        if (!isMounted) return;

        setOrder(orderData);

        if (orderData?.payment && isWalletOrRefund) {
          const amount =
            user.role === "artist"
              ? orderData.payment.artist_payout_satang
              : orderData.payment.refund_satang;
          setPayment(amount);
        }
      } catch (error: unknown) {
        console.error("Failed to fetch order:", error);
      }
    }

    loadOrder();

    return () => {
      isMounted = false;
    };
  }, [orderId, notification.type, user.role, isWalletOrRefund]);

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
          alt="Notification Icon"
          width={80}
          height={80}
          className={notification.is_read ? "opacity-50" : "opacity-100"}
        />
        <div className="flex-1 min-w-0 space-y-4">
          <div className="flex min-w-0 space-x-5">
            {/* If this is order noti, show order's status */}
            {(notification.type === "ORDER_CREATED" ||
              notification.type === "ORDER_ACCEPTED" ||
              notification.type === "ORDER_DECLINED" ||
              notification.type === "ORDER_CANCELED" ||
              notification.type === "ORDER_COMPLETED") &&
              order && (
                <div
                  className={`${statusColor[order.status]} opacity-${notification.is_read ? "50" : "100"} flex w-50 p-1.5 border rounded-full justify-center`}
                >
                  <p className="text-caption text-primary-500">
                    {order.status}
                  </p>
                </div>
              )}
            {/* Noti title */}
            <p className={`${textColor} text-left text-h3 truncate`}>
              {notification.title}
            </p>
            {isWalletOrRefund && payment !== null && (
              <p
                className={`${notification.is_read ? "text-accent-300" : "text-accent-500"} text-left text-h3 truncate`}
              >
                [{payment < 0 ? "-" : "+"}
                {(payment / 100).toLocaleString()}THB]
              </p>
            )}
          </div>
          {/* Noti Message */}
          <p className={`${textColor} text-left text-caption truncate`}>
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
