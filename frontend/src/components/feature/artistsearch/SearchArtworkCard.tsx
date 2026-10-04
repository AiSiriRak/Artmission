"use client";

import Image from "next/image";

import type { ArtistSearchItem } from "@/lib/api/artists";
import { WhiteCard } from "@/components/ui/WhiteCard";

interface SearchArtworkCardProps {
  item: ArtistSearchItem;
}

function formatPriceThb(priceSatang: number): string {
  return `${(priceSatang / 100).toLocaleString("en-US")} THB`;
}

function formatReviewScore(score: number): string {
  return score.toFixed(1);
}

export function SearchArtworkCard({ item }: SearchArtworkCardProps) {
  return (
    <div className="relative w-full min-w-0 hover:brightness-95 transition-[filter]">
      <WhiteCard
        margin=""
        roundsize="rounded-[45px]"
        padding="p-0"
        className="w-full min-w-0 max-w-none overflow-hidden"
      >
        <div className="relative w-full aspect-square flex items-center justify-center bg-white">
          <Image
            src={item.image_url || "/icons/emptyimage.svg"}
            alt={item.artwork_name}
            width={378}
            height={378}
            className="object-contain w-full h-full p-10"
          />
        </div>

        <div className="flex w-full flex-col gap-1.5 p-8.75">
          <p className="truncate text-h2 text-primary-500">
            {formatPriceThb(item.price_satang)}
          </p>

          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0 flex flex-col">
              <p className="truncate text-h3 text-primary-500">
                {item.artwork_name}
              </p>
              <p className="truncate text-body text-primary-500">
                {item.artist_name}
              </p>
            </div>

            {item.review_score !== null && (
              <div className="flex shrink-0 items-center gap-1 pt-0.5">
                <span className="text-h3 text-primary-500">
                  {formatReviewScore(item.review_score)}
                </span>
                <Image src="/icons/star.svg" alt="" width={20} height={20} />
              </div>
            )}
          </div>
        </div>
      </WhiteCard>
    </div>
  );
}
