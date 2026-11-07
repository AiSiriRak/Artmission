"use client";

import { Button } from "@/components/ui/Button";
import Image from "next/image";
import DownloadArtworkButton from "./DownloadArtworkButton";

interface FinalArtworkActionsProps {
  hasReviewed: boolean;
  imageUrl: string | null;
  fileName: string;
  onReview: () => void;
}

export default function FinalArtworkActions({
  hasReviewed,
  imageUrl,
  fileName,
  onReview,
}: FinalArtworkActionsProps) {
  return (
    <div className="mt-6 flex flex-col items-center gap-2">
      <div className="flex flex-wrap items-center justify-center gap-3">
        {!hasReviewed && (
          <>
            <Button
              type="button"
              variant="dark"
              onClick={onReview}
            >
              <span
                role="img"
                aria-label="edit icon"
                className="inline-block h-6 w-6 bg-white"
                style={{
                  maskImage: "url('/icons/edit.svg')",
                  maskRepeat: "no-repeat",
                  maskPosition: "center",
                  maskSize: "contain",
                }}
              />
              Review
            </Button>

            <div className="flex flex-col items-center gap-1">
              <Button
                type="button"
                disabled
              >
                <span
                  role="img"
                  aria-label="download icon"
                  className="inline-block w-6 h-6 bg-white"
                  style={{
                    maskImage: "url('/icons/download.svg')",
                    maskRepeat: "no-repeat",
                    maskPosition: "center",
                    maskSize: "contain",
                  }}
                />
                Download
              </Button>
            </div>
          </>
        )}

        {hasReviewed && imageUrl && (
          <DownloadArtworkButton
            imageUrl={imageUrl}
            fileName={fileName}
          />
        )}
      </div>

      {!hasReviewed && (
        <div className="flex flex-row items-center justify-center pt-2 gap-2">
          <Image
            src="/icons/fail.svg"
            alt=""
            width={16}
            height={16}
            className="h-4 w-4 shrink-0"
          />
          <p className="text-caption">
            Please review the artwork before downloading.
          </p>
        </div>
      )}
    </div>
  );
}