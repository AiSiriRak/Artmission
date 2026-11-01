"use client";

import { useState } from "react";
import { toast } from "sonner";

import { createMockNotification } from "@/lib/api/notification";
import { NotificationType } from "@/lib/api/types";
import { NOTI_TYPE } from "@/lib/types";

import { Button } from "@/components/ui/Button";

const mockEvents = [
  {
    title: "Payment noti",
    message:
      "a very long messageeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    type: NOTI_TYPE.payment,
    payment: {
      amount: 50,
    },
  },
  {
    title: "Order noti",
    message:
      "a very long messageeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    type: NOTI_TYPE.order_update,
    order: {
      status: "PENDING",
    },
  },
  {
    title: "Review noti",
    message:
      "a very long messageeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    type: NOTI_TYPE.review,
  },
  {
    title: "Pospone noti",
    message:
      "a very long messageeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    type: NOTI_TYPE.pospone,
  },
] as const;

export default function NotiTest() {
  const [loading, setLoading] = useState(false);

  const handleCreateNotification = async (
    event: (typeof mockEvents)[number],
  ) => {
    setLoading(true);

    try {
      await createMockNotification({
        type: event.type as NotificationType,
        title: event.title,
        message: event.message,
        payment: "payment" in event ? event.payment : undefined,
        order: "order" in event ? event.order : undefined,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleMockNotification = () => {
    toast.info("New Order Received!", {
      description: "You have received a new order from a customer.",
      duration: 5000,
    });
    console.log("asadsad");
  };

  return (
    <div className="flex flex-wrap gap-3">
      {mockEvents.map((event) => (
        <Button
          key={event.type}
          disabled={loading}
          onClick={() => {
            handleCreateNotification(event);
            handleMockNotification;
          }}
          variant="light"
        >
          + {event.title}
        </Button>
      ))}
    </div>
  );
}
