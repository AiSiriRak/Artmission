"use client";

import { WhiteCard } from "@/components/ui/WhiteCard";
import Image from "next/image";

import type { OrderCancellation } from "@/lib/api/order-details";
import { CUSTOMER_REASONS } from "../customer/CustomerCancelOrderPopUp";
import { reasons as ARTIST_REASONS } from "../artist/popups/CancelOrderPopup";

type CancelActor = "customer" | "artist" | "neutral";

interface CancelReasonInfo {
  text: string;
  actor: CancelActor;
}

// format cancel reason to display-friendly text and its tone
// neutral: Mutual agreement, Other, Artist not confirm, Overdue deadline
// customer: Placed order by mistake, No longer needed, Artist cannot meet new requirements, Artist unresponsive, Taking too long
// artist: Unable to complete, Unable to meet deadline, Customer change requirements. Customer unresponsive, Personal or Technical issue
function formatCancelReason(reason: string): CancelReasonInfo {
  switch (reason) {
    // neutral
    case "MUTUAL_AGREEMENT":
      return {
        text: "Both participants agreed to cancel",
        actor: "neutral",
      };
    case "OTHER":
      return {
        text: "Other reason",
        actor: "neutral",
      };
    case "ARTIST_DID_NOT_CONFIRM_IN_TIME":
      return {
        text: "Artist did not confirm the order in time.",
        actor: "neutral",
      };
    case "NOT_SUBMITTED_WITHIN_DEADLINE":
      return {
        text: "Artist did not submit the deliverable on time.",
        actor: "neutral",
      };
  }

  // customer
  const customerMatch = CUSTOMER_REASONS.find(
    (item) => item.value === reason,
  );

  if (customerMatch) {
    return {
      text: customerMatch.label,
      actor: "customer",
    };
  }

  // artist
  const artistMatch = ARTIST_REASONS.find(
    (item) => item.value === reason,
  );

  if (artistMatch) {
    return {
      text: artistMatch.label,
      actor: "artist",
    };
  }

  return {
    text: reason || "Not specified",
    actor: "neutral",
  };
}

const CANCEL_STATEMENTS: Record<CancelActor, string> = {
  neutral: "This order has been cancelled.",
  customer: "The customer cancelled this order.",
  artist: "The artist cancelled this order.",
};

const mockCancellation = {
  order_id: "642edf72-357c-4419-8816-fd64d80574ca",
  reason: "PLACED_ORDER_BY_MISTAKE",
  explanation:
    "I am so sorry bro.",
  created_at: "",
};

interface OrderCancelSectionProps {
  cancellation?: OrderCancellation | null;
}

export default function OrderCancelSection({
  cancellation,
}: OrderCancelSectionProps) {
  if (!cancellation) return null;

  const cancelReason = formatCancelReason(cancellation.reason);
  const cancelStatement = CANCEL_STATEMENTS[cancelReason.actor];

  return (
    <section className="mb-10">
      <WhiteCard
        margin=""
        padding="px-5 pt-3 pb-3 sm:px-7 pt-5 pb-5"
        roundsize="rounded-3xl"
        className="max-w-none bg-status-cancel-pale shadow-none"
      >
        <div className="flex items-start gap-4">
          <div className="flex flex-col">
            <div className="pt-2" />
            <Image
              src="/icons/Bell_light.svg"
              alt=""
              width={24}
              height={24}
              className="h-6 w-6 shrink-0"
            />
          </div>

          <div className="min-w-0">
            <div className="flex flex-wrap items-baseline gap-x-3">
              <h3 className="text-h4">Order Canceled</h3>
              <h1 className="text-caption">{cancelStatement}</h1>
            </div>
            <dl className="mt-2">
              <div className="flex flex-wrap items-baseline gap-x-2">
                <dt className="text-body font-bold">Reason: </dt>
                <dd className="text-caption">{cancelReason.text}</dd>
              </div>
              <div className="flex flex-wrap items-baseline gap-x-2">
                <dt className="text-body font-bold">Explanation:</dt>
                <dd className="text-caption">{cancellation.explanation}</dd>
              </div>
            </dl>
          </div>
        </div>
      </WhiteCard>
    </section>
  );
}