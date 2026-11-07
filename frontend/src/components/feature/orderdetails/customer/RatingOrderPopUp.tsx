"use client";

import { useEffect, useState } from "react";
import Image from "next/image";

import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";
import { TextInput } from "@/components/ui/TextInput";

interface RatingOrderPopupProps {
  isOpen: boolean;
  status: string;
  onCancel: () => void;
  onConfirm: (rating: number, comment: string | null) => void;
}

export function RatingOrderPopup({ isOpen, onConfirm }: RatingOrderPopupProps) {
  const [rating, setRating] = useState(0);
  const [review, setReview] = useState("");
  const [error, setError] = useState("");

  const handleConfirm = () => {
    if (rating === 0) {
      setError("Please select a rating.");
      return;
    }

    setError("");
    onConfirm(rating, review.trim() || null);
  };

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
                  <Image
                    key={index}
                    src={
                      index < rating
                        ? "/icons/Star_filled.svg"
                        : "/icons/Star_light.svg"
                    }
                    alt={index < rating ? "Filled Star" : "Light Star"}
                    width={40}
                    height={40}
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

            {error && (
              <p className="mt-2 text-small text-error">
                {error}
              </p>
            )}

            {/* Buttons */}
            <div className="mt-[20px] flex w-full justify-center">
              <Button
                variant="accent-500"
                className="w-full py-[12px]"
                onClick={handleConfirm}
                disabled={rating === 0}
              >
                Submit
              </Button>
            </div>
          </div>
        )}
      </WhiteCard>
    </div>
  );
}
