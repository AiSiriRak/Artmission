"use client";

import { useEffect, useState } from "react";
import { WhiteCard } from "@/components/ui/WhiteCard";

export default function SubmitDraft() {
  const [file, setFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState("");

  useEffect(() => {
    if (!file) return;

    const url = URL.createObjectURL(file);
    setPreviewUrl(url);

    return () => URL.revokeObjectURL(url);
  }, [file]);

  return (
    <section className="mt-6">
      <h2 className="mb-4 text-h3">Submit Draft</h2>

      <WhiteCard
        margin=""
        padding="p-6 sm:p-12"
        roundsize="rounded-3xl"
        className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
      >
        <label className="relative flex h-[294px] w-full max-w-[940px] mx-auto cursor-pointer flex-col items-center justify-center gap-5 rounded-[25px] p-6 text-secondary-600 focus-within:outline">
        
        {/* dash */}
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
            strokeDasharray="40 40" //ความยาว, ช่องว่าง
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
              if (selectedFile) setFile(selectedFile);
            }}
          />

          {previewUrl ? (
            <img
              src={previewUrl}
              alt="Draft preview"
              className="max-h-80 w-full object-contain"
            />
          ) : (
            <>
              <span className="text-h3">Add your artwork</span>
              <span
                aria-hidden="true"
                className="flex h-16 w-16 items-center justify-center rounded border border-secondary-600 text-h1"
              >
                +
              </span>
            </>
          )}
        </label>

        {/* TODO: connect with API */}
      </WhiteCard>
    </section>
  );
}