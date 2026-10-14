"use client";

import { useId, useState } from "react";
import Image from "next/image";
import { Button } from "@/components/ui/Button";
import { SelectInput } from "@/components/ui/SelectInput";
import ArtistOrderPopup from "./ArtistOrderPopup";

interface CancelOrderPopupProps {
  orderSummary: string;
  onClose: () => void;
  onConfirm: (value: { reason: string; description: string }) => void;
}

// TODO: Map these UI values to Backend reason codes when the API is connected.
const reasons = [
  { value: "cannot_complete", label: "I’m unable to complete the project" },
  { value: "cannot_meet_deadline", label: "I’m unable to meet the deadline" },
  { value: "requirements_changed", label: "The project requirements have changed" },
  { value: "client_unresponsive", label: "The client is unresponsive" },
  { value: "mutual_agreement", label: "The client and I agreed to cancel" },
  { value: "personal_or_technical", label: "I have a personal or technical issue" },
  { value: "other", label: "Other" },
];

export default function CancelOrderPopup({ orderSummary, onClose, onConfirm }: CancelOrderPopupProps) {
  const [reason, setReason] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState("");
  const descriptionId = useId();

  return (
    <ArtistOrderPopup
      title="Cancel this order?"
      form
      onClose={onClose}
      icon={<Image src="/icons/Error_Circle.svg" alt="" width={64} height={64} className="h-16 w-16" />}
    >
      <p className="mx-auto mt-3 text-center text-small">
        Canceling will end this project and notify the client.<br />
        This action cannot be undone.
      </p>
      <div className="my-3 rounded-xl bg-neutral-400 px-4 py-2 text-small">{orderSummary}</div>
      <form onSubmit={(event) => {
        event.preventDefault();
        if (!reason || !description.trim()) { setError("กรุณาเลือกเหตุผลและกรอกคำอธิบาย"); return; }
        onConfirm({ reason, description: description.trim() });
      }}>
        <p className="mb-2 text-caption">Reason for cancellation <span className="text-error">*</span></p>
        <SelectInput options={reasons} value={reason} onChange={setReason} formatSelectedLabel={(label) => reason ? label : "Select reason for cancellation"} buttonClassName="rounded-lg border-neutral-400" />
        <label htmlFor={descriptionId} className="mb-2 mt-3 block text-caption">Please explain your reason <span className="text-error">*</span></label>
        <textarea id={descriptionId} required maxLength={300} value={description} onChange={(event) => setDescription(event.target.value)} placeholder="Briefly explain why you're cancelling..." className="h-20 w-full resize-none rounded-xl border-2 border-neutral-400 p-3 text-small" />
        <p className="mt-1 text-right text-small text-neutral">{description.length}/300</p>
        <div className="mt-3 flex items-center gap-3 rounded-xl bg-accent-200 px-4 py-3 text-small text-error">
          <Image src="/icons/pink-error-triangle.svg" alt="" width={30} height={27} className="shrink-0" />
          <p>The client will be notified immediately. Any refund will be handled according to the order policy.</p>
        </div>
        {error && <p role="alert" className="mt-3 text-small text-error">{error}</p>}
        <div className="mt-4 flex justify-end gap-6">
          <Button type="button" variant="transparent" className="w-[135px] rounded-xl" onClick={onClose}>Keep order</Button>
          <Button type="submit" variant="error" className="w-[135px] rounded-xl">Cancel order</Button>
        </div>
      </form>
    </ArtistOrderPopup>
  );
}
