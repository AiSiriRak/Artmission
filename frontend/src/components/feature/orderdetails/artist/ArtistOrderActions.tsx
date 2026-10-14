import { Button } from "@/components/ui/Button";
import type { OrderStatus } from "@/lib/api/types";

interface ArtistOrderActionsProps {
  status: OrderStatus;
  onReject: () => void;
  onAccept: () => void;
  onCancel: () => void;
  onPostpone: () => void;
}

export default function ArtistOrderActions({ status, onReject, onAccept, onCancel, onPostpone }: ArtistOrderActionsProps) {
  // TODO: connect these actions with the API when ready.
  switch (status) {
    case "PENDING":
      return (
        <>
          <Button type="button" className="!border w-32" onClick={onReject}>Reject</Button>
          <Button type="button" variant="dark" className="w-32" onClick={onAccept}>Accept</Button>
        </>
      );
    case "NOT_PAID":
      return <Button type="button" variant="red" className="min-w-32" onClick={onCancel}>Cancel Order</Button>;
    case "IN_PROCESS":
      return (
        <>
          <Button type="button" variant="transparent" className="min-w-32" onClick={onPostpone}>Postpone Order</Button>
          <Button type="button" variant="red" className="min-w-32" onClick={onCancel}>Cancel Order</Button>
        </>
      );
    case "SUCCESS":
    case "CANCEL":
      return null;
  }
}
