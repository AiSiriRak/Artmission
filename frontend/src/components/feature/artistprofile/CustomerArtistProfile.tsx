"use client";

import { useEffect, useState } from "react";

import MainLayout from "@/components/feature/main/MainLayout";
import ProfileImage from "./ProfileImage";
import TagList from "./TagList";
import { Loading } from "@/components/ui/Loading";

import { getArtistProfile } from "@/lib/api/artists";
import { isApiError } from "@/lib/api/error";
import type { ArtistProfile } from "@/lib/api/types";


interface CustomerArtistProfileProps {
  artistId: string;
}

type ProfileState =
  | { status: "loading" }
  | { status: "success"; profile: ArtistProfile }
  | { status: "error"; message: string };

export default function CustomerArtistProfile({
  artistId,
}: CustomerArtistProfileProps) {
  const [state, setState] = useState<ProfileState>({
    status: "loading",
  });

  useEffect(() => {
    let active = true;

    async function loadProfile() {
      try {
        const profile = await getArtistProfile(artistId);

        if (active) {
          setState({ status: "success", profile });
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

  const { profile } = state;

  const categories = (profile.categories ?? []).map(
    (category) => category.label,
  );

  const styles = (profile.styles ?? []).map(
    (style) => style.label,
  );

  return (
    <MainLayout usertype="customer">
      <div className="w-full bg-white text-primary-500">
        <div className="h-48 bg-secondary-300 border border-neutral" />

        <div className="relative mx-auto max-w-5xl px-8 pb-12">
          <div className="relative z-10 -mt-16 mb-6 sm:-mt-20">
            <ProfileImage
              imageUrl={profile.profile_url}
              className="w-36 h-36 md:w-44 md:h-44"
            />
          </div>

          <h1 className="mb-6 text-h2 font-bold break-words">
            {profile.artist_name}
          </h1>

          <p className="mb-8 whitespace-pre-wrap break-words text-body text-primary-400">
            {profile.description?.trim() || "ยังไม่มีคำอธิบาย"}
          </p>

          <div className="flex flex-wrap gap-x-16 gap-y-6">
            <div>
              <h2 className="mb-3 text-body font-bold">
                Category:
              </h2>
              <TagList items={categories} variant="category" />
            </div>

            <div>
              <h2 className="mb-3 text-body font-bold">
                Style:
              </h2>
              <TagList items={styles} variant="style" />
            </div>
          </div>
        </div>
      </div>
    </MainLayout>
  );
}