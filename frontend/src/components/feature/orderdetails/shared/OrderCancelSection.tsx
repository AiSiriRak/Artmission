"use client";

import { WhiteCard } from "@/components/ui/WhiteCard";

import type { OrderCancellation } from "@/lib/api/order-details";
import { CUSTOMER_REASONS } from "../customer/CustomerCancelOrderPopUp";
import { reasons as ARTIST_REASONS } from "../artist/popups/CancelOrderPopup";

type CancelActor = "customer" | "artist" | "neutral";

interface CancelReasonInfo {
	text: string;
	actor: CancelActor;
}

// format cancel reason to display-friendly text and its tone
// neutral: Mutual agreement, Other, Artist not confirm, Overdue deadline
// customer: Placed order by mistake, No longer needed, Artist cannot meet new requirements, Artist unresponsive, Taking too long
// artist: Unable to complete, Unable to meet deadline, Customer change requirements. Customer unresponsive, Personal or Technical issue
function formatCancelReason(reason: string): CancelReasonInfo {
	switch (reason) {
		// neutral
		case "MUTUAL_AGREEMENT":
			return {
				text: "Both participants agreed to cancel",
				actor: "neutral",
			};
		case "OTHER":
			return {
				text: "Other reason",
				actor: "neutral",
			};
		case "ARTIST_DID_NOT_CONFIRM_IN_TIME":
			return {
				text: "Artist did not confirm the order in time.",
				actor: "neutral",
			};
		case "NOT_SUBMITTED_WITHIN_DEADLINE":
			return {
				text: "Artist did not submit the deliverable on time.",
				actor: "neutral",
			};
	}

	// customer
	const customerMatch = CUSTOMER_REASONS.find(
		(item) => item.value === reason,
	);

	if (customerMatch) {
		return {
			text: customerMatch.label,
			actor: "customer",
		};
	}

	// artist
	const artistMatch = ARTIST_REASONS.find(
		(item) => item.value === reason,
	);

	if (artistMatch) {
		return {
			text: artistMatch.label,
			actor: "artist",
		};
	}

	return {
		text: reason || "Not specified",
		actor: "neutral",
	};
}

const CANCEL_STATEMENTS: Record<CancelActor, string> = {
	neutral: "This order has been cancelled.",
	customer: "The customer cancelled this order.",
	artist: "The artist cancelled this order.",
};

const mockCancellation = {
	order_id: "642edf72-357c-4419-8816-fd64d80574ca",
	reason: "PLACED_ORDER_BY_MISTAKE",
	explanation:
		"I am so sorry bro.",
	created_at: "",
};

interface OrderCancelSectionProps {
	cancellation?: OrderCancellation | null;
}

export default function OrderCancelSection({
	cancellation = mockCancellation,
}: OrderCancelSectionProps) {
	if (!cancellation) return null;

	const cancelReason = formatCancelReason(cancellation.reason);
	const cancelStatement = CANCEL_STATEMENTS[cancelReason.actor];

	return (
		<section className="mb-10">
			<WhiteCard
				margin=""
				padding="p-6 sm:p-8"
				roundsize="rounded-3xl"
				className="max-w-none bg-status-cancel-pale shadow-none"
			>
				<div className="flex items-start gap-4">
					<img
						src="/icons/edit.svg" // TODO: add bell-light icon
						alt=""
						className="mt-1 h-6 w-6 shrink-0"
					/>
					<div className="min-w-0 flex-1">
						<div className="flex flex-wrap items-baseline gap-x-3">
							<h3 className="text-h3">{cancelStatement}</h3>
						</div>
						<dl className="mt-6 space-y-3">
							<div className="flex flex-wrap items-baseline gap-x-2">
								<dt className="text-h3">Reason: </dt>
								<dd className="text-body">{cancelReason.text}</dd>
							</div>
							<div>
								<dt className="text-h3">Explanation</dt>
								<dd className="text-body">{cancellation.explanation}</dd>
							</div>
						</dl>
					</div>
				</div>
			</WhiteCard>
		</section>
	);
}