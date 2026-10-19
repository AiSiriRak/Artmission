"use client";

import { useEffect, useState } from "react";
import { routes } from "@/lib/routes";

import { format, formatDistanceToNow } from "date-fns";
import { Order } from "@/lib/api/types";
import { getArtistProfile } from "@/lib/api/artists";

import { ArtworkCard } from "@/components/ui/ArtworkCard";

interface OrderCard {
  status: {
    value: string;
    label: string;
  }[];
  order: Order;
  canOpenDetails?: boolean;
}

export function OrderCard({ status, order, canOpenDetails = false }: OrderCard) {
  const [artistName, setArtistName] = useState("");
  const statusLabel =
    status.find((item) => item.value === order.status)?.label ?? order.status;

  useEffect(() => {
    async function loadArtist() {
      const artistProfile = await getArtistProfile(order.artist_id);
      setArtistName(artistProfile.artist_name);
    }
    loadArtist();
  }, [order.artist_id]);

  const statusColor: Record<string, string> = {
    PENDING: "bg-status-pending-pale",
    NOT_PAID: "bg-status-notpaid-pale",
    IN_PROCESS: "bg-status-inprocess-pale",
    SUCCESS: "bg-status-success-pale",
    CANCEL: "bg-status-cancel-pale",
  };

  return (
    <ArtworkCard
      href={canOpenDetails ? routes.order.detail(order.id) : undefined}
      imageUrl={order.deliverable_preview_url}
      imageAlt=""
      overlay={
        <div
          className={`${statusColor[order.status]} absolute z-20 right-[22px] top-[24px] flex w-30 p-1.5 border rounded-full justify-center`}
        >
          <p className="text-caption text-primary-500">{statusLabel}</p>
        </div>
      }
    >
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
          {order.name}
        </p>
        <p className=" truncate text-left text-body text-primary-500">
          {/* WILL BE REPLACE WITH ARTIST NAME */}
          {artistName}
        </p>
      </div>
      {/* Deadline */}
      <p className="text-left text-caption text-accent-500 pb-4">
        deadline -
        {order.deadline_at
          ? format(new Date(order.deadline_at), " d MMM yyyy")
          : " None"}
      </p>
    </ArtworkCard>
  );
}
