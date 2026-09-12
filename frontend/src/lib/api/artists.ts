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
  
  // แกะ Array ออกมา
  const rawList = Array.isArray(res) ? res : (res as any).artworks || [];

  // ✅ ทำ Data Mapping: แปลง artwork_id ให้เป็น id ทุกตัว
  const normalizedList = rawList.map((art: any) => ({
    ...art,
    id: art.id || art.artwork_id, // บังคับให้มี id เสมอ
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
        // แคสต์ให้เป็น any เพื่อให้ TypeScript ยอมให้เช็ค instanceof ได้โดยไม่แจ้ง error
        const val = value as any;
        
        if (val instanceof Blob) {
          // กรณีเป็นไฟล์ (File หรือ Blob)
          formData.append(key, val);
        } else {
          // กรณีเป็นข้อความ ตัวเลข หรือ boolean ให้แปลงเป็น String
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