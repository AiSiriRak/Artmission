"use client";

import { useEffect } from "react";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";

interface ThankYouOrderPopup {
  isOpen: boolean;
  status: string;
  onClose: () => void;
}

export function ThankYouOrderPopup({ isOpen, onClose}: ThankYouOrderPopup) {
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
        className="!max-w-[420px] flex flex-col items-center"
        padding="px-[30px] py-[40px]"
      >
        {isOpen && (
            <div className="flex flex-col items-center w-full">
            {/* Success Icon */}
                <div className="mt-[40px]">
                    <img 
                        src="/icons/success_pink.svg" 
                        alt="Success" 
                        className="w-[64px] h-[64px] object-contain" 
                    />
                </div>
            {/* Title */}
            <p className="mt-[25px] text-center text-h3 text-primary-500">
                 Thanks For Your Order!
            </p>
            
            {/* Description */}
            <p className="mt-[25px] text-center text-small text-primary-500 w-full">
              Your commission has been successfully completed. You can download the final artwork.
            </p>
            {/* Buttons */}
            <div className="mt-[25px] flex w-full justify-center">
              <Button variant="transparent" className="w-full" onClick={onClose}>
                Close
              </Button>
            </div>
          </div>
        )}
      </WhiteCard>
    </div>
  );
}
