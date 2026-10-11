"use client";

import type { ReactNode } from "react";
import type { ArtistProfile, Artwork } from "@/lib/api/types";
import type { ReviewData } from "@/app/artist-profile/types";

import ProfileImage from "./ProfileImage";
import TagList from "./TagList";
import ArtworkCard from "./ArtworkCard";
import ReviewList from "./ReviewList";

interface ArtistProfileViewProps {
  profile: ArtistProfile;
  artworks: Artwork[];
  reviews: ReviewData[];
  reviewScore: number | null;
  onArtworkClick: (artwork: Artwork) => void;
  actions?: ReactNode;
  reviewActions?: ReactNode;
}

function formatPrice(satang: number) {
  return (satang / 100).toLocaleString();
}

export default function ArtistProfileView({
  profile,
  artworks,
  reviews,
  reviewScore,
  onArtworkClick,
  actions,
  reviewActions,
}: ArtistProfileViewProps) {
  const categories = [...new Set(artworks.map((artwork) => artwork.category).filter(Boolean))];
  const styles = [...new Set(artworks.flatMap((artwork) => artwork.styles ?? []).filter(Boolean))];
  const prices = artworks.map((artwork) => artwork.price_satang).filter((price) => Number.isFinite(price) && price > 0);

  let priceRange = "N/A";

  if (prices.length > 0) {
    const minPrice = Math.min(...prices);
    const maxPrice = Math.max(...prices);

    priceRange = (minPrice === maxPrice) ? `${formatPrice(minPrice)} THB` : `${formatPrice(minPrice)} - ${formatPrice(maxPrice)} THB`;
  }

  return (
    <div className="w-full">
      {/* Profile */}
      <div className="w-full">
        <div className="w-full h-48 bg-secondary-300 border border-neutral" />

        <div className="max-w-5xl mx-auto px-8 relative">
          <div className="flex justify-between items-end -mt-16 sm:-mt-20 mb-6 relative z-10">
            <ProfileImage
              imageUrl={profile.profile_url || ""}
              className="w-36 h-36 md:w-44 md:h-44"
            />

            {actions}
          </div>

          <div className="mb-6">
            <h1 className="text-h2 font-bold text-gray-900">
              {profile.artist_name || "Unknown Artist"}
            </h1>
          </div>

          <p className="text-gray-700 text-sm md:text-base leading-relaxed mb-8 whitespace-pre-wrap">
            {profile.description}
          </p>

          <div className="flex flex-wrap gap-x-16 gap-y-6 mb-12">
            <div>
              <span className="block text-body font-bold text-gray-800 mb-3">
                Category:
              </span>

              {categories.length > 0 ? (
                <TagList items={categories} variant="category" />
              ) : (
                <span className="text-gray-400 text-sm">None</span>
              )}
            </div>

            <div>
              <span className="block text-body font-bold text-gray-800 mb-3">
                Style:
              </span>

              {styles.length > 0 ? (
                <TagList items={styles} variant="style" />
              ) : (
                <span className="text-gray-400 text-sm">None</span>
              )}
            </div>

            <div>
              <span className="block text-body font-bold text-gray-800 mb-3">
                Price Range:
              </span>

              <span className="font-bold text-lg text-gray-800">
                {priceRange}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Artworks */}
      <div className="max-w-5xl mx-auto px-8 pb-10">
        <div className="flex items-center gap-2 bg-secondary-300 border-l-4 border-accent-500 px-4 py-2.5 mb-6 rounded-r-md text-h3 font-bold text-gray-900">
          <span>Artworks</span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {artworks.map((artwork) => (
            <ArtworkCard
              key={artwork.id}
              artwork={artwork}
              showEditControls={false}
              onClick={() => onArtworkClick(artwork)}
            />
          ))}
        </div>
      </div>

      {/* Reviews */}
      <div className="max-w-5xl mx-auto px-8 pb-20">
        <div className="flex items-center gap-2 bg-secondary-300 border-l-4 border-accent-500 px-4 py-2.5 mb-6 rounded-r-md text-h3 font-bold text-gray-900">
          <span>Reviews</span>
          <span className="text-red-400 text-base ml-1">☆</span>
          <span>
            {reviewScore == null ? "—" : reviewScore.toFixed(1)}/5
          </span>
        </div>

        <ReviewList reviews={reviews} />

        {reviewActions && (
          <div className="mt-6">
            {reviewActions}
          </div>
        )}
      </div>
    </div>
  );
}