import { Button } from "@/components/ui/Button";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { createOrder } from "@/lib/api/orders";
import type { Artwork } from "@/lib/api/types";
import { routes } from "@/lib/routes";

// calculate the minimum date (YYYY-MM-DD) based on the minimum_deadline_days (number)
const getMinDeadlineDate = (minDays: number): string => {
  const currentDate = new Date();
  currentDate.setDate(currentDate.getDate() + minDays + 1); // Add 1 to ensure the minimum date is at least minDays from today
  return currentDate.toISOString().split("T")[0]; // Format as YYYY-MM-DD
};

// calculate the estimate deadline by user's selected date and current date
const getEstimatedDeliveryDate = (deadline: string): string => {
   if (!deadline) return "-";

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const deadlineDate = new Date(`${deadline}T00:00:00`);
  const diffTime = deadlineDate.getTime() - today.getTime();
  const daysDiff = Math.round(diffTime / (1000 * 60 * 60 * 24));

  return `${daysDiff} days`;
};

interface OrderRequestSectionProps {
  artwork: Artwork;
}

export function OrderRequestSection({ artwork }: OrderRequestSectionProps) {
  const router = useRouter();
  const minDeadline = getMinDeadlineDate(artwork.minimum_deadline_days);
  const [deadline, setDeadline] = useState(minDeadline);
  const [orderName, setOrderName] = useState("");
  const [details, setDetails] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const estimatedDelivery = getEstimatedDeliveryDate(deadline);

// handle error message display hide when user starts typing in the input fields

  const handleCreateOrder = async () => {
    if (!orderName.trim() || !details.trim()) {
      setError("Please enter the order name and request details.");
      return;
    }
    if (!deadline || deadline < minDeadline) {
      setError(`Deadline must be at least ${artwork.minimum_deadline_days} days from today.`);
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      await createOrder({
        artwork_id: artwork.id,
        name: orderName.trim(),
        customer_description: details.trim(),
        deadline_at: new Date(`${deadline}T00:00:00`).toISOString(),
      });
      router.push(routes.order.history);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create order.");
      setSubmitting(false);
    }
  };

  return (
    <div className="flex justify-between items-center border border-primary-500 rounded-3xl bg-secondary-200 shadow-sm">
      {/* Order Request Section */}
      <div className="w-full m-6">
        <div className="p-4 flex flex-col flex-1 gap-6">
          {/* Artwork Info section*/}
          <div className="flex flex-col gap-2">
            <h1 className="text-h1">{artwork.name}</h1>
            <div className="flex flex-wrap gap-2">
              <p className="px-4 py-1.5 bg-accent-200 text-primary-400 rounded-full text-sm font-semibold shadow-sm">
                {artwork.category}
              </p>
              {(artwork.styles ?? []).map((style, index) => (
                <span
                  key={index}
                  className="px-4 py-1.5 bg-secondary-600 text-gray-900 rounded-full text-sm font-semibold shadow-sm"
                >
                  {style}
                </span>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <h3 className="text-h3">Order Name</h3>
            <input
              type="text"
              value={orderName}
              onChange={(e) => {
                setOrderName(e.target.value);
                setError(null);
              }}
              className="border border-gray-300 rounded-md bg-white p-2"
              placeholder="Enter order name"
            />
          </div>

          <div className="flex flex-col gap-2">
            <h3 className="text-h3">Request Details</h3>
            <textarea
              value={details}
              onChange={(e) => {
                setDetails(e.target.value);
                setError(null);
              }}
              className="border border-gray-300 rounded-md bg-white p-2"
              placeholder="Enter request details"
              rows={4}
            ></textarea>
          </div>

          <div className="flex gap-4">
            <h3 className="text-h3 pt-2">Deadline Date</h3>

            <div className="flex flex-col items-center gap-2">
              <input
                type="date"
                lang="en-GB"
                min={minDeadline}
                value={deadline}
                onChange={(e) => {
                  setDeadline(e.target.value);
                  setError(null);
                }}
                className="border border-gray-300 rounded-md bg-white p-2"
              />
              <p className="text-sm text-gray-400">
                Minimum {artwork.minimum_deadline_days} days required
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* create vertical line */}
      <div className="w-px self-stretch border border-primary-500 mx-6"></div>

      {/* Order Summary Section */}
      <div className="w-160 flex flex-col gap-6 m-10">
        <h1 className="text-h1">Order Summary</h1>
        <div className="flex flex-col gap-4">
          <div className="flex justify-between">
            <p className="text-h3">Total Price</p>
            <p className="text-h3 font-bold">{(artwork.price_satang / 100).toLocaleString()} THB</p>
          </div>
          <div className="flex justify-between">
            <p className="text-h3">Estimated Delivery</p>
            {/* use deadline date to calurate esstimated deliverry */}
            <p className="text-h3 font-bold">{estimatedDelivery}</p>
          </div>
        </div>
        <Button variant="dark" onClick={handleCreateOrder} disabled={submitting}>
          {submitting ? "Creating..." : "Create Order"}
        </Button>

        {error && <p className="text-sm text-center text-red-600">{error}</p>}
      </div>
    </div>
  );
}
