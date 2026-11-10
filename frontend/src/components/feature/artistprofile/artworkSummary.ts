import type { Artwork } from "@/lib/api/types";
import { formatPriceFromSatang } from "@/lib/format-price";

export function deriveArtworkSummary(artworks: Artwork[]) {
  const categories = [
    ...new Set(artworks.map((art) => art.category).filter(Boolean)),
  ];
  const styles = [
    ...new Set(artworks.flatMap((art) => art.styles || [])),
  ].filter(Boolean);
  const prices = artworks
    .map((art) => art.price_satang)
    .filter((price) => Number.isFinite(price) && price > 0);
  const minPrice = prices.length > 0 ? Math.min(...prices) : 0;
  const maxPrice = prices.length > 0 ? Math.max(...prices) : 0;
  const priceRangeText =
    prices.length === 0
      ? "N/A"
      : minPrice === maxPrice
        ? `${formatPriceFromSatang(minPrice)} THB`
        : `${formatPriceFromSatang(minPrice)} - ${formatPriceFromSatang(maxPrice)} THB`;

  return { categories, styles, priceRangeText };
}
