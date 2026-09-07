import { useRef } from "react";
import { Button } from "@/components/ui/Button";

interface ArtworkSampleGalleryProps {
  isEditing: boolean;
  images: string[];
  onUpload: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onRemove: (index: number) => void;
}

export default function ArtworkSampleGallery({ 
  isEditing, 
  images, 
  onUpload, 
  onRemove 
}: ArtworkSampleGalleryProps) {
  
  // 📍 สร้าง ref ไว้ข้างในนี้เลย Component จะจัดการตัวเองได้ ไม่พึ่งพาไฟล์แม่
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleAddSampleClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className="mb-10">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-body font-bold text-primary-500">Artwork Samples</h3>
        
        {isEditing && (
          <>
            <Button 
              variant="transparent"
              onClick={handleAddSampleClick} // เรียกใช้ฟังก์ชันด้านบน
              className="text-button"
            >
              + Add Sample
            </Button>
            
            {/* ซ่อน input ไว้ตรงนี้ได้เลย เพราะมันผูกกับ ref ในไฟล์นี้แล้ว */}
            <input 
              type="file" 
              accept="image/*" 
              ref={fileInputRef} 
              onChange={onUpload} 
              className="hidden" 
            />
          </>
        )}
      </div>
      
      <div className="flex flex-wrap gap-4">
        {images.map((imgUrl, index) => (
          <div key={index} className="w-48 h-48 bg-white border border-gray-200 rounded-xl relative p-2 flex items-center justify-center">
            {isEditing && (
              <button 
                onClick={() => onRemove(index)}
                className="absolute top-2 right-2 bg-white shadow-md rounded-full w-6 h-6 flex items-center justify-center text-gray-500 hover:text-red-500 text-xs cursor-pointer z-10"
              >
                ✕
              </button>
            )}
            <img src={imgUrl} alt={`sample-${index}`} className="max-w-full max-h-full object-contain" />
          </div>
        ))}
      </div>
    </div>
  );
}