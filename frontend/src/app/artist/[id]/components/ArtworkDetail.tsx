// ArtworkDetail.tsx
import { useState, useRef } from "react";
import { ArtworkData } from "../types";
import { Button } from "@/components/ui/Button";

interface ArtworkDetailProps {
  artwork?: ArtworkData | null;
  onBack: () => void;
  isCustomerMode: boolean;
  onSave?: (savedArtwork: ArtworkData) => void;
  onDelete?: (id: number | string) => void;
}

// จำลองฐานข้อมูล Category และ Style ที่มีในระบบ (นำไปต่อ API ได้)
const AVAILABLE_CATEGORIES = ["Book", "Comic", "Game", "Animation", "Portrait", "Illustration", "Other"];
const AVAILABLE_STYLES = ["Pixel Art", "Pixel", "Pixel 8 bit", "Pixel 16 bit", "Pixel 32 bit", "Water Color", "Cartoon", "Graphic", "Anime", "Realism"];

export default function ArtworkDetail({ artwork, onBack, isCustomerMode, onSave, onDelete }: ArtworkDetailProps) {
  const [isEditing, setIsEditing] = useState(!artwork);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  
  const [savedData, setSavedData] = useState({
    name: artwork?.name || "",
    category: artwork?.category || "",
    style: artwork?.style || "",
    description: artwork?.description || "", 
    deadline: artwork?.deadline || 1,
    price: artwork?.price || 0,
  });

  const [savedImages, setSavedImages] = useState<string[]>(
    artwork?.images && artwork.images.length > 0 
      ? artwork.images 
      : [artwork?.coverImage || "/placeholder.jpg"]
  );

  const [formData, setFormData] = useState({ ...savedData });
  const [images, setImages] = useState<string[]>([...savedImages]);

  // 📌 2. State สำหรับระบบ Search และ Dropdown
  const [catSearch, setCatSearch] = useState("");
  const [styleSearch, setStyleSearch] = useState("");
  const [showCatDropdown, setShowCatDropdown] = useState(false);
  const [showStyleDropdown, setShowStyleDropdown] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      const imageUrl = URL.createObjectURL(file);
      setImages((prev) => [...prev, imageUrl]);
    }
  };

  const handleRemoveImage = (indexToRemove: number) => {
    setImages((prev) => prev.filter((_, index) => index !== indexToRemove));
  };

  const handleEditClick = () => {
    setFormData({ ...savedData });
    setImages([...savedImages]);
    setIsEditing(true);
  };

  const handleSave = () => {
    setSavedData({ ...formData });
    setSavedImages([...images]);
    setIsEditing(false);

    if (onSave) {
        onSave({
            ...artwork, 
            id: artwork?.id || Date.now(), 
            name: formData.name,
            category: formData.category,
            style: formData.style,
            price: formData.price,
            coverImage: images.length > 0 ? images[0] : "/placeholder.jpg",
            description: formData.description,
            deadline: Number(formData.deadline),
            images: [...images],
        });
    }
  };

  const handleCancel = () => {
    setIsEditing(false);
  };

  const confirmDelete = () => {
    if (onDelete && artwork?.id) {
      onDelete(artwork.id); 
    }
  };

  const displayImages = isEditing ? images : savedImages;

  // 📌 3. ฟังก์ชันและตัวแปรจัดการ Tags (Category & Style)
  const currentStyles = formData.style ? formData.style.split(',').map(s => s.trim()).filter(Boolean) : [];
  
  const filteredCategories = AVAILABLE_CATEGORIES.filter(c => c.toLowerCase().includes(catSearch.toLowerCase()));
  const filteredStyles = AVAILABLE_STYLES.filter(s => s.toLowerCase().includes(styleSearch.toLowerCase()));

  const addStyle = (style: string) => {
    if (!currentStyles.includes(style)) {
      setFormData(prev => ({ ...prev, style: [...currentStyles, style].join(', ') }));
    }
    setStyleSearch("");
    setShowStyleDropdown(false);
  };

  const removeStyle = (styleToRemove: string) => {
    setFormData(prev => ({ 
      ...prev, 
      style: currentStyles.filter(s => s !== styleToRemove).join(', ') 
    }));
  };

  const selectCategory = (cat: string) => {
    setFormData(prev => ({ ...prev, category: cat }));
    setCatSearch("");
    setShowCatDropdown(false);
  };

  return (
<>
      <div className="max-w-5xl mx-auto px-8 pt-10 pb-20">
        <Button 
          variant="light"
          onClick={onBack} 
          className="mb-6 text-button flex items-center gap-2"
        >
          <span>←</span> Back
        </Button>

        <div className="border border-primary-500 rounded-3xl p-10 bg-secondary-200 shadow-sm">
          
          <div className="flex justify-between items-start mb-8">
            <div className="w-full max-w-xl">
              {isEditing ? (
                <div className="space-y-1 mb-6">
                  <label className="text-body font-bold text-primary-500 block">Artwork Name</label>
                  <input 
                    type="text" 
                    name="name"
                    value={formData.name} 
                    onChange={handleInputChange}
                    placeholder="e.g., Pet Portrait" 
                    className="border border-primary-500 p-2.5 w-full rounded-lg bg-white" 
                  />
                </div>
              ) : (
                <div className="mb-4">
                  <h1 className="text-h1 font-bold text-gray-900 mb-4">{savedData.name || "Untitled"}</h1>
                  
                  {/* 📍 นำ Tags มาเรียงต่อกันใต้ชื่อตอน View Mode */}
                  <div className="flex flex-wrap gap-2">
                    {/* แท็ก Category */}
                    {savedData.category && (
                      <span className="px-4 py-1.5 bg-accent-200 text-primary-400 rounded-full text-sm font-semibold shadow-sm">
                        {savedData.category}
                      </span>
                    )}
                    {/* แท็ก Style */}
                    {savedData.style && savedData.style.split(',').map(s => s.trim()).filter(Boolean).map(style => (
                      <span key={style} className="px-4 py-1.5 bg-secondary-600 text-gray-900 rounded-full text-sm font-semibold shadow-sm">
                        {style}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {!isCustomerMode && (
              isEditing ? (
                <div className="flex gap-3">
                  <Button 
                    variant="light" 
                    className="text-cursor" 
                    onClick={handleCancel}
                  >
                    Cancel
                  </Button>
                  <Button variant="dark" className="text-cursor" onClick={handleSave}>Save</Button>
                </div>
              ) : (
                <Button 
                  onClick={handleEditClick} 
                  variant="transparent" 
                  icon={
                    <img 
                    src="/icons/edit.svg" 
                    alt="Edit" 
                    className="w-4 h-4 object-contain" 
                    />
                  }
                  className="text-button"
                >
                  Edit
                </Button>
              )
            )}
          </div>

          {/* 📍 ครอบ Grid ทั้งหมดด้วย isEditing เพื่อให้ซ่อนตอนโหมด View */}
          {isEditing && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
               {/* --- Category Section --- */}
               <div className="relative">
                  <label className="text-body font-bold text-primary-500 block mb-3">Category:</label>
                  <input 
                    type="text" 
                    value={catSearch}
                    onChange={(e) => { setCatSearch(e.target.value); setShowCatDropdown(true); }}
                    onFocus={() => setShowCatDropdown(true)}
                    onBlur={() => setTimeout(() => setShowCatDropdown(false), 200)}
                    placeholder="Search category..." 
                    className="border border-primary-500 p-2.5 w-full max-w-xs rounded-lg bg-white outline-none focus:border-black" 
                  />
                  
                  {/* Category Dropdown */}
                  {showCatDropdown && (
                    <div className="absolute z-10 w-full max-w-xs mt-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                      {filteredCategories.length > 0 ? filteredCategories.map(cat => (
                        <div 
                          key={cat} 
                          className="px-4 py-2 hover:bg-gray-100 cursor-pointer text-sm text-gray-800"
                          onClick={() => selectCategory(cat)}
                        >
                          {cat}
                        </div>
                      )) : (
                        <div className="px-4 py-2 text-sm text-gray-400">No results found</div>
                      )}
                    </div>
                  )}

                  {/* Selected Category Tag */}
                  {formData.category && (
                    <div className="flex mt-3">
                      <span className="px-4 py-1.5 bg-accent-200 text-primary-500 rounded-full text-sm font-semibold flex items-center gap-2 shadow-sm">
                        {formData.category}
                        <button onClick={() => setFormData(prev => ({...prev, category: ""}))} className="text-gray-600 hover:text-black cursor-pointer leading-none">✕</button>
                      </span>
                    </div>
                  )}
               </div>

               {/* --- Style Section --- */}
               <div className="relative">
                  <label className="text-body font-bold text-gray-900 block mb-3">Style:</label>
                  <input 
                    type="text" 
                    value={styleSearch}
                    onChange={(e) => { setStyleSearch(e.target.value); setShowStyleDropdown(true); }}
                    onFocus={() => setShowStyleDropdown(true)}
                    onBlur={() => setTimeout(() => setShowStyleDropdown(false), 200)}
                    placeholder="Search style..." 
                    className="border border-primary-500 p-2.5 w-full max-w-xs rounded-lg bg-white outline-none focus:border-black" 
                  />

                  {/* Style Dropdown */}
                  {showStyleDropdown && (
                    <div className="absolute z-10 w-full max-w-xs mt-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                      {filteredStyles.length > 0 ? filteredStyles.map(style => (
                        <div 
                          key={style} 
                          className="px-4 py-2 hover:bg-gray-100 cursor-pointer text-sm text-primary-400"
                          onClick={() => addStyle(style)}
                        >
                          {style}
                        </div>
                      )) : (
                        <div className="px-4 py-2 text-sm text-gray-400">No results found</div>
                      )}
                    </div>
                  )}

                  {/* Selected Style Tags */}
                  {currentStyles.length > 0 && (
                    <div className="flex flex-wrap gap-2 mt-3">
                      {currentStyles.map(style => (
                        <span key={style} className="px-4 py-1.5 bg-secondary-600 text-gray-900 rounded-full text-sm font-semibold flex items-center gap-2 shadow-sm">
                          {style}
                          <button onClick={() => removeStyle(style)} className="text-gray-700 hover:text-black cursor-pointer leading-none">✕</button>
                        </span>
                      ))}
                    </div>
                  )}
               </div>
            </div>
          )}

          <div className="mb-10">
            <label className="text-body font-bold text-primary-500 block mb-2">Description</label>
            {isEditing ? (
              <textarea 
                name="description"
                value={formData.description}
                onChange={handleInputChange}
                rows={4} 
                className="border border-primary-500 p-3 w-full rounded-lg bg-white resize-none" 
                placeholder="Describe your artwork..."
              />
            ) : (
              // ลบ bg-white, border, padding, และ rounded ออก
              <div className="text-gray-700 text-caption leading-relaxed">
                {savedData.description}
              </div>
            )}
          </div>

          <div className="mb-10">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-body font-bold text-primary-500">Artwork Samples</h3>
              {isEditing && (
                <>
                  <Button 
                     variant="transparent"
                    onClick={() => fileInputRef.current?.click()}
                    className="text-button"
                  >

                    + Add Sample
                  </Button>
                  <input 
                    type="file" 
                    accept="image/*" 
                    ref={fileInputRef} 
                    onChange={handleImageUpload} 
                    className="hidden" 
                  />
                </>
              )}
            </div>
            
            <div className="flex flex-wrap gap-4">
              {displayImages.map((imgUrl, index) => (
                <div key={index} className="w-48 h-48 bg-white border border-gray-200 rounded-xl relative p-2 flex items-center justify-center">
                  {isEditing && (
                    <button 
                      onClick={() => handleRemoveImage(index)}
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

          <div className="flex flex-wrap items-end gap-6 mb-4">
            <div>
              <label className="text-body font-bold text-primary-500 block mb-2">Minimum deadline:</label>
              {isEditing ? (
                <div className="flex items-center gap-2">
                  <input 
                    type="number" 
                    name="deadline"
                    value={formData.deadline} 
                    onChange={handleInputChange}
                    className="border border-primary-500 p-2.5 w-24 rounded-lg bg-white" 
                  />
                  <span className="text-sm text-gray-600">days</span>
                </div>
              ) : (
                <div className="bg-primary-400 text-secondary-200 px-6 py-2.5 rounded-lg font-semibold flex items-center gap-2">
                  ⏳ {savedData.deadline} days
                </div>
              )}
            </div>
            
            <div>
              <label className="text-body font-bold text-primary-500 block mb-2">Price:</label>
              {isEditing ? (
                <div className="flex items-center gap-2">
                  <input 
                    type="number" 
                    name="price"
                    value={formData.price} 
                    onChange={handleInputChange}
                    className="border border-primary-500 p-2.5 w-32 rounded-lg bg-white" 
                  />
                  <span className="text-sm text-gray-600">THB</span>
                </div>
              ) : (
                <div className="bg-secondary-600 text-primary-400 px-6 py-2.5 rounded-lg font-semibold flex items-center gap-2">
                  💵 {savedData.price} THB
                </div>
              )}
            </div>
            
            {/* ปุ่ม Delete Artwork */}
            {isEditing && artwork && (
              <div className="flex justify-end mt-12 w-full">
                <Button 
                  variant="error"
                  onClick={() => setShowDeleteConfirm(true)}
                  icon={
                    
                    <img 
                        src="/icons/delete.svg" 
                        alt="Delete" 
                        className="w-4 h-4 object-contain " 
                    />
                    } 
                >
                  Delete Artwork
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 📍 ปรับแก้ Modal ให้ตรงกับรูปภาพ image_60da03.png */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm">
          <div className="bg-white rounded-[24px] p-8 max-w-sm w-full mx-4 shadow-xl">
            <h3 className="text-[22px] font-bold text-gray-900 mb-3 text-center leading-tight">
              Are you sure want to<br />delete artwork sample?
            </h3>
            <p className="text-gray-600 text-sm text-center mb-8">
              This artwork will be permanently deleted.<br />This action cannot be undone.
            </p>
            <div className="flex gap-4 justify-center">
              <button 
                className="rounded-[14px] px-6 py-3 border border-gray-300 font-semibold cursor-pointer w-full text-gray-800 hover:bg-gray-50 transition-colors"
                onClick={() => setShowDeleteConfirm(false)}
              >
                No, Keep it
              </button>
              <button 
                onClick={confirmDelete}
                className="bg-[#FF3333] hover:bg-red-600 text-white rounded-[14px] px-6 py-3 font-semibold cursor-pointer transition-colors w-full"
              >
                Yes, Delete !
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}