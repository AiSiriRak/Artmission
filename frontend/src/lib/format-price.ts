export function formatPriceFromSatang(priceSatang: number): string {
  const price = priceSatang / 100;
  return Number.isInteger(price)
    ? price.toLocaleString("en-US")
    : price.toLocaleString("en-US", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      });
}
