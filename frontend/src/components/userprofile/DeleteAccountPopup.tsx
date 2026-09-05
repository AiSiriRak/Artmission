"use client";

import { useEffect } from "react";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";

interface DeleteAccountPopup {
  isOpen: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

export function DeleteAccountPopup({
  isOpen,
  onCancel,
  onConfirm,
}: DeleteAccountPopup) {
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

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/25">
      <WhiteCard>
        {/* Title */}
        <p className="text-center text-h3 text-primary-500">Delete Account</p>
        {/* Description */}
        <p className="mt-8 text-center text-small text-primary-500">
          This action cannot be undone. This will permanently delete your entire
          account. You can no longer access your previous orders, and all your
          order history will be deleted.
        </p>
        {/* Buttons */}
        <div className="mt-8 flex w-full justify-between">
          <Button variant="light" onClick={onCancel}>
            Cancel
          </Button>
          <Button variant="red" onClick={onConfirm}>
            Delete Account
          </Button>
        </div>
      </WhiteCard>
    </div>
  );
}
