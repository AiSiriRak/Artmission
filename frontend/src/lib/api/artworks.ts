import { apiFetch } from "./client";
import type { 
  Artwork, 
  CreateArtworkInput 
} from "./types";

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
 * ลบผลงานศิลปะ
 * DELETE /artworks/{artwork_id}
 */
export async function deleteArtwork(artworkId: string): Promise<void> {
  await apiFetch<void>(`/artworks/${artworkId}`, {
    method: "DELETE",
  });
}