import OrderDetails from "./OrderDetails";
import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";

export default function ArtistOrderDetails() {
  return (
    <OrderDetails
      usertype="artist"
      status="PENDING"
      orderActions={
        <>
          <Button
            type="button"
            variant="light"
            className="!border w-32"
          >
            Reject
          </Button>

          <Button
            type="button"
            variant="dark"
            className="w-32"
          >
            Accept
          </Button>
        </>
      }

      // artist info
      participantInfo={
        <section>
          <h2 className="mb-4 text-h3 font-bold text-primary-500">
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

              <dl className="grid min-w-0 flex-1 grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-4 gap-y-2 text-primary-500">
                <dt className="text-h3 font-bold">Name</dt>
                <dd className="break-words text-body">Username</dd>

                <dt className="text-h3 font-bold">Email</dt>
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