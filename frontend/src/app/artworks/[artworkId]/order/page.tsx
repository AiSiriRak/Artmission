"use client"

import { Suspense, useEffect, useState } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";

import MainLayout from "@/components/feature/main/MainLayout";
import { Button } from "@/components/ui/Button";
import { Loading } from "@/components/ui/Loading";
import { OrderRequestSection } from "@/components/feature/order-placement/OrderRequestSection";
import { getArtistArtworks } from "@/lib/api/artists";
import type { Artwork } from "@/lib/api/types";

function OrderPlacementContent() {
  const router = useRouter();
  const { artworkId } = useParams<{ artworkId: string }>();
  const artistId = useSearchParams().get("artistId");

  const [artwork, setArtwork] = useState<Artwork | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!artistId) {
      setError("Missing artist for this artwork.");
      return;
    }

    let cancelled = false;
    getArtistArtworks(artistId)
      .then((artworks) => {
        if (cancelled) return;
        const found = artworks.find((item) => item.id === artworkId);
        if (found) setArtwork(found);
        else setError("Artwork not found.");
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load artwork.");
        }
      });

    return () => {
      cancelled = true;
    };
  }, [artworkId, artistId]);

  return (
    <MainLayout usertype="customer">
        
        <div className="p-20">
            <Button
              variant="light"
              onClick={() => router.back()}
              className="text-button mb-10 flex items-center gap-2 !border"
            >
              <span>←</span> Back
            </Button>

            <h2 className="text-h2 mb-10 font-bold text-gray-900">Order Detail</h2>

            {error && <p className="text-body text-red-600">{error}</p>}
            {!error && !artwork && <Loading />}
            {artwork && <OrderRequestSection artwork={artwork} />}

        </div>
        </MainLayout>
  );
}

export default function OrderPlacementPage() {
  return (
    <Suspense fallback={<Loading />}>
      <OrderPlacementContent />
    </Suspense>
  );
}