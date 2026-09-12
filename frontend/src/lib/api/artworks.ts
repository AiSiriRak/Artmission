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

/**
 * อัปโหลดรูปภาพไปยังเซิร์ฟเวอร์
 * POST /upload (⚠️ คุณต้องเปลี่ยน URL นี้ให้ตรงกับ API ของ Backend คุณ)
 */
export async function uploadImage(file: File): Promise<string> {
  const formData = new FormData();
  // "file" คือชื่อฟิลด์ที่ Backend ต้องการ (ถ้า Backend ใช้ชื่อ "image" ก็ให้เปลี่ยนเป็น "image")
  formData.append("file", file); 

  // สมมติว่า Backend ตอบกลับมาเป็น { "url": "https://..." } 
  // ⚠️ เปลี่ยน "/upload" เป็น Path ที่ Backend คุณใช้รับไฟล์จริงๆ
  const response = await apiFetch<{ url: string }>("/upload", {
    method: "POST",
    body: formData,
  });

  return response.url;
}
