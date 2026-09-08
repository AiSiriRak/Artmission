"use client";

import Image from "next/image";

import { format, formatDistanceToNow } from "date-fns";
import { Order } from "@/lib/api/types";

import { WhiteCard } from "@/components/ui/WhiteCard";

interface OrderCard {
  status: {
    value: string;
    label: string;
  }[];
  order: Order;
}

export function OrderCard({ status, order }: OrderCard) {
  const statusLabel =
    status.find((item) => item.value === order.status)?.label ?? order.status;
  const statusColor: Record<string, string> = {
    PENDING: "bg-status-pending-pale",
    NOT_PAID: "bg-status-notpaid-pale",
    IN_PROCESS: "bg-status-inprocess-pale",
    SUCCESS: "bg-status-success-pale",
    CANCEL: "bg-status-cancel-pale",
  };

  return (
    <div className="relative w-full min-w-0 hover:brightness-90">
      <div>
        {/* Status Banner */}
        <div
          className={`${statusColor[order.status]} absolute z-20 right-[22px] top-[24px] flex w-30 p-1.5 border rounded-full justify-center`}
        >
          <p className="text-caption text-primary-500">{statusLabel}</p>
        </div>
        <WhiteCard
          margin=""
          roundsize="rounded-4xl"
          padding="p-[35px]"
          className="w-full min-w-0 max-w-none"
        >
          <div className="relative w-full min-w-0 ">
            {/* Image */}
            <div className="flex w-full h-80 justify-center">
              {/* Empty Image */}
              <Image
                src={"/icons/emptyimage.svg"}
                alt={""}
                width={100}
                height={100}
                className="relative z-10"
              />
            </div>
          </div>
          <div className="space-y-2 ">
            {/* Last Update */}
            <div className="mt-1.5 flex items-center space-x-2">
              <p className="text-left text-small text-neutral">Last Updated</p>{" "}
              <div className="w-1.5 h-1.5 flex item-center bg-neutral rounded-full" />
              <p className="text-left text-small text-neutral">
                {formatDistanceToNow(new Date(order.created_at), {
                  addSuffix: true,
                })}
              </p>
            </div>
            {/* Order Name and Artist */}
            <div className="min-w-0">
              <p className="truncate text-left text-h3 text-primary-500">
                {order.id}
              </p>
              <p className=" truncate text-left text-body text-primary-500">
                {order.artist_id}
              </p>
            </div>
            {/* Deadline */}
            <p className="text-left text-caption text-accent-500">
              deadline - {format(new Date(order.created_at), "d MMM yyyy")}
            </p>
          </div>
        </WhiteCard>
      </div>
    </div>
  );
}
