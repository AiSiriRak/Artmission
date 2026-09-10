import ProfileManager from "@/components/feature/artistprofile/ProfileManager";
import MainLayout from "@/components/feature/main/MainLayout";
import { getArtistProfile } from "@/lib/api/artists";
import { getAccount } from "@/lib/api/users"; // 1. นำเข้า getAccount เพื่อหา ID ของผู้ใช้ปัจจุบัน
import type { ArtistProfile, Artwork } from "@/lib/api/types";

export default async function ArtistProfilePage() {
  let artistProfile: ArtistProfile | null = null;
  let artworks: Artwork[] = []; 

  try {
    // 1. ดึงข้อมูล Account ของผู้ใช้
    const account = await getAccount();

    // 2. ใช้ account.id ส่งไปยัง GET /artists/{artist_id}
    if (account?.id) {
      artistProfile = await getArtistProfile(account.id);
    } else {
      console.warn("Account ID not found");
    }
    
  } catch (error: any) {
    console.warn("Cannot fetch artist profile", error);
    if (error.body?.errors) {
      console.dir(error.body.errors, { depth: null });
    }
  }

  // Fallback กรณีดึงข้อมูลไม่สำเร็จ
  if (!artistProfile) {
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