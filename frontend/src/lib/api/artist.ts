import { apiFetch } from "./client";
import { getAccount } from "./users";
import {
  ArtistProfile,
  Artwork,
  CreateArtworkInput,
  UpdateArtistInput,
} from "./types";

// ดึงข้อมูลโปรไฟล์ (ดึงข้อมูล Account ร่วมกับ Artist)
export async function getArtistProfile(): Promise<ArtistProfile> {
  const user = await getAccount();
  
  try {
    const artist = await apiFetch<ArtistProfile>("/artists/me");
    return {
      ...artist,
      // เปลี่ยนจาก name เป็น artist_name ตาม Schema ใหม่
      artist_name: artist.artist_name || user.username, 
    };
  } catch {
    // โครงสร้าง Fallback ใหม่ที่ตรงกับ ArtistProfileView
    return {
      artist_id: user.id, // เปลี่ยนจาก id เป็น artist_id
      artist_name: user.username, // เปลี่ยนจาก name เป็น artist_name
      description: "",
      categories: [], // ฟิลด์ใหม่ที่ Backend คืนค่า
      styles: [], // ฟิลด์ใหม่ที่ Backend คืนค่า
      max_price_satang: null, // ฟิลด์ใหม่
      min_price_satang: null, // ฟิลด์ใหม่
      review_score: null, // ฟิลด์ใหม่
    };
  }
}

// อัปเดตข้อมูลโปรไฟล์ศิลปิน
export async function updateArtistProfile(
  data: UpdateArtistInput
): Promise<ArtistProfile> {
  // Backend ทำ API เส้นนี้เสร็จแล้ว จึงสามารถยิงตรงและลบ Mock data ออกได้เลย
  return await apiFetch<ArtistProfile>("/artists/me", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
}

// CRUD ผลงาน Artwork
export async function getArtworks(artistId: string): Promise<Artwork[]> {
  return apiFetch<Artwork[]>(`/artists/${artistId}/artworks`);
}

export async function createArtwork(data: CreateArtworkInput): Promise<Artwork> {
  return apiFetch<Artwork>("/artworks", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
}

export async function deleteArtwork(id: string | number): Promise<void> {
  return apiFetch<void>(`/artworks/${id}`, {
    method: "DELETE",
  });
}