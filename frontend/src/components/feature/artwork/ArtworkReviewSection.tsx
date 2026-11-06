import ReviewList from "../review/ReviewList";
import { getArtworkReviews } from "@/lib/api/reviews";

interface ArtworkReviewSectionProps {
  artworkId: string;
}

export default function ArtworkReviewSection({
  artworkId,
}: ArtworkReviewSectionProps) {
  const reviews = getArtworkReviews(artworkId);

  return (
    <div className="mt-12">
      <div className="mb-6 flex items-center gap-3">
        <h2 className="text-xl font-bold text-gray-900">
          Order Reviews
        </h2>

        <span className="text-sm font-normal text-gray-400">
          {reviews.length} reviews
        </span>
      </div>

      <ReviewList reviews={reviews} />
    </div>
  );
}