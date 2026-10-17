import { useState } from "react";
import { Button } from "@/components/ui/Button";
import type { Draft } from "@/components/feature/orderdetails/shared/DraftHistory";

interface CustomerDraftActionsProps {
  draft: Draft;
  /** false once the revision limit is used up. */
  canRequestChanges: boolean;
  onRequestChanges: (draft: Draft, comment: string) => void;
  onAccept: (draft: Draft) => void;
}

export default function CustomerDraftActions({
  draft,
  canRequestChanges,
  onRequestChanges,
  onAccept,
}: CustomerDraftActionsProps) {
  const [isCommenting, setIsCommenting] = useState(false);
  const [comment, setComment] = useState("");
  const [error, setError] = useState<string | null>(null);

  if (draft.status !== "WAITING_FOR_CUSTOMER_COMMENT") return null;

  if (!isCommenting) {
    return (
      <div className="mt-8 items-center flex flex-col">
        <div className="flex gap-4 mb-4">
          {canRequestChanges && (
            <Button
              type="button"
              variant="light"
              onClick={() => setIsCommenting(true)}
            >
              Request Changes
            </Button>
          )}
          <Button type="button" variant="dark" onClick={() => onAccept(draft)}>
            Accept Commission
          </Button>
        </div>
        <div className="flex items-center gap-2 mb-4">
          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-red-500 text-subtle font-bold text-white">
            i
          </span>
          <div className="text-center text-small text-muted">
            Please comment within 3 days. The artwork will be automatically
            accepted if there is no response.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="mt-8 flex w-full flex-col">
      <p className="mb-2 text-body font-bold">Customer Comment</p>
      <div className="mb-4 w-full">
        <textarea
          value={comment}
          onChange={(e) => {
            setComment(e.target.value);
            setError(null);
          }}
          className="w-full rounded-md border border-gray-300 bg-white p-2"
          placeholder="Describe your comment"
          rows={4}
        />
        {error && <p className="mt-1 text-sm text-red-500">{error}</p>}
        <div className="mt-4 flex items-center justify-center gap-4">
          <Button
            type="button"
            variant="light"
            onClick={() => {
              setComment("");
              setError(null);
              setIsCommenting(false);
            }}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="dark"
            onClick={() => {
              if (!comment.trim()) {
                setError("Comment cannot be empty");
                return;
              }
              onRequestChanges(draft, comment.trim());
              setIsCommenting(false);
              setComment("");
            }}
          >
            Submit
          </Button>
        </div>
      </div>
    </div>
  );
}
