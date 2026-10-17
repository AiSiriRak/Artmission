"use client";

import { useEffect, useState } from "react";
import OrderDescription from "@/components/feature/orderdetails/shared/OrderDescription";
import OrderDetails from "@/components/feature/orderdetails/shared/OrderDetails";
import OrderInfo from "@/components/feature/orderdetails/shared/OrderInfo";
import DraftHistory, {
  type Draft,
  type DraftStatus,
} from "@/components/feature/orderdetails/shared/DraftHistory";
import {
  confirmDeliverable,
  getOrderById,
  type OrderDeliverable,
  type OrderDetail,
} from "@/lib/api/orders";
import { CustomerCancelOrderPopUp } from "./CustomerCancelOrderPopUp";
import { CustomerCommentConfirmPopUp } from "./CustomerCommentConfirmPopUp";
import { RatingOrderPopup } from "./RatingOrderPopUp";
import { ThankYouOrderPopup } from "./ThankYouOrderPopUp";
import CustomerDraftActions from "./CustomerDraftActions";
import DownloadArtworkButton from "./DownloadArtworkButton";
import CustomerOrderActions from "./CustomerOrderActions";
import { getArtistProfile } from "@/lib/api/artists";
import type { ArtistProfile } from "@/lib/api/types";
import { ArtistMiniProfile } from "./ArtistMiniProfile";
import CustomerPostponeSection from "./CustomerPostponeSection";

const DAY_MS = 24 * 60 * 60 * 1000;

const draftStatusByDecision: Record<OrderDeliverable["decision"], DraftStatus> = {
  WAIT: "WAITING_FOR_CUSTOMER_COMMENT",
  REJECTED: "CHANGE_REQUESTED",
  APPROVED: "COMMISSION_ACCEPTED",
};

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString("en-GB", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

function getDeadlineNote(deadlineAt: string, status: OrderDetail["status"]) {
  if (status === "SUCCESS" || status === "CANCEL") return undefined;
  const daysLeft = Math.ceil((new Date(deadlineAt).getTime() - Date.now()) / DAY_MS);
  if (daysLeft < 0) return `${Math.abs(daysLeft)} days overdue`;
  return `${daysLeft} days left`;
}

function toDrafts(deliverables: OrderDeliverable[] | null): Draft[] {
  return [...(deliverables ?? [])]
    .map((item) => ({
      id: item.id,
      draftNumber: item.version,
      status: draftStatusByDecision[item.decision],
      submitDate: formatDate(item.created_at),
      commentDate: item.decision === "WAIT" ? undefined : formatDate(item.updated_at),
      previewImageUrl: item.preview_image_url,
      customerComment: item.comment ?? undefined,
    }));
}

const REVISION_LIMIT = 2;

type PendingAction =
  | { kind: "request-changes"; draft: Draft; comment: string }
  | { kind: "accept"; draft: Draft };

export default function CustomerOrderDetail({ order: initialOrder }: { order: OrderDetail }) {
  const [order, setOrder] = useState(initialOrder);
  const [drafts, setDrafts] = useState<Draft[]>(() => toDrafts(initialOrder.deliverables));
  const [pending, setPending] = useState<PendingAction | null>(null);
  const [cancelOpen, setCancelOpen] = useState(false);
  const [reviewStep, setReviewStep] = useState<"none" | "rating" | "thanks">(() =>
    initialOrder.status === "SUCCESS" &&
    (initialOrder.deliverables ?? []).some((d) => d.decision === "APPROVED") &&
    !localStorage.getItem(`order-rating-shown:${initialOrder.id}`)
      ? "rating"
      : "none",
  );
  const [popupStatus, setPopupStatus] = useState<"default" | "success" | "fail">("default");
  const [popupError, setPopupError] = useState<string | null>(null);

  const [artistProfile, setArtistProfile] = useState<ArtistProfile | null>(null);

  useEffect(() => {
    let cancelled = false;
    getArtistProfile(initialOrder.artist_id)
      .then((profile) => {
        if (!cancelled) setArtistProfile(profile);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [initialOrder.artist_id]);

  const revisionsUsed = drafts.filter((d) => d.status === "CHANGE_REQUESTED").length;
  const closePopup = () => {
    setPending(null);
    setPopupStatus("default");
    setPopupError(null);
  };

  const handleConfirm = async () => {
    if (!pending) return;

    const { draft } = pending;
    try {
      await confirmDeliverable(order.id, draft.id, {
        Accept: pending.kind === "accept",
        comment: pending.kind === "request-changes" ? pending.comment : undefined,
      });
      const fresh = await getOrderById(order.id);
      setOrder(fresh);
      setDrafts(toDrafts(fresh.deliverables));
      if (pending.kind === "accept") {
        closePopup();
        setReviewStep("rating");
      } else {
        setPopupStatus("success");
      }
    } catch (err) {
      setPopupError(err instanceof Error ? err.message : null);
      setPopupStatus("fail");
    }
  };

  // The rating popup is shown once for an order that is already completed and
  // not yet rated (e.g. the final draft was auto-accepted). The backend has no
  // review endpoint, so "already shown" is remembered per order in localStorage.
  const ratingKey = `order-rating-shown:${order.id}`;

  useEffect(() => {
    if (reviewStep === "rating") localStorage.setItem(ratingKey, "1");
  }, [reviewStep, ratingKey]);

  return (
    <>
    
      <OrderDetails
        usertype="customer"
        postponeSection={
          order.active_delay_request && order.status !== "CANCEL" ? (
            <CustomerPostponeSection
              orderId={order.id}
              delayRequest={order.active_delay_request}
              onDecided={async () => {
                const fresh = await getOrderById(order.id);
                setOrder(fresh);
                setDrafts(toDrafts(fresh.deliverables));
              }}
            />
          ) : null
        }
        description={
          <OrderDescription
            orderNumber={order.id}
            orderName={order.name}
            artworkName={order.artwork_snapshot.artwork_name}
            category={order.artwork_snapshot.category}
            styles={order.artwork_snapshot.styles ?? []}
            requestDetail={order.customer_description}
          />
        }
        info={
          <OrderInfo
            status={order.status}
            priceSatang={order.price_satang}
            orderDate={formatDate(order.created_at)}
            deadlineDate={formatDate(order.deadline_at)}
            deadlineNote={getDeadlineNote(order.deadline_at, order.status)}
            revisions={{ used: revisionsUsed, limit: REVISION_LIMIT }}
            actions={
              <CustomerOrderActions
                orderId={order.id}
                status={order.status}
                onCancelOrder={() => setCancelOpen(true)}
                onPaymentBypassed={async () => {
                  const fresh = await getOrderById(order.id);
                  setOrder(fresh);
                  setDrafts(toDrafts(fresh.deliverables));
                }}
              />
            }
          />
        }
        userInfo={
          <div className="flex flex-col gap-2">
            <h2 className="mb-1 text-h3">Artist Info</h2>
            <ArtistMiniProfile
              artist={{
                artist_id: order.artist_id,
                artist_name:
                  artistProfile?.artist_name ?? order.other_party.name,
                profile_url: artistProfile?.profile_url ?? null,
                review_score:
                  artistProfile?.review_score ??
                  order.other_party.artist_review_score ??
                  null,
              }}
            />
          </div>
        }
      >
        {drafts.length > 0 && (
          <DraftHistory
            drafts={drafts}
            renderFinalArtworkActions={() =>
              order.original_image_url ? (
                <DownloadArtworkButton
                  imageUrl={order.original_image_url}
                  fileName={`${order.name}-final`}
                />
              ) : null
            }
            renderDraftActions={(draft) => (
              <CustomerDraftActions
                draft={draft}
                canRequestChanges={revisionsUsed < REVISION_LIMIT}
                onRequestChanges={(d, comment) =>
                  setPending({ kind: "request-changes", draft: d, comment })
                }
                onAccept={(d) => setPending({ kind: "accept", draft: d })}
              />
            )}
          />
        )}
      </OrderDetails>
      <CustomerCommentConfirmPopUp
        isOpen={pending !== null}
        kind={pending?.kind ?? "accept"}
        revisionNumber={revisionsUsed + 1}
        revisionLimit={REVISION_LIMIT}
        status={popupStatus}
        errorMessage={popupError}
        onCancel={closePopup}
        onConfirm={handleConfirm}
      />
      <RatingOrderPopup
        isOpen={reviewStep === "rating"}
        status="default"
        onCancel={() => setReviewStep("none")}
        // The backend has no review endpoint yet, so the rating is not persisted.
        onConfirm={() => setReviewStep("thanks")}
      />
      <ThankYouOrderPopup
        isOpen={reviewStep === "thanks"}
        status="success"
        onClose={() => setReviewStep("none")}
      />
      <CustomerCancelOrderPopUp
        isOpen={cancelOpen}
        orderId={order.id}
        orderLabel={`Order #${order.id.slice(0, 8)} · ${order.name} · ฿${(order.price_satang / 100).toLocaleString()}`}
        onClose={() => setCancelOpen(false)}
        onCancelled={() => setOrder((prev) => ({ ...prev, status: "CANCEL" }))}
      />
    </>
  );
}
