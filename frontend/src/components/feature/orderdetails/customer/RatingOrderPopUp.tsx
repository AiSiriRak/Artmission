"use client";

import { useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";
import { TextInput } from "@/components/ui/TextInput";

interface RatingOrderPopup {
  isOpen: boolean;
  status: string;
  onCancel: () => void;
  onConfirm: () => void;
}

export function RatingOrderPopup({ isOpen, onConfirm }: RatingOrderPopup) {
    const [rating, setRating] = useState(0); 
    const [review, setReview] = useState("");
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
        className="!max-w-[600px] w-full flex flex-col items-center"
        padding="p-[40px]"
      >
        {isOpen && (
          <div className="flex flex-col items-center w-full">
            {/* Title */}
            <p className="text-center text-h2 text-primary-500">
              Review Your Order
            </p>
            {/* Star */}
            <div className="mt-[24px]">
                <div className="flex items-center gap-2">
                    {[...Array(5)].map((_, index) => (
                        <img
                            key={index}
                            src={
                            index < rating
                                ? "/icons/Star_filled.svg"
                                : "/icons/Star_light.svg"
                            }
                            alt={index < rating ? "Filled Star" : "Light Star"}
                            className="cursor-pointer w-[40px] h-[40px] object-contain object-center transition-transform hover:scale-110"
                            onClick={() => setRating(index + 1)}
                        />
                    ))}
                </div>
            </div>
            {/* TextInput */}
            <div className="mt-[24px]">
                <TextInput
                    multiline
                    value={review}
                    onChange={setReview}
                    placeholder="Describe your Review"
                    style={{ width: '520px', height: '146px' }}
                    className="rounded-lg border-2 border-primary-500 px-4 py-3 text-small placeholder:text-neutral-500"
                    />
            </div>
            {/* Buttons */}
            <div className="mt-[20px] flex w-full justify-center">
              <Button variant="accent-500" className="w-full py-[12px]" onClick={onConfirm}>
                Submit
              </Button>
            </div>
          </div>
        )}
      </WhiteCard>
    </div>
  );
}
