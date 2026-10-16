"use client";

import { useEffect } from "react";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";

interface CustomerConfirmPopUp {
  isOpen: boolean;
  title: string;
  description: string;
  cancelText: string;
  confirmText: string;
  onCancel: () => void;
  onConfirm: () => void;
}

export function CustomerConfirmPopUp({ isOpen, title, description, cancelText, confirmText, onCancel, onConfirm }: CustomerConfirmPopUp) {
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
      <WhiteCard
        className="!w-max flex flex-col items-center"
        padding="px-15 py-8"
      >
        {isOpen && (
          <div className="mt-6 space-y-1 flex flex-col items-center">
            {/* Title */}
            <p className="text-center text-h3 text-primary-500">
              {title}
            </p>
            {/* Description */}
            <p className="mt-8 text-center text-small text-primary-500 w-50">
              {description}
            </p>
            {/* Buttons */}
            <div className="mt-8 flex w-full justify-between">
              <Button variant="light" onClick={onCancel}>
                {cancelText}
              </Button>
              <Button variant="dark" onClick={onConfirm}>
                {confirmText}
              </Button>
            </div>
          </div>
        )}
      </WhiteCard>
    </div>
  );
}
