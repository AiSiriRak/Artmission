import { apiFetch } from "./client";
import type { 
  Artwork, 
  CreateArtworkInput, 
  UpdateArtworkInput 
} from "./types";

/**
 * ดึงข้อมูลผลงานศิลปะชิ้นเดียวตาม artworkId
 * GET /artworks/{artworkId}
 */
export async function getArtwork(artworkId: string): Promise<Artwork> {
  return apiFetch<Artwork>(`/artworks/${artworkId}`);
}

/**
 * สร้างรายการผลงานศิลปะชิ้นใหม่
 * POST /artworks
 */
export async function createArtwork(
  data: CreateArtworkInput,
): Promise<Artwork> {
  return apiFetch<Artwork>("/artworks", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

/**
 * อัปเดตข้อมูลผลงานศิลปะ
 * PUT /artworks/{artworkId}
 */
export async function updateArtwork(
  artworkId: string,
  data: UpdateArtworkInput,
): Promise<Artwork> {
  return apiFetch<Artwork>(`/artworks/${artworkId}`, {
    method: "PUT", // หาก Backend ใช้ PATCH สามารถเปลี่ยนตรงนี้เป็น "PATCH" ได้ครับ
    body: JSON.stringify(data),
  });
}

/**
 * ลบผลงานศิลปะ
 * DELETE /artworks/{artworkId}
 */
export async function deleteArtwork(artworkId: string): Promise<void> {
  await apiFetch<void>(`/artworks/${artworkId}`, {
    method: "DELETE",
  });
}