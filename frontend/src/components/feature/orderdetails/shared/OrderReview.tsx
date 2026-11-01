"use client";

import { WhiteCard } from "@/components/ui/WhiteCard";

import { Review, mockReviews } from "@/lib/api/reviews";

function ReviewStars({ rating, size = 24 }: { rating: number; size?: number }) {
	return (
		<div className="flex items-center gap-2">
			{[...Array(5)].map((_, index) => {
				const filled = index < rating;

				return (
					<img
						key={index}
						src={
							filled
								? "/icons/Star_filled.svg"
								: "/icons/Star_light.svg"
						}
						alt=""
						style={{
							width: filled ? size : size * 1.3,
							height: filled ? size : size * 1.3,
						}}
						className="shrink-0 object-contain"
					/>
				);
			})}
		</div>
	);
}

interface OrderReviewProps {
	review?: Review | null;
}

export default function OrderCancelSection({
	review = mockReviews[1],
}: OrderReviewProps) {
	if (!review) return null;

	return (
		<div className="flex flex-col gap-2">
			<h2 className="mb-1 text-h3">Review</h2>
			<section>
				<WhiteCard
					margin=""
					padding="p-6 sm:p-8"
					roundsize="rounded-3xl"
					className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
				>
					<dl className="flex min-w-0 flex-1 flex-col gap-2">
						<dd className="flex items-center gap-2">
							<ReviewStars rating={review.rating} size={28} />
							<span className="text-small text-primary-400/60">
								{review.rating}/5
							</span>
						</dd>
					</dl>
					<dl className="flex min-w-0 flex-1 flex-col gap-2 mt-2">
						<dd className="break-words text-body">
							{review.comment}
						</dd>
					</dl>
				</WhiteCard>
			</section>
		</div>
	);
}