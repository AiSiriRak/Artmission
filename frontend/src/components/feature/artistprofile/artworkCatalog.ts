import { getCategories, getStyles, type CatalogReference } from "@/lib/api/artworks";

export const ALLOWED_CATEGORIES = [
  "NOVEL COVER",
  "ILLUSTRATION",
  "CHARACTER DESIGN",
  "BACKGROUND DESIGN",
  "STORYBOARD",
  "CARTOON PANEL",
  "LOGO DESIGN",
  "SOCIAL MEDIA POST",
] as const;

export const ALLOWED_STYLES = [
  "CARTOON",
  "REALISM",
  "SEMI-REALISM",
  "WATERCOLOR",
  "ACRYLIC",
  "SKETCH",
  "ABSTRACT",
  "PIXEL ART",
  "FANTASY",
] as const;

const MOCK_CATEGORY_IDS = [
  "00000000-0000-0000-0000-000000000001",
  "00000000-0000-0000-0000-000000000002",
  "00000000-0000-0000-0000-000000000003",
  "00000000-0000-0000-0000-000000000004",
  "00000000-0000-0000-0000-000000000005",
  "00000000-0000-0000-0000-000000000006",
  "00000000-0000-0000-0000-000000000007",
  "00000000-0000-0000-0000-000000000008",
] as const;

const MOCK_STYLE_IDS = [
  "11111111-0000-0000-0000-000000000001",
  "11111111-0000-0000-0000-000000000002",
  "11111111-0000-0000-0000-000000000003",
  "11111111-0000-0000-0000-000000000004",
  "11111111-0000-0000-0000-000000000005",
  "11111111-0000-0000-0000-000000000006",
  "11111111-0000-0000-0000-000000000007",
  "11111111-0000-0000-0000-000000000008",
  "11111111-0000-0000-0000-000000000009",
] as const;

export type CatalogOption = {
  id: string;
  label: string;
  disabled: boolean;
};

function normalizeLabel(value: string) {
  return value.trim().toLowerCase();
}

const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function isUuid(value: string) {
  return UUID_PATTERN.test(value);
}

export function placeholderCatalog(labels: readonly string[]): CatalogOption[] {
  return labels.map((label) => ({ id: "", label, disabled: true }));
}

export function mapAllowedCatalog(
  allowed: readonly string[],
  fetched: CatalogReference[],
): CatalogOption[] {
  const idByLabel = new Map<string, string>();
  for (const item of fetched) {
    const key = normalizeLabel(item.label);
    if (!key || !isUuid(item.id) || idByLabel.has(key)) continue;
    idByLabel.set(key, item.id);
  }

  return allowed.map((label) => {
    const id = idByLabel.get(normalizeLabel(label)) ?? "";
    return { id, label, disabled: id === "" };
  });
}

function mockReferences(
  labels: readonly string[],
  ids: readonly string[],
): CatalogReference[] {
  return labels.map((label, index) => ({ id: ids[index], label }));
}

function catalogWithMockFallback(
  allowed: readonly string[],
  fetched: CatalogReference[],
  mockIds: readonly string[],
  kind: string,
  alreadyWarned: boolean,
): CatalogReference[] {
  const matched = mapAllowedCatalog(allowed, fetched);
  if (matched.every((option) => option.id !== "")) return fetched;
  if (!alreadyWarned) {
    console.warn(
      `Mock UUIDs are being used for ${kind} because the API is unavailable or returned no matching rows.`,
    );
  }
  return [...fetched, ...mockReferences(allowed, mockIds)];
}

function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(
      () => reject(new Error("catalog request timed out")),
      ms,
    );
    promise.then(
      (value) => {
        clearTimeout(timer);
        resolve(value);
      },
      (error: unknown) => {
        clearTimeout(timer);
        reject(error instanceof Error ? error : new Error(String(error)));
      },
    );
  });
}

export function resolveCatalogId(value: string, options: CatalogOption[]) {
  const enabled = options.filter((option) => !option.disabled && option.id);
  if (!value || enabled.length === 0) return value;
  if (enabled.some((option) => option.id === value)) return value;
  const key = normalizeLabel(value);
  return (
    enabled.find((option) => normalizeLabel(option.label) === key)?.id ?? ""
  );
}

export function resolveCatalogIds(values: string[], options: CatalogOption[]) {
  return [
    ...new Set(
      values
        .map((value) => resolveCatalogId(value, options))
        .filter((value) => value !== ""),
    ),
  ];
}

export function displayLabel(options: CatalogOption[], value: string) {
  const key = normalizeLabel(value);
  const match = options.find(
    (option) => option.id === value || normalizeLabel(option.label) === key,
  );
  return match?.label ?? value;
}

export async function loadArtworkCatalog(): Promise<{
  categories: CatalogOption[];
  styles: CatalogOption[];
}> {
  let categories: CatalogReference[] = [];
  let styles: CatalogReference[] = [];
  let unavailable = false;

  try {
    [categories, styles] = await withTimeout(
      Promise.all([getCategories(), getStyles()]),
      8000,
    );
  } catch (error) {
    unavailable = true;
    console.warn(
      "Mock UUIDs are being used because the category and style API is unavailable.",
      error,
    );
  }

  categories = catalogWithMockFallback(
    ALLOWED_CATEGORIES,
    categories,
    MOCK_CATEGORY_IDS,
    "categories",
    unavailable,
  );
  styles = catalogWithMockFallback(
    ALLOWED_STYLES,
    styles,
    MOCK_STYLE_IDS,
    "styles",
    unavailable,
  );

  return {
    categories: mapAllowedCatalog(ALLOWED_CATEGORIES, categories),
    styles: mapAllowedCatalog(ALLOWED_STYLES, styles),
  };
}
