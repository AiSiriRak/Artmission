"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import OrderDetails from "../shared/OrderDetails";
import OrderDescription from "../shared/OrderDescription";
import OrderInfo from "../shared/OrderInfo";
import DraftHistory, { type Draft } from "../shared/DraftHistory";
import MainLayout from "@/components/feature/main/MainLayout";
import { Button } from "@/components/ui/Button";
import { Loading } from "@/components/ui/Loading";
import { getOrderById, confirmDeliverable, type OrderDetail } from "@/lib/api/orders";
import { getArtistProfile } from "@/lib/api/artists";
import type { ArtistProfile } from "@/lib/api/types";
import { isApiError } from "@/lib/api/error";
import { startAutoRefresh } from "@/lib/auto-refresh";
import { routes } from "@/lib/routes";
import CustomerOrderActions from "./CustomerOrderActions";
import CustomerDraftActions from "./CustomerDraftActions";
import CustomerPostponeSection from "./CustomerPostponeSection";
import DownloadArtworkButton from "./DownloadArtworkButton";
import { CustomerCancelOrderPopUp } from "./CustomerCancelOrderPopUp";
import { CustomerCommentConfirmPopUp } from "./CustomerCommentConfirmPopup";
import { ArtistMiniProfile } from "./ArtistMiniProfile";

const decisions: Record<NonNullable<OrderDetail["deliverables"]>[number]["decision"], Draft["status"]> = {
  WAIT: "WAITING_FOR_CUSTOMER_REVIEW", REJECTED: "CHANGE_REQUESTED", APPROVED: "COMMISSION_ACCEPTED",
};

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" });
}

export default function CustomerOrderDetail({ orderId }: { orderId: string }) {
  const router = useRouter();
  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [artist, setArtist] = useState<ArtistProfile | null>(null);
  const [error, setError] = useState("");
  const [cancelOpen, setCancelOpen] = useState(false);
  const [pending, setPending] = useState<{ draft: Draft; accept: boolean; comment: string } | null>(null);
  const [popupStatus, setPopupStatus] = useState<"default" | "success" | "fail">("default");
  const [popupError, setPopupError] = useState("");
  const [busy, setBusy] = useState(false);
  const active = useRef(false);
  const requestVersion = useRef(0);
  const submitting = useRef(false);

  const reload = useCallback(async () => {
    const version = ++requestVersion.current;
    try {
      const latest = await getOrderById(orderId, { signal: AbortSignal.timeout(15000), cache: "no-store" });
      if (active.current && version === requestVersion.current) {
        setOrder(latest);
        setError("");
      }
    } catch (error: unknown) {
      if (!active.current || version !== requestVersion.current) return;
      if (isApiError(error) && error.status === 401) router.replace(routes.login);
      setError(error instanceof Error ? error.message : "โหลดออร์เดอร์ไม่สำเร็จ");
    }
  }, [orderId, router]);

  useEffect(() => {
    active.current = true;
    let refreshing = false;
    async function refresh() {
      if (refreshing || submitting.current) return;
      refreshing = true;
      try { await reload(); } finally { refreshing = false; }
    }
    void refresh();
    const stop = startAutoRefresh(() => { void refresh(); });
    return () => { active.current = false; requestVersion.current += 1; stop(); };
  }, [reload]);

  const artistId = order?.artist_id;
  useEffect(() => {
    let active = true;
    if (artistId) void getArtistProfile(artistId).then((profile) => {
      if (active) setArtist(profile);
    }).catch(() => { /* The order still provides the artist's name. */ });
    return () => { active = false; };
  }, [artistId]);

  async function handleConfirm() {
    if (!pending || submitting.current) return;
    submitting.current = true;
    requestVersion.current += 1;
    setBusy(true);
    try {
      await confirmDeliverable(orderId, pending.draft.id, { accept: pending.accept, comment: pending.comment });
      if (!active.current) return;
      setOrder((previous) => previous && ({
        ...previous,
        status: pending.accept ? "SUCCESS" : previous.status,
        deliverables: previous.deliverables?.map((draft) => draft.id === pending.draft.id
          ? { ...draft, decision: pending.accept ? "APPROVED" : "REJECTED", comment: pending.comment, updated_at: new Date().toISOString() }
          : draft) ?? null,
      }));
      setPopupStatus("success");
      await reload();
    } catch (error: unknown) {
      if (!active.current) return;
      setPopupError(error instanceof Error ? error.message : "ทำรายการไม่สำเร็จ");
      setPopupStatus("fail");
      if (isApiError(error) && error.status === 401) router.replace(routes.login);
    } finally {
      submitting.current = false;
      if (active.current) setBusy(false);
    }
  }

  if (!order) return <MainLayout usertype="customer">{error
    ? <div className="flex flex-col items-center gap-4 p-12"><p role="alert">{error}</p><Button onClick={() => { void reload(); }}>Try again</Button></div>
    : <Loading />}</MainLayout>;

  const drafts: Draft[] = [...(order.deliverables ?? [])].sort((a, b) => b.version - a.version).map((item) => ({
    id: item.id, draftNumber: item.version, status: decisions[item.decision],
    submitDate: formatDate(item.created_at), reviewDate: item.decision === "WAIT" ? undefined : formatDate(item.updated_at),
    imageUrl: item.preview_image_url, customerReview: item.comment ?? undefined,
  }));
  const revisionsUsed = drafts.filter((draft) => draft.status === "CHANGE_REQUESTED").length;
  const latestDraft = drafts[0];
  const openDecision = (draft: Draft, accept: boolean, comment = "") => {
    setPopupStatus("default"); setPopupError(""); setPending({ draft, accept, comment });
  };

  return <>
    <OrderDetails usertype="customer"
      notice={order.status === "IN_PROCESS" && order.active_delay_request?.customer_decision === "WAIT"
        ? <CustomerPostponeSection key={order.active_delay_request.id} orderId={orderId} delayRequest={order.active_delay_request} onDecided={reload} /> : undefined}
      description={<OrderDescription orderNumber={order.id} artworkName={order.artwork_snapshot.artwork_name || order.name}
        categoryName="" categories={order.artwork_snapshot.category ? [order.artwork_snapshot.category] : []}
        styles={order.artwork_snapshot.styles ?? []} requestDetail={order.customer_description} />}
      info={<><OrderInfo status={order.status} priceSatang={order.price_satang} orderDate={formatDate(order.created_at)}
        deadlineDate={formatDate(order.deadline_at)} revisions={{ used: revisionsUsed, limit: 2 }}
        actions={order.status !== "CANCEL" && order.status !== "SUCCESS" ? <CustomerOrderActions status={order.status} onCancelOrder={() => setCancelOpen(true)} /> : undefined} />
        {error && <p role="status" className="mt-3 text-small text-error">{error}</p>}</>}
      userInfo={<><h2 className="mb-4 text-h3">Artist Info</h2><ArtistMiniProfile artist={{ artist_id: order.artist_id,
        artist_name: artist?.artist_name ?? order.other_party.name, profile_url: artist?.profile_url ?? null,
        review_score: artist?.review_score ?? order.other_party.artist_review_score ?? null }} /></>}
    >
      {drafts.length > 0 && <DraftHistory drafts={drafts}
        renderDraftActions={(draft) => order.status === "IN_PROCESS" && draft.id === latestDraft?.id
          ? <CustomerDraftActions key={draft.id} draft={draft} canRequestChanges={revisionsUsed < 2}
            onAccept={(draft) => openDecision(draft, true)} onRequestChanges={(draft, comment) => openDecision(draft, false, comment)} /> : null}
        // TODO(backend): Enable only after an authorized original-file download API exists; preview_image_url is watermarked.
        renderFinalArtworkActions={() => order.status === "SUCCESS" ? <DownloadArtworkButton fileName={`${order.name}-final`} /> : null} />}
    </OrderDetails>
    <CustomerCommentConfirmPopUp isOpen={Boolean(pending)} kind={pending?.accept ? "accept" : "request-changes"}
      revisionNumber={revisionsUsed + 1} revisionLimit={2} status={popupStatus} errorMessage={popupError} busy={busy}
      onCancel={() => { if (!submitting.current) setPending(null); }} onConfirm={() => { void handleConfirm(); }} />
    <CustomerCancelOrderPopUp isOpen={cancelOpen} orderId={orderId} orderLabel={`Order #${order.id} · ${order.name}`}
      onClose={() => setCancelOpen(false)} onCancelled={() => { void reload(); }} />
  </>;
}
