"use client";
import { useState } from "react";
import { ArtistData, ArtworkData } from "../types";
import ProfileSidebar from "./ProfileSidebar";
import EditorField from "./EditorField";
import ArtworkCard from "./ArtworkCard";
import TagList from "./TagList";
import ProfileImage from "./ProfileImage";
import ArtworkDetail from "./ArtworkDetail";
import { Button } from "@/components/ui/Button";


export default function ProfileManager({ 
  initialProfile, initialArtworks 
}: { 
  initialProfile: ArtistData, initialArtworks: ArtworkData[] 
}) {
  const [isCustomerMode, setIsCustomerMode] = useState(false);
  const [isEditing, setIsEditing] = useState(false);

  const [selectedArtwork, setSelectedArtwork] = useState<ArtworkData | 'new' | null>(null);

  const [artworks, setArtworks] = useState<ArtworkData[]>(initialArtworks);
  
  const [savedData, setSavedData] = useState<ArtistData>(initialProfile);
  const [artistData, setArtistData] = useState<ArtistData>(initialProfile);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setArtistData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSave = () => {
    setSavedData(artistData);
    setIsEditing(false);
  };

  const handleSaveArtwork = (savedArtwork: ArtworkData) => {
    setArtworks((prevArtworks) => {
      const exists = prevArtworks.some((item) => item.id === savedArtwork.id);
      if (exists) {
        // แก้ไข Artwork เดิม
        return prevArtworks.map((item) => item.id === savedArtwork.id ? savedArtwork : item);
      } else {
        // เพิ่ม Artwork ใหม่
        return [...prevArtworks, savedArtwork];
      }
    });
  };

  const handleDeleteArtwork = (idToDelete: number | string) => {
    setArtworks((prevArtworks) => prevArtworks.filter(art => art.id !== idToDelete));
    setSelectedArtwork(null); // ปิดหน้า Detail และกลับสู่หน้าหลัก
  };

  const handleToggleMode = () => {
    setIsCustomerMode(!isCustomerMode);
    if (!isCustomerMode) setIsEditing(false);
    setSelectedArtwork(null);
  };

  const handleImageChange = (newImageUrl: string) => {
    setArtistData((prev) => ({ ...prev, profileImage: newImageUrl }));
  };

  // --- 📍 ลอจิกดึงข้อมูลอัตโนมัติจากผลงานทั้งหมด ---
  // 1. ดึง Category ไม่ซ้ำ
  const derivedCategories = [...new Set(artworks.map(art => art.category).filter(Boolean))];
  
  // 2. ดึง Style ไม่ซ้ำ (รองรับการคั่นด้วยเครื่องหมาย comma)
  const derivedStyles = [...new Set(
    artworks.flatMap(art => 
      art.style ? art.style.split(',').map(s => s.trim()) : []
    ).filter(Boolean)
  )];

  // 3. คำนวณช่วงราคา (Price Range)
  const prices = artworks.map(art => Number(art.price)).filter(p => !isNaN(p) && p > 0);
  const minPrice = prices.length > 0 ? Math.min(...prices) : 0;
  const maxPrice = prices.length > 0 ? Math.max(...prices) : 0;
  const priceRangeText = prices.length === 0 
    ? "N/A" 
    : minPrice === maxPrice 
      ? `${minPrice.toLocaleString()} THB` 
      : `${minPrice.toLocaleString()} - ${maxPrice.toLocaleString()} THB`;
  // ----------------------------------------------

  const showEditControls = !isCustomerMode && !isEditing;

  if (selectedArtwork) {
    return (
      <ArtworkDetail 
        artwork={selectedArtwork === 'new' ? null : selectedArtwork} 
        onBack={() => setSelectedArtwork(null)}
        isCustomerMode={isCustomerMode}
        onSave={handleSaveArtwork}
        onDelete={handleDeleteArtwork}
      />
    );
  }

  return (
    <div className="w-full">
      {/* ------------------------------------------- */}
      {/* ส่วนที่ 1: Profile Layout (แยกตามโหมด) */}
      {/* ------------------------------------------- */}
      {!isCustomerMode ? (
        
        // --- 1.1 โหมด ARTIST (มุมมองเจ้าของ) ---
        <div className="max-w-5xl mx-auto px-8 pt-10">
          <h1 className="text-2xl font-bold mb-10">Your Information</h1>
          <div className="flex flex-col md:flex-row gap-10 mb-12">
            
            <ProfileSidebar 
              imageUrl={artistData.profileImage}
              isEditing={isEditing}
              onEdit={() => setIsEditing(true)}
              onSave={handleSave}
              onCancel={() => { setIsEditing(false); setArtistData(savedData); }}
              onToggleView={handleToggleMode}
              onImageChange={handleImageChange}
            />

            <div className="flex-1">
              <EditorField label="Profile Name" name="name" value={artistData.name} isEditing={false} onChange={handleChange} />
              <EditorField label="Description" name="description" value={artistData.description} isEditing={isEditing} onChange={handleChange} isTextArea />
              
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

      ) : (

        // --- 1.2 โหมด VIEW AS (มุมมองลูกค้า) ---
        <div className="w-full">
          {/* แบนเนอร์สีครีม (ขยายเต็มจอ) */}
          <div className="w-full h-48 bg-secondary-300 border border-neutral"></div>
          
          <div className="max-w-5xl mx-auto px-8 relative">
            
            {/* 1. โซนรูปโปรไฟล์ และ ปุ่ม (จัด Flex ให้อยู่ซ้าย-ขวา) */}
            <div className="flex justify-between items-end -mt-16 sm:-mt-20 mb-6 relative z-10">
              
              <ProfileImage 
                imageUrl={artistData.profileImage}
                className="w-36 h-36 md:w-44 md:h-44"
              />

              {/* ปุ่ม Exit View As: ปรับ padding ให้น้อยลง (px-4 py-1.5) เพื่อให้ปุ่มดูเล็กกะทัดรัด และดันขึ้นด้านบนเล็กน้อยด้วย mb-6 */}
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
                className="px-4 py-1.5 text-xs sm:text-sm flex items-center gap-2 rounded-full cursor-pointer font-semibold mb-8 shadow-md"
              >
                Exit Preview
              </Button>
            </div>

            {/* 2. โซนชื่อศิลปิน (แยกบรรทัดลงมาอยู่ด้านล่างรูปและปุ่ม) */}
            <div className="mb-6">
              <h1 className="text-h2 font-bold text-gray-900">{artistData.name}</h1>
            </div>

            {/* รายละเอียด */}
            <p className="text-gray-700 text-sm md:text-base leading-relaxed mb-8">
              {artistData.description}
            </p>

            {/* ข้อมูล Tag และ ราคา (จัดเป็นแนวนอน) */}
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

      {/* ------------------------------------------- */}
      {/* ส่วนที่ 2: Artwork (ใช้ร่วมกันทั้ง 2 โหมด) */}
      {/* ------------------------------------------- */}
      <div className="max-w-5xl mx-auto px-8 pb-10">
        <hr className="border-gray-200 mb-10" />
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-bold">
            {isCustomerMode ? "Artwork" : "Your Artwork"}{" "}
            <span className="text-sm font-normal text-gray-400 ml-2">{artworks.length} samples</span>
          </h2>
        </div>
        
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          
          {/* กรอบปะ เพิ่ม Artwork (ซ่อนเมื่อไม่ได้อยู่ในสถานะที่แก้ไขได้) */}
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

          {artworks.map((art) => (
            <ArtworkCard 
              key={art.id} 
              artwork={art} 
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
    </div>
  );
}