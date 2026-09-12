"use client";

import Image from "next/image";
import { useEffect } from "react";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";

interface LogoutPopup {
  isOpen: boolean;
  status: string;
  onCancel: () => void;
  onConfirm: () => void;
}

export function LogoutPopup({ isOpen, onCancel, onConfirm }: LogoutPopup) {
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
              Are You Logging Out?
            </p>
            {/* Description */}
            <p className="mt-8 text-center text-small text-primary-500 w-50">
              After you log out, you can always log back in at any time.
            </p>
            {/* Buttons */}
            <div className="mt-8 flex w-full justify-between">
              <Button variant="light" onClick={onCancel}>
                Cancel
              </Button>
              <Button variant="red" onClick={onConfirm}>
                Log out
              </Button>
            </div>
          </div>
        )}
      </WhiteCard>
    </div>
  );
}
