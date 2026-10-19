import Image from "next/image";
import Link from "next/link";
import type { ReactNode } from "react";

import { WhiteCard } from "@/components/ui/WhiteCard";

interface ArtworkCardProps {
  href?: string;
  imageUrl: string | null;
  imageAlt: string;
  /** Rendered absolutely-positioned over the top-right corner of the card. */
  overlay?: ReactNode;
  children: ReactNode;
}

/**
 * Shared visual shell for image-led cards (artwork search results, order
 * history, …). Keep width/rounding/image sizing here so every usage stays
 * visually consistent; pass domain-specific content via `children`.
 */
export function ArtworkCard({
  href,
  imageUrl,
  imageAlt,
  overlay,
  children,
}: ArtworkCardProps) {
  const content = (
    <div className="relative w-[320px] min-w-0 hover:brightness-90">
      {overlay}
      <WhiteCard
        margin=""
        roundsize="rounded-4xl"
        padding=""
        className="w-full min-w-0 max-w-none overflow-hidden"
      >
        <div className="flex h-80 w-full items-center justify-center bg-white">
          <Image
            src={imageUrl || "/icons/emptyimage.svg"}
            alt={imageAlt}
            width={320}
            height={320}
            className="h-full w-full object-cover"
          />
        </div>
        <div className="space-y-2 p-6">{children}</div>
      </WhiteCard>
    </div>
  );

  return href ? (
    <Link href={href} className="block">
      {content}
    </Link>
  ) : (
    content
  );
}
