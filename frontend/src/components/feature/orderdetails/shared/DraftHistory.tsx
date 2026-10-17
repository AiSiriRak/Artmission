import type { ReactNode } from "react";
import Image from "next/image";
import { WhiteCard } from "@/components/ui/WhiteCard";

export type DraftStatus =
  | "WAITING_FOR_CUSTOMER_REVIEW"
  | "CHANGE_REQUESTED"
  | "COMMISSION_ACCEPTED";

export interface Draft {
  /** Backend id. Used as the list key. */
  id: string;
  draftNumber: number;
  status: DraftStatus;
  submitDate: string;
  reviewDate?: string;
  imageUrl: string;
  imageAlt?: string;
  customerReview?: string;
  reviewDeadline?: string;
}

interface DraftHistoryProps {
  drafts: Draft[];
  renderDraftActions?: (draft: Draft) => ReactNode;
  renderFinalArtworkActions?: (draft: Draft) => ReactNode;
}

const statusView: Record<DraftStatus, { label: string; className: string }> = {
  WAITING_FOR_CUSTOMER_REVIEW: {
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
      <dt className="text-h3">{label}</dt>
      <dd className="text-body">{value}</dd>
    </div>
  );
}

function ArtworkPreview({ src, alt }: { src: string; alt: string }) {
  return (
    <div className="mx-auto mt-8 w-full max-w-[940px] p-6">
      <div className="mx-auto w-fit max-w-full border border-neutral bg-white">
        <img src={src} alt={alt} className="block h-auto w-auto max-h-[632px] max-w-full object-contain" />
      </div>
    </div>
  );
}

function DraftStatusBadge({ status }: { status: DraftStatus }) {
  const view = statusView[status];

  return (
    <div
      className={`${view.className} ml-auto flex w-[252px] max-w-full shrink-0 justify-center rounded-full border border-primary-500 p-1.5`}
    >
      <p className="text-center text-caption text-primary-500">{view.label}</p>
    </div>
  );
}

export default function DraftHistory({ drafts, renderDraftActions, renderFinalArtworkActions }: DraftHistoryProps) {
  const approvedDraft = drafts.find(
    (draft) => draft.status === "COMMISSION_ACCEPTED",
  );

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
            {renderFinalArtworkActions?.(approvedDraft)}
          </HistoryCard>
        </section>
      )}

      <section>
        <h2 className="mb-8 rounded-r-md border-l-4 border-accent-500 bg-secondary-300 px-4 py-2.5 text-h3">
          Draft History
        </h2>

        <div className="space-y-10">
          {drafts.map((draft) => {
            const reviewDate = draft.reviewDate?.trim();
            const customerReview = draft.customerReview?.trim();

            return (
              <article key={draft.id} className="min-w-0">
                <h3 className="mb-4 text-h3">#{draft.draftNumber} Draft</h3>
                <HistoryCard>
                  <div className="flex flex-wrap items-start justify-between gap-x-4 gap-y-3">
                    <dl className="space-y-1">
                      <DateLine label="Submit Date" value={draft.submitDate} />
                      {reviewDate && (
                        <DateLine label="Review Date" value={reviewDate} />
                      )}
                    </dl>
                    <DraftStatusBadge status={draft.status} />
                  </div>
                  <ArtworkPreview
                    src={draft.imageUrl}
                    alt={
                      draft.imageAlt ?? `Draft ${draft.draftNumber} artwork`
                    }
                  />
                  {customerReview && (
                    <div className="mt-8">
                      <p className="mb-2 text-h3">Customer Review</p>
                      <p className="whitespace-pre-wrap break-words text-caption">
                        {customerReview}
                      </p>
                    </div>
                  )}
                  {draft.status === "WAITING_FOR_CUSTOMER_REVIEW" && draft.reviewDeadline && (
                    <p className="mt-8 flex items-start justify-center gap-2 text-small">
                      <Image src="/icons/clock.svg" alt="" width={16} height={16} className="shrink-0" />
                      The customer has review until {draft.reviewDeadline}. If no response, the artwork will be automatically accepted.
                    </p>
                  )}
                  {renderDraftActions?.(draft)}
                </HistoryCard>
              </article>
            );
          })}
        </div>
      </section>
    </div>
  );
}
