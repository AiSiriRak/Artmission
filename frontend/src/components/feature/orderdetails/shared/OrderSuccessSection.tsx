"use client";

import { WhiteCard } from "@/components/ui/WhiteCard";

// format order success method
// manual-success by customer on draft 1 or 2
// auto-success by system on draft 3
function formatSuccessMethod(draftCount: number, review: boolean): string {
  if (draftCount === 3) {
    return "The artist has submitted the last draft. It automatically becomes final artwork.";
  } else {
    if (review) {
      return "The customer has accepted and reviewed the artwork.";
    } else {
      return "The customer has accepted the artwork.";
    }
  }
}

interface OrderSuccessSectionProps {
  draftCount: number;
  review: boolean;
}

export default function OrderSuccessSection({
  draftCount,
  review,
}: OrderSuccessSectionProps) {
  return (
    <section className="mb-10">
      <WhiteCard
        margin=""
        padding="p-6 sm:p-8"
        roundsize="rounded-3xl"
        className="max-w-none bg-status-success-pale shadow-none"
      >
        <div className="flex items-start gap-4">
          <img
            src="/icons/Bell_light.svg"
            alt=""
            className="h-8 w-8 shrink-0"
          />
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-baseline gap-x-3">
              <h3 className="text-h3">Order Completed</h3>
            </div>
            <dl className="mt-4 space-y-3">
              <div className="flex flex-wrap items-baseline gap-x-2 text-body">
                {formatSuccessMethod(draftCount, review)}
              </div>
            </dl>
          </div>
        </div>
      </WhiteCard>
    </section>
  );
}