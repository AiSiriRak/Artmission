import { useState, useEffect, useRef } from "react";
import type {
  Artwork,
  CreateArtworkInput,
  UpdateArtworkInput,
} from "@/lib/api/types";
import { Button } from "@/components/ui/Button";
import ReviewList from "./ReviewList";
import DeleteArtworkModal from "./DeleteArtworkModal";
import ArtworkSampleGallery from "./ArtworkSampleGallery";
import CatalogPicker from "./CatalogPicker";
import { ValidateArtworkForm } from "./ArtworkValidation";
import {
  ALLOWED_CATEGORIES,
  ALLOWED_STYLES,
  displayLabel,
  isUuid,
  loadArtworkCatalog,
  placeholderCatalog,
  resolveCatalogId,
  resolveCatalogIds,
  type CatalogOption,
} from "./artworkCatalog";

interface ReviewData {
  id: number;
  reviewerName: string;
  timeAgo: string;
  orderName: string;
  rating: number;
  comment: string;
}

interface ArtworkDetailProps {
  artwork?: (Artwork & { id?: string | number }) | null;
  onBack: () => void;
  isCustomerMode: boolean;
  onSave?: (
    payload: CreateArtworkInput | UpdateArtworkInput | Record<string, unknown>,
    artworkId?: string,
  ) => void | Promise<void>;
  onDelete?: (artworkId: string) => void | Promise<void>;
}

const mockArtworkReviews: ReviewData[] = [
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
];

interface ImageItem {
  previewUrl: string;
  file?: File;
}

interface ArtworkDraft {
  name: string;
  category: string;
  styles: string[];
  description: string;
  minimum_deadline_days: number;
  price: number;
}

export default function ArtworkDetail({
  artwork,
  onBack,
  isCustomerMode,
  onSave,
  onDelete,
}: ArtworkDetailProps) {
  const [isEditing, setIsEditing] = useState(!artwork);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const initialPriceTHB = artwork?.price_satang
    ? artwork.price_satang / 100
    : 0;

  const initialImages: ImageItem[] =
    artwork?.artwork_samples && artwork.artwork_samples.length > 0
      ? artwork.artwork_samples.map((sample) => ({
          previewUrl: sample.image_url,
        }))
      : [];

  const [savedData, setSavedData] = useState<ArtworkDraft>({
    name: artwork?.name || "",
    category: artwork?.category || "",
    styles: artwork?.styles || [],
    description: artwork?.description || "",
    minimum_deadline_days: artwork?.minimum_deadline_days || 1,
    price: initialPriceTHB,
  });

  const [savedImages, setSavedImages] = useState<ImageItem[]>(initialImages);
  const [formData, setFormData] = useState({ ...savedData });
  const [images, setImages] = useState<ImageItem[]>([...savedImages]);

  const [availableCategories, setAvailableCategories] = useState<
    CatalogOption[]
  >(() => placeholderCatalog(ALLOWED_CATEGORIES));
  const [availableStyles, setAvailableStyles] = useState<CatalogOption[]>(() =>
    placeholderCatalog(ALLOWED_STYLES),
  );
  const catalogRef = useRef({
    categories: placeholderCatalog(ALLOWED_CATEGORIES),
    styles: placeholderCatalog(ALLOWED_STYLES),
  });
  const [errors, setErrors] = useState({
    name: "",
    description: "",
    category: "",
    styles: "",
    minimum_deadline_days: "",
    price: "",
  });
  const [hasSubmitted, setHasSubmitted] = useState(false);

  useEffect(() => {
    catalogRef.current = {
      categories: availableCategories,
      styles: availableStyles,
    };

    // eslint-disable-next-line react-hooks/set-state-in-effect
    setFormData((prev) => {
      const category = resolveCatalogId(prev.category, availableCategories);
      const styles = resolveCatalogIds(prev.styles, availableStyles);
      if (
        category === prev.category &&
        styles.join("\0") === prev.styles.join("\0")
      ) {
        return prev;
      }
      return { ...prev, category, styles };
    });
    setSavedData((prev) => {
      const category = resolveCatalogId(prev.category, availableCategories);
      const styles = resolveCatalogIds(prev.styles, availableStyles);
      if (
        category === prev.category &&
        styles.join("\0") === prev.styles.join("\0")
      ) {
        return prev;
      }
      return { ...prev, category, styles };
    });
  }, [availableCategories, availableStyles]);

  useEffect(() => {
    if (artwork) {
      const priceTHB = artwork.price_satang ? artwork.price_satang / 100 : 0;
      const imgArray: ImageItem[] =
        artwork.artwork_samples && artwork.artwork_samples.length > 0
          ? artwork.artwork_samples.map((sample) => ({
              previewUrl: sample.image_url,
            }))
          : [{ previewUrl: "/placeholder.jpg" }];

      const newData = {
        name: artwork.name || "",
        category: resolveCatalogId(
          artwork.category || "",
          catalogRef.current.categories,
        ),
        styles: resolveCatalogIds(
          artwork.styles || [],
          catalogRef.current.styles,
        ),
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

  useEffect(() => {
    let cancelled = false;

    void loadArtworkCatalog().then((catalog) => {
      if (cancelled) return;
      setAvailableCategories(catalog.categories);
      setAvailableStyles(catalog.styles);
    });

    return () => {
      cancelled = true;
    };
  }, []);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const { name, value } = e.target;
    const newFormData = { ...formData, [name]: value };
    setFormData(newFormData);
    handleErrorChange(newFormData);
  };

  const handleErrorChange = (newFormData: typeof formData) => {
    if (hasSubmitted) {
      const newErrors = ValidateArtworkForm(newFormData);
      setErrors(newErrors);
    }
  };

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      const previewUrl = URL.createObjectURL(file);
      setImages((prev) => [...prev, { previewUrl, file }]);
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

  const handleNumberKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (["e", "E", "+", "-"].includes(e.key)) {
      e.preventDefault();
    }
  };

  const handleSave = () => {
    const categoryId = resolveCatalogId(formData.category, availableCategories);
    const styleIds = resolveCatalogIds(formData.styles, availableStyles).filter(
      isUuid,
    );
    const categoryToSend = isUuid(categoryId) ? categoryId : "";
    const resolvedForm = {
      ...formData,
      category: categoryToSend,
      styles: styleIds,
    };

    setHasSubmitted(true);

    const newErrors = ValidateArtworkForm(resolvedForm);
    setErrors(newErrors);

    const hasErrors = Object.values(newErrors).some(Boolean);
    if (hasErrors) return;

    setFormData(resolvedForm);
    setSavedData({ ...resolvedForm });
    setSavedImages([...images]);
    setIsEditing(false);

    if (onSave) {
      const actualId =
        artwork?.id || (artwork as Record<string, unknown>)?.artwork_id;
      const updating = !!actualId;
      const artworkId = actualId ? String(actualId) : undefined;

      const newImages = images
        .filter((img) => img.previewUrl !== "/placeholder.jpg" && img.file)
        .map((img) => img.file);

      const payload: Record<string, unknown> = {
        name: resolvedForm.name,
        description: resolvedForm.description,
        price_satang: Math.round(Number(resolvedForm.price || 0) * 100),
        minimum_deadline_days: Number(resolvedForm.minimum_deadline_days),
        category_id: categoryToSend,
        style_ids: styleIds,
      };

      if (updating) {
        const originalUrls = (artwork?.artwork_samples || [])
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          .map((s: any) => s.image_url || (typeof s === "string" ? s : ""))
          .filter(Boolean);
        const currentUrls = images
          .filter((img) => !img.file)
          .map((img) => img.previewUrl);

        const deletedUrls = originalUrls.filter(
          (url: string) => !currentUrls.includes(url),
        );

        payload.uploaded_samples = newImages.length > 0 ? newImages : undefined;
        payload.deleted_sample_urls =
          deletedUrls.length > 0 ? deletedUrls : null;
      } else {
        payload.artwork_samples = newImages.length > 0 ? newImages : undefined;
      }

      onSave(payload, artworkId);
    }
  };

  const confirmDelete = () => {
    if (onDelete && artwork?.id) {
      onDelete(String(artwork.id));
    }
  };

  const selectCategory = (categoryId: string) => {
    if (!categoryId) return;
    setFormData((prev) => ({ ...prev, category: categoryId }));
    setErrors((prev) => ({ ...prev, category: "" }));
    handleErrorChange({ ...formData, category: categoryId });
  };

  const addStyle = (styleId: string) => {
    if (!styleId || formData.styles.includes(styleId)) return;
    const nextStyles = [...formData.styles, styleId];
    setFormData((prev) => ({ ...prev, styles: nextStyles }));
    setErrors((prev) => ({ ...prev, styles: "" }));
    handleErrorChange({ ...formData, styles: nextStyles });
  };

  const removeStyle = (styleToRemove: string) => {
    setFormData((prev) => ({
      ...prev,
      styles: prev.styles.filter((style) => style !== styleToRemove),
    }));
    handleErrorChange({
      ...formData,
      styles: formData.styles.filter((style) => style !== styleToRemove),
    });
  };

  const displayImages = (isEditing ? images : savedImages).map(
    (img) => img.previewUrl,
  );

  return (
    <>
      <div className="max-w-5xl mx-auto px-8 pt-10 pb-20">
        <Button
          variant="light"
          onClick={onBack}
          className="mb-6 text-button flex items-center gap-2 !border"
        >
          <span>←</span> Back
        </Button>

        <div className="border border-primary-500 rounded-3xl p-10 bg-secondary-200 shadow-sm">
          <div className="flex justify-between items-start mb-8">
            <div className="w-full max-w-xl">
              {isEditing ? (
                <div className="space-y-1 mb-6">
                  <label className="text-body font-bold text-primary-500 block">
                    Artwork Name
                  </label>
                  <input
                    type="text"
                    name="name"
                    value={formData.name}
                    onChange={handleInputChange}
                    placeholder="e.g., Pet Portrait"
                    className="border border-primary-500 p-2.5 w-full rounded-lg bg-white"
                  />
                  {errors.name && (
                    <p className="text-red-500 text-sm mt-1">{errors.name}</p>
                  )}
                </div>
              ) : (
                <div className="mb-4">
                  <h1 className="text-h1 font-bold text-gray-900 mb-4">
                    {savedData.name || "Untitled"}
                  </h1>
                  <div className="flex flex-wrap gap-2">
                    {savedData.category && (
                      <span className="px-4 py-1.5 bg-accent-200 text-primary-400 rounded-full text-sm font-semibold shadow-sm">
                        {displayLabel(availableCategories, savedData.category)}
                      </span>
                    )}
                    {savedData.styles &&
                      savedData.styles.map((style) => (
                        <span
                          key={style}
                          className="px-4 py-1.5 bg-secondary-600 text-gray-900 rounded-full text-sm font-semibold shadow-sm"
                        >
                          {displayLabel(availableStyles, style)}
                        </span>
                      ))}
                  </div>
                </div>
              )}
            </div>

            {!isCustomerMode &&
              (isEditing ? (
                <div className="flex gap-3">
                  <Button
                    variant="light"
                    className="text-cursor !border"
                    onClick={() => setIsEditing(false)}
                  >
                    Cancel
                  </Button>
                  <Button
                    variant="dark"
                    className="text-cursor"
                    onClick={handleSave}
                  >
                    Save
                  </Button>
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
              ))}
          </div>

          {isEditing && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
              <CatalogPicker
                variant="category"
                options={availableCategories}
                selectedIds={formData.category ? [formData.category] : []}
                error={errors.category}
                onSelect={selectCategory}
                onRemove={() => {
                  setFormData((prev) => ({ ...prev, category: "" }));
                  handleErrorChange({ ...formData, category: "" });
                }}
              />
              <CatalogPicker
                variant="style"
                options={availableStyles}
                selectedIds={formData.styles}
                error={errors.styles}
                onSelect={addStyle}
                onRemove={removeStyle}
              />
            </div>
          )}

          <div className="mb-10">
            <label className="text-body font-bold text-primary-500 block mb-2">
              Description
            </label>
            {isEditing ? (
              <div className="flex flex-col gap-1">
                <textarea
                  name="description"
                  value={formData.description}
                  onChange={handleInputChange}
                  rows={4}
                  className="border border-primary-500 p-3 w-full rounded-lg bg-white resize-none"
                  placeholder="Describe your artwork..."
                />
                {errors.description && (
                  <p className="text-red-500 text-sm mt-1">
                    {errors.description}
                  </p>
                )}
              </div>
            ) : (
              <div className="text-gray-700 text-caption leading-relaxed">
                {savedData.description}
              </div>
            )}
          </div>

          <ArtworkSampleGallery
            isEditing={isEditing}
            images={displayImages}
            hasSubmitted={hasSubmitted}
            onUpload={handleImageUpload}
            onRemove={handleRemoveImage}
          />

          <div className="flex flex-wrap items-start gap-6 mb-4 mt-6">
            <div>
              <label className="text-body font-bold text-primary-500 block mb-2">
                Minimum deadline:
              </label>
              {isEditing ? (
                <div className="flex flex-col gap-1">
                  <div className="flex items-center gap-2">
                    <input
                      type="number"
                      name="minimum_deadline_days"
                      value={formData.minimum_deadline_days}
                      onChange={handleInputChange}
                      onKeyDown={handleNumberKeyDown}
                      min={1}
                      className="border border-primary-500 p-2.5 w-24 rounded-lg bg-white"
                    />
                    <span className="text-sm text-gray-600">days</span>
                  </div>
                  {errors.minimum_deadline_days && (
                    <p className="text-red-500 text-sm mt-1">
                      {errors.minimum_deadline_days}
                    </p>
                  )}
                </div>
              ) : (
                <div className="bg-primary-400 text-secondary-200 px-6 py-2.5 rounded-lg font-semibold flex items-center gap-2">
                  <img
                    src="/icons/alarm-clock.svg"
                    alt="Deadline"
                    className="w-4 h-4 object-contain brightness-0 invert"
                  />
                  {savedData.minimum_deadline_days} days
                </div>
              )}
            </div>

            <div>
              <label className="text-body font-bold text-primary-500 block mb-2">
                Price:
              </label>
              {isEditing ? (
                <div className="flex flex-col gap-1">
                  <div className="flex items-center gap-2">
                    <input
                      type="number"
                      name="price"
                      value={formData.price}
                      onChange={handleInputChange}
                      onKeyDown={handleNumberKeyDown}
                      min={0}
                      className="border border-primary-500 p-2.5 w-32 rounded-lg bg-white"
                    />
                    <span className="text-sm text-gray-600">THB</span>
                  </div>
                  {errors.price && (
                    <p className="text-red-500 text-sm mt-1">{errors.price}</p>
                  )}
                </div>
              ) : (
                <div className="bg-secondary-600 text-primary-400 px-6 py-2.5 rounded-lg font-semibold flex items-center gap-2">
                  <img
                    src="/icons/banknote.svg"
                    alt="Price"
                    className="w-4 h-4 object-contain"
                  />
                  {savedData.price.toLocaleString()} THB
                </div>
              )}
            </div>

            {isEditing && artwork?.id && (
              <div className="flex justify-end mt-12 w-full">
                <Button
                  variant="error"
                  onClick={() => setShowDeleteConfirm(true)}
                  icon={
                    <img
                      src="/icons/delete.svg"
                      alt="Delete"
                      className="w-4 h-4 object-contain"
                    />
                  }
                >
                  Delete Artwork
                </Button>
              </div>
            )}
          </div>
        </div>

        <div className="mt-12">
          <div className="flex items-center gap-3 mb-6">
            <h2 className="text-xl font-bold text-gray-900">Order Reviews</h2>
            <span className="text-sm font-normal text-gray-400">
              {mockArtworkReviews.length} reviews
            </span>
          </div>
          {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
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
