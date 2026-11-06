import type { ReviewData } from "@/lib/api/reviews";
import ReviewCard from "./ReviewCard";

interface ReviewListProps {
  title?: string;
  reviews: ReviewData[];
}

export default function ReviewList({ title, reviews }: ReviewListProps) {
  if (!reviews || reviews.length === 0) return null;

  return (
    <div className="border border-gray-200 rounded-3xl p-6 md:p-8 bg-white flex flex-col gap-4">
      {/* แสดงหัวข้อเฉพาะเมื่อมีการส่ง title เข้ามา */}
      {title && (
        <h3 className="text-lg font-bold text-gray-900 mb-2">{title}</h3>
      )}

      {/* รายการ Reviews */}
      <div className="flex flex-col gap-4">
        {reviews.map((review) => (
          <ReviewCard key={review.id} review={review} />
        ))}
      </div>
    </div>
  );
}
