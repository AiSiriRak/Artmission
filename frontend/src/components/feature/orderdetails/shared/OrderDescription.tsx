import { WhiteCard } from "@/components/ui/WhiteCard";
import TagList from "@/components/feature/artistprofile/TagList";

interface OrderDescriptionProps {
  orderNumber: string;
  artworkName: string;
  categoryName: string;
  categories: string[];
  styles: string[];
  requestDetail: string;
}

function formatTagLabel(value: string) {
  const label = value.trim().toLowerCase();
  return label.charAt(0).toUpperCase() + label.slice(1);
}

export default function OrderDescription({
  orderNumber,
  artworkName,
  categoryName,
  categories,
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
        <h3 className="mb-4 break-words text-h1">{artworkName}</h3>
        <div className="mb-8 flex flex-wrap items-center gap-3">
          {categoryName && <span className="text-h3">{categoryName}</span>}
          <TagList items={categories.map(formatTagLabel)} variant="category" />
          <TagList items={styles.map(formatTagLabel)} variant="style" />
        </div>
        <h4 className="mb-3 text-h3">Request Detail</h4>
        <p className="whitespace-pre-wrap break-words text-caption">{requestDetail}</p>
      </WhiteCard>
    </>
  );
}
