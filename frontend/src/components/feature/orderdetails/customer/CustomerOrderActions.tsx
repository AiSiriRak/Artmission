import { Button } from "@/components/ui/Button";
import type { OrderStatus } from "@/lib/api/types";

interface CustomerOrderActionsProps {
  status: OrderStatus;
  onCancelOrder: () => void;
}

function CancelOrderButton({ onClick }: { onClick: () => void }) {
  return (
    <Button type="button" variant="red" onClick={onClick}>
      Cancel Order
    </Button>
  );
}

function GoToPaymentButton() {
  // The backend has no payment endpoint yet, so this stays disabled.
  return (
    <Button type="button" variant="light" disabled title="Payment is not available yet">
      Go to payment
    </Button>
  );
}

export default function CustomerOrderActions({
  status,
  onCancelOrder,
}: CustomerOrderActionsProps) {
  switch (status) {
    case "PENDING":
    case "IN_PROCESS":
      return <CancelOrderButton onClick={onCancelOrder} />;
    case "NOT_PAID":
      return (
        <>
          <GoToPaymentButton />
          <CancelOrderButton onClick={onCancelOrder} />
        </>
      );
    case "SUCCESS":
    case "CANCEL":
      return null;
  }
}
