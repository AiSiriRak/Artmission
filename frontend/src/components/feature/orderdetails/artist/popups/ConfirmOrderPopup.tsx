import { Button } from "@/components/ui/Button";
import ArtistOrderPopup from "./ArtistOrderPopup";

export type ConfirmOrderAction = "reject" | "accept" | "submit";

const content = {
  reject: {
    title: "Reject this Order?",
    description: "Are you sure you want to decline this request? The customer will be notified and this order will be cancelled.",
    button: "Reject",
  },
  accept: {
    title: "Accept this Order?",
    description: "By accepting this request, the customer will be notified to proceed with payment to start the commission.",
    button: "Accept",
  },
  submit: {
    title: "Submit Draft to Customer?",
    description: "The customer will be notified and given 3 days to review and approve the draft.",
    button: "Submit",
  },
};

interface ConfirmOrderPopupProps {
  action: ConfirmOrderAction;
  onClose: () => void;
  onConfirm: () => void;
  busy?: boolean;
  error?: string;
}

export default function ConfirmOrderPopup({ action, onClose, onConfirm, busy = false, error }: ConfirmOrderPopupProps) {
  const view = content[action];
  return (
    <ArtistOrderPopup title={view.title} onClose={onClose}>
      <p className="mx-auto my-6 max-w-[256px] text-center text-small">{view.description}</p>
      {error && <p role="alert" className="mb-4 text-center text-small text-error">{error}</p>}
      <div className="flex justify-center gap-4">
        <Button type="button" variant="transparent" className="w-[138px] rounded-xl" disabled={busy} onClick={onClose}>Cancel</Button>
        <Button type="button" variant="dark" className="w-[138px] rounded-xl" disabled={busy} onClick={onConfirm}>{busy ? "Please wait..." : view.button}</Button>
      </div>
    </ArtistOrderPopup>
  );
}
