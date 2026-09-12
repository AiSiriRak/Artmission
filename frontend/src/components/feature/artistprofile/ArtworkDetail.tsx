import { useState, useEffect } from "react";
import type { Artwork, CreateArtworkInput, } from "@/lib/api/types";
import { Button } from "@/components/ui/Button";
import ReviewList from "./ReviewList";
import DeleteArtworkModal from "./DeleteArtworkModal"; 
import ArtworkSampleGallery from "./ArtworkSampleGallery"; 

interface ReviewData {
  id: number;
  reviewerName: string;
  timeAgo: string;
  orderName: string;
  rating: number;
  comment: string;
}

interface ArtworkDetailProps {
  // เชื่อมกับ Artwork ตรงๆ และเผื่อฟิลด์ id ไว้กรณี Backend ตกหล่นจาก Schema
  artwork?: (Artwork & { id?: string | number }) | null;
  onBack: () => void;
  isCustomerMode: boolean;
  onSave?: (payload: CreateArtworkInput, artworkId?: string) => void | Promise<void>;
  onDelete?: (artworkId: string) => void | Promise<void>;
}

const AVAILABLE_CATEGORIES = ["Book", "Comic", "Game", "Animation", "Portrait", "Illustration", "Other"];
const AVAILABLE_STYLES = ["Pixel Art", "Pixel", "Pixel 8 bit", "Pixel 16 bit", "Pixel 32 bit", "Water Color", "Cartoon", "Graphic", "Anime", "Realism"];

const mockArtworkReviews: ReviewData[] = [
  { id: 1, reviewerName: "Name", timeAgo: "2 hrs ago", orderName: "Pixel Art", rating: 3.5, comment: "งานน่ารักมากๆๆๆ ❤️❤️❤️" },
  { id: 2, reviewerName: "Name", timeAgo: "2 hrs ago", orderName: "Pixel Art", rating: 3.5, comment: "งานน่ารักมากๆๆๆ ❤️❤️❤️" },
];

export default function ArtworkDetail({ artwork, onBack, isCustomerMode, onSave, onDelete }: ArtworkDetailProps) {
  const [isEditing, setIsEditing] = useState(!artwork);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  
  // 1. Map ค่าตั้งต้นให้ตรงกับ Schema
  const initialPriceTHB = artwork?.price_satang ? artwork.price_satang / 100 : 0;
  
  // แปลง artwork_samples จาก Object Array -> String Array ไว้ใช้ใน UI
  const initialImages = artwork?.artwork_samples && artwork.artwork_samples.length > 0 
    ? artwork.artwork_samples.map(sample => sample.image_url) 
    : [];

  const [savedData, setSavedData] = useState({
    name: artwork?.name || "",
    category: artwork?.category || "",
    styles: artwork?.styles || [], // ใช้ Array โดยตรงตาม Schema
    description: artwork?.description || "", 
    minimum_deadline_days: artwork?.minimum_deadline_days || 1,
    price: initialPriceTHB,
  });

  const [savedImages, setSavedImages] = useState<string[]>(initialImages);
  const [formData, setFormData] = useState({ ...savedData });
  const [images, setImages] = useState<string[]>([...savedImages]);

  const [catSearch, setCatSearch] = useState("");
  const [styleSearch, setStyleSearch] = useState("");
  const [showCatDropdown, setShowCatDropdown] = useState(false);
  const [showStyleDropdown, setShowStyleDropdown] = useState(false);

  // เมื่อ Props เปลี่ยน ให้เซ็ต State ใหม่
  useEffect(() => {
    if (artwork) {
      const priceTHB = artwork.price_satang ? artwork.price_satang / 100 : 0;
      const imgArray = artwork.artwork_samples && artwork.artwork_samples.length > 0 
        ? artwork.artwork_samples.map(sample => sample.image_url) 
        : ["/placeholder.jpg"];
      
      const newData = {
        name: artwork.name || "",
        category: artwork.category || "",
        styles: artwork.styles || [],
        description: artwork.description || "",
        minimum_deadline_days: artwork.minimum_deadline_days || 1,
        price: priceTHB,
      };

      setSavedData(newData);
      setFormData(newData);
      setSavedImages(imgArray);
      setImages(imgArray);
    }
  }, [artwork]);

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
      // 2. จัด Payload ส่งกลับให้ตรงกับ Schema CreateArtworkInput
      const payload: CreateArtworkInput = {
        name: formData.name,
        description: formData.description,
        price_satang: Math.round(Number(formData.price || 0) * 100),
        minimum_deadline_days: Number(formData.minimum_deadline_days),
        category: formData.category,
        styles: formData.styles.length > 0 ? formData.styles : null,
        artwork_samples: images
          .filter(url => url !== "/placeholder.jpg")
          .map(url => ({ image_url: url })),
      };

      const artworkId = artwork?.id ? String(artwork.id) : undefined;
      
      // ✅ ส่ง Payload ที่ Type ตรงเป๊ะกลับไปให้ Parent จัดการ
      onSave(payload, artworkId);
    }
  };

  const confirmDelete = () => {
    if (onDelete && artwork?.id) {
      onDelete(String(artwork.id)); 
    }
  };

  const displayImages = isEditing ? images : savedImages;

  // กรอง Style / Category
  const filteredCategories = AVAILABLE_CATEGORIES.filter(c => c.toLowerCase().includes(catSearch.toLowerCase()));
  const filteredStyles = AVAILABLE_STYLES.filter(s => s.toLowerCase().includes(styleSearch.toLowerCase()));

  const addStyle = (style: string) => {
    if (!formData.styles.includes(style)) {
      setFormData(prev => ({ ...prev, styles: [...prev.styles, style] }));
    }
    setStyleSearch("");
    setShowStyleDropdown(false);
  };

  const removeStyle = (styleToRemove: string) => {
    setFormData(prev => ({ 
      ...prev, 
      styles: prev.styles.filter(s => s !== styleToRemove) 
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
        <Button variant="light" onClick={onBack} className="mb-6 text-button flex items-center gap-2 !border">
          <span>←</span> Back
        </Button>

        <div className="border border-primary-500 rounded-3xl p-10 bg-secondary-200 shadow-sm">
          
          {/* Header Section */}
          <div className="flex justify-between items-start mb-8">
            <div className="w-full max-w-xl">
              {isEditing ? (
                <div className="space-y-1 mb-6">
                  <label className="text-body font-bold text-primary-500 block">Artwork Name</label>
                  <input type="text" name="name" value={formData.name} onChange={handleInputChange} placeholder="e.g., Pet Portrait" className="border border-primary-500 p-2.5 w-full rounded-lg bg-white" />
                </div>
              ) : (
                <div className="mb-4">
                  <h1 className="text-h1 font-bold text-gray-900 mb-4">{savedData.name || "Untitled"}</h1>
                  <div className="flex flex-wrap gap-2">
                    {savedData.category && <span className="px-4 py-1.5 bg-accent-200 text-primary-400 rounded-full text-sm font-semibold shadow-sm">{savedData.category}</span>}
                    {savedData.styles && savedData.styles.map(style => (
                      <span key={style} className="px-4 py-1.5 bg-secondary-600 text-gray-900 rounded-full text-sm font-semibold shadow-sm">{style}</span>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {!isCustomerMode && (
              isEditing ? (
                <div className="flex gap-3">
                  <Button variant="light" className="text-cursor !border" onClick={() => setIsEditing(false)}>Cancel</Button>
                  <Button variant="dark" className="text-cursor" onClick={handleSave}>Save</Button>
                </div>
              ) : (
                <Button onClick={handleEditClick} variant="transparent" icon={<img src="/icons/edit.svg" alt="Edit" className="w-4 h-4 object-contain" />} className="text-button">Edit</Button>
              )
            )}
          </div>

          {/* Tags Dropdown Section */}
          {isEditing && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
               <div className="relative">
                  <label className="text-body font-bold text-primary-500 block mb-3">Category:</label>
                  <input type="text" value={catSearch} onChange={(e) => { setCatSearch(e.target.value); setShowCatDropdown(true); }} onFocus={() => setShowCatDropdown(true)} onBlur={() => setTimeout(() => setShowCatDropdown(false), 200)} placeholder="Search category..." className="border border-primary-500 p-2.5 w-full max-w-xs rounded-lg bg-white outline-none focus:border-black" />
                  
                  {showCatDropdown && (
                    <div className="absolute z-10 w-full max-w-xs mt-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                      {filteredCategories.length > 0 ? filteredCategories.map(cat => (
                        <div key={cat} onMouseDown={() => selectCategory(cat)} className="px-4 py-2 hover:bg-gray-100 cursor-pointer text-sm text-gray-800">{cat}</div>
                      )) : <div className="px-4 py-2 text-sm text-gray-400">No results found</div>}
                    </div>
                  )}

                  {formData.category && (
                    <div className="flex mt-3">
                      <span className="px-4 py-1.5 bg-accent-200 text-primary-500 rounded-full text-sm font-semibold flex items-center gap-2 shadow-sm">
                        {formData.category}
                        <button onClick={() => setFormData(prev => ({...prev, category: ""}))} className="text-gray-600 hover:text-black cursor-pointer leading-none">✕</button>
                      </span>
                    </div>
                  )}
               </div>

               <div className="relative">
                  <label className="text-body font-bold text-gray-900 block mb-3">Style:</label>
                  <input type="text" value={styleSearch} onChange={(e) => { setStyleSearch(e.target.value); setShowStyleDropdown(true); }} onFocus={() => setShowStyleDropdown(true)} onBlur={() => setTimeout(() => setShowStyleDropdown(false), 200)} placeholder="Search style..." className="border border-primary-500 p-2.5 w-full max-w-xs rounded-lg bg-white outline-none focus:border-black" />

                  {showStyleDropdown && (
                    <div className="absolute z-10 w-full max-w-xs mt-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                      {filteredStyles.length > 0 ? filteredStyles.map(style => (
                        <div key={style} onMouseDown={() => addStyle(style)} className="px-4 py-2 hover:bg-gray-100 cursor-pointer text-sm text-primary-400">{style}</div>
                      )) : <div className="px-4 py-2 text-sm text-gray-400">No results found</div>}
                    </div>
                  )}

                  {formData.styles.length > 0 && (
                    <div className="flex flex-wrap gap-2 mt-3">
                      {formData.styles.map(style => (
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

          {/* Description Section */}
          <div className="mb-10">
            <label className="text-body font-bold text-primary-500 block mb-2">Description</label>
            {isEditing ? (
              <textarea name="description" value={formData.description} onChange={handleInputChange} rows={4} className="border border-primary-500 p-3 w-full rounded-lg bg-white resize-none" placeholder="Describe your artwork..." />
            ) : (
              <div className="text-gray-700 text-caption leading-relaxed">{savedData.description}</div>
            )}
          </div>

          <ArtworkSampleGallery 
            isEditing={isEditing}
            images={displayImages}
            onUpload={handleImageUpload}
            onRemove={handleRemoveImage}
          />

          {/* Pricing & Deadline Section */}
          <div className="flex flex-wrap items-end gap-6 mb-4 mt-6">
            <div>
              <label className="text-body font-bold text-primary-500 block mb-2">Minimum deadline:</label>
              {isEditing ? (
                <div className="flex items-center gap-2">
                  <input type="number" name="minimum_deadline_days" value={formData.minimum_deadline_days} onChange={handleInputChange} className="border border-primary-500 p-2.5 w-24 rounded-lg bg-white" />
                  <span className="text-sm text-gray-600">days</span>
                </div>
              ) : (
                <div className="bg-primary-400 text-secondary-200 px-6 py-2.5 rounded-lg font-semibold flex items-center gap-2">⏳ {savedData.minimum_deadline_days} days</div>
              )}
            </div>
            
            <div>
              <label className="text-body font-bold text-primary-500 block mb-2">Price:</label>
              {isEditing ? (
                <div className="flex items-center gap-2">
                  <input type="number" name="price" value={formData.price} onChange={handleInputChange} className="border border-primary-500 p-2.5 w-32 rounded-lg bg-white" />
                  <span className="text-sm text-gray-600">THB</span>
                </div>
              ) : (
                <div className="bg-secondary-600 text-primary-400 px-6 py-2.5 rounded-lg font-semibold flex items-center gap-2">💵 {savedData.price.toLocaleString()} THB</div>
              )}
            </div>
            
            {isEditing && artwork?.id && (
              <div className="flex justify-end mt-12 w-full">
                <Button variant="error" onClick={() => setShowDeleteConfirm(true)} icon={<img src="/icons/delete.svg" alt="Delete" className="w-4 h-4 object-contain" />}>
                  Delete Artwork
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* Reviews Section */}
        <div className="mt-12">
          <div className="flex items-center gap-3 mb-6">
            <h2 className="text-xl font-bold text-gray-900">Order Reviews</h2>
            <span className="text-sm font-normal text-gray-400">{mockArtworkReviews.length} reviews</span>
          </div>
          <ReviewList reviews={mockArtworkReviews as any} />
        </div>
      </div>

      <DeleteArtworkModal 
        isOpen={showDeleteConfirm} 
        onClose={() => setShowDeleteConfirm(false)} 
        onConfirm={confirmDelete} 
      />
    </>
  );
}