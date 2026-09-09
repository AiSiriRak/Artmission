import { ReviewData } from "../../../app/artist/types"; 

export default function ReviewCard({ review }: { review: ReviewData }) {
  return (
    <div className="bg-secondary-300 rounded-2xl p-5 flex items-start gap-4">
      
      {/* รูปโปรไฟล์คนรีวิว */}
      <div className="w-12 h-12 rounded-full bg-gray-300 flex-shrink-0 overflow-hidden">
        {review.avatarUrl && (
          <img 
            src={review.avatarUrl} 
            alt={review.reviewerName || "Reviewer"} 
            className="w-full h-full object-cover" 
          />
        )}
      </div>

      {/* ข้อมูลรีวิว */}
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
  );
}