import Image from "next/image";

import type {
  ReviewData,
  ArtistReviewData,
} from "@/lib/api/reviews";

export function ReviewStars({
  rating,
  size = 24,
}: {
  rating: number;
  size?: number;
}) {
  return (
    <div className="flex items-center gap-2">
      {[...Array(5)].map((_, index) => {
        const filled = index < rating;
        const starSize = filled ? size : size * 1.3;

        return (
          <Image
            key={index}
            src={
              filled
                ? "/icons/Star_filled.svg"
                : "/icons/Star_light.svg"
            }
            alt=""
            width={starSize}
            height={starSize}
            style={{
              width: starSize,
              height: starSize,
            }}
            className="shrink-0 object-contain"
          />
        );
      })}
    </div>
  );
}

export default function ReviewCard({ review }: { review: ReviewData | ArtistReviewData }) {
  return (
    <div className="bg-secondary-300 rounded-2xl p-5 flex items-start gap-4">
      {/* Profile Image */}
      <div className="w-12 h-12 rounded-full bg-gray-300 flex-shrink-0 overflow-hidden">
        <Image
          src={review.avatarUrl || "/default-avatar.png"}
          alt={review.reviewerName || "Reviewer"}
          width={48}
          height={48}
          className="w-auto h-full object-cover"
        />
      </div>

      {/* Review Information */}
      <div className="flex-1 text-sm text-gray-800 leading-relaxed gap-4">
        <div className="flex items-center gap-2 mb-1">
          <span className="font-bold text-gray-900">
            {review.reviewerName || "Name"}
          </span>
          <span className="text-gray-400 text-xs">
            • {review.timeAgo || "2 hrs ago"}
          </span>
        </div>

        {/* Artwork Name - Artist page only */}
        {"artworkName" in review && (
          <p className="font-semibold text-gray-900">
            Artwork:{" "}
            <span className="font-normal">
              {review.artworkName}
            </span>
          </p>
        )}

        <p className="font-semibold text-gray-900">
          Order:{" "}
          <span className="font-normal">{review.orderName || "Pixel Art"}</span>
        </p>
        <div className="flex items-center gap-2">
          <span className="font-semibold text-gray-900">Rating:</span>
          <ReviewStars rating={review.rating} size={14} />
        </div>

        <p className="mt-2 text-gray-700">{review.comment}</p>
      </div>
    </div>
  );
}
