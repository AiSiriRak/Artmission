import OrderDetails from "./OrderDetails";
import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";
import type { OrderStatus } from "@/lib/api/types";

// TODO: connect with API
const mockStatus: OrderStatus = "PENDING";

function renderArtistActions(status: OrderStatus) {
  // TODO: connect with API
  switch (status) {
    case "PENDING":
      return (
        <>
          <Button type="button" className="!border w-32">
            Reject
          </Button>

          <Button type="button" variant="dark" className="w-32">
            Accept
          </Button>
        </>
      );

    case "NOT_PAID":
      return (
        <Button type="button" variant="red" className="min-w-32">
          Cancel Order
        </Button>
      );

    case "IN_PROCESS":
      return (
        <>
          <Button type="button" className="!border min-w-32">
            Postpone Order
          </Button>

          <Button type="button" variant="red" className="min-w-32">
            Cancel Order
          </Button>
        </>
      );

    case "SUCCESS":
    case "CANCEL":
      return null;
  }
}

export default function ArtistOrderDetails() {
  return (
    <OrderDetails
      usertype="artist"
      status={mockStatus}
      orderActions={renderArtistActions(mockStatus)}

      participantInfo={
        <section>
          <h2 className="mb-4 text-h3">
            Customer Info
          </h2>

          <WhiteCard
            margin=""
            padding="p-6 sm:p-8"
            roundsize="rounded-3xl"
            className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
          >
            <div className="flex flex-wrap items-center gap-6">
              <div
                aria-hidden="true"
                className="h-24 w-24 shrink-0 rounded-full bg-neutral"
              />

              <dl className="grid min-w-0 flex-1 grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-4 gap-y-2">
                <dt className="text-h3">Name</dt>
                <dd className="break-words text-body">Username</dd>

                <dt className="text-h3">Email</dt>
                <dd className="break-words text-body">
                  useremail@gmail.com
                </dd>
              </dl>
            </div>
          </WhiteCard>
        </section>
      }

    />
  );
}