import { ReviewData } from "../../../app/artist/[id]/types";

interface ReviewListProps {
  title?: string; // 📍 ใส่ ? เพื่อให้ title เป็น optional (ไม่ส่งมาได้)
  reviews: ReviewData[];
}

export default function ReviewList({ title, reviews }: ReviewListProps) {
  return (
    <div className="border border-gray-200 rounded-3xl p-6 md:p-8 bg-white flex flex-col gap-4">
      
      {/* แสดงหัวข้อเฉพาะเมื่อมีการส่ง title เข้ามาเท่านั้น */}
      {title && (
        <h3 className="text-lg font-bold text-gray-900 mb-2">{title}</h3>
      )}

      {/* รายการ Reviews */}
      {reviews.map((review) => (
        <div 
          key={review.id} 
          className="bg-[#F7F2EC] rounded-2xl p-5 flex items-start gap-4"
        >
          {/* Avatar วงกลมสีเทา */}
          <div className="w-12 h-12 rounded-full bg-gray-400 shrink-0"></div>

          {/* รายละเอียด Review */}
          <div className="flex-1 text-sm text-gray-800 leading-relaxed">
            <div className="flex items-center gap-2 mb-1">
              <span className="font-bold text-gray-900">{review.reviewerName || "Name"}</span>
              <span className="text-gray-400 text-xs">• {review.timeAgo || "2 hrs ago"}</span>
            </div>
            
            <p className="font-semibold text-gray-900">
              Order: <span className="font-normal">{review.orderName || "Pixel Art"}</span>
            </p>
            <p className="font-semibold text-gray-900 mb-2">
              Rating: <span className="font-normal">{review.rating}/5</span>
            </p>
            
            <p className="text-gray-700">{review.comment}</p>
          </div>
        </div>
      ))}
    </div>
  );
}