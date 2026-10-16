import { WhiteCard } from "@/components/ui/WhiteCard";
import TagList from "@/components/feature/artistprofile/TagList";

interface OrderDescriptionProps {
  orderNumber: string;
  orderName: string;
  artworkName: string;
  category: string;
  styles: string[];
  requestDetail: string;
}

function formatTagLabel(value: string) {
  const label = value.trim().toLowerCase();
  return label.charAt(0).toUpperCase() + label.slice(1);
}

export default function OrderDescription({
  orderNumber,
  orderName,
  artworkName,
  category,
  styles,
  requestDetail,
}: OrderDescriptionProps) {
  return (
    <>
      <h2 className="mb-4 text-h3">Order #{orderNumber}</h2>
      <WhiteCard
        margin=""
        padding="p-6 sm:p-9"
        roundsize="rounded-3xl"
        className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
      >
        <h3 className="mb-4 break-words text-h1">{orderName}</h3>
        <div className="mb-8 flex flex-wrap items-center gap-3">
          <span className="text-h3">{artworkName}</span>
          <TagList items={[category]} variant="category" />
          <TagList items={styles} variant="style" />
        </div>
        <h4 className="mb-3 text-h3">Request Detail</h4>
        <p className="whitespace-pre-wrap break-words text-caption">{requestDetail}</p>
      </WhiteCard>
    </>
  );
}
