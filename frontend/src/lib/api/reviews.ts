import { apiFetch } from "./client";
import { format, formatDistanceToNow } from "date-fns";

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

export interface ReviewData {
  id: string | number;
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

export const mockReviews: Review[] = [
  {
    id: "review-004",
    order_id: "order-004",
    customer_id: "customer-004",
    artist_id: "artist-001",
    rating: 5,
    comment: "Absolutely beautiful! Exceeded my expectations.",
    created_at: "2026-10-14T09:20:00Z",
  },
  {
    id: "review-005",
    order_id: "order-005",
    customer_id: "customer-005",
    artist_id: "artist-001",
    rating: 4,
    comment: "Really nice artwork. Minor details could be improved.",
    created_at: "2026-10-12T14:30:00Z",
  },
  {
    id: "review-006",
    order_id: "order-006",
    customer_id: "customer-006",
    artist_id: "artist-001",
    rating: 5,
    comment: "The final result was exactly what I wanted!",
    created_at: "2026-10-10T11:45:00Z",
  },
  {
    id: "review-007",
    order_id: "order-007",
    customer_id: "customer-007",
    artist_id: "artist-001",
    rating: 3,
    comment: "Good overall, but the delivery took longer than expected.",
    created_at: "2026-10-08T16:00:00Z",
  },
  {
    id: "review-008",
    order_id: "order-008",
    customer_id: "customer-008",
    artist_id: "artist-001",
    rating: 5,
    comment: "Fantastic attention to detail. Highly recommended!",
    created_at: "2026-10-06T10:10:00Z",
  },
  {
    id: "review-009",
    order_id: "order-009",
    customer_id: "customer-009",
    artist_id: "artist-001",
    rating: 4,
    comment: "Very talented artist. I love the color choices.",
    created_at: "2026-10-04T13:25:00Z",
  },
  {
    id: "review-010",
    order_id: "order-010",
    customer_id: "customer-010",
    artist_id: "artist-001",
    rating: 5,
    comment: "Amazing experience from start to finish.",
    created_at: "2026-10-02T08:40:00Z",
  },
  {
    id: "review-011",
    order_id: "order-011",
    customer_id: "customer-011",
    artist_id: "artist-001",
    rating: 2,
    comment: "The artwork did not fully match the original request.",
    created_at: "2026-09-30T15:15:00Z",
  },
  {
    id: "review-012",
    order_id: "order-012",
    customer_id: "customer-012",
    artist_id: "artist-001",
    rating: 5,
    comment: "Beautiful work! Communication was excellent.",
    created_at: "2026-09-28T12:30:00Z",
  },
  {
    id: "review-013",
    order_id: "order-013",
    customer_id: "customer-013",
    artist_id: "artist-001",
    rating: 4,
    comment: "Great artwork and delivered on time.",
    created_at: "2026-09-26T09:00:00Z",
  },
  {
    id: "review-014",
    order_id: "order-014",
    customer_id: "customer-014",
    artist_id: "artist-001",
    rating: 5,
    comment: "I couldn't be happier with the result!",
    created_at: "2026-09-24T17:20:00Z",
  },
  {
    id: "review-015",
    order_id: "order-015",
    customer_id: "customer-015",
    artist_id: "artist-001",
    rating: 3,
    comment: "Decent artwork, but I expected a little more detail.",
    created_at: "2026-09-22T10:50:00Z",
  },
  {
    id: "review-016",
    order_id: "order-016",
    customer_id: "customer-016",
    artist_id: "artist-001",
    rating: 5,
    comment: "Wonderful artist! Very patient with revisions.",
    created_at: "2026-09-20T14:10:00Z",
  },
  {
    id: "review-017",
    order_id: "order-017",
    customer_id: "customer-017",
    artist_id: "artist-001",
    rating: 4,
    comment: "Lovely artwork. Would definitely commission again.",
    created_at: "2026-09-18T11:35:00Z",
  },
  {
    id: "review-018",
    order_id: "order-018",
    customer_id: "customer-018",
    artist_id: "artist-001",
    rating: 5,
    comment: "Exceeded my expectations. Every detail is perfect.",
    created_at: "2026-09-16T08:25:00Z",
  },
  {
    id: "review-019",
    order_id: "order-019",
    customer_id: "customer-019",
    artist_id: "artist-001",
    rating: 1,
    comment: "Unfortunately, the final artwork was quite different from what I requested.",
    created_at: "2026-09-14T16:45:00Z",
  },
  {
    id: "review-020",
    order_id: "order-020",
    customer_id: "customer-020",
    artist_id: "artist-001",
    rating: 4,
    comment: "Good quality and friendly communication.",
    created_at: "2026-09-12T13:15:00Z",
  },
  {
    id: "review-021",
    order_id: "order-021",
    customer_id: "customer-021",
    artist_id: "artist-001",
    rating: 5,
    comment: "Absolutely stunning! The artist captured exactly what I imagined.",
    created_at: "2026-09-10T09:40:00Z",
  },
  {
    id: "review-022",
    order_id: "order-022",
    customer_id: "customer-022",
    artist_id: "artist-001",
    rating: 3,
    comment: "",
    created_at: "2026-09-08T15:30:00Z",
  },
  {
    id: "review-023",
    order_id: "order-023",
    customer_id: "customer-023",
    artist_id: "artist-001",
    rating: 5,
    comment: "Professional, creative, and very easy to work with!",
    created_at: "2026-09-06T10:05:00Z",
  },
];

interface MockArtwork {
  id: string;
  name: string;
}

export const mockArtworks: MockArtwork[] = [
  {
    id: "artwork-001",
    name: "Sunset in the Mountains",
  },
  {
    id: "artwork-002",
    name: "A Quiet Afternoon",
  },
  {
    id: "artwork-003",
    name: "Dreamy Garden",
  },
];

const mockReviewArtworkMap: Record<string, string> = {
  "review-004": "artwork-001",
  "review-005": "artwork-002",
  "review-006": "artwork-003",
  "review-007": "artwork-001",
  "review-008": "artwork-002",
  "review-009": "artwork-003",
  "review-010": "artwork-001",
  "review-011": "artwork-002",
  "review-012": "artwork-003",
  "review-013": "artwork-001",
  "review-014": "artwork-002",
  "review-015": "artwork-003",
  "review-016": "artwork-001",
  "review-017": "artwork-002",
  "review-018": "artwork-003",
  "review-019": "artwork-001",
  "review-020": "artwork-002",
  "review-021": "artwork-003",
  "review-022": "artwork-001",
  "review-023": "artwork-002",
};

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

export function getArtistReviews(
  artistId: string,
): ArtistReviewData[] {
  return mockReviews
    // .filter((review) => review.artist_id === artistId)
    .map((review) => {
      const artworkId = mockReviewArtworkMap[review.id];
      const artwork = mockArtworks.find(
        (item) => item.id === artworkId,
      );

      return {
        id: review.id,
        reviewerName: "Customer",
        timeAgo: formatDistanceToNow(
          new Date(review.created_at),
        ),
        orderName: "Artwork Order",
        artworkName: artwork?.name ?? "Unknown Artwork",
        rating: review.rating,
        comment: review.comment ?? "",
      };
    });
}

export function getArtworkReviews(
  artworkId: string,
): ReviewData[] {
  return mockReviews
    .map((review) => ({
      id: review.id,
      reviewerName: "Customer",
      timeAgo: formatDistanceToNow(
        new Date(review.created_at),
      ),
      orderName: "Artwork Order",
      rating: review.rating,
      comment: review.comment ?? "",
    }));
}