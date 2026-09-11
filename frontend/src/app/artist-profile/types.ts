export interface ReviewData {
  id: string | number;
  reviewerName: string;
  timeAgo: string;
  orderName: string;
  rating: number;
  comment: string;
  avatarUrl?: string;
}