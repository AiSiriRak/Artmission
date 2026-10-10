"use client";

import type { ReactNode } from "react";
import { useRouter } from "next/navigation";

import MainLayout from "@/components/feature/main/MainLayout";
import { Button } from "@/components/ui/Button";
import { WhiteCard } from "@/components/ui/WhiteCard";
import TagList from "@/components/feature/artistprofile/TagList";

import type { OrderStatus } from "@/lib/api/types";

interface OrderDetailsProps {
  usertype: "artist" | "customer";
  status: OrderStatus;
  children?: ReactNode;
  orderActions?: ReactNode;
  participantInfo?: ReactNode;
}

const statusColors: Record<OrderStatus, string> = {
  PENDING: "bg-status-pending-pale",
  NOT_PAID: "bg-status-notpaid-pale",
  IN_PROCESS: "bg-status-inprocess-pale",
  SUCCESS: "bg-status-success-pale",
  CANCEL: "bg-status-cancel-pale",
};

const statusLabels: Record<OrderStatus, string> = {
  PENDING: "PENDING",
  NOT_PAID: "NOT PAID",
  IN_PROCESS: "IN PROCESS",
  SUCCESS: "SUCCESS",
  CANCEL: "CANCEL",
};

export default function OrderDetails({
  usertype,
  status,
  children,
  orderActions,
  participantInfo,
}: OrderDetailsProps) {
  const router = useRouter();

  return (
    <MainLayout usertype={usertype}>
      <div className="mx-auto max-w-[1680px] px-6 py-10 sm:px-12 lg:px-20 text-primary-500">
        <Button
          type="button"
          className="!border w-24"
          onClick={() => router.back()}
          icon={
            <img
              src="/icons/Arrow_left.svg"
              alt=""
              className="w-4 h-4 object-contain"
            />
          }
        >
          Back
        </Button>

        <h1 className="mt-10 mb-8 text-h2">
          Order Detail
        </h1>

        {/* left side: order details */}
        <div className="grid grid-cols-1 items-start gap-10 lg:grid-cols-[minmax(0,1.85fr)_minmax(0,1fr)]">
          <section className="min-w-0">
            <h2 className="mb-4 text-h3">
              Order #1024
            </h2>

            <WhiteCard
              margin=""
              padding="p-6 sm:p-9"
              roundsize="rounded-3xl"
              className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
            >
              <h3 className="mb-4 break-words text-h1">
                Cutie Catto
              </h3>

              <div className="mb-8 flex flex-wrap items-center gap-3">
                <span className="text-h3">
                  Pet Portrait
                </span>

                <TagList items={["Digital Art"]} variant="category" />
                <TagList items={["Realism", "Sketch"]} variant="style" />
              </div>

              <h4 className="mb-3 text-h3">
                Request Detail
              </h4>

              <p className="whitespace-pre-wrap break-words text-small">
                Please draw a portrait of my cat in a realistic pencil sketch style.
                Use a simple background and keep the details of the fur and eyes.
                The final artwork should be suitable for printing.
              </p>
            </WhiteCard>

            {children}
          </section>

          {/* right side: order info */}
          <section className="min-w-0">
            <h2 className="mb-4 text-h3">
              Order Info
            </h2>

            <WhiteCard
              margin=""
              padding="p-0"
              roundsize="rounded-3xl"
              className="max-w-none overflow-hidden border border-primary-500 bg-secondary-200 shadow-none"
            >
              <div className="flex items-center justify-between gap-4 border-b border-primary-500 px-6 py-5">
                <span className="text-h3">
                  Status
                </span>

                <span
                  className={`rounded-full border border-primary-500 px-6 py-1 text-caption ${statusColors[status]}`}
                >
                  {statusLabels[status]}
                </span>
              </div>

              <dl className="space-y-3 px-6 py-6">
                <div className="flex items-baseline justify-between gap-4">
                  <dt className="text-h3">Total Amount</dt>
                  <dd className="text-body text-right">3,000 THB</dd>
                </div>

                <div className="flex items-baseline justify-between gap-4">
                  <dt className="text-h3">Order Date</dt>
                  <dd className="text-body text-right">02 Sep 2026</dd>
                </div>

                <div className="flex items-baseline justify-between gap-4">
                  <dt className="text-h3">Deadline</dt>

                  <dd className="text-right">
                    <p className="text-body">10 Sep 2026</p>
                    <p className="text-small text-neutral">Due in 7 days</p>
                  </dd>
                </div>
              </dl>

              {orderActions && (
                <div className="flex flex-wrap justify-center gap-6 border-t border-primary-500 px-6 py-4">
                  {orderActions}
                </div>
              )}
            </WhiteCard>

            {/* artist info or customer info */}
            {participantInfo && (
                <div className="mt-10">
                  {participantInfo}
                </div>
              )}
          </section>

        </div>
      </div>
    </MainLayout>
  );
}