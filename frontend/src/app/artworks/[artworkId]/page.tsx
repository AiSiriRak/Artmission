"use client"

import { Suspense, useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";

import { ArtistMiniProfile } from "@/components/feature/artwork/ArtistMiniProfile";
import ArtworkDetailCard from "@/components/feature/artwork/ArtworkDetailCard";
import ArtworkOrderSection from "@/components/feature/artwork/ArtworkOrderSection";
import MainLayout from "@/components/feature/main/MainLayout"
import { Button } from "@/components/ui/Button";
import { Loading } from "@/components/ui/Loading";
import { getArtistProfile } from "@/lib/api/artists";
import { getArtworkById } from "@/lib/api/artworks";
import type { ArtistProfile, Artwork } from "@/lib/api/types";

function ArtworkDetailContent() {
    const router = useRouter();
    const { artworkId } = useParams<{ artworkId: string }>();

    const [artwork, setArtwork] = useState<Artwork | null>(null);
    const [artist, setArtist] = useState<ArtistProfile | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        let cancelled = false;
        (async () => {
            try {
                const found = await getArtworkById(artworkId);
                const profile = await getArtistProfile(found.artist_id);
                if (cancelled) return;

                setArtwork(found);
                setArtist(profile);
            } catch (err) {
                if (!cancelled) {
                    setError(err instanceof Error ? err.message : "Failed to load artwork.");
                }
            }
        })();

        return () => {
            cancelled = true;
        };
    }, [artworkId]);

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

            <h2 className="text-h2 mb-10 font-bold text-gray-900">Artwork Detail Page</h2>

            {error && <p className="text-body text-red-600">{error}</p>}
            {!error && (!artwork || !artist) && <Loading />}

            {artwork && artist && (
              <div className="flex items-start gap-10">
                <ArtworkDetailCard artwork={artwork} />
                <div className="sticky top-28 flex flex-col gap-y-4 gap-x-auto">
                  <ArtistMiniProfile artist={artist} />
                  <ArtworkOrderSection 
                    minimumDeadlineDays={artwork.minimum_deadline_days}
                    priceSatang={artwork.price_satang}
                    artworkId={artwork.id}
                  />
                </div>
              </div>
            )}
        </div>
        </MainLayout>
    );
}

export default function ArtworkDetailPage() {
    return (
        <Suspense fallback={<Loading />}>
            <ArtworkDetailContent />
        </Suspense>
    );
}