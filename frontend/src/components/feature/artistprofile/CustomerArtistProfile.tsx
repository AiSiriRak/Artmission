"use client";

import { useEffect, useState } from "react";
import Image from "next/image";

import { useRouter } from "next/navigation";

import MainLayout from "@/components/feature/main/MainLayout";
import ArtistProfileView from "./ArtistProfileView";
import { Loading } from "@/components/ui/Loading";
import { Button } from "@/components/ui/Button";

import { getArtistProfile, getArtistArtworks } from "@/lib/api/artists";
import { isApiError } from "@/lib/api/error";
import type { ArtistProfile, Artwork } from "@/lib/api/types";
import {
  getArtistReviews,
  type ArtistReviewData,
} from "@/lib/api/reviews";

interface CustomerArtistProfileProps {
  artistId: string;
}

type ProfileState =
  | { status: "loading" }
  | { status: "success"; profile: ArtistProfile; artworks: Artwork[]; }
  | { status: "error"; message: string };

export default function CustomerArtistProfile({
  artistId,
}: CustomerArtistProfileProps) {

  const router = useRouter();

  const [state, setState] = useState<ProfileState>({
    status: "loading",
  });

  useEffect(() => {
    let active = true;

    async function loadProfile() {
      try {
        const [profile, artworks] = await Promise.all([
          getArtistProfile(artistId),
          getArtistArtworks(artistId),
        ]);

        if (active) {
          setState({ status: "success", profile, artworks });
        }
      } catch (error: unknown) {
        if (!active) return;

        const message =
          isApiError(error) && error.status === 404
            ? "ไม่พบข้อมูลศิลปิน"
            : "โหลดข้อมูลไม่สำเร็จ กรุณาตรวจสอบ Backend และ ID ศิลปิน";

        setState({ status: "error", message });
      }
    }

    void loadProfile();

    return () => {
      active = false;
    };
  }, [artistId]);

  const [reviews, setReviews] = useState<ArtistReviewData[]>([]);

  useEffect(() => {
    let cancelled = false;

    const fetchReviews = async () => {
      try {
        const data = await getArtistReviews(artistId, 100);

        if (!cancelled) {
          setReviews(data);
        }
      } catch (error) {
        console.error("Failed to fetch artist reviews:", error);

        if (!cancelled) {
          setReviews([]);
        }
      }
    };

    fetchReviews();

    return () => {
      cancelled = true;
    };
  }, [artistId]);

  if (state.status === "loading") {
    return (
      <MainLayout usertype="customer">
        <Loading />
      </MainLayout>
    );
  }

  if (state.status === "error") {
    return (
      <MainLayout usertype="customer">
        <p role="alert" className="p-12 text-center text-error">
          {state.message}
        </p>
      </MainLayout>
    );
  }

  const { profile, artworks } = state;

  return (
    <MainLayout usertype="customer">
      <div className="w-full bg-white text-primary-500">
        <ArtistProfileView
          profile={profile}
          artworks={artworks}
          reviews={reviews}
          reviewScore={profile.review_score}
          onArtworkClick={(artwork) => {
            router.push(`/artworks/${encodeURIComponent(artwork.id)}`);
          }}
          actions={
            <Button
              type="button"
              variant="light"
              className="!border w-24"
              onClick={() => router.back()}
              icon={
                <Image
                  src="/icons/Arrow_left.svg"
                  alt=""
                  width={16}
                  height={16}
                  className="h-4 w-4 object-contain"
                />
              }
            >
              Back
            </Button>
          }
        />
      </div>
    </MainLayout>
  );
}