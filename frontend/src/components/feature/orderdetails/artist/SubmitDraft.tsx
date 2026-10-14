"use client";

import { useEffect, useState } from "react";
import { WhiteCard } from "@/components/ui/WhiteCard";
import { Button } from "@/components/ui/Button";
import ConfirmOrderPopup from "./popups/ConfirmOrderPopup";

export default function SubmitDraft() {
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [draft, setDraft] = useState<{
    file: File;
    previewUrl: string;
  } | null>(null);

  useEffect(() => {
    if (!draft) return;
    return () => URL.revokeObjectURL(draft.previewUrl);
  }, [draft]);

  return (
    <section className="mt-6">
      <h2 className="mb-4 text-h3">Submit Draft</h2>

      <WhiteCard
        margin=""
        padding="p-6 sm:p-12"
        roundsize="rounded-3xl"
        className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
      >
        {draft ? (
          <div className="mx-auto w-full max-w-[940px] p-6">
            <div className="relative mx-auto w-fit max-w-full">
              <button
                type="button"
                aria-label="Remove draft image"
                onClick={() => setDraft(null)}
                className="absolute top-2 right-2 bg-white shadow-md rounded-full w-6 h-6 flex items-center justify-center text-gray-500 hover:text-red-500 text-xs cursor-pointer z-10"
              >
                ✕
              </button>
              {/* Local file previews use a temporary browser object URL. */}
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={draft.previewUrl}
                alt="Draft preview"
                className="block h-auto w-auto max-h-[632px] max-w-full object-contain"
              />
            </div>
          </div>
        ) : (
          <label className="relative flex h-[294px] w-full max-w-[940px] mx-auto cursor-pointer flex-col items-center justify-center gap-5 rounded-[25px] p-6 text-secondary-600">
            <svg
              aria-hidden="true"
              className="pointer-events-none absolute inset-0 h-full w-full"
              viewBox="0 0 940 294"
              preserveAspectRatio="none"
            >
              <rect
                x="1"
                y="1"
                width="938"
                height="292"
                rx="25"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeDasharray="40 40"
                vectorEffect="non-scaling-stroke"
              />
            </svg>

            <input
              type="file"
              accept="image/*"
              aria-label="Choose draft image"
              className="sr-only"
              onChange={(event) => {
                const selectedFile = event.target.files?.[0];
                if (!selectedFile || !selectedFile.type.startsWith("image/")) return;

                setDraft({
                  file: selectedFile,
                  previewUrl: URL.createObjectURL(selectedFile),
                });
              }}
            />

            <span className="text-h3">Add your artwork</span>
            <span
              aria-hidden="true"
              className="flex h-16 w-16 items-center justify-center rounded border border-secondary-600 text-h1"
            >
              +
            </span>
          </label>
        )}

        {draft && (
          <div className="mt-1 flex justify-center">
            {/* TODO: Connect Submit to the draft API when Backend is ready. */}
            <Button type="button" variant="dark" onClick={() => setIsConfirmOpen(true)}>
              Submit
            </Button>
          </div>
        )}
      </WhiteCard>
      {isConfirmOpen && draft && (
        <ConfirmOrderPopup
          action="submit"
          onClose={() => setIsConfirmOpen(false)}
          onConfirm={() => {
            // TODO: Submit draft.file to the API before clearing the selected image.
            setIsConfirmOpen(false);
          }}
        />
      )}
    </section>
  );
}
