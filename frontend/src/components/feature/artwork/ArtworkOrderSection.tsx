import Link from "next/link";
import Image from "next/image";
import { formatPriceFromSatang } from "@/lib/format-price";

interface ArtworkOrderSectionProps {
  minimumDeadlineDays: number;
  priceSatang: number;
  artworkId: string;
}

export default function ArtworkOrderSection({
  minimumDeadlineDays,
  priceSatang,
  artworkId,
}: ArtworkOrderSectionProps) {
  return (
    <div className="w-100 border border-primary-500 rounded-3xl p-10 bg-secondary-200 shadow-sm">
      <div className="flex flex-col gap-6">
        <div className="flex gap-6">
          <div className="flex items-center gap-2">
            <Image src="/icons/alarm-clock.svg" alt="timer icon" width={24} height={24} className="w-6 h-6" />
            <p className="text-h4">{minimumDeadlineDays} days</p>
          </div>
          <div className="flex items-center gap-2">
            <Image src="/icons/banknote.svg" alt="price icon" width={24} height={24} className="w-6 h-6" />
            <p className="text-h4">{formatPriceFromSatang(priceSatang)} THB</p>
          </div>
        </div>
        {/* Additional order details can be added here */}
        {/* link to order placement page */}
        <Link
          href={`/artworks/${encodeURIComponent(artworkId)}/order`}
          className="flex w-full items-center justify-center gap-2 rounded-lg bg-accent-500 px-5 py-2 text-h3 text-white shadow-md transition-colors hover:brightness-90"
        >
          Request to Order
        </Link>
      </div>
    </div>
  );
}