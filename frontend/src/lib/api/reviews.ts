import { apiFetch } from "./client";
import type { } from "./types";

const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function isUuid(value: string) {
  return UUID_PATTERN.test(value);
}

function uuidList(value: unknown): string[] {
  let items: unknown[] = [];
  if (Array.isArray(value)) {
    items = value;
  } else if (typeof value === "string" && value.trim().startsWith("[")) {
    try {
      const parsed = JSON.parse(value) as unknown;
      if (Array.isArray(parsed)) items = parsed;
    } catch {
      items = [];
    }
  }

  return items.flatMap((item) => {
    if (typeof item !== "string") return [];
    const id = item.trim();
    return isUuid(id) ? [id] : [];
  });
}

export interface Review {
  id: string;
  order_id: string;
  customer_id: string;
  artist_id: string;
  rating: number;
  comment?: string;
  created_at: string;
}

export const mockReviews: Review[] = [
  {
    id: "review-001",
    order_id: "order-001",
    customer_id: "customer-001",
    artist_id: "artist-001",
    rating: 5,
    comment: "Amazing artwork! The artist did a great job.",
    created_at: "2026-10-20T10:30:00Z",
  },
  {
    id: "review-002",
    order_id: "order-002",
    customer_id: "customer-002",
    artist_id: "artist-001",
    rating: 4,
    comment: "Great work and good communication.",
    created_at: "2026-10-18T08:15:00Z",
  },
  {
    id: "review-003",
    order_id: "order-003",
    customer_id: "customer-003",
    artist_id: "artist-001",
    rating: 5,
    comment: "",
    created_at: "2026-10-15T12:00:00Z",
  },
];

export async function getReviews(
  artistId: string,
): Promise<Review[]> {
  await new Promise((resolve) => setTimeout(resolve, 500));

  return mockReviews.filter(
    (review) => review.artist_id === artistId,
  );
}

// Get review by order ID
export async function getReviewByOrderId(
  orderId: string,
): Promise<Review | null> {
  await new Promise((resolve) => setTimeout(resolve, 300));

  return (
    mockReviews.find((review) => review.order_id === orderId) ?? null
  );
}

// Create mock review
export async function createReview(
  review: Review,
): Promise<Review> {
  await new Promise((resolve) => setTimeout(resolve, 300));

  const existingIndex = mockReviews.findIndex(
    (item) => item.order_id === review.order_id,
  );

  if (existingIndex !== -1) {
    mockReviews[existingIndex] = review;
  } else {
    mockReviews.push(review);
  }

  return review;
}