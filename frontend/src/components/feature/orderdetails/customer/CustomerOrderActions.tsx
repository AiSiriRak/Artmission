"use client";

import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/Button";
import type { OrderStatus } from "@/lib/api/types";
import { routes } from "@/lib/routes";

interface CustomerOrderActionsProps {
  orderId: string;
  status: OrderStatus;
  onCancelOrder: () => void;
}

export default function CustomerOrderActions({ orderId, status, onCancelOrder }: CustomerOrderActionsProps) {
  const router = useRouter();
  if (status === "SUCCESS" || status === "CANCEL") return null;
  return (
    <div className="flex flex-wrap items-start gap-2">
      {status === "NOT_PAID" && (
        <Button type="button" variant="dark" onClick={() => router.push(routes.order.payment(orderId))}>
          Go to Payment
        </Button>
      )}
      <Button type="button" variant="red" onClick={onCancelOrder}>Cancel Order</Button>
    </div>
  );
}
