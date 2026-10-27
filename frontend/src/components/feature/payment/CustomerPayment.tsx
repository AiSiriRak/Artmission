"use client";

/* eslint-disable @next/next/no-img-element -- Local Figma assets and temporary file previews. */
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import MainLayout from "@/components/feature/main/MainLayout";
import { Button } from "@/components/ui/Button";
import type { OrderDetail } from "@/lib/api/orders";
import { routes } from "@/lib/routes";

const QR_LIFETIME_MS = 5 * 60 * 1000;
const card = "rounded-[25px] border border-primary-500 bg-secondary-200";
const asset = "/images/payment/";

export default function CustomerPayment({ order }: { order: OrderDetail }) {
  // Preview only: use the QR API's expires_at when the integration is available.
  const [expiresAt, setExpiresAt] = useState<number | null>(null);
  const [remaining, setRemaining] = useState(300);
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [dragging, setDragging] = useState(false);
  const input = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const deadline = expiresAt ?? Date.now() + QR_LIFETIME_MS;
    const interval = window.setInterval(() => {
      setRemaining(Math.max(0, Math.ceil((deadline - Date.now()) / 1000)));
    }, 250);
    return () => window.clearInterval(interval);
  }, [expiresAt]);

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
      setError("Invalid file. Please choose a PNG or JPG image.");
    } else if (selected.size === 0 || selected.size > 5 * 1024 * 1024) {
      setError("Please choose a non-empty image no larger than 5 MB.");
    } else {
      setFile(selected);
      setPreview(URL.createObjectURL(selected));
    }
  }

  const expired = remaining === 0;
  const time = `${Math.floor(remaining / 60).toString().padStart(2, "0")}:${(remaining % 60).toString().padStart(2, "0")}`;
  const amount = `${(order.price_satang / 100).toLocaleString("en-US")} THB`;

  return (
    <MainLayout usertype="customer">
      <div className="mx-auto max-w-[1920px] px-6 py-10 text-primary-500 sm:px-12 lg:px-40">
        <Link href={routes.order.detail(order.id)} className="inline-flex h-10 items-center justify-center gap-2 rounded-xl border px-6 text-button focus-visible:outline-2 focus-visible:outline-offset-4">
          <img src="/icons/Arrow_left.svg" width={24} height={24} alt="" />Back
        </Link>
        <h1 className="mb-8 mt-6 text-h2">Payment</h1>
        <div className="mb-10 rounded-[25px] bg-secondary-300 px-6 py-8 sm:px-12" role="note">
          <div className="flex items-center gap-3">
            <img src={`${asset}wallet.svg`} width={36} height={36} alt="" />
            <h2 className="text-h3 sm:text-h2">Payment preview</h2>
          </div>
          <p className="mt-2 text-body">Payment integration is not connected yet. Do not pay using this sample QR.</p>
        </div>
        <div className="grid items-start gap-12 lg:grid-cols-[minmax(0,1.85fr)_minmax(0,1fr)]">
          <div className="min-w-0 space-y-12">
            <section aria-labelledby="scan-heading">
              <h2 id="scan-heading" className="mb-4 text-h3">1. Scan to Pay</h2>
              <div className={`${card} flex flex-col items-center gap-4 p-6 text-center sm:p-8`}>
                {expired ? (
                  <div className="flex aspect-square w-full max-w-[500px] flex-col items-center justify-center gap-4" role="status">
                    <h3 className="text-h3">QR code expired</h3>
                    <p className="text-body">Request a new QR to continue.</p>
                  </div>
                ) : (
                  <img src={`${asset}promptpay-preview.png`} width={500} height={500} className="aspect-square w-full max-w-[500px] object-contain" alt="Sample PromptPay QR for design preview only. Do not scan to pay." />
                )}
                <div>
                  <h3 className="text-h3">Total Amount</h3>
                  <p className="mt-2 text-h2 sm:text-h1">{amount}</p>
                </div>
                <Button type="button" variant="dark" disabled title="Available when a payment QR has been generated" icon={<img src={`${asset}download.svg`} width={24} height={24} alt="" />}>Download QR</Button>
                <p className="text-small">{expired ? "Already paid? Attach and submit your payment slip below. You don’t need a new QR." : "Already paid? Attach your payment slip below."}</p>
              </div>
            </section>
            <section aria-labelledby="upload-heading">
              <h2 id="upload-heading" className="mb-4 text-h3">2. Upload your slip</h2>
              <div className={`${card} flex flex-col items-center gap-4 p-6 sm:p-8`}>
                <input ref={input} type="file" accept="image/png,image/jpeg" className="sr-only" tabIndex={-1} aria-label="Choose payment slip" onChange={(event) => { chooseFile(event.target.files); event.target.value = ""; }} />
                <button
                  type="button"
                  className={`flex min-h-[320px] w-full flex-col items-center justify-center gap-6 rounded-2xl border-2 border-dashed p-5 text-center focus-visible:outline-2 focus-visible:outline-offset-4 sm:min-h-[400px] ${dragging ? "border-accent-500 bg-accent-200" : "border-secondary-600"}`}
                  aria-describedby={error ? "slip-error" : "slip-help"}
                  onClick={() => input.current?.click()}
                  onDragOver={(event) => { event.preventDefault(); setDragging(true); }}
                  onDragLeave={() => setDragging(false)}
                  onDrop={(event) => { event.preventDefault(); setDragging(false); chooseFile(event.dataTransfer.files); }}
                >
                  {file && preview ? (
                    <>
                      <img src={preview} alt="Selected payment slip" className="max-h-64 max-w-full object-contain" onError={() => { setFile(null); setPreview(null); setError("This image could not be opened. Please choose another PNG or JPG."); }} />
                      <span className="max-w-full break-all text-body">{file.name}</span>
                      <span className="text-small">Click to choose another slip</span>
                    </>
                  ) : (
                    <>
                      <img src={`${asset}upload.svg`} width={150} height={150} alt="" />
                      <span className="text-body">Drag &amp; drop your slip here, or click to choose</span>
                    </>
                  )}
                  <span id="slip-help" className="text-small">PNG / JPG · Maximum 5 MB</span>
                </button>
                {error && <p id="slip-error" role="alert" className="text-small text-error">{error}</p>}
                {file && <Button type="button" onClick={() => { setFile(null); setPreview(null); setError(null); }}>Remove slip</Button>}
                <Button type="button" variant="dark" disabled className="w-[138px]">Submit</Button>
                <p className="text-center text-small">We’ll verify your slip and notify the artist.</p>
                <p className="text-center text-small">Slip submission will be available when payment integration is connected.</p>
              </div>
            </section>
          </div>
          <section className="min-w-0" aria-labelledby="payment-info-heading">
            <h2 id="payment-info-heading" className="mb-4 text-h3">Payment Info</h2>
            <div className={`${card} overflow-hidden`}>
              <div className="flex flex-col items-center gap-6 px-6 py-8">
                <div className="flex items-center justify-center gap-2">
                  <img src={`${asset}hourglass.svg`} width={48} height={48} alt="" />
                  <p className="text-h3 xl:text-h2" role="timer" aria-label="Preview QR countdown">{expired ? "QR expired" : <>QR expires in <span className="tabular-nums">{time}</span></>}</p>
                </div>
                <Button type="button" className="!border !rounded-xl" onClick={() => { setRemaining(300); setExpiresAt(Date.now() + QR_LIFETIME_MS); }}>Request new QR</Button>
                <p className="text-center text-small">Preview timer only · No payment QR is generated</p>
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
                <ol className="list-decimal pl-6 text-body">
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
