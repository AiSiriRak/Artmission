"use client";

import { format } from "date-fns";
import { WhiteCard } from "@/components/ui/WhiteCard";

import { Review } from "@/lib/api/reviews";
import { ReviewStars } from "../../review/ReviewCard";

interface OrderReviewProps {
  review?: Review | null;
}

export default function OrderCancelSection({
  review,
}: OrderReviewProps) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center gap-3">
        <h2 className="text-h3">Review</h2>
        {review && (
          <ReviewStars rating={review.rating} size={22} />
        )}
        {review && (
          <span className="text-small text-primary-400/60">
            {format(new Date(review.created_at), "d MMM yyyy")}
          </span>
        )}
      </div>

      {review && review.comment && review.comment.trim() !== "" && (
        <section>
          <WhiteCard
            margin=""
            padding="p-6 sm:p-8"
            roundsize="rounded-3xl"
            className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
          >
            <div className="break-words text-body">
              {review.comment}
            </div>
          </WhiteCard>
        </section>
      )}

      {!review && (
        <section>
          <WhiteCard
            margin=""
            padding="p-6 sm:p-8"
            roundsize="rounded-3xl"
            className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
          >
            <div className="mt-1 flex flex-col items-center justify-center text-center">
              <img
                src="/icons/message-circle-off.svg"
                alt=""
                className="h-18 w-18 shrink-0"
              />
              <div className="mt-2 break-words text-body text-neutral">
                waiting for customer's review
              </div>
            </div>
          </WhiteCard>
        </section>
      )}
    </div>
  );
}