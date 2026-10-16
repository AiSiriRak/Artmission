
"use client";

import Image from "next/image";
import { useEffect } from "react";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";

export type CustomerConfirmKind = "request-changes" | "accept";

interface CustomerCommentConfirmPopUpProps {
  isOpen: boolean;
  kind: CustomerConfirmKind;
  /** 1-based revision this request would use (request-changes only). */
  revisionNumber?: number;
  revisionLimit?: number;
  status?: "default" | "success" | "fail";
  errorMessage?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
}

const copy: Record<
  CustomerConfirmKind,
  { title: string; confirm: string; success: string; fail: string }
> = {
  "request-changes": {
    title: "Request Changes",
    confirm: "Confirm",
    success: "Your change request has been sent to the artist.",
    fail: "Unable to send your change request.",
  },
  accept: {
    title: "Accept Commission",
    confirm: "Accept",
    success: "The commission has been accepted.",
    fail: "Unable to accept the commission.",
  },
};

export function CustomerCommentConfirmPopUp({
  isOpen,
  kind,
  revisionNumber = 1,
  revisionLimit = 2,
  status = "default",
  errorMessage,
  onCancel,
  onConfirm,
}: CustomerCommentConfirmPopUpProps) {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  }, [isOpen]);
  if (!isOpen) {
    return null;
  }

  const text = copy[kind];

  const description =
    kind === "request-changes"
      ? revisionNumber >= revisionLimit
        ? `This is your last available revision (${revisionNumber} of ${revisionLimit}). You will not be able to request further changes after this.`
        : `This will count as ${revisionNumber} of your ${revisionLimit} available revisions. The artist will be notified to revise the artwork.`
      : "Once accepted, the order is completed and you can no longer request changes.";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/25">
      <WhiteCard>
        {status === "success" ? (
          <div className="space-y-5">
            <Image
              src="/icons/success.svg"
              alt=""
              width={64}
              height={64}
              className="mx-auto"
            />
            <p className="text-center text-h3 text-primary-500">Success</p>
            <p className="text-center text-small text-primary-500">
              {text.success}
            </p>
            <Button variant="light" onClick={onCancel} className="w-full">
              Close
            </Button>
          </div>
        ) : status === "fail" ? (
          <div className="space-y-6">
            <Image
              src="/icons/fail.svg"
              alt=""
              width={64}
              height={64}
              className="mx-auto"
            />
            <p className="text-center text-h3 text-primary-500">{text.fail}</p>
            {errorMessage && (
              <p className="text-center text-small text-primary-500">
                {errorMessage}
              </p>
            )}
            <Button variant="light" onClick={onCancel} className="w-full">
              Close
            </Button>
          </div>
        ) : (
          <div className="mt-6 space-y-4.5">
            <p className="text-center text-h3 text-primary-500">{text.title}</p>
            <p className="mt-8 text-center text-small text-primary-500">
              {description}
            </p>
            <div className="mt-8 flex w-full justify-between gap-4">
              <Button variant="light" onClick={onCancel}>
                Cancel
              </Button>
              <Button
                variant="dark"
                onClick={onConfirm}
              >
                {text.confirm}
              </Button>
            </div>
          </div>
        )}
      </WhiteCard>
    </div>
  );
}

