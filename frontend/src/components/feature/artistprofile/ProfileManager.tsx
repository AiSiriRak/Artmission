"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { updateArtistProfile, getArtistArtworks } from "@/lib/api/artists";
import {
  createArtwork,
  updateArtwork,
  deleteArtwork,
} from "@/lib/api/artworks";
import type {
  UpdateArtistInput,
  ArtistProfile,
  Artwork,
  CreateArtworkInput,
  UpdateArtworkInput,
} from "@/lib/api/types";

// Temporary Type of Review
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
import ArtworkDetail from "./ArtworkDetail";
import ReviewList from "./ReviewList";
import ArtistProfileView from "./ArtistProfileView";

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
  initialReviews = [],
}: ProfileManagerProps) {
  const router = useRouter();

  const [isCustomerMode, setIsCustomerMode] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const [selectedArtwork, setSelectedArtwork] = useState<
    Artwork | "new" | null
  >(null);

  const [savedData, setSavedData] = useState<ArtistProfile>(initialProfile);
  const [artistData, setArtistData] = useState<ArtistProfile>(initialProfile);
  const [artworks, setArtworks] = useState<Artwork[]>(initialArtworks);

  const [newProfileImageFile, setNewProfileImageFile] = useState<File | null>(
    null,
  );

  // Mock Review - Removed setReviews as it was never used
  const [reviews] = useState<ReviewData[]>(
    initialReviews.length > 0
      ? initialReviews
      : [
          {
            id: 1,
            reviewerName: "Name",
            timeAgo: "2 hrs ago",
            orderName: "Pixel Art",
            rating: 3.5,
            comment: "งานน่ารักมากๆๆๆ ❤️❤️❤️",
          },
          {
            id: 2,
            reviewerName: "Name",
            timeAgo: "2 hrs ago",
            orderName: "Pixel Art",
            rating: 3.5,
            comment: "งานน่ารักมากๆๆๆ ❤️❤️❤️",
          },
          {
            id: 3,
            reviewerName: "Name",
            timeAgo: "2 hrs ago",
            orderName: "Pixel Art",
            rating: 3.5,
            comment: "งานน่ารักมากๆๆๆ ❤️❤️❤️",
          },
        ],
  );

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    setArtistData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const payload: Record<string, unknown> = {
        description: artistData.description || "",
      };

      if (newProfileImageFile) {
        payload.profile_image = newProfileImageFile;
      }

      const updatedProfile = await updateArtistProfile(
        payload as UpdateArtistInput,
      );

      const updatedData: ArtistProfile = {
        ...artistData,
        ...updatedProfile,
        description: updatedProfile.description || artistData.description,
        min_price_satang: updatedProfile.min_price_satang,
        max_price_satang: updatedProfile.max_price_satang,
      };

      setSavedData(updatedData);
      setArtistData(updatedData);
      setIsEditing(false);
      setNewProfileImageFile(null);

      router.refresh();
    } catch (error: unknown) {
      console.error("Save failed:", error);

      if (typeof error === "object" && error !== null && "body" in error) {
        console.log(
          "Detailed Validation Error:",
          JSON.stringify(error.body, null, 2),
        );
      }

      alert("เกิดข้อผิดพลาดในการบันทึกข้อมูล");
    } finally {
      setIsSaving(false);
    }
  };

  const refreshArtworks = async (artistId: string) => {
    const res = await getArtistArtworks(artistId);
    const artworksList = Array.isArray(res)
      ? res
      : (res as Record<string, unknown>).artworks ||
        (res as Record<string, unknown>).items ||
        (res as Record<string, unknown>).data ||
        [];

    setArtworks(artworksList as Artwork[]);
  };

  const handleSaveArtwork = async (
    payload: CreateArtworkInput | UpdateArtworkInput | Record<string, unknown>,
    artworkId?: string,
  ) => {
    setIsSaving(true);
    try {
      if (artworkId) {
        // ถ้าเป็นการแก้ไข (มี ID ส่งมา) -> ใช้ Update API แทนการลบแล้วสร้างใหม่
        await updateArtwork(artworkId, payload as UpdateArtworkInput);
      } else {
        await createArtwork(payload as CreateArtworkInput);
      }

      const artistId =
        artistData.artist_id || (artistData as Record<string, unknown>).id;
      if (artistId) {
        await refreshArtworks(String(artistId));
      }

      setSelectedArtwork(null);
    } catch (error: unknown) {
      console.error("Failed to save artwork:", error);

      if (typeof error === "object" && error !== null && "body" in error) {
        console.log(
          "Detailed Artwork Error:",
          JSON.stringify(error.body, null, 2),
        );
      }

      alert("เกิดข้อผิดพลาดในการบันทึกผลงาน");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteArtwork = async (artworkId: string) => {
    const isConfirmed = window.confirm("คุณต้องการลบผลงานนี้ใช่หรือไม่?");
    if (!isConfirmed) return;

    setIsSaving(true);
    try {
      await deleteArtwork(artworkId);

      const artistId =
        artistData.artist_id || (artistData as Record<string, unknown>).id;
      if (artistId) {
        await refreshArtworks(String(artistId));
      } else {
        setArtworks((prev) =>
          prev.filter((art) => String(art.id) !== artworkId),
        );
      }

      setSelectedArtwork(null);
    } catch (error: unknown) {
      console.error("Failed to delete artwork:", error);
      alert("เกิดข้อผิดพลาดในการลบผลงาน");
    } finally {
      setIsSaving(false);
    }
  };

  const handleToggleMode = () => {
    setIsCustomerMode(!isCustomerMode);
    if (!isCustomerMode) setIsEditing(false);
    setSelectedArtwork(null);
  };

  const handleImageChange = (newImageUrl: string, file?: File) => {
    setArtistData((prev) => ({ ...prev, profile_url: newImageUrl }));
    if (file) {
      setNewProfileImageFile(file);
    }
  };

  // --- ลอจิกดึงข้อมูลอัตโนมัติ (อิงตาม ArtworkView schema) ---
  const derivedCategories = [
    ...new Set(artworks.map((art) => art.category).filter(Boolean)),
  ];

  const derivedStyles = [
    ...new Set(artworks.flatMap((art) => art.styles || [])),
  ].filter(Boolean);

  // Price from Satang to Bath
  const prices = artworks
    .map((art) => Number(art.price_satang ? art.price_satang / 100 : 0))
    .filter((p) => !isNaN(p) && p > 0);

  const minPrice = prices.length > 0 ? Math.min(...prices) : 0;
  const maxPrice = prices.length > 0 ? Math.max(...prices) : 0;

  const priceRangeText =
    prices.length === 0
      ? "N/A"
      : minPrice === maxPrice
        ? `${minPrice.toLocaleString()} THB`
        : `${minPrice.toLocaleString()} - ${maxPrice.toLocaleString()} THB`;

  const showEditControls = !isCustomerMode && !isEditing;
  const avgRating =
    reviews.length > 0
      ? (
          reviews.reduce((sum, item) => sum + item.rating, 0) / reviews.length
        ).toFixed(1)
      : "0.0";

  if (isSaving) {
    return <Loading />;
  }

  // Show artwork details
  if (selectedArtwork) {
    return (
      <ArtworkDetail
        artwork={selectedArtwork === "new" ? null : selectedArtwork}
        onBack={() => setSelectedArtwork(null)}
        isCustomerMode={isCustomerMode}
        onSave={handleSaveArtwork}
        onDelete={handleDeleteArtwork}
      />
    );
  }

  if (isCustomerMode) {
    return (
      <ArtistProfileView
        profile={artistData}
        artworks={artworks}
        reviews={reviews}
        reviewScore={Number(avgRating)}
        onArtworkClick={(artwork) => {
          setSelectedArtwork(artwork);
        }}
        actions={
          <Button
            onClick={handleToggleMode}
            variant="dark"
            icon={
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src="/icons/exit.svg"
                alt="Exit"
                className="w-4 h-4 object-contain"
              />
            }
          >
            Exit Preview
          </Button>
        }
      />
    );
  }

  return (
    <div className="w-full">
      {/* -- Profile -- */}
      <div className="max-w-5xl mx-auto px-8 pt-10">
        <h1 className="text-h3 font-bold mb-10">Your Information</h1>

        <div className="mb-12">
          <div className="flex flex-col md:flex-row gap-10">
            <ProfileSidebar
              imageUrl={artistData.profile_url || ""}
              isEditing={isEditing}
              onEdit={() => setIsEditing(true)}
              onSave={handleSave}
              onCancel={() => {
                setIsEditing(false);
                setArtistData(savedData);
              }}
              onToggleView={handleToggleMode}
              onImageChange={handleImageChange}
            />

            <div className="flex-1">
              <EditorField
                label="Profile Name"
                name="artist_name"
                value={artistData.artist_name || ""}
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
                  <span className="block text-body font-bold text-gray-900 mb-3">
                    Category:
                  </span>
                  {derivedCategories.length > 0 ? (
                    <TagList items={derivedCategories} variant="category" />
                  ) : (
                    <span className="text-gray-400 text-sm">
                      No categories
                    </span>
                  )}
                </div>
                <div>
                  <span className="block text-body font-bold text-gray-900 mb-3">
                    Style:
                  </span>
                  {derivedStyles.length > 0 ? (
                    <TagList items={derivedStyles} variant="style" />
                  ) : (
                    <span className="text-gray-400 text-sm">No styles</span>
                  )}
                </div>
              </div>

              <div className="mt-8">
                <span className="block text-body font-bold text-gray-900 mb-3">
                  Price Range:
                </span>
                <span className="text-gray-700 text-base">
                  {priceRangeText}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* -- Artwork -- */}
      <div className="max-w-5xl mx-auto px-8 pb-10">
        <hr className="border-gray-200 mb-10" />
        <div className="flex justify-between items-center mb-6">
          <div className="flex items-center gap-3">
            <h2 className="text-h3 font-bold text-gray-900">
              Your Artwork
            </h2>
            <span className="text-sm font-normal text-gray-400">
              {artworks.length} samples
            </span>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {showEditControls && (
            <div
              onClick={() => setSelectedArtwork("new")}
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
              key={art.id || art.name || idx}
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

      {/* -- Review -- */}
      <div className="max-w-5xl mx-auto px-8 pb-20">
        <hr className="border-gray-200 mb-6" />
        <div className="flex justify-between items-center mb-6">
          <div className="flex items-center gap-3">
            <h2 className="text-h3 font-bold text-gray-900">
              Your Reviews
            </h2>
            <span className="text-sm font-normal text-gray-400">
              {reviews.length} reviews
            </span>
          </div>
          <div className="flex items-center gap-1 font-bold text-gray-900 text-base">
            <span className="text-red-400 text-lg">☆</span>
            <span>{avgRating}/5</span>
          </div>
        </div>

        <ReviewList reviews={reviews} />
      </div>
    </div>
  );
}
