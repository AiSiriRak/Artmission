"use client";

import { useEffect, useState } from "react";
import ProfileManager from "@/components/feature/artistprofile/ProfileManager";
import MainLayout from "@/components/feature/main/MainLayout";
import { getArtistProfile } from "@/lib/api/artists";
import { getAccount } from "@/lib/api/users";
import type { ArtistProfile, Artwork } from "@/lib/api/types";
import { Loading } from "@/components/ui/Loading"; // สมมติว่ามี Loading Component

export default function ArtistProfilePage() {
  const [artistProfile, setArtistProfile] = useState<ArtistProfile | null>(null);
  const [artworks, setArtworks] = useState<Artwork[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    async function loadData() {
      try {
        const account = await getAccount();
        if (account?.id) {
          const profile = await getArtistProfile(account.id);
          setArtistProfile(profile);
          // อนาคตเพิ่ม: const works = await getArtworks(account.id); setArtworks(works);
        }
      } catch (err) {
        console.warn("Cannot fetch artist profile", err);
        setError(true);
      } finally {
        setIsLoading(false);
      }
    }
    loadData();
  }, []);

  if (isLoading) return <Loading />; // รอโหลดข้อมูล

  if (error || !artistProfile) {
    return (
      <MainLayout page="Artist Profile" usertype="artist">
        <div className="w-full min-h-screen flex items-center justify-center bg-white text-black">
          <p className="text-xl">ไม่พบข้อมูลศิลปิน หรือเกิดข้อผิดพลาดในการโหลดข้อมูล</p>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout page="Artist Profile" usertype="artist">
      <div className="w-full bg-white text-black">
        <ProfileManager 
          initialProfile={artistProfile} 
          initialArtworks={artworks} 
        />
      </div>
    </MainLayout>
  );
}