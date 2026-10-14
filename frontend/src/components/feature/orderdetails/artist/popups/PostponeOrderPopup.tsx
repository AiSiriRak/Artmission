"use client";

import { useId, useState } from "react";
import Image from "next/image";
import { Button } from "@/components/ui/Button";
import { SelectInput } from "@/components/ui/SelectInput";
import ArtistOrderPopup from "./ArtistOrderPopup";

export interface PostponeRequest {
  deadline: string;
  reason: string;
  description: string;
}

interface PostponeOrderPopupProps {
  orderSummary: string;
  currentDeadline: string;
  currentDeadlineLabel: string;
  onClose: () => void;
  onConfirm: (request: PostponeRequest) => void;
}

const reasons = [
  { value: "details", label: "I need more time for details" },
  { value: "feedback", label: "I am waiting for additional customer feedback" },
  { value: "emergency", label: "I have a personal emergency / illness" },
  { value: "workload", label: "I am experiencing an unexpected high workload" },
  { value: "other", label: "Other" },
];

export default function PostponeOrderPopup({ orderSummary, currentDeadline, currentDeadlineLabel, onClose, onConfirm }: PostponeOrderPopupProps) {
  const [deadline, setDeadline] = useState("");
  const [reason, setReason] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState("");
  const fieldId = useId();
  const earliestDate = new Date(currentDeadline + "T00:00:00Z");
  earliestDate.setUTCDate(earliestDate.getUTCDate() + 1);
  const minDate = earliestDate.toISOString().slice(0, 10);

  return (
    <ArtistOrderPopup title="Request to Postpone Order?" form onClose={onClose} icon={
      <Image src="/icons/clock.svg" alt="" width={64} height={64} className="h-16 w-16" />
    }>
      <p className="mx-auto mt-3 text-center text-small">
        This will send a request to extend the deadline.<br />
        The new date must be approved by the customer.
      </p>
      <div className="my-2 rounded-xl bg-neutral-400 px-4 py-2 text-small"><p>{orderSummary}</p><p>Current Deadline · {currentDeadlineLabel}</p></div>
      <form onSubmit={(event) => {
        event.preventDefault();
        if (!deadline || deadline <= currentDeadline || !reason || !description.trim()) {
          setError("กรุณาเลือกวันหลัง Deadline เดิม พร้อมเหตุผลและคำอธิบาย"); return;
        }
        onConfirm({ deadline, reason, description: description.trim() });
      }}>
        <label htmlFor={fieldId + "-date"} className="mb-2 block text-caption">New Proposed Deadline <span className="text-error">*</span></label>
        <div className="relative flex items-center gap-2 rounded-lg border-2 border-neutral-400 px-3 py-2 focus-within:border-primary-500">
          <Image src="/icons/calendar.svg" alt="" width={16} height={16} className="shrink-0" />
          <span aria-hidden="true" className={`text-small ${deadline ? "text-primary-500" : "text-neutral"}`}>
            {deadline
              ? new Date(deadline + "T00:00:00").toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" })
              : "Choose Date"}
          </span>
          <input
            id={fieldId + "-date"}
            type="date"
            required
            min={minDate}
            value={deadline}
            onChange={(event) => setDeadline(event.target.value)}
            onClick={(event) => event.currentTarget.showPicker?.()}
            className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
          />
        </div>
        <p className="mt-1 text-small text-neutral">* Must be later than the current deadline</p>
        <p className="mb-2 mt-3 text-caption">Reason for Postponement <span className="text-error">*</span></p>
        <SelectInput options={reasons} value={reason} onChange={setReason} formatSelectedLabel={(label) => reason ? label : "Select reason for postpone"} buttonClassName="rounded-lg border-neutral-400" />
        <label htmlFor={fieldId + "-description"} className="mb-2 mt-3 block text-caption">Please explain your reason <span className="text-error">*</span></label>
        <textarea id={fieldId + "-description"} required maxLength={300} value={description} onChange={(event) => setDescription(event.target.value)} placeholder="Briefly explain why you need more time..." className="h-18 w-full resize-none rounded-xl border-2 border-neutral-400 p-3 text-small" />
        <p className="mt-1 text-right text-small text-neutral">{description.length}/300</p>
        <div className="mt-3 flex items-center gap-3 rounded-xl bg-neutral-400 px-4 py-2 text-small">
          <Image src="/icons/Info.svg" alt="" width={32} height={32} className="h-8 w-8 shrink-0" />
          <p>The customer has 48 hours to respond. If no response, the request will be automatically declined and the original deadline will remain.</p>
        </div>
        {error && <p role="alert" className="mt-3 text-small text-error">{error}</p>}
        <div className="mt-3 flex justify-end gap-6">
          <Button type="button" variant="transparent" className="w-[138px] rounded-xl" onClick={onClose}>Cancel</Button>
          <Button type="submit" variant="dark" className="w-[138px] rounded-xl">Send Request</Button>
        </div>
      </form>
    </ArtistOrderPopup>
  );
}
