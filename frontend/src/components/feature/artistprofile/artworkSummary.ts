import type { Artwork } from "@/lib/api/types";

export function deriveArtworkSummary(artworks: Artwork[]) {
  const categories = [
    ...new Set(artworks.map((art) => art.category).filter(Boolean)),
  ];
  const styles = [
    ...new Set(artworks.flatMap((art) => art.styles || [])),
  ].filter(Boolean);
  const prices = artworks
    .map((art) => Number(art.price_satang ? art.price_satang / 100 : 0))
    .filter((price) => !isNaN(price) && price > 0);
  const minPrice = prices.length > 0 ? Math.min(...prices) : 0;
  const maxPrice = prices.length > 0 ? Math.max(...prices) : 0;
  const priceRangeText =
    prices.length === 0
      ? "N/A"
      : minPrice === maxPrice
        ? `${minPrice.toLocaleString()} THB`
        : `${minPrice.toLocaleString()} - ${maxPrice.toLocaleString()} THB`;

  return { categories, styles, priceRangeText };
}
