"use client";

import { useState } from "react";
import { Button } from "@/components/ui/Button";
import type { OrderStatus } from "@/lib/api/types";
import { bypassPayment } from "@/lib/api/orders";
import { CustomerConfirmPopUp } from "./CustomerConfirmPopUp";

interface CustomerOrderActionsProps {
  orderId: string;
  status: OrderStatus;
  onCancelOrder: () => void;
  onPaymentBypassed: () => Promise<void>;
}

function CancelOrderButton({ onClick }: { onClick: () => void }) {
  return (
    <Button type="button" variant="red" onClick={onClick}>
      Cancel Order
    </Button>
  );
}

function GoToPaymentButton({ onClick }: { onClick: () => void }) {
  return (
    <Button type="button" variant="dark" onClick={onClick}>
      Go to Payment
    </Button>
  );
}

export default function CustomerOrderActions({
  orderId,
  status,
  onCancelOrder,
  onPaymentBypassed,
}: CustomerOrderActionsProps) {
  const [paymentConfirmationOpen, setPaymentConfirmationOpen] = useState(false);
  const [submittingPayment, setSubmittingPayment] = useState(false);
  const [paymentError, setPaymentError] = useState<string | null>(null);

  const confirmPaymentBypass = async () => {
    setSubmittingPayment(true);
    setPaymentError(null);
    try {
      await bypassPayment(orderId);
      setPaymentConfirmationOpen(false);
      try {
        await onPaymentBypassed();
      } catch {
        setPaymentError("Payment was bypassed, but the order could not be refreshed.");
      }
    } catch (error) {
      setPaymentError(
        error instanceof Error ? error.message : "Could not bypass payment.",
      );
    } finally {
      setSubmittingPayment(false);
    }
  };

  switch (status) {
    case "PENDING":
    case "IN_PROCESS":
      return <CancelOrderButton onClick={onCancelOrder} />;
    case "NOT_PAID":
      return (
        <div className="flex flex-col items-start gap-2">
          <div className="flex items-start gap-2">
            <GoToPaymentButton
              onClick={() => {
                setPaymentError(null);
                setPaymentConfirmationOpen(true);
              }}
            />
            <CancelOrderButton onClick={onCancelOrder} />
            <CustomerConfirmPopUp
              isOpen={paymentConfirmationOpen}
              title="Simulate Payment?"
              description="This testing action will bypass payment and move the order to In Process."
              cancelText="Cancel"
              confirmText={submittingPayment ? "Processing..." : "Confirm"}
              confirmDisabled={submittingPayment}
              errorMessage={paymentError}
              onCancel={() => setPaymentConfirmationOpen(false)}
              onConfirm={confirmPaymentBypass}
            />
          </div>

          {paymentError && (
            <p role="alert" className="max-w-xs text-small text-error">
              {paymentError}
            </p>
          )}
        </div>
      );
    case "SUCCESS":
    case "CANCEL":
      return null;
  }
}
