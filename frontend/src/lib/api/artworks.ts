import { apiFetch } from "./client";
import type { Artwork, CreateArtworkInput, UpdateArtworkInput } from "./types";

const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function isUuid(value: string) {
  return UUID_PATTERN.test(value);
}

function uuidList(value: unknown): string[] {
  let items: unknown[] = [];
  if (Array.isArray(value)) {
    items = value;
  } else if (typeof value === "string" && value.trim().startsWith("[")) {
    try {
      const parsed = JSON.parse(value) as unknown;
      if (Array.isArray(parsed)) items = parsed;
    } catch {
      items = [];
    }
  }

  return items.flatMap((item) => {
    if (typeof item !== "string") return [];
    const id = item.trim();
    return isUuid(id) ? [id] : [];
  });
}

/**
 * Multipart body for artwork create/update.
 * category_id is one UUID string. style_ids is one JSON array of UUID strings.
 */
function buildFormData(data: Record<string, unknown>): FormData {
  const formData = new FormData();
  Object.entries(data).forEach(([key, value]) => {
    if (value === undefined || value === null) return;

    if (key === "category_id") {
      const raw = Array.isArray(value) ? value[0] : value;
      const id = typeof raw === "string" ? raw.trim() : "";
      if (isUuid(id)) formData.append("category_id", id);
      return;
    }

    if (key === "style_ids" || key === "deleted_sample_urls") {
      const list =
        key === "style_ids"
          ? uuidList(value)
          : Array.isArray(value)
            ? value.map(String)
            : [];
      formData.append(key, JSON.stringify(list));
      return;
    }

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
      return;
    }

    if (value instanceof Blob) {
      formData.append(key, value);
      return;
    }

    formData.append(key, String(value));
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

export type CatalogReference = {
  id: string;
  label: string;
};

function readCatalogList(body: unknown, key: string): CatalogReference[] {
  const list = Array.isArray(body)
    ? body
    : body &&
        typeof body === "object" &&
        Array.isArray((body as Record<string, unknown>)[key])
      ? ((body as Record<string, unknown>)[key] as unknown[])
      : [];

  return list.flatMap((item) => {
    if (!item || typeof item !== "object") return [];
    const record = item as Record<string, unknown>;
    if (typeof record.id !== "string" || typeof record.label !== "string") {
      return [];
    }
    return [{ id: record.id, label: record.label }];
  });
}

/**
 * ดึงรายการหมวดหมู่พร้อม UUID จากฐานข้อมูล
 * GET /categories
 */
export async function getCategories(): Promise<CatalogReference[]> {
  const body = await apiFetch<unknown>("/categories");
  return readCatalogList(body, "categories");
}

/**
 * ดึงรายการสไตล์พร้อม UUID จากฐานข้อมูล
 * GET /styles
 */
export async function getStyles(): Promise<CatalogReference[]> {
  const body = await apiFetch<unknown>("/styles");
  return readCatalogList(body, "styles");
}
