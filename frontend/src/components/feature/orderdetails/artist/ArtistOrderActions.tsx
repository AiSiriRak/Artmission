import { Button } from "@/components/ui/Button";
import type { OrderStatus } from "@/lib/api/types";

interface ArtistOrderActionsProps {
  status: OrderStatus;
}

export default function ArtistOrderActions({ status }: ArtistOrderActionsProps) {
  // TODO: connect these actions with the API when ready.
  switch (status) {
    case "PENDING":
      return (
        <>
          <Button type="button" className="!border w-32">Reject</Button>
          <Button type="button" variant="dark" className="w-32">Accept</Button>
        </>
      );
    case "NOT_PAID":
      return <Button type="button" variant="red" className="min-w-32">Cancel Order</Button>;
    case "IN_PROCESS":
      return (
        <>
          <Button type="button" variant="transparent" className="min-w-32">Postpone Order</Button>
          <Button type="button" variant="red" className="min-w-32">Cancel Order</Button>
        </>
      );
    case "SUCCESS":
    case "CANCEL":
      return null;
  }
}
