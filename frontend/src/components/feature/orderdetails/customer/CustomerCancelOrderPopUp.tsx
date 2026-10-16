"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { SelectInput } from "@/components/ui/SelectInput";
import { WhiteCard } from "@/components/ui/WhiteCard";
import { cancelOrder } from "@/lib/api/orders";
import type { CancelOrderInput } from "@/lib/api/types";

type CancelReason = CancelOrderInput["reason"];

// Reasons the backend accepts from a customer (order.CancelReason.IsValid).
const CUSTOMER_REASONS: { value: CancelReason; label: string }[] = [
  { value: "PLACED_ORDER_BY_MISTAKE", label: "I placed this order by mistake" },
  { value: "NO_LONGER_NEEDED", label: "I no longer need this artwork" },
  {
    value: "ARTIST_UNABLE_TO_MEET_NEW_REQUIREMENTS",
    label: "The artist can't meet my new requirements",
  },
  { value: "ARTIST_UNRESPONSIVE", label: "The artist is unresponsive" },
  { value: "TAKING_TOO_LONG", label: "The order is taking too long" },
  { value: "MUTUAL_AGREEMENT", label: "The artist and I agreed to cancel" },
  { value: "OTHER", label: "Other" },
];

interface CustomerCancelOrderPopUpProps {
  isOpen: boolean;
  orderId: string;
  orderLabel: string;
  onClose: () => void;
  /** Called after the backend confirmed the cancellation. */
  onCancelled: () => void;
}

export function CustomerCancelOrderPopUp({
  isOpen,
  orderId,
  orderLabel,
  onClose,
  onCancelled,
}: CustomerCancelOrderPopUpProps) {
  const [reason, setReason] = useState<string>("");
  const [explanation, setExplanation] = useState("");
  const [validation, setValidation] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [status, setStatus] = useState<"default" | "success" | "fail">(
    "default",
  );
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    document.body.style.overflow = isOpen ? "hidden" : "";
    return () => {
      document.body.style.overflow = "";
    };
  }, [isOpen]);

  if (!isOpen) return null;

  const close = () => {
    if (status === "success") onCancelled();
    setReason("");
    setExplanation("");
    setValidation(null);
    setStatus("default");
    setErrorMessage(null);
    onClose();
  };

  const submit = async () => {
    if (!reason) {
      setValidation("Please select a reason for cancellation");
      return;
    }
    if (!explanation.trim()) {
      setValidation("Please explain your reason");
      return;
    }
    setValidation(null);
    setSubmitting(true);
    try {
      await cancelOrder(orderId, {
        reason: reason as CancelReason,
        explanation: explanation.trim(),
      });
      setStatus("success");
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : null);
      setStatus("fail");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/25">
      <WhiteCard>
        {status === "success" ? (
          <div className="space-y-5">
            <Image
              src="/icons/success.svg"
              alt=""
              width={64}
              height={64}
              className="mx-auto"
            />
            <p className="text-center text-h3 text-primary-500">Success</p>
            <p className="text-center text-small text-primary-500">
              Your order has been cancelled.
            </p>
            <Button variant="light" onClick={close} className="w-full">
              Close
            </Button>
          </div>
        ) : status === "fail" ? (
          <div className="space-y-6">
            <Image
              src="/icons/fail.svg"
              alt=""
              width={64}
              height={64}
              className="mx-auto"
            />
            <p className="text-center text-h3 text-primary-500">
              Unable to cancel this order.
            </p>
            {errorMessage && (
              <p className="text-center text-small text-primary-500">
                {errorMessage}
              </p>
            )}
            <Button variant="light" onClick={close} className="w-full">
              Close
            </Button>
          </div>
        ) : (
          <div className="mt-6 w-full space-y-4">
            <span className="mx-auto flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-subtle font-bold text-white">
              i
            </span>
            <h2 className="text-center text-h2 text-primary-500">
              Cancel this order?
            </h2>
            <p className="text-center text-small text-primary-500">
              {orderLabel}
            </p>

            <div>
              <p className="mb-1 text-body font-bold">
                Reason for cancellation *
              </p>
              <SelectInput
                value={reason}
                onChange={(value) => {
                  setReason(value);
                  setValidation(null);
                }}
                options={CUSTOMER_REASONS}
                formatSelectedLabel={(label) =>
                  reason ? label : "Select reason for cancellation"
                }
              />
            </div>

            <div>
              <p className="mb-1 text-body font-bold">
                Please explain your reason *
              </p>
              <textarea
                value={explanation}
                onChange={(e) => {
                  setExplanation(e.target.value);
                  setValidation(null);
                }}
                rows={4}
                placeholder="Briefly explain why you're cancelling…"
                className="w-full rounded-md border border-gray-300 bg-white p-2"
              />
              {validation && (
                <p className="mt-1 text-sm text-red-500">{validation}</p>
              )}
            </div>

            <div className="rounded-md bg-red-50 p-3 text-small text-red-600">
              The artist will be notified immediately after cancellation. This
              action cannot be undone.
            </div>

            <div className="flex w-full justify-between gap-4">
              <Button variant="light" onClick={close} disabled={submitting}>
                Keep Order
              </Button>
              <Button variant="red" onClick={submit} disabled={submitting}>
                {submitting ? "Cancelling..." : "Cancel Order"}
              </Button>
            </div>
          </div>
        )}
      </WhiteCard>
    </div>
  );
}
