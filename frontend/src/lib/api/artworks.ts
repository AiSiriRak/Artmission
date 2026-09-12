import { apiFetch } from "./client";
import type { Artwork, CreateArtworkInput, UpdateArtworkInput } from "./types";

/**
 * Helper function สำหรับแปลง Object เป็น FormData
 * รองรับการจัดการ Array เช่น styles หรือไฟล์รูปภาพหลายรูป
 */
function buildFormData(data: Record<string, unknown>): FormData {
  const formData = new FormData();
  Object.entries(data).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      if (Array.isArray(value)) {
        const isFileField =
          key === "artwork_samples" ||
          key === "uploaded_samples" ||
          value.some((item) => item instanceof Blob);

        if (isFileField) {
          value.forEach((item) => {
            formData.append(key, item as Blob | string);
          });
        } else {
          formData.append(key, JSON.stringify(value));
        }
      } else if (value instanceof Blob) {
        formData.append(key, value);
      } else {
        formData.append(key, String(value));
      }
    }
  });
  return formData;
}

/**
 * สร้างรายการผลงานศิลปะชิ้นใหม่ พร้อมแนบไฟล์รูป
 * POST /artworks
 */
export async function createArtwork(
  data: CreateArtworkInput | FormData,
): Promise<Artwork> {
  const body =
    data instanceof FormData
      ? data
      : buildFormData(data as Record<string, unknown>);
  return apiFetch<Artwork>("/artworks", {
    method: "POST",
    body,
  });
}

/**
 * แก้ไขผลงานศิลปะ พร้อมอัปเดตไฟล์รูป
 * PUT /artworks/{artwork_id}
 */
export async function updateArtwork(
  artworkId: string,
  data: UpdateArtworkInput | FormData,
): Promise<Artwork> {
  const body =
    data instanceof FormData
      ? data
      : buildFormData(data as Record<string, unknown>);
  return apiFetch<Artwork>(`/artworks/${artworkId}`, {
    method: "PUT",
    body,
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
