import { ReviewData } from "../../../app/artist/[id]/types"; // 📍 หมายเหตุ: เช็ค path ตรงนี้ให้ตรงกับไฟล์ types ของคุณด้วยนะครับ

export default function ReviewCard({ review }: { review: ReviewData }) {
  return (
    <div className="flex gap-4 p-6 rounded-xl bg-secondary-400 mb-4 last:mb-0">
      
      {/* รูปโปรไฟล์คนรีวิว (ถ้าไม่มีให้แสดงวงกลมสีเทา) */}
      <div className="w-12 h-12 rounded-full bg-gray-300 flex-shrink-0 overflow-hidden">
        {review.avatarUrl && (
          <img 
            src={review.avatarUrl} 
            alt={review.reviewerName} 
            className="w-full h-full object-cover" 
          />
        )}
      </div>

      {/* ข้อมูลรีวิว */}
      <div className="flex flex-col text-sm text-gray-900 font-bold">
        <div className="mb-1">
          {review.reviewerName} <span className="text-gray-400 font-normal text-xs ml-2">• {review.timeAgo}</span>
        </div>
        <div className="mb-1">Order: {review.orderName}</div>
        <div className="mb-2">Rating: {review.rating}/5</div>
        <p className="font-normal">{review.comment}</p>
      </div>
      
    </div>
  );
}