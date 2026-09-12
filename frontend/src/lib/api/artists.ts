import { apiFetch } from "./client";
import { getAccount } from "./users";
import type { 
  ArtistProfile, 
  UpdateArtistInput, 
  GetArtistArtworksOutput 
} from "./types";

/**
 * ดึงข้อมูลโปรไฟล์สาธารณะของศิลปินตาม artist_id
 * GET /artists/{artist_id}
 */
export async function getArtistProfile(
  artistId: string,
  limit?: number,
  offset?: number
): Promise<ArtistProfile> {
  const params = new URLSearchParams();
  if (limit !== undefined) params.append("limit", limit.toString());
  if (offset !== undefined) params.append("offset", offset.toString());

  const queryString = params.toString() ? `?${params.toString()}` : "";
  return apiFetch<ArtistProfile>(`/artists/${artistId}${queryString}`);
}

/**
 * ดึงผลงานศิลปะทั้งหมดของศิลปิน
 * GET /artists/{artist_id}/artworks
 */
export async function getArtistArtworks(artistId: string) {
  const res = await apiFetch(`/artists/${artistId}/artworks`);
  
  const rawList = Array.isArray(res) ? res : (res as any).artworks || [];

  const normalizedList = rawList.map((art: any) => ({
    ...art,
    id: art.id || art.artwork_id,
  }));

  return normalizedList;
}
/**
 * อัปเดตข้อมูลโปรไฟล์ศิลปินของตนเอง (รองรับทั้ง Object และ FormData)
 * PUT /artists/me
 */
export async function updateArtistProfile(
  data: UpdateArtistInput | FormData,
): Promise<ArtistProfile> {
  let formData: FormData;

  if (data instanceof FormData) {
    formData = data;
  } else {
    formData = new FormData();
    Object.entries(data).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {

        const val = value as any;
        if (val instanceof Blob) {
          formData.append(key, val);
        } else {
          formData.append(key, String(val));
        }
      }
    });
  }

  return apiFetch<ArtistProfile>("/artists/me", {
    method: "PUT",
    body: formData,
  });
}

// Re-export getAccount ไว้กรณีที่ต้องการใช้ผ่านโมดูลนี้
export { getAccount };