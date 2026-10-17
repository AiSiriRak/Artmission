"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import DraftHistory, { type Draft } from "../shared/DraftHistory";
import { cancelOrder, confirmOrder, getOrderDetail, requestOrderPostpone, submitOrderDraft, type OrderDetail, type OrderDeliverable } from "@/lib/api/order-details";
import { getAccount } from "@/lib/api/users";
import { isApiError } from "@/lib/api/error";
import { routes } from "@/lib/routes";
import { startAutoRefresh } from "@/lib/auto-refresh";
import { Button } from "@/components/ui/Button";
import { Loading } from "@/components/ui/Loading";
import MainLayout from "@/components/feature/main/MainLayout";
import OrderDetails from "../shared/OrderDetails";
import OrderDescription from "../shared/OrderDescription";
import OrderInfo from "../shared/OrderInfo";
import UserInfo from "../shared/UserInfo";
import ArtistOrderActions from "./ArtistOrderActions";
import SubmitDraft from "./SubmitDraft";
import ConfirmOrderPopup from "./popups/ConfirmOrderPopup";
import CancelOrderPopup from "./popups/CancelOrderPopup";
import PostponeOrderPopup from "./popups/PostponeOrderPopup";
import PostponeNotice from "./PostponeNotice";

function formatDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "—" : date.toLocaleDateString("en-GB", {
    day: "2-digit", month: "short", year: "numeric",
  });
}

export function mapOrderDrafts(deliverables: OrderDeliverable[]): Draft[] {
  const decisions: Record<OrderDeliverable["decision"], Draft["status"]> = {
    WAIT: "WAITING_FOR_CUSTOMER_REVIEW",
    REJECTED: "CHANGE_REQUESTED",
    APPROVED: "COMMISSION_ACCEPTED",
  };
  return [...deliverables].sort((a, b) => b.version - a.version).map((item) => ({
    id: item.id,
    draftNumber: item.version,
    status: decisions[item.decision],
    submitDate: formatDate(item.created_at),
    reviewDate: item.decision === "WAIT" ? undefined : formatDate(item.updated_at),
    imageUrl: item.preview_image_url,
    customerReview: item.comment ?? undefined,
    // TODO: Backend must return the actual review expiry before showing auto-acceptance.
  }));
}

function errorMessage(error: unknown) {
  if (isApiError(error)) {
    if (error.status === 401) return "กรุณาเข้าสู่ระบบใหม่";
    if (error.status === 403) return "คุณไม่มีสิทธิ์ทำรายการกับออร์เดอร์นี้";
    if (error.status === 404) return "ไม่พบออร์เดอร์นี้";
    if (error.message) return error.message;
  }
  return error instanceof Error ? error.message : "ทำรายการไม่สำเร็จ กรุณาลองใหม่";
}

export default function ArtistOrderDetails() {
  const { orderId } = useParams<{ orderId: string }>();
  const router = useRouter();
  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [loadedId, setLoadedId] = useState("");
  const [loadError, setLoadError] = useState("");
  const [reloadKey, setReloadKey] = useState(0);
  const [popup, setPopup] = useState<"reject" | "accept" | "cancel" | "postpone" | null>(null);
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState("");
  const [refreshWarning, setRefreshWarning] = useState("");
  const mutationPending = useRef(false);
  const mutationVersion = useRef(0);
  const activeOrderId = useRef<string | null>(null);

  useEffect(() => {
    activeOrderId.current = orderId;
    let active = true;
    let ready = false;
    let refreshing = false;
    const refreshController = new AbortController();
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 15000);
    async function load() {
      try {
        const [account, detail] = await Promise.all([getAccount(), getOrderDetail(orderId, { signal: controller.signal })]);
        if (!active) return;
        if (account.role !== "artist") throw new Error("หน้านี้สำหรับบัญชี Artist เท่านั้น");
        if (!detail) throw new Error("ไม่พบข้อมูลออร์เดอร์");
        setOrder(detail);
        setLoadError("");
        setPopup(null);
        setActionError("");
        setRefreshWarning("");
        ready = true;
      } catch (error: unknown) {
        if (!active) return;
        setOrder(null);
        setLoadError(controller.signal.aborted ? "รอข้อมูลนานเกินไป กรุณาลองใหม่" : errorMessage(error));
        if (isApiError(error) && error.status === 401) router.replace(routes.login);
      } finally {
        window.clearTimeout(timeout);
        if (active) setLoadedId(orderId);
      }
    }
    void load();
    async function refresh() {
      if (!ready || refreshing || mutationPending.current) return;
      refreshing = true;
      const version = mutationVersion.current;
      try {
        const latest = await getOrderDetail(orderId, {
          signal: AbortSignal.any([refreshController.signal, AbortSignal.timeout(15000)]),
          cache: "no-store",
        });
        // Ignore a response fetched before an Accept/Cancel/Submit action.
        if (!active || mutationPending.current || version !== mutationVersion.current) return;
        setOrder(latest);
        setRefreshWarning("");
      } catch (error: unknown) {
        if (!active || mutationPending.current || version !== mutationVersion.current) return;
        if (isApiError(error) && error.status === 401) router.replace(routes.login);
        else setRefreshWarning("โหลดข้อมูลล่าสุดไม่สำเร็จ ระบบจะลองใหม่อัตโนมัติ");
      } finally {
        refreshing = false;
      }
    }
    const stopRefresh = startAutoRefresh(() => { void refresh(); });
    return () => {
      active = false;
      activeOrderId.current = null;
      controller.abort();
      refreshController.abort();
      stopRefresh();
      window.clearTimeout(timeout);
    };
  }, [orderId, reloadKey, router]);

  function closePopup() {
    if (!mutationPending.current) setPopup(null);
  }

  function openPopup(value: "reject" | "accept" | "cancel" | "postpone") {
    if (mutationPending.current) return;
    setActionError("");
    setPopup(value);
  }

  async function runMutation(action: () => Promise<(previous: OrderDetail) => OrderDetail>) {
    if (mutationPending.current) throw new Error("กำลังทำรายการ กรุณารอสักครู่");
    mutationPending.current = true;
    mutationVersion.current += 1;
    setBusy(true);
    setActionError("");
    setRefreshWarning("");
    try {
      const update = await action();
      if (activeOrderId.current !== orderId) return;
      setOrder((previous) => previous?.id === orderId ? update(previous) : previous);
      setPopup(null);
      try {
        const latest = await getOrderDetail(orderId, { signal: AbortSignal.timeout(15000) });
        if (activeOrderId.current === orderId) setOrder(latest);
      } catch {
        if (activeOrderId.current === orderId) setRefreshWarning("ทำรายการสำเร็จแล้ว แต่โหลดข้อมูลล่าสุดไม่สำเร็จ กรุณารีเฟรชหน้า");
      }
    } catch (error: unknown) {
      if (activeOrderId.current === orderId) {
        setActionError(errorMessage(error));
        if (isApiError(error) && error.status === 401) router.replace(routes.login);
      }
      throw error;
    } finally {
      mutationVersion.current += 1;
      mutationPending.current = false;
      setBusy(false);
    }
  }

  if (loadedId !== orderId) return <MainLayout usertype="artist"><Loading /></MainLayout>;
  if (!order || loadError) {
    return (
      <MainLayout usertype="artist">
        <div className="flex flex-col items-center gap-6 p-12">
          <p role="alert" className="text-body text-error">{loadError || "ไม่พบออร์เดอร์นี้"}</p>
          <Button type="button" onClick={() => { setLoadedId(""); setReloadKey((value) => value + 1); }}>ลองใหม่</Button>
        </div>
      </MainLayout>
    );
  }

  const drafts = mapOrderDrafts(order.deliverables ?? []);
  const latestDeliverable = [...(order.deliverables ?? [])].sort((a, b) => b.version - a.version)[0];
  const canSubmit = order.status === "IN_PROCESS" && latestDeliverable?.decision !== "WAIT";
  const hasActions = ["PENDING", "NOT_PAID", "IN_PROCESS"].includes(order.status);
  const pendingPostpone = order.status === "IN_PROCESS" && order.active_delay_request?.customer_decision === "WAIT"
    ? order.active_delay_request : null;
  const deadline = new Date(order.deadline_at);
  const deadlineInput = [deadline.getFullYear(), String(deadline.getMonth() + 1).padStart(2, "0"), String(deadline.getDate()).padStart(2, "0")].join("-");
  const summary = "Order #" + order.id + " · " + order.name + " · ฿" + (order.price_satang / 100).toLocaleString("en-US");

  return (
    <>
      <OrderDetails
        usertype="artist"
        notice={pendingPostpone ? <PostponeNotice deadline={pendingPostpone.deadline_request_at} /> : undefined}
        description={
          <OrderDescription orderNumber={order.id}
            artworkName={order.artwork_snapshot.artwork_name || order.name}
            categoryName=""
            categories={order.artwork_snapshot.category ? [order.artwork_snapshot.category] : []}
            styles={order.artwork_snapshot.styles ?? []}
            requestDetail={order.customer_description} />
        }
        info={
          <>
            <OrderInfo status={order.status} priceSatang={order.price_satang}
              orderDate={formatDate(order.created_at)} deadlineDate={formatDate(order.deadline_at)}
              actions={hasActions ? (
                <ArtistOrderActions status={order.status} busy={busy} postponePending={Boolean(pendingPostpone)}
                  onReject={() => openPopup("reject")} onAccept={() => openPopup("accept")}
                  onCancel={() => openPopup("cancel")} onPostpone={() => openPopup("postpone")} />
              ) : undefined} />
            {/* TODO: Display revisions when the API provides the revision allowance. */}
            {refreshWarning && <p role="status" className="mt-3 text-small text-error">{refreshWarning}</p>}
          </>
        }
        userInfo={<UserInfo title="Customer Info" name={order.other_party.name} email={order.other_party.email} />}
      >
        {canSubmit && (
          <SubmitDraft disabled={busy} onSubmit={(file) => runMutation(async () => {
            const deliverable = await submitOrderDraft(orderId, file);
            return (previous) => ({ ...previous, deliverables: [
              ...(previous.deliverables ?? []).filter((item) => item.id !== deliverable.id), deliverable,
            ] });
          })} />
        )}
        {drafts.length > 0 && <DraftHistory drafts={drafts} />}
      </OrderDetails>
      {(popup === "reject" || popup === "accept") && (
        <ConfirmOrderPopup action={popup} busy={busy} error={actionError} onClose={closePopup}
          onConfirm={() => {
            void runMutation(async () => {
              const result = await confirmOrder(orderId, popup === "accept");
              return (previous) => ({ ...previous, status: result.status });
            }).catch(() => { /* Error is shown in the popup. */ });
          }} />
      )}
      {popup === "cancel" && (
        <CancelOrderPopup orderSummary={summary} busy={busy} serverError={actionError} onClose={closePopup}
          onConfirm={({ reason, description }) => {
            void runMutation(async () => {
              await cancelOrder(orderId, { reason, explanation: description });
              return (previous) => ({ ...previous, status: "CANCEL" });
            }).catch(() => { /* Error is shown in the popup. */ });
          }} />
      )}
      {popup === "postpone" && (
        <PostponeOrderPopup orderSummary={summary} currentDeadline={deadlineInput}
          currentDeadlineLabel={formatDate(order.deadline_at)} onClose={closePopup} busy={busy} serverError={actionError}
          onConfirm={({ deadline: proposedDate, reason, description }) => {
            void runMutation(async () => {
              // Keep the deadline's local time when changing its calendar date.
              const requestedDeadline = new Date(order.deadline_at);
              const [year, month, day] = proposedDate.split("-").map(Number);
              requestedDeadline.setFullYear(year, month - 1, day);
              const request = await requestOrderPostpone(orderId, {
                deadline_request_at: requestedDeadline.toISOString(),
                reason,
                explanation: description,
              });
              return (previous) => ({ ...previous, active_delay_request: request });
            }).catch(() => { /* Error is shown in the popup. */ });
          }} />
      )}
    </>
  );
}
