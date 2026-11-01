"use client";

import { Button } from "@/components/ui/Button";
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
	if (!hasReviewed) {
		return (
			<div className="mt-6 flex flex-col items-center gap-2">
				<Button
					type="button"
					variant="dark"
					onClick={onReview}
				>
					<span
						role="img"
						aria-label="edit icon"
						className="inline-block w-6 h-6 bg-white"
						style={{
							maskImage: "url('/icons/edit.svg')",
							maskRepeat: "no-repeat",
							maskPosition: "center",
							maskSize: "contain",
						}}
					/>
					Review
				</Button>
			</div>
		);
	}

	if (!imageUrl) return null;

	return (
		<DownloadArtworkButton
			imageUrl={imageUrl}
			fileName={fileName}
		/>
	);
}