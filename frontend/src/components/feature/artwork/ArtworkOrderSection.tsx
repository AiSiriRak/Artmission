import { Button } from "@/components/ui/Button";

const mockArtworkData = {
  minimum_deadline_days: 7,
  price: 1000,
};

export default function ArtworkOrderSection() {
  return (
    <div className="w-160 border border-primary-500 rounded-3xl p-10 bg-secondary-200 shadow-sm">
        <div className="flex flex-col gap-6">
            <div className="flex gap-4">
                <div className="flex items-center gap-1">
                    <img src="/icons/alarm-clock.svg" alt="timer icon" className="w-6 h-6" />
                    <p className="text-h3">{mockArtworkData.minimum_deadline_days} days</p>
                </div>
                <div className="flex items-center gap-1">
                    <img src="/icons/banknote.svg" alt="price icon" className="w-6 h-6" />
                    <p className="text-h3">{mockArtworkData.price.toLocaleString()} THB</p>
                </div>
            </div>
            <Button variant="accent-500" className="text-h3 shadow-md w-full">
                Request to Order
            </Button>
        </div>
    </div>
  );
}