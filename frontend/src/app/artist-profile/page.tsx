"use client";

import { useEffect, useState } from "react";
import ProfileManager from "@/components/feature/artistprofile/ProfileManager";
import MainLayout from "@/components/feature/main/MainLayout";
import { getArtistProfile, getArtistArtworks } from "@/lib/api/artists";
import { getAccount } from "@/lib/api/users";
import type { ArtistProfile, Artwork } from "@/lib/api/types";
import { Loading } from "@/components/ui/Loading";

export default function ArtistProfilePage() {
  const [artistProfile, setArtistProfile] = useState<ArtistProfile | null>(null);
  const [artworks, setArtworks] = useState<Artwork[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    async function loadData() {
      try {
        setIsLoading(true);
        const account = await getAccount();
        if (!account?.id) {
          throw new Error("User account not found or not logged in.");
        }

        const [profile, artworksData] = await Promise.all([
          getArtistProfile(account.id),
          getArtistArtworks(account.id)
        ]);
        
        if (!profile) {
          throw new Error("Profile not found.");
        }

        setArtistProfile(profile);

        // each artwork in the array
        const artworksList = Array.isArray(artworksData) 
          ? artworksData 
          : (artworksData as any).artworks || (artworksData as any).items || (artworksData as any).data || [];
          
        setArtworks(artworksList);
      } catch (err) {
        console.warn("Cannot fetch artist profile", err);
        setError(true);
      } finally {
        setIsLoading(false);
      }
    }
    loadData();
  }, []);

  if (isLoading) return <Loading />; 

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