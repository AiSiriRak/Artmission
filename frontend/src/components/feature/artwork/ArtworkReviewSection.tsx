"use client";

import { useEffect, useState } from "react";

import ReviewList from "../review/ReviewList";
import {
  getArtworkReviews,
  type ReviewData,
} from "@/lib/api/reviews";

interface ArtworkReviewSectionProps {
  artworkId: string;
}

export default function ArtworkReviewSection({
  artworkId,
}: ArtworkReviewSectionProps) {
  const [reviews, setReviews] = useState<ReviewData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    const fetchReviews = async () => {
      setLoading(true);

      try {
        const data = await getArtworkReviews(artworkId);

        if (!cancelled) {
          setReviews(data);
        }
      } catch (error) {
        console.error("Failed to fetch artwork reviews:", error);

        if (!cancelled) {
          setReviews([]);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    fetchReviews();

    return () => {
      cancelled = true;
    };

  }, [artworkId]);

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