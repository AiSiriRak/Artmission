"use client";

import Image from "next/image";
import { useId, useState } from "react";

import { Button } from "@/components/ui/Button";
import { SelectInput } from "@/components/ui/SelectInput";
import { cancelOrder } from "@/lib/api/orders";
import type { CancelOrderInput } from "@/lib/api/types";
import ArtistOrderPopup from "../artist/popups/ArtistOrderPopup";

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
  const explanationId = useId();
  const [reason, setReason] = useState<string>("");
  const [explanation, setExplanation] = useState("");
  const [validation, setValidation] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [status, setStatus] = useState<"default" | "success" | "fail">(
    "default",
  );
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  if (!isOpen) return null;

  const close = () => {
    if (submitting) return;
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
    <ArtistOrderPopup
      title={
        status === "success"
          ? "Success"
          : status === "fail"
            ? "Unable to cancel this order."
            : "Cancel this order?"
      }
      form
      onClose={close}
      icon={
        <Image
          src={
            status === "success"
              ? "/icons/success.svg"
              : status === "fail"
                ? "/icons/fail.svg"
                : "/icons/Error_Circle.svg"
          }
          alt=""
          width={64}
          height={64}
          className="h-16 w-16"
        />
      }
    >
      {status === "success" ? (
        <div className="space-y-5">
          <p className="text-center text-small">
            Your order has been cancelled.
          </p>
          <Button variant="light" onClick={close} className="w-full">
            Close
          </Button>
        </div>
      ) : status === "fail" ? (
        <div className="space-y-6">
          {errorMessage && (
            <p className="text-center text-small">{errorMessage}</p>
          )}
          <Button variant="light" onClick={close} className="w-full">
            Close
          </Button>
        </div>
      ) : (
        <>
          <p className="mx-auto mt-3 text-center text-small">
            The artist will be notified and work on this order will stop. This action cannot be undone.
          </p>
          <div className="my-3 rounded-xl bg-neutral-400 px-4 py-2 text-small">
            {orderLabel}
          </div>
          <form
            onSubmit={(event) => {
              event.preventDefault();
              void submit();
            }}
          >
            <p className="mb-2 text-caption">
              Reason for cancellation <span className="text-error">*</span>
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
              buttonClassName="rounded-lg border-neutral-400"
            />
            <label
              htmlFor={explanationId}
              className="mb-2 mt-3 block text-caption"
            >
              Please explain your reason <span className="text-error">*</span>
            </label>
            <textarea
              id={explanationId}
              value={explanation}
              onChange={(event) => {
                setExplanation(event.target.value);
                setValidation(null);
              }}
              rows={4}
              placeholder="Briefly explain why you're cancelling..."
              className="h-20 w-full resize-none rounded-xl border-2 border-neutral-400 p-3 text-small"
            />
            {validation && (
              <p role="alert" className="mt-1 text-small text-error">
                {validation}
              </p>
            )}
            <div className="mt-3 flex items-center gap-3 rounded-xl bg-status-notpaid-pale px-4 py-3 text-small text-blue-900">
              <Image
                src="/icons/info.svg"
                alt=""
                width={30}
                height={27}
                className="shrink-0"
                style={{ filter: "invert(34%) sepia(92%) saturate(749%) hue-rotate(180deg) brightness(95%) contrast(92%)" }}
              />
              <div className="flex-1">
                <p className="text-small font-bold">Refund information</p>
                <p>If you cancel this order now, you’ll receive a 50% refund of the order total.</p>
               </div>
            </div>
            <div className="mt-3 flex items-center gap-3 rounded-xl bg-accent-200 px-4 py-3 text-small text-error">
              <Image
                src="/icons/pink-error-triangle.svg"
                alt=""
                width={30}
                height={27}
                className="shrink-0"
              />
              <p>The artist will be notified immediately after cancellation.</p>
            </div>
            <div className="mt-4 flex justify-end gap-6">
              <Button
                type="button"
                variant="transparent"
                className="w-[135px] rounded-xl"
                disabled={submitting}
                onClick={close}
              >
                Keep order
              </Button>
              <Button
                type="submit"
                variant="error"
                className="w-[135px] rounded-xl"
                disabled={submitting}
              >
                {submitting ? "Cancelling..." : "Cancel order"}
              </Button>
            </div>
          </form>
        </>
      )}
    </ArtistOrderPopup>
  );
}
