import { Button } from "@/components/ui/Button";

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
            <div className="flex gap-4">
                <div className="flex items-center gap-1">
                    <img src="/icons/alarm-clock.svg" alt="timer icon" className="w-6 h-6" />
                    <p className="text-h3">{minimumDeadlineDays} days</p>
                </div>
                <div className="flex items-center gap-1">
                    <img src="/icons/banknote.svg" alt="price icon" className="w-6 h-6" />
                    <p className="text-h3">{(priceSatang / 100).toLocaleString()} THB</p>
                </div>
            </div>
            {/* Additional order details can be added here */}
            {/* link to order placement page */}
            <a href={`/artworks/${artworkId}/order`} className="w-full">
              <Button variant="accent-500" className="text-h3 shadow-md w-full">
                  Request to Order
              </Button>
            </a>
        </div>
    </div>
  );
}