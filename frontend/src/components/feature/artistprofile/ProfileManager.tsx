"use client";

import { useState } from "react";
import { useRouter } from "next/navigation"; 
import { updateArtistProfile } from "@/lib/api/artists"; 
import type { UpdateArtistInput, ArtistProfile, Artwork } from "@/lib/api/types";

// กำหนด Type ของ Review ไว้ในนี้ชั่วคราว (จนกว่า Backend จะมี Review Type)
export interface ReviewData {
  id: string | number;
  reviewerName: string;
  timeAgo: string;
  orderName: string;
  rating: number;
  comment: string;
  avatarUrl?: string;
}

// Custom / Local Components
import ProfileSidebar from "./ProfileSidebar";
import EditorField from "./EditorField";
import ArtworkCard from "./ArtworkCard";
import TagList from "./TagList";
import ProfileImage from "./ProfileImage";
import ArtworkDetail from "./ArtworkDetail";
import ReviewList from "./ReviewList";

// Shared UI Components
import { Button } from "@/components/ui/Button";
import { Loading } from "@/components/ui/Loading";

interface ProfileManagerProps {
  initialProfile: ArtistProfile;
  initialArtworks: Artwork[];
  initialReviews?: ReviewData[];
}

export default function ProfileManager({ 
  initialProfile, 
  initialArtworks, 
  initialReviews = []
}: ProfileManagerProps) {

  const router = useRouter();

  const [isCustomerMode, setIsCustomerMode] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false); 

  const [selectedArtwork, setSelectedArtwork] = useState<Artwork | 'new' | null>(null);
  const [artworks, setArtworks] = useState<Artwork[]>(initialArtworks);

  const [reviews, setReviews] = useState<ReviewData[]>(
    initialReviews.length > 0 ? initialReviews : [
      { id: 1, reviewerName: "Name", timeAgo: "2 hrs ago", orderName: "Pixel Art", rating: 3.5, comment: "งานน่ารักมากๆๆๆ ❤️❤️❤️" },
      { id: 2, reviewerName: "Name", timeAgo: "2 hrs ago", orderName: "Pixel Art", rating: 3.5, comment: "งานน่ารักมากๆๆๆ ❤️❤️❤️" },
      { id: 3, reviewerName: "Name", timeAgo: "2 hrs ago", orderName: "Pixel Art", rating: 3.5, comment: "งานน่ารักมากๆๆๆ ❤️❤️❤️" },
    ]
  );
  
  const [savedData, setSavedData] = useState<ArtistProfile>(initialProfile);
  const [artistData, setArtistData] = useState<ArtistProfile>(initialProfile);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setArtistData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      // ส่งเฉพาะฟิลด์ที่ UpdateArtistInput อนุญาตให้อัปเดต
      const payload: UpdateArtistInput = {
        description: artistData.description || "",
      };

      const updatedProfile = await updateArtistProfile(payload);

      const updatedData: ArtistProfile = {
        ...artistData,
        description: updatedProfile.description || artistData.description,
        min_price_satang: updatedProfile.min_price_satang,
        max_price_satang: updatedProfile.max_price_satang,
      };

      setSavedData(updatedData);
      setArtistData(updatedData);
      setIsEditing(false);
      
      router.refresh(); 
      alert("บันทึกข้อมูลเรียบร้อยแล้ว");
      
    
    } catch (error: any) {
      console.error("Save failed:", error);

      // 🛑 เพิ่มบรรทัดนี้เพื่อปริ้นต์รายละเอียด Validation Error จาก Backend ออกมาดู
      if (error.body) {
        console.log("Detailed Validation Error:", JSON.stringify(error.body, null, 2));
      }

      alert("เกิดข้อผิดพลาดในการบันทึกข้อมูล");
    } finally {
      setIsSaving(false);
    }
  };

  const handleSaveArtwork = (savedArtwork: Artwork) => {
    setArtworks((prevArtworks) => {
      const exists = prevArtworks.some((item) => item.name === savedArtwork.name);
      if (exists) {
        return prevArtworks.map((item) => 
          item.name === savedArtwork.name ? savedArtwork : item
        );
      } else {
        return [...prevArtworks, savedArtwork];
      }
    });
  };

  const handleDeleteArtwork = (nameToDelete: string | number) => {
    setArtworks((prevArtworks) => prevArtworks.filter(art => art.name !== nameToDelete));
    setSelectedArtwork(null);
  };

  const handleToggleMode = () => {
    setIsCustomerMode(!isCustomerMode);
    if (!isCustomerMode) setIsEditing(false);
    setSelectedArtwork(null);
  };

  const handleImageChange = (newImageUrl: string) => {
    setArtistData((prev) => ({ ...prev, profile_url: newImageUrl } as ArtistProfile));
  };

  // --- ลอจิกดึงข้อมูลอัตโนมัติ (อิงตาม ArtworkView schema) ---
  const derivedCategories = [...new Set(artworks.map(art => art.category).filter(Boolean))];
  
  const derivedStyles = [...new Set(
    artworks.flatMap(art => art.styles || [])
  )].filter(Boolean);

  // คำนวณราคาจาก price_satang เป็นบาท
  const prices = artworks
    .map(art => Number(art.price_satang ? art.price_satang / 100 : 0))
    .filter(p => !isNaN(p) && p > 0);
    
  const minPrice = prices.length > 0 ? Math.min(...prices) : 0;
  const maxPrice = prices.length > 0 ? Math.max(...prices) : 0;
  
  const priceRangeText = prices.length === 0 
    ? "N/A" 
    : minPrice === maxPrice 
      ? `${minPrice.toLocaleString()} THB` 
      : `${minPrice.toLocaleString()} - ${maxPrice.toLocaleString()} THB`;

  const showEditControls = !isCustomerMode && !isEditing;
  const avgRating = reviews.length > 0 
    ? (reviews.reduce((sum, item) => sum + item.rating, 0) / reviews.length).toFixed(1)
    : "0.0";

  if (isSaving) {
    return <Loading />;
  }

  if (selectedArtwork) {
    return (
      <ArtworkDetail 
        artwork={(selectedArtwork === 'new' ? null : selectedArtwork) as any}
        onBack={() => setSelectedArtwork(null)}
        isCustomerMode={isCustomerMode}
        onSave={handleSaveArtwork as any}
        onDelete={handleDeleteArtwork as any}
      />
    );
  }

  return (
    <div className="w-full">
      
      {/* -- Profile -- */}

      {!isCustomerMode ? (
        
        // Editor mode
        <div className="max-w-5xl mx-auto px-8 pt-10">
          <h1 className="text-h3 font-bold mb-10">Your Information</h1>

          <div className="mb-12">
            <div className="flex flex-col md:flex-row gap-10">
              <ProfileSidebar 
                imageUrl={(artistData as any).profile_url || (artistData as any).profileImage || ""}
                isEditing={isEditing}
                onEdit={() => setIsEditing(true)}
                onSave={handleSave}
                onCancel={() => { setIsEditing(false); setArtistData(savedData); }}
                onToggleView={handleToggleMode}
                onImageChange={handleImageChange}
              />

              <div className="flex-1">
                <EditorField 
                  label="Profile Name" 
                  name="artist_name" 
                  value={(artistData as any).artist_name || (artistData as any).name || ""} 
                  isEditing={false} 
                  onChange={handleChange} 
                />
                <EditorField 
                  label="Description" 
                  name="description" 
                  value={artistData.description || ""} 
                  isEditing={isEditing} 
                  onChange={handleChange} 
                  isTextArea 
                />
                
                <div className="grid grid-cols-2 gap-6 mt-8">
                  <div>
                    <span className="block text-body font-bold text-gray-900 mb-3">Category:</span>
                    {derivedCategories.length > 0 ? (
                      <TagList items={derivedCategories} variant="category" />
                    ) : (
                      <span className="text-gray-400 text-sm">No categories</span>
                    )}
                  </div>
                  <div>
                    <span className="block text-body font-bold text-gray-900 mb-3">Style:</span>
                    {derivedStyles.length > 0 ? (
                      <TagList items={derivedStyles} variant="style" />
                    ) : (
                      <span className="text-gray-400 text-sm">No styles</span>
                    )}
                  </div>
                </div>

                <div className="mt-8">
                  <span className="block text-body font-bold text-gray-900 mb-3">Price Range:</span>
                  <span className="text-gray-700 text-base">{priceRangeText}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

      ) : (

        // Preview mode
        <div className="w-full">
          <div className="w-full h-48 bg-secondary-300 border border-neutral"></div>
          
          <div className="max-w-5xl mx-auto px-8 relative">
            <div className="flex justify-between items-end -mt-16 sm:-mt-20 mb-6 relative z-10">
              <ProfileImage 
                imageUrl={(artistData as any).profile_url || (artistData as any).profileImage || ""}
                className="w-36 h-36 md:w-44 md:h-44"
              />

              <Button 
                onClick={handleToggleMode} 
                variant="dark" 
                icon={
                  <img 
                    src="/icons/exit.svg" 
                    alt="Exit" 
                    className="w-4 h-4 object-contain" 
                  />
                } 
              >
                Exit Preview
              </Button>
            </div>

            <div className="mb-6">
              <h1 className="text-h2 font-bold text-gray-900">
                {(artistData as any).artist_name || (artistData as any).name || "Unknown Artist"}
              </h1>
            </div>

            <p className="text-gray-700 text-sm md:text-base leading-relaxed mb-8 whitespace-pre-wrap">
              {artistData.description}
            </p>

            <div className="flex flex-wrap gap-x-16 gap-y-6 mb-12">
              <div>
                <span className="block text-body font-bold text-gray-800 mb-3">Category:</span>
                {derivedCategories.length > 0 ? (
                  <TagList items={derivedCategories} variant="category"/>
                ) : (
                  <span className="text-gray-400 text-sm">None</span>
                )}
              </div>
              <div>
                <span className="block text-body font-bold text-gray-800 mb-3">Style:</span>
                {derivedStyles.length > 0 ? (
                  <TagList items={derivedStyles} variant="style"/>
                ) : (
                  <span className="text-gray-400 text-sm">None</span>
                )}
              </div>
              <div>
                <span className="block text-body font-bold text-gray-800 mb-3">Price Range:</span>
                <span className="font-bold text-lg text-gray-800">{priceRangeText}</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* -- Artwork -- */}
      <div className="max-w-5xl mx-auto px-8 pb-10">
        {isCustomerMode ? (
          <div className="flex items-center gap-2 bg-secondary-300 border-l-4 border-accent-500 px-4 py-2.5 mb-6 rounded-r-md text-h3 font-bold text-gray-900">
            <span>Artworks</span>
          </div>
        ) : (
          <>
            <hr className="border-gray-200 mb-10" />
            <div className="flex justify-between items-center mb-6">
              <div className="flex items-center gap-3">
                <h2 className="text-h3 font-bold text-gray-900">Your Artwork</h2>
                <span className="text-sm font-normal text-gray-400">{artworks.length} samples</span>
              </div>
            </div>
          </>
        )}
        
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {showEditControls && (
            <div 
              onClick={() => setSelectedArtwork('new')}
              className="border-2 border-dashed border-accent-300 rounded-2xl flex flex-col items-center justify-center text-accent-500 cursor-pointer hover:bg-accent-50 transition-colors h-full min-h-[250px]"
            >
              <span className="font-bold mb-3">Add your artwork</span>
              <div className="w-12 h-12 border border-accent-300 rounded-md flex items-center justify-center text-2xl">
                +
              </div>
            </div>
          )}

          {artworks.map((art, idx) => (
            <ArtworkCard 
              key={(art as any).id || art.name || idx} 
              artwork={art as any} 
              showEditControls={showEditControls}
              onClick={() => {
                if (!isEditing) {
                  setSelectedArtwork(art);
                }
              }}
            />
          ))}
        </div>
      </div>

      {/* -- Review -- */}
      <div className="max-w-5xl mx-auto px-8 pb-20">
        {isCustomerMode ? (
          <div className="flex items-center gap-2 bg-secondary-300 border-l-4 border-accent-500 px-4 py-2.5 mb-6 rounded-r-md text-h3 font-bold text-gray-900">
            <span>Reviews</span>
            <span className="text-red-400 text-base ml-1">☆</span>
            <span>{avgRating}/5</span>
          </div>
        ) : (
          <>
            <hr className="border-gray-200 mb-6" />
            <div className="flex justify-between items-center mb-6">
              <div className="flex items-center gap-3">
                <h2 className="text-h3 font-bold text-gray-900">Your Reviews</h2>
                <span className="text-sm font-normal text-gray-400">{reviews.length} reviews</span>
              </div>
              <div className="flex items-center gap-1 font-bold text-gray-900 text-base">
                <span className="text-red-400 text-lg">☆</span>
                <span>{avgRating}/5</span>
              </div>
            </div>
          </>
        )}

        <ReviewList reviews={reviews as any} />
      </div>

    </div>
  );
}