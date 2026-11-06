"use client";

/* eslint-disable @next/next/no-img-element -- Local Figma assets and temporary file previews. */
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { QRCodeCanvas } from "qrcode.react";
import MainLayout from "@/components/feature/main/MainLayout";
import { Button } from "@/components/ui/Button";
import type { OrderDetail } from "@/lib/api/orders";
import { routes } from "@/lib/routes";
import { ApiError } from "@/lib/api/client";
import { loadPaymentQR, verifyPaymentSlip, type PaymentQR } from "@/lib/api/payments";
import { paymentDisplayDeadline, remainingSeconds } from "@/lib/payment-countdown";

const card = "rounded-[25px] border border-primary-500 bg-secondary-200";
const asset = "/images/payment/";

export default function CustomerPayment({ order }: { order: OrderDetail }) {
  const router = useRouter();
  const [qr, setQr] = useState<PaymentQR | null>(null);
  const [remaining, setRemaining] = useState<number | null>(null);
  const [displayDeadline, setDisplayDeadline] = useState<number | null>(null);
  const [backendRemaining, setBackendRemaining] = useState<number | null>(null);
  const [qrLoading, setQrLoading] = useState(true);
  const [qrError, setQrError] = useState<string | null>(null);
  const [serverExpired, setServerExpired] = useState(false);
  const [loadingAttempt, setLoadingAttempt] = useState(0);
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [dragging, setDragging] = useState(false);
  const input = useRef<HTMLInputElement>(null);
  const chooseButton = useRef<HTMLButtonElement>(null);
  const qrCanvas = useRef<HTMLCanvasElement>(null);
  const pendingQR = useRef<{ orderId: string; promise: Promise<PaymentQR> } | null>(null);
  const submittingRef = useRef(false);

  useEffect(() => {
    let active = true;
    // Reuse the in-flight request during React's development effect replay.
    const promise = pendingQR.current?.orderId === order.id ? pendingQR.current.promise : loadPaymentQR(order.id);
    pendingQR.current = { orderId: order.id, promise };
    promise.then((result) => {
      if (!active) return;
      setQr(result);
      const now = Date.now();
      const deadline = paymentDisplayDeadline(result.qr_expire, now);
      setDisplayDeadline(deadline);
      setRemaining(remainingSeconds(deadline, now));
      setBackendRemaining(remainingSeconds(Date.parse(result.qr_expire), now));
      setQrError(null);
      setServerExpired(false);
    }).catch((error: unknown) => {
      if (!active) return;
      if (error instanceof ApiError && error.status === 410) setServerExpired(true);
      else setQrError(error instanceof Error ? error.message : "Could not load the payment QR.");
    }).finally(() => {
      if (pendingQR.current?.promise === promise) pendingQR.current = null;
      if (active) setQrLoading(false);
    });
    return () => { active = false; };
  }, [order.id, loadingAttempt]);

  useEffect(() => {
    if (!qr || displayDeadline === null) return;
    const interval = window.setInterval(() => {
      const now = Date.now();
      setRemaining(remainingSeconds(displayDeadline, now));
      setBackendRemaining(remainingSeconds(Date.parse(qr.qr_expire), now));
    }, 1000);
    return () => window.clearInterval(interval);
  }, [qr, displayDeadline]);

  useEffect(() => {
    if (!preview) return;
    return () => URL.revokeObjectURL(preview);
  }, [preview]);

  function chooseFile(files: FileList | null) {
    setError(null);
    setFile(null);
    setPreview(null);
    if (!files?.length) return;
    if (files.length !== 1) { setError("Please attach one payment slip at a time."); return; }
    const selected = files[0];
    if (!["image/png", "image/jpeg"].includes(selected.type)) {
      setError("Choose a PNG or JPG image.");
    } else if (selected.size > 5 * 1024 * 1024) {
      setError("This file is too large. Maximum size is 5 MB.");
    } else if (selected.size === 0) {
      setError("This file is empty. Choose another image.");
    } else {
      setFile(selected);
      setPreview(URL.createObjectURL(selected));
    }
  }

  const paymentExpired = serverExpired || backendRemaining === 0;
  const expired = paymentExpired || remaining === 0;
  const seconds = remaining ?? 0;
  const time = `${Math.floor(seconds / 3600) ? `${Math.floor(seconds / 3600)}:` : ""}${Math.floor((seconds % 3600) / 60).toString().padStart(2, "0")}:${(seconds % 60).toString().padStart(2, "0")}`;
  const amount = `${(order.price_satang / 100).toLocaleString("en-US")} THB`;

  async function submitSlip() {
    if (!file) { setError("Please select a slip before submitting."); return; }
    if (submittingRef.current || paymentExpired || !qr) return;
    if (Date.parse(qr.qr_expire) <= Date.now()) { setServerExpired(true); return; }
    submittingRef.current = true;
    setSubmitting(true);
    setError(null);
    try {
      const result = await verifyPaymentSlip(order.id, file);
      if (!result.success) setError(result.message);
      else router.replace(routes.order.detail(order.id));
    } catch (error) {
      if (error instanceof ApiError && error.status === 410) setServerExpired(true);
      setError(error instanceof ApiError && error.status >= 500
        ? "We couldn’t confirm whether your payment was recorded. Check your order status before resubmitting, or contact support."
        : error instanceof Error ? error.message : "Could not verify this slip. Please check your order status before trying again.");
    } finally {
      submittingRef.current = false;
      setSubmitting(false);
    }
  }

  function retryQR() {
    setQrLoading(true);
    setQrError(null);
    setLoadingAttempt((value) => value + 1);
  }

  function downloadQR() {
    if (!qrCanvas.current || expired) return;
    const link = document.createElement("a");
    link.download = `payment-${order.id}.png`;
    link.href = qrCanvas.current.toDataURL("image/png");
    link.click();
  }

  return (
    <MainLayout usertype="customer">
      <div className="mx-auto max-w-[1920px] px-6 py-10 text-primary-500 sm:px-12 lg:px-40">
        <Link href={routes.order.detail(order.id)} replace className="inline-flex h-10 items-center justify-center gap-2 rounded-xl border px-6 text-button focus-visible:outline-2 focus-visible:outline-offset-4">
          <img src="/icons/Arrow_left.svg" width={24} height={24} alt="" />Back
        </Link>
        <h1 className="mb-8 mt-6 text-h2">Payment</h1>
        <div className="mb-10 rounded-[25px] bg-secondary-300 px-6 py-8 sm:px-12" role="note">
          <div className="flex items-center gap-3">
            <img src={`${asset}wallet.svg`} width={36} height={36} alt="" />
            <h2 className="text-h3 sm:text-h2">{qr ? `Pay by ${new Date(qr.qr_expire).toLocaleString("en-GB", { timeZone: "Asia/Bangkok", day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" })} (UTC+7)` : "Order payment"}</h2>
          </div>
          <p className="mt-2 text-body">{paymentExpired ? "This payment has expired. Check your order details for its current status." : "Unpaid orders are automatically cancelled after the payment deadline."}</p>
        </div>
        <div className="grid items-start gap-12 lg:grid-cols-[minmax(0,1.85fr)_minmax(0,1fr)]">
          <div className="min-w-0 space-y-12">
            <section aria-labelledby="scan-heading">
              <h2 id="scan-heading" className="mb-4 text-h3">1. Scan to Pay</h2>
              <div className={`${card} flex flex-col items-center gap-4 p-6 text-center sm:p-8`}>
                {expired ? (
                  <div className="flex aspect-square w-full max-w-[500px] flex-col items-center justify-center gap-4" role="status">
                    <h3 className="text-h3">QR code expired</h3>
                    <p className="text-body">{paymentExpired ? "This payment can no longer be used. Please check your order details." : "Request a new QR to continue."}</p>
                  </div>
                ) : qrLoading ? (
                  <div className="flex aspect-square w-full max-w-[500px] items-center justify-center" role="status">Loading payment QR…</div>
                ) : qr ? (
                  <div className="w-full max-w-[500px] bg-white p-4">
                    <QRCodeCanvas ref={qrCanvas} value={qr.qr_payload} size={500} marginSize={4} style={{ width: "100%", height: "auto" }} role="img" aria-label="PromptPay payment QR code" />
                  </div>
                ) : (
                  <div className="flex aspect-square w-full max-w-[500px] flex-col items-center justify-center gap-4">
                    <p role="alert" className="text-error">{qrError || "Payment QR is unavailable."}</p>
                    <Button type="button" onClick={retryQR}>Try again</Button>
                  </div>
                )}
                {qr && qrError && <p role="alert" className="text-error">{qrError}</p>}
                <div>
                  <h3 className="text-h3">Total Amount</h3>
                  <p className="mt-2 text-h2 sm:text-h1">{amount}</p>
                </div>
                <Button type="button" variant="dark" disabled={!qr || expired || qrLoading} onClick={downloadQR} icon={<img src={`${asset}download.svg`} width={24} height={24} alt="" />}>Download QR</Button>
                <p className="text-small">{paymentExpired ? "This payment is no longer available." : "Already paid? Attach your payment slip below."}</p>
              </div>
            </section>
            <section aria-labelledby="upload-heading">
              <h2 id="upload-heading" className="mb-4 text-h3">2. Upload your slip</h2>
              <div className={`${card} flex flex-col items-center gap-4 p-6 sm:p-8`}>
                <input ref={input} type="file" accept="image/png,image/jpeg" disabled={submitting} className="sr-only" tabIndex={-1} aria-label="Choose payment slip" onChange={(event) => { chooseFile(event.target.files); event.target.value = ""; }} />
                <div
                  className={`flex min-h-[320px] w-full flex-col items-center justify-center gap-6 rounded-2xl border-2 border-dashed p-5 text-center focus-visible:outline-2 focus-visible:outline-offset-4 sm:min-h-[400px] ${dragging ? "border-accent-500 bg-accent-200" : "border-secondary-600"}`}
                  onDragOver={(event) => { event.preventDefault(); setDragging(true); }}
                  onDragLeave={() => setDragging(false)}
                  onDrop={(event) => { event.preventDefault(); setDragging(false); if (!submitting) chooseFile(event.dataTransfer.files); }}
                >
                  {file && preview ? (
                    <>
                      <div className="relative max-w-full">
                        <img src={preview} alt="Selected payment slip" className="max-h-64 max-w-full object-contain" onError={() => { setFile(null); setPreview(null); setError("This image could not be opened. Please choose another PNG or JPG."); }} />
                        <button type="button" aria-label="Remove slip" disabled={submitting} className="absolute right-2 top-2 flex size-10 items-center justify-center rounded-full border border-primary-500 bg-white focus-visible:outline-2 focus-visible:outline-offset-2 disabled:opacity-50" onClick={() => { setFile(null); setPreview(null); setError(null); requestAnimationFrame(() => chooseButton.current?.focus()); }}>
                          <img src="/icons/close_round.svg" width={24} height={24} alt="" />
                        </button>
                      </div>
                      <span className="max-w-full break-all text-body">{file.name}</span>
                      <button type="button" disabled={submitting} className="text-small underline focus-visible:outline-2 focus-visible:outline-offset-4" onClick={() => input.current?.click()}>Choose another slip</button>
                    </>
                  ) : (
                    <button ref={chooseButton} type="button" disabled={submitting} className="flex w-full flex-col items-center gap-6 focus-visible:outline-2 focus-visible:outline-offset-4" aria-describedby={error ? "slip-error" : "slip-help"} onClick={() => input.current?.click()}>
                      <img src={`${asset}upload.svg`} width={150} height={150} alt="" />
                      <span className="text-body">Drag &amp; drop your slip here, or click to choose</span>
                    </button>
                  )}
                  <span id="slip-help" className="text-small">PNG / JPG · Maximum 5 MB</span>
                </div>
                {error && <p id="slip-error" role="alert" className="text-small text-error">{error}</p>}
                <Button type="button" variant="dark" disabled={submitting || paymentExpired || qrLoading || !qr} className="w-[138px]" onClick={() => void submitSlip()}>{submitting ? "Verifying…" : "Submit"}</Button>
                <p className="text-center text-small">We’ll verify your slip and notify the artist.</p>
              </div>
            </section>
          </div>
          <section className="min-w-0" aria-labelledby="payment-info-heading">
            <h2 id="payment-info-heading" className="mb-4 text-h3">Payment Info</h2>
            <div className={`${card} overflow-hidden`}>
              <div className="flex flex-col items-center gap-6 px-6 py-8">
                <div className="flex items-center justify-center gap-2">
                  <img src={`${asset}hourglass.svg`} width={48} height={48} alt="" />
                  <p className="text-h3 xl:text-h2" role="timer" aria-label="QR countdown">{expired ? "QR expired" : qr ? <>QR expires in <span className="tabular-nums">{time}</span></> : qrLoading ? "Loading…" : "QR unavailable"}</p>
                </div>
                <Button type="button" disabled={qrLoading || submitting} onClick={retryQR}>{qrLoading ? "Loading…" : "Request new QR"}</Button>
                {paymentExpired && <Link href={routes.order.detail(order.id)} replace className="text-button underline">View order details</Link>}
              </div>
              <dl className="space-y-3 border-y px-6 py-6">
                {[
                  ["Order", order.name],
                  ["Artist", order.other_party.name],
                  ["Total Amount", amount],
                  ["Order Date", new Date(order.created_at).toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric", timeZone: "Asia/Bangkok" })],
                ].map(([label, value]) => (
                  <div key={label} className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
                    <dt className="text-h3">{label}</dt><dd className="min-w-0 break-words text-right text-body">{value}</dd>
                  </div>
                ))}
              </dl>
              <div className="px-6 py-8">
                <h3 className="mb-2 text-h3">How to pay</h3>
                <ol className="list-decimal pl-6 text-body text-[17px]">
                  <li>Open your banking app and scan the QR.</li>
                  <li>Confirm the amount and pay.</li>
                  <li>Attach the slip and submit it.</li>
                </ol>
              </div>
            </div>
          </section>
        </div>
      </div>
    </MainLayout>
  );
}
