import { apiFetch } from "./client";
import { getAccount } from "./users";
import type { ArtistProfile, UpdateArtistInput } from "./types";

/**
 * ดึงข้อมูลโปรไฟล์สาธารณะของศิลปินตาม artist_id
 * GET /artists/{artist_id}
 */
export async function getArtistProfile(
  artistId: string,
): Promise<ArtistProfile> {
  return apiFetch<ArtistProfile>(`/artists/${artistId}`);
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
        formData.append(key, value as string | Blob);
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