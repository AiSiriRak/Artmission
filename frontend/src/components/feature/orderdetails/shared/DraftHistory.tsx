import type { ReactNode } from "react";
import { useState } from "react";
import { WhiteCard } from "@/components/ui/WhiteCard";
import { Button } from "@/components/ui/Button";

export type DraftStatus =
  | "WAITING_FOR_CUSTOMER_COMMENT"
  | "CHANGE_REQUESTED"
  | "COMMISSION_ACCEPTED";

export interface Draft {
  /** Backend id. Used as the list key. */
  id: string;
  draftNumber: number;
  status: DraftStatus;
  submitDate: string;
  commentDate?: string;
  imageUrl: string;
  imageAlt?: string;
  customerComment?: string;
}

interface DraftHistoryProps {
  drafts: Draft[];
  usertype?: "artist" | "customer";
  /** Customer only: false once the revision limit is used up. */
  canRequestChanges?: boolean;
  onRequestChanges?: (draft: Draft, comment: string) => void;
  onAccept?: (draft: Draft) => void;
}

const statusView: Record<DraftStatus, { label: string; className: string }> = {
  WAITING_FOR_CUSTOMER_COMMENT: {
    label: "Waiting for Customer Review",
    className: "bg-status-notpaid-pale",
  },
  CHANGE_REQUESTED: {
    label: "Change Requested",
    className: "bg-status-pending-pale",
  },
  COMMISSION_ACCEPTED: {
    label: "Commission Accepted",
    className: "bg-status-success-pale",
  },
};

function HistoryCard({ children }: { children: ReactNode }) {
  return (
    <WhiteCard
      margin=""
      padding="p-6 sm:p-8"
      roundsize="rounded-3xl"
      className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
    >
      {children}
    </WhiteCard>
  );
}

function DateLine({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-wrap items-baseline gap-x-4">
      <dt className="text-body font-bold">{label}</dt>
      <dd className="text-body">{value}</dd>
    </div>
  );
}

function ArtworkPreview({ src, alt }: { src: string; alt: string }) {
  return (
    <div className="mx-auto mt-8 w-full max-w-56 border border-primary-500 bg-white p-3">
      <img
        src={src}
        alt={alt}
        className="aspect-square w-full object-contain"
      />
    </div>
  );
}

function DraftStatusBadge({ status }: { status: DraftStatus }) {
  const view = statusView[status];

  return (
    <span
      className={`inline-flex shrink-0 items-center whitespace-nowrap rounded-full px-4 py-1 text-caption ${view.className}`}
    >
      {view.label}
    </span>
  );
}

export default function DraftHistory({
  drafts,
  usertype,
  canRequestChanges = true,
  onRequestChanges,
  onAccept,
}: DraftHistoryProps) {
  const approvedDraft = drafts.find(
    (draft) => draft.status === "COMMISSION_ACCEPTED",
  );

  const [isCommenting, setIsCommenting] = useState(false);
  const [comment, setComment] = useState("");
  const [error, setError] = useState<string | null>(null);

  return (
    <div className="mt-10 min-w-0">
      {approvedDraft && (
        <section className="mb-10">
          <h2 className="mb-4 text-h3">Final Artwork</h2>
          <HistoryCard>
            <dl>
              <DateLine label="Submit Date" value={approvedDraft.submitDate} />
            </dl>
            <ArtworkPreview
              src={approvedDraft.imageUrl}
              alt={approvedDraft.imageAlt ?? "Final artwork"}
            />
          </HistoryCard>
        </section>
      )}

      <section>
        <h2 className="mb-8 rounded-r-md border-l-4 border-accent-500 bg-secondary-300 px-4 py-2.5 text-h3 font-bold">
          Draft History
        </h2>

        <div className="space-y-10">
          {drafts.map((draft) => {
            const commentDate = draft.commentDate?.trim();
            const customerComment = draft.customerComment?.trim();

            return (
              <article key={draft.id} className="min-w-0">
                <h3 className="mb-4 text-h3">#{draft.draftNumber} Draft</h3>
                <HistoryCard>
                  <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-3">
                    <dl className="space-y-1">
                      <DateLine label="Submit Date" value={draft.submitDate} />
                      {commentDate && (
                        <DateLine label="Comment Date" value={commentDate} />
                      )}
                    </dl>
                    <DraftStatusBadge status={draft.status} />
                  </div>
                  <ArtworkPreview
                    src={draft.imageUrl}
                    alt={draft.imageAlt ?? `Draft ${draft.draftNumber} artwork`}
                  />

                  {/* Display customer no comment yet draft.status = "WAITING_FOR_CUSTOMER_COMMENT" */}
                  {usertype === "customer" &&
                    draft.status === "WAITING_FOR_CUSTOMER_COMMENT" &&
                    (!isCommenting ? (
                      // Display action buttons for requesting changes or accepting the commission
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
                          <Button
                            type="button"
                            variant="dark"
                            onClick={() => onAccept?.(draft)}
                          >
                            Accept Commission
                          </Button>
                        </div>
                        <div className="flex items-center gap-2 mb-4">
                          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-red-500 text-subtle font-bold text-white">
                            i
                          </span>
                          <div className="text-center text-small text-muted">
                            Please comment within 3 days. Artwork will be
                            automatically accepted on 05 Sep 2026 if no
                            response.
                          </div>
                        </div>
                      </div>
                    ) : (
                      // Display the comment form for the customer to submit their comment
                      <div className="mt-8 flex w-full flex-col">
                        <p className="mb-2 text-body font-bold">
                          Customer Comment
                        </p>

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

                          {error && (
                            <p className="mt-1 text-sm text-red-500">{error}</p>
                          )}

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
                                onRequestChanges?.(draft, comment.trim());
                                setIsCommenting(false);
                                setComment("");
                              }}
                            >
                              Submit
                            </Button>
                          </div>
                        </div>
                      </div>
                    ))}

                  {customerComment && (
                    <div className="mt-8">
                      <p className="mb-2 text-body font-bold">
                        Customer Comment
                      </p>
                      <p className="whitespace-pre-wrap break-words text-small">
                        {customerComment}
                      </p>
                    </div>
                  )}
                </HistoryCard>
              </article>
            );
          })}
        </div>
      </section>
    </div>
  );
}
