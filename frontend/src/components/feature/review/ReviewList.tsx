import { useState } from "react";
import { Button } from "@/components/ui/Button";

import type { ReviewData } from "@/lib/api/reviews";
import ReviewCard from "./ReviewCard";

interface ReviewListProps {
  title?: string;
  reviews: ReviewData[];
}

const INITIAL_COUNT = 5;
const LOAD_MORE_COUNT = 5;

export default function ReviewList({ title, reviews }: ReviewListProps) {
  const [visibleCount, setVisibleCount] = useState(INITIAL_COUNT);

  if (!reviews || reviews.length === 0) return null;

  const visibleReviews = reviews.slice(0, visibleCount);
  const hasMore = visibleCount < reviews.length;

  return (
    <div className="border border-gray-200 rounded-3xl p-6 md:p-8 bg-white flex flex-col gap-4">
      {/* Show title if exist */}
      {title && (
        <h3 className="text-lg font-bold text-gray-900 mb-2">{title}</h3>
      )}

      {/* Reviews list */}
      <div className="flex flex-col gap-4">
        {visibleReviews.map((review) => (
          <ReviewCard key={review.id} review={review} />
        ))}
      </div>

      {/* Show more */}
      {hasMore && (
        <div className="flex justify-center mt-2">
          <Button
            variant="light"
            onClick={() =>
              setVisibleCount((prev) => prev + LOAD_MORE_COUNT)
            }
          >
            Show More
          </Button>
        </div>
      )}
    </div>
  );
}
