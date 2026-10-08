import { Button } from "@/components/ui/Button";
import { useState } from "react";

const mockArtwork = {
  name: "The Starry Night",
  //   description:
  //     "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
  category: "Painting",
  styles: ["Abstract", "Modern"],
  minimum_deadline_days: 7,
  price: 1000,
  //   images: ["/default-avatar.png", "/default-avatar.png", "/default-avatar.png", "/default-avatar.png", "/default-avatar.png"],
};

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


export function OrderRequestSection() {
  const [deadline, setDeadline] = useState(getMinDeadlineDate(mockArtwork.minimum_deadline_days));
  const estimatedDelivery = getEstimatedDeliveryDate(deadline);

  return (
    <div className="flex justify-between items-center border border-primary-500 rounded-3xl bg-secondary-200 shadow-sm">
      {/* Order Request Section */}
      <div className="w-full m-6">
        <div className="p-4 flex flex-col flex-1 gap-6">
          {/* Todo: replace with actual artwork name load from backend */}

          {/* Artwork Info section*/}
          <div className="flex flex-col gap-2">
            <h1 className="text-h1">{mockArtwork.name}</h1>
            <div className="flex flex-wrap gap-2">
              <p className="px-4 py-1.5 bg-accent-200 text-primary-400 rounded-full text-sm font-semibold shadow-sm">
                {mockArtwork.category}
              </p>
              {mockArtwork.styles.map((style, index) => (
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
              className="border border-gray-300 rounded-md bg-white p-2"
              placeholder="Enter order name"
            />
          </div>

          <div className="flex flex-col gap-2">
            <h3 className="text-h3">Request Details</h3>
            <textarea
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
                min={getMinDeadlineDate(mockArtwork.minimum_deadline_days)}
                onChange={(e) => setDeadline(e.target.value)}
                className="border border-gray-300 rounded-md bg-white p-2"
              />
              <p className="text-sm text-gray-400">
                Minimum {mockArtwork.minimum_deadline_days} days required
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
            <p className="text-h3 font-bold">{mockArtwork.price.toLocaleString()} THB</p>
          </div>
          <div className="flex justify-between">
            <p className="text-h3">Estimated Delivery</p>
            {/* use deadline date to calurate esstimated deliverry */}
            <p className="text-h3 font-bold">{estimatedDelivery}</p>
          </div>
        </div>
        <Button variant="dark" >
          Create Order
        </Button>
      </div>
    </div>
  );
}
