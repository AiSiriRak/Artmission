import { useRef } from "react";
import Image from "next/image";

import { Button } from "@/components/ui/Button";

interface ArtworkSampleGalleryProps {
  isEditing: boolean;
  images: string[];
  hasSubmitted: boolean;
  onUpload: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onRemove: (index: number) => void;
}

export default function ArtworkSampleGallery({
  isEditing,
  images,
  hasSubmitted,
  onUpload,
  onRemove,
}: ArtworkSampleGalleryProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleAddSampleClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className="mb-10">
      <div className="flex justify-between items-center mb-4">
        <div className="flex flex-col">
          <h3 className="text-body font-bold text-primary-500">
            Artwork Samples
          </h3>
          <h4 className="text-sm text-gray-600">
            The first image will be used as the cover.
          </h4>
          {hasSubmitted && images.length === 0 && (
            <p className="text-red-500 text-sm mt-1">
              Please upload at least one artwork sample image
            </p>
          )}
        </div>

        {isEditing && (
          <>
            <Button variant="transparent" onClick={handleAddSampleClick}>
              + Add Sample
            </Button>

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

      <div className="flex gap-4 overflow-x-auto pb-3">
        {images.map((imgUrl, index) => (
          <div
            key={index}
            className="relative h-48 w-fit flex-none overflow-hidden rounded-xl"
          >
            {isEditing && (
              <button
                onClick={() => onRemove(index)}
                className="absolute top-2 right-2 bg-white shadow-md rounded-full w-6 h-6 flex items-center justify-center text-gray-500 hover:text-red-500 text-xs cursor-pointer z-10"
              >
                ✕
              </button>
            )}

            <Image
              src={imgUrl}
              alt={`sample-${index}`}
              width={400}
              height={400}
              className="h-48 w-auto max-w-none object-contain"
            />
          </div>
        ))}
      </div>
    </div>
  );
}
