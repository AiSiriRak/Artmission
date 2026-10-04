/**
 * Artist Search mock.
 *
 * - `search` matches artist name only
 * - results are artworks (multiple rows can share the same artist)
 * - price / category / styles filter the artwork
 * - min_rating filters by the artist's review score
 */

export const ARTIST_SEARCH_CATEGORIES = [
  "Book",
  "Comic",
  "Game",
  "Animation",
  "Portrait",
  "Illustration",
  "Other",
] as const;

export const ARTIST_SEARCH_STYLES = [
  "Pixel Art",
  "Water Color",
  "Cartoon",
  "Graphic",
  "Anime",
  "Realism",
  "Digital art",
] as const;

export type ArtistSearchSort =
  "price_asc" | "price_desc" | "review_score_asc" | "review_score_desc";

export type ArtistSearchParams = {
  /** Matches artist name only. */
  search?: string;
  sort?: ArtistSearchSort;
  /** Artwork price in satang (100 satang = 1 THB). */
  min_price_satang?: number;
  /** Artwork price in satang (100 satang = 1 THB). */
  max_price_satang?: number;
  /** Single artwork category label. */
  category?: string;
  /** Multi-select artwork style labels; artwork must match at least one. */
  styles?: string[];
  /** Inclusive minimum artist review score (1–5). Omit = any rating. */
  min_rating?: number;
  limit?: number;
  offset?: number;
};

/** One search result row = one artwork (+ denormalized artist fields for the card). */
export type ArtistSearchItem = {
  artwork_id: string;
  artwork_name: string;
  image_url: string | null;
  price_satang: number;
  category: string;
  styles: string[];
  artist_id: string;
  artist_name: string;
  review_score: number | null;
};

export type ArtistSearchResult = {
  artworks: ArtistSearchItem[];
  total: number;
};

type MockArtist = {
  artist_id: string;
  artist_name: string;
  review_score: number | null;
};

type MockArtwork = {
  artwork_id: string;
  artwork_name: string;
  image_url: string | null;
  price_satang: number;
  category: string;
  styles: string[];
  artist_id: string;
};

const MOCK_DELAY_MS = { min: 400, max: 800 } as const;

const MOCK_ARTISTS: MockArtist[] = [
  {
    artist_id: "artist-001",
    artist_name: "Leonardo da Vinci",
    review_score: 4.8,
  },
  {
    artist_id: "artist-002",
    artist_name: "Pixel Sage",
    review_score: 4.5,
  },
  {
    artist_id: "artist-003",
    artist_name: "Aiko Nakamura",
    review_score: 4.2,
  },
  {
    artist_id: "artist-004",
    artist_name: "Mira Chen",
    review_score: 3.9,
  },
  {
    artist_id: "artist-005",
    artist_name: "Bit Weaver",
    review_score: 4.7,
  },
  {
    artist_id: "artist-006",
    artist_name: "Sofia Reyes",
    review_score: 3.5,
  },
  {
    artist_id: "artist-007",
    artist_name: "Kenji Park",
    review_score: 4.0,
  },
  {
    artist_id: "artist-008",
    artist_name: "Nova Ink",
    review_score: 2.8,
  },
  {
    artist_id: "artist-009",
    artist_name: "Elena Vargas",
    review_score: 5.0,
  },
  {
    artist_id: "artist-010",
    artist_name: "Retro Byte",
    review_score: 3.2,
  },
  {
    artist_id: "artist-011",
    artist_name: "Hannah Cole",
    review_score: null,
  },
  {
    artist_id: "artist-012",
    artist_name: "Theo Blanc",
    review_score: 4.4,
  },
];

const MOCK_ARTWORKS: MockArtwork[] = [
  {
    artwork_id: "artwork-001",
    artwork_name: "Portrait Study 01",
    image_url: null,
    price_satang: 150_000,
    category: "Portrait",
    styles: ["Realism", "Digital art"],
    artist_id: "artist-001",
  },
  {
    artwork_id: "artwork-002",
    artwork_name: "Portrait Study 02",
    image_url: null,
    price_satang: 280_000,
    category: "Illustration",
    styles: ["Realism"],
    artist_id: "artist-001",
  },
  {
    artwork_id: "artwork-003",
    artwork_name: "Pixel Art 01",
    image_url: null,
    price_satang: 50_000,
    category: "Game",
    styles: ["Pixel Art", "Cartoon"],
    artist_id: "artist-002",
  },
  {
    artwork_id: "artwork-004",
    artwork_name: "Pixel Art 02",
    image_url: null,
    price_satang: 75_000,
    category: "Animation",
    styles: ["Pixel Art"],
    artist_id: "artist-002",
  },
  {
    artwork_id: "artwork-005",
    artwork_name: "Pixel Village",
    image_url: null,
    price_satang: 120_000,
    category: "Game",
    styles: ["Pixel Art", "Graphic"],
    artist_id: "artist-002",
  },
  {
    artwork_id: "artwork-006",
    artwork_name: "Sakura Panel",
    image_url: null,
    price_satang: 90_000,
    category: "Comic",
    styles: ["Anime", "Digital art"],
    artist_id: "artist-003",
  },
  {
    artwork_id: "artwork-007",
    artwork_name: "Mecha Cover",
    image_url: null,
    price_satang: 210_000,
    category: "Illustration",
    styles: ["Anime"],
    artist_id: "artist-003",
  },
  {
    artwork_id: "artwork-008",
    artwork_name: "Forest Cover",
    image_url: null,
    price_satang: 180_000,
    category: "Book",
    styles: ["Water Color", "Cartoon"],
    artist_id: "artist-004",
  },
  {
    artwork_id: "artwork-009",
    artwork_name: "Dungeon Tileset",
    image_url: null,
    price_satang: 30_000,
    category: "Game",
    styles: ["Pixel Art", "Graphic"],
    artist_id: "artist-005",
  },
  {
    artwork_id: "artwork-010",
    artwork_name: "Boss Sprite Sheet",
    image_url: null,
    price_satang: 95_000,
    category: "Game",
    styles: ["Pixel Art"],
    artist_id: "artist-005",
  },
  {
    artwork_id: "artwork-011",
    artwork_name: "Oil Portrait",
    image_url: null,
    price_satang: 450_000,
    category: "Portrait",
    styles: ["Realism", "Digital art"],
    artist_id: "artist-006",
  },
  {
    artwork_id: "artwork-012",
    artwork_name: "Character Turnaround",
    image_url: null,
    price_satang: 160_000,
    category: "Animation",
    styles: ["Anime", "Cartoon"],
    artist_id: "artist-007",
  },
  {
    artwork_id: "artwork-013",
    artwork_name: "Idle Cycle",
    image_url: null,
    price_satang: 110_000,
    category: "Game",
    styles: ["Cartoon"],
    artist_id: "artist-007",
  },
  {
    artwork_id: "artwork-014",
    artwork_name: "Ink Splash",
    image_url: null,
    price_satang: 40_000,
    category: "Comic",
    styles: ["Graphic", "Cartoon"],
    artist_id: "artist-008",
  },
  {
    artwork_id: "artwork-015",
    artwork_name: "Botanical Spread",
    image_url: null,
    price_satang: 520_000,
    category: "Book",
    styles: ["Water Color", "Realism"],
    artist_id: "artist-009",
  },
  {
    artwork_id: "artwork-016",
    artwork_name: "Wildflower Study",
    image_url: null,
    price_satang: 310_000,
    category: "Illustration",
    styles: ["Water Color"],
    artist_id: "artist-009",
  },
  {
    artwork_id: "artwork-017",
    artwork_name: "8-bit Hero",
    image_url: null,
    price_satang: 25_000,
    category: "Game",
    styles: ["Pixel Art"],
    artist_id: "artist-010",
  },
  {
    artwork_id: "artwork-018",
    artwork_name: "Soft Glow",
    image_url: null,
    price_satang: 85_000,
    category: "Portrait",
    styles: ["Digital art", "Anime"],
    artist_id: "artist-011",
  },
  {
    artwork_id: "artwork-019",
    artwork_name: "Poster Series A",
    image_url: null,
    price_satang: 220_000,
    category: "Illustration",
    styles: ["Graphic", "Digital art"],
    artist_id: "artist-012",
  },
  {
    artwork_id: "artwork-020",
    artwork_name: "Poster Series B",
    image_url: null,
    price_satang: 240_000,
    category: "Illustration",
    styles: ["Graphic"],
    artist_id: "artist-012",
  },
];

const ARTISTS_BY_ID = new Map(
  MOCK_ARTISTS.map((artist) => [artist.artist_id, artist]),
);

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function randomDelayMs(): number {
  const { min, max } = MOCK_DELAY_MS;
  return min + Math.floor(Math.random() * (max - min + 1));
}

function normalize(value: string): string {
  return value.trim().toLowerCase();
}

function toSearchItem(artwork: MockArtwork): ArtistSearchItem | null {
  const artist = ARTISTS_BY_ID.get(artwork.artist_id);
  if (!artist) return null;

  return {
    artwork_id: artwork.artwork_id,
    artwork_name: artwork.artwork_name,
    image_url: artwork.image_url,
    price_satang: artwork.price_satang,
    category: artwork.category,
    styles: artwork.styles,
    artist_id: artist.artist_id,
    artist_name: artist.artist_name,
    review_score: artist.review_score,
  };
}

function matchesArtistName(item: ArtistSearchItem, search?: string): boolean {
  if (!search?.trim()) return true;
  return normalize(item.artist_name).includes(normalize(search));
}

function matchesPrice(
  item: ArtistSearchItem,
  minPrice?: number,
  maxPrice?: number,
): boolean {
  if (minPrice !== undefined && item.price_satang < minPrice) return false;
  if (maxPrice !== undefined && item.price_satang > maxPrice) return false;
  return true;
}

function matchesCategory(item: ArtistSearchItem, category?: string): boolean {
  if (!category?.trim()) return true;
  return normalize(item.category) === normalize(category);
}

function matchesStyles(item: ArtistSearchItem, styles?: string[]): boolean {
  if (!styles || styles.length === 0) return true;

  const wanted = new Set(styles.map(normalize));
  return item.styles.some((style) => wanted.has(normalize(style)));
}

function matchesMinRating(item: ArtistSearchItem, minRating?: number): boolean {
  if (minRating === undefined) return true;
  if (item.review_score === null) return false;
  return item.review_score >= minRating;
}

function compareItems(
  a: ArtistSearchItem,
  b: ArtistSearchItem,
  sort: ArtistSearchSort,
): number {
  switch (sort) {
    case "price_asc":
      return a.price_satang - b.price_satang;
    case "price_desc":
      return b.price_satang - a.price_satang;
    case "review_score_asc": {
      const scoreA = a.review_score ?? -1;
      const scoreB = b.review_score ?? -1;
      return scoreA - scoreB;
    }
    case "review_score_desc": {
      const scoreA = a.review_score ?? -1;
      const scoreB = b.review_score ?? -1;
      return scoreB - scoreA;
    }
    default:
      return 0;
  }
}

/**
 * In-memory artist-name search that returns artworks.
 * Simulates network latency until the real endpoint lands.
 */
export async function searchArtistsMock(
  params: ArtistSearchParams = {},
): Promise<ArtistSearchResult> {
  await delay(randomDelayMs());

  const sort = params.sort ?? "price_asc";
  const limit = params.limit ?? 20;
  const offset = params.offset ?? 0;

  const filtered = MOCK_ARTWORKS.map(toSearchItem)
    .filter((item): item is ArtistSearchItem => item !== null)
    .filter(
      (item) =>
        matchesArtistName(item, params.search) &&
        matchesPrice(item, params.min_price_satang, params.max_price_satang) &&
        matchesCategory(item, params.category) &&
        matchesStyles(item, params.styles) &&
        matchesMinRating(item, params.min_rating),
    )
    .sort((a, b) => compareItems(a, b, sort));

  return {
    artworks: filtered.slice(offset, offset + limit),
    total: filtered.length,
  };
}
