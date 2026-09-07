"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";

interface DeleteAccountPopup {
  isOpen: boolean;
  status: string;
  onCancel: () => void;
  onConfirm: () => void;
}

export function DeleteAccountPopup({
  isOpen,
  onCancel,
  onConfirm,
  status = "default",
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
        {status == "success" ? (
          // === Deleted Success ===
          <div className="space-y-5">
            {" "}
            {/* Icon */}
            <Image
              src="/icons/success.svg"
              alt=""
              width={64}
              height={64}
              className="mx-auto"
            />
            {/* Title */}
            <p className="text-center text-h3 text-primary-500">
              Account Deleted
            </p>
            {/* Description */}
            <p className="text-center text-small text-primary-500">
              Your account has been successfully deleted.
            </p>
            {/* Button */}
            <Button variant="light" onClick={onCancel} className="w-full">
              Close
            </Button>
          </div>
        ) : (
          <div>
            {status == "fail" ? (
              // === Deleted Fail ===
              <div className="space-y-6">
                {/* Icon */}
                <Image
                  src="/icons/fail.svg"
                  alt=""
                  width={64}
                  height={64}
                  className="mx-auto"
                />
                {/* Title */}
                <p className="text-center text-h3 text-primary-500">
                  Unable to Delete Account
                </p>
                {/* Description */}
                <p className="   text-center text-small text-primary-500">
                  Your account cannot be deleted at the moment, as there are
                  still ongoing orders.
                </p>
                {/* Button */}
                <Button variant="light" onClick={onCancel} className="w-full">
                  Close
                </Button>
              </div>
            ) : (
              // === Deleted Confirm ===
              <div className="mt-6 space-y-4.5">
                {/* Title */}
                <p className="text-center text-h3 text-primary-500">
                  Delete Account
                </p>
                {/* Description */}
                <p className="mt-8 text-center text-small text-primary-500">
                  This action cannot be undone. This will permanently delete
                  your entire account. You can no longer access your previous
                  orders, and all your order history will be deleted.
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
              </div>
            )}
          </div>
        )}
      </WhiteCard>
    </div>
  );
}
