"use client";

import Image from "next/image";

import type { SearchArtwork } from "@/lib/api/types";
import { ArtworkCard } from "@/components/ui/ArtworkCard";

interface SearchArtworkCardProps {
  item: SearchArtwork;
}

function formatPriceThb(priceSatang: number): string {
  return `${(priceSatang / 100).toLocaleString("en-US")} THB`;
}

function formatReviewScore(score: number): string {
  return score.toFixed(1);
}

export function SearchArtworkCard({ item }: SearchArtworkCardProps) {
  const imageUrl = item.artwork_samples?.[0]?.image_url ?? null;

  return (
    <ArtworkCard
      href={`/artworks/${item.id}`}
      imageUrl={imageUrl}
      imageAlt={item.name}
    >
      <p className="truncate text-h3 text-primary-500">
        {formatPriceThb(item.price_satang)}
      </p>

      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col">
          <p className="truncate text-body font-medium text-primary-500">
            {item.name}
          </p>
          <p className="truncate text-small text-primary-500">
            {item.artist.artist_name}
          </p>
        </div>

        {item.artist.review_score !== null && (
          <div className="flex shrink-0 items-center gap-1 pt-0.5">
            <span className="text-body font-medium text-primary-500">
              {formatReviewScore(item.artist.review_score)}
            </span>
            <Image src="/icons/star.svg" alt="" width={20} height={20} />
          </div>
        )}
      </div>
    </ArtworkCard>
  );
}
