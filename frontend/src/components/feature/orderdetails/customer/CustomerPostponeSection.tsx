"use client";

import { useRef, useState } from "react";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";
import { confirmDelayRequest } from "@/lib/api/orders";
import type { DelayRequest } from "@/lib/api/types";
import { CustomerConfirmPopUp } from "./CustomerConfirmPopUp";

interface CustomerPostponeSectionProps {
  orderId: string;
  delayRequest: DelayRequest;
  /** Called after the backend accepted the decision; the parent should reload the order. */
  onDecided: () => void | Promise<void>;
}

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString("en-GB", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

// format postpone reason to display-friendly text
// "NEED_MORE_TIME" | "WAITING_FOR_FEEDBACK" | "PERSONAL_REASON" | "HIGH_WORKLOAD" | "TECHNICAL_ISSUE" | "OTHER";
function formatReason(reason: string): string {
  switch (reason) {
    case "NEED_MORE_TIME":
      return "Need more time";
    case "WAITING_FOR_FEEDBACK":
      return "Waiting for feedback";
    case "PERSONAL_REASON":
      return "Personal reason";
    case "HIGH_WORKLOAD":
      return "High workload";
    case "TECHNICAL_ISSUE":
      return "Technical issue";
    case "OTHER":
      return "Other";
    default:
      return reason;
  }
}

export default function CustomerPostponeSection({
  orderId,
  delayRequest,
  onDecided,
}: CustomerPostponeSectionProps) {
  const [pendingDecision, setPendingDecision] = useState<
    "accept" | "decline" | null
  >(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const decisionPending = useRef(false);

  if (delayRequest.customer_decision !== "WAIT") return null;

  const closePopup = () => {
    if (!submitting) setPendingDecision(null);
  };

  const decide = async () => {
    if (!pendingDecision || decisionPending.current) return;
    decisionPending.current = true;
    setSubmitting(true);
    setError(null);
    try {
      await confirmDelayRequest(orderId, delayRequest.id, {
        accept: pendingDecision === "accept",
      });
      await onDecided();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      decisionPending.current = false;
      setSubmitting(false);
      setPendingDecision(null);
    }
  };

  const popupCopy =
    pendingDecision === "decline"
      ? {
          title: "Decline Postpone Request?",
          description:
            "The delivery date will remain unchanged. The artist will be notified to deliver the art by the original deadline.",
          confirmText: "Decline",
        }
      : {
          title: "Accept Postpone Request?",
          description:
            "You are agreeing to extend the delivery date. The artist will be notified to proceed with the new deadline.",
          confirmText: "Accept",
        };

  return (
    <section className="mb-10">
      <WhiteCard
        margin=""
        padding="p-6 sm:p-8"
        roundsize="rounded-3xl"
        className="max-w-none bg-status-pending-pale shadow-none"
      >
        <div className="flex items-start gap-4">
          <img
            src="/icons/alarm-clock.svg"
            alt=""
            className="mt-1 h-6 w-6 shrink-0"
          />
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-baseline gap-x-3">
              <h3 className="text-h3">Postpone Request from Artist</h3>
              <span className="text-caption text-red-500">
                * Auto declines in 48 hours if no response.
              </span>
            </div>
            <p className="mt-1 text-small">
              The artist has requested to extend the deadline.
            </p>

            <dl className="mt-6 space-y-3">
              <div>
                <dt className="text-h3">New Deadline Date</dt>
                <dd className="text-body">
                  {formatDate(delayRequest.deadline_request_at)}
                </dd>
              </div>
              <div>
                <dt className="text-h3">Reason</dt>
                <dd className="text-body">{formatReason(delayRequest.reason)}</dd>
              </div>
              <div>
                <dt className="text-h3">Details</dt>
                <dd className="whitespace-pre-wrap break-words text-body">
                  {delayRequest.explanation}
                </dd>
              </div>
            </dl>

            {error && <p className="mt-4 text-sm text-red-500">{error}</p>}

            <div className="mt-6 flex justify-end gap-4">
              <Button
                type="button"
                variant="light"
                disabled={submitting}
                onClick={() => setPendingDecision("decline")}
              >
                Decline
              </Button>
              <Button
                type="button"
                variant="dark"
                disabled={submitting}
                onClick={() => setPendingDecision("accept")}
              >
                Accept Extension
              </Button>
            </div>
          </div>
        </div>
      </WhiteCard>

      <CustomerConfirmPopUp
        isOpen={pendingDecision !== null}
        title={popupCopy.title}
        description={popupCopy.description}
        cancelText="Cancel"
        confirmText={submitting ? "Saving..." : popupCopy.confirmText}
        busy={submitting}
        onCancel={closePopup}
        onConfirm={decide}
      />
    </section>
  );
}
