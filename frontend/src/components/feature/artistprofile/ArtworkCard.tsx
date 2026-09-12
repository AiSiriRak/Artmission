import type { Artwork } from "@/lib/api/types";

interface ArtworkCardProps {
  artwork: Artwork;
  showEditControls: boolean;
  onClick: () => void;
}

export default function ArtworkCard({ artwork, showEditControls, onClick }: ArtworkCardProps) {
  // ดึงรูปภาพแรกจาก artwork_samples ตาม Schema ใหม่ ถ้าไม่มีให้ใช้ placeholder
  const imageUrl = artwork.artwork_samples?.[0]?.image_url || "/placeholder.png";
  
  // แปลงราคาจากสตางค์เป็นบาทให้ปลอดภัย
  const displayPrice = artwork.price_satang != null
    ? (artwork.price_satang / 100).toLocaleString()
    : "0";

  return (
    <div 
      className="relative rounded-xl border border-gray-200 overflow-hidden cursor-pointer hover:shadow-md transition-all group h-full flex flex-col"
      onClick={onClick}
    >
      <div className="relative">
        <img 
          src={imageUrl} 
          alt={artwork.name || "Artwork"} 
          className="w-full h-56 object-cover" 
        />
        
        {showEditControls && (
          <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity z-10">
            <img 
              src="/icons/edit.svg" 
              alt="Edit" 
              className="w-6 h-6 object-contain drop-shadow-md hover:scale-110 transition-transform" 
            />
          </div>
        )}
      </div>
      
      <div className="p-4 flex flex-col flex-1 bg-neutral-200">
        <div className="flex justify-between items-start mb-1">
          <h3 className="font-bold text-gray-900">{artwork.name || "Untitled"}</h3>
        </div>
        <p className="text-xs text-accent-500 mb-1">{artwork.category || "Uncategorized"}</p>
        <div className="mt-auto flex justify-end items-center">
          <p className="font-bold text-lg">{displayPrice} THB</p>
        </div>
      </div>
    </div>
  );
}