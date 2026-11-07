import { apiFetch } from "./client";
import { formatDistanceToNow } from "date-fns";

import type {
  Review,
  ReviewDetail,
  ArtistReviewDetailView,
  ArtworkReviewsOutput,
  ArtistReviewsOutput,
  SubmitReviewInput,
} from "./types";

export type {
  Review,
  SubmitReviewInput,
}

// UI types
export interface ReviewData {
  id: string;
  reviewerName: string;
  timeAgo: string;
  orderName: string;
  rating: number;
  comment: string;
  avatarUrl?: string;
}

export interface ArtistReviewData extends ReviewData {
  artworkName: string;
}

// Map API Response to UI data
function mapReview(
  review: ReviewDetail
): ReviewData {
  return {
    id: review.id,
    reviewerName: review.username,
    timeAgo: formatDistanceToNow(new Date(review.created_at), {
      addSuffix: true,
    }),
    orderName: review.order,
    rating: review.rating,
    comment: review.comment ?? "",
  };
}

function mapArtistReview(
  review: ArtistReviewDetailView,
): ArtistReviewData {
  return {
    id: review.id,
    reviewerName: review.username,
    timeAgo: formatDistanceToNow(new Date(review.created_at), {
      addSuffix: true,
    }),
    artworkName: review.artwork_name,
    orderName: review.order,
    rating: review.rating,
    comment: review.comment ?? "",
  };
}

// Create review
export async function createReview(
  orderId: string,
  data: SubmitReviewInput,
): Promise<Review> {
  return apiFetch<Review>(
    `/orders/${encodeURIComponent(orderId)}/review`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    },
  );
}

// Get review by order ID
export async function getReviewByOrderId(
  orderId: string,
): Promise<Review> {
  return apiFetch<Review>(
    `/orders/${encodeURIComponent(orderId)}/review`,
  );
}

// Get reviews by artwork
export async function getArtworkReviews(
  artworkId: string,
  limit?: number,
  offset?: number,
): Promise<ReviewData[]> {
  const params = new URLSearchParams();

  if (limit !== undefined) params.set("limit", String(limit));
  if (offset !== undefined) params.set("offset", String(offset));

  const query = params.toString();
  const path = `/artworks/${encodeURIComponent(artworkId)}/reviews${query ? `?${query}` : ""}`;

  const result = await apiFetch<ArtworkReviewsOutput>(path);

  return (result.reviews ?? []).map(mapReview);
}

// Get reviews by artist
export async function getArtistReviews(
  artistId: string,
  limit?: number,
  offset?: number,
): Promise<ArtistReviewData[]> {
  const params = new URLSearchParams();

  if (limit !== undefined) params.set("limit", String(limit));
  if (offset !== undefined) params.set("offset", String(offset));

  const query = params.toString();
  const path = `/artists/${encodeURIComponent(artistId)}/reviews${query ? `?${query}` : ""}`;

  const result = await apiFetch<ArtistReviewsOutput>(path);

  return (result.reviews ?? []).map(mapArtistReview);
}