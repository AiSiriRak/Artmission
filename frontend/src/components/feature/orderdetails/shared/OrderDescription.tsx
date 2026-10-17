import { WhiteCard } from "@/components/ui/WhiteCard";
import TagList from "@/components/feature/artistprofile/TagList";

interface OrderDescriptionProps {
  orderNumber: string;
  orderName: string;
  artworkName: string;
  categoryName: string;
  categories: string[];
  styles: string[];
  requestDetail: string;
}

export default function OrderDescription({
  orderNumber,
  orderName,
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
        <h3 className="mb-4 break-words text-h1">{orderName}</h3>
        <div className="mb-8 flex flex-wrap items-center gap-3">
          <span className="break-words text-h3">{artworkName}</span>
          {categoryName && <span className="text-h3">{categoryName}</span>}
          <TagList items={categories} variant="category" />
          <TagList items={styles} variant="style" />
        </div>
        <h4 className="mb-3 text-h3">Request Detail</h4>
        <p className="whitespace-pre-wrap break-words text-caption">{requestDetail}</p>
      </WhiteCard>
    </>
  );
}
