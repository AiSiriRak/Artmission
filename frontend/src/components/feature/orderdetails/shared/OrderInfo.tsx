import type { ReactNode } from "react";
import type { OrderStatus } from "@/lib/api/types";
import { WhiteCard } from "@/components/ui/WhiteCard";

interface OrderInfoProps {
  status: OrderStatus;
  priceSatang: number;
  orderDate: string;
  deadlineDate: string;
  deadlineNote?: string;
  revisions?: { used: number; limit: number };
  actions?: ReactNode;
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

export default function OrderInfo({
  status,
  priceSatang,
  orderDate,
  deadlineDate,
  deadlineNote,
  revisions,
  actions,
}: OrderInfoProps) {
  return (
    <>
      <h2 className="mb-4 text-h3">Order Info</h2>
      <WhiteCard
        margin=""
        padding="p-0"
        roundsize="rounded-3xl"
        className="max-w-none overflow-hidden border border-primary-500 bg-secondary-200 shadow-none"
      >
        <div className="flex items-center justify-between gap-4 border-b border-primary-500 px-6 py-5">
          <span className="text-h3">Status</span>
          <span className={`rounded-full border border-primary-500 px-6 py-1 text-caption ${statusColors[status]}`}>
            {statusLabels[status]}
          </span>
        </div>
        <dl className="space-y-3 px-6 py-6">
          <div className="flex items-baseline justify-between gap-4">
            <dt className="text-h3">Total Amount</dt>
            <dd className="text-body text-right">{(priceSatang / 100).toLocaleString("en-US")} THB</dd>
          </div>
          <div className="flex items-baseline justify-between gap-4">
            <dt className="text-h3">Order Date</dt>
            <dd className="text-body text-right">{orderDate}</dd>
          </div>
          <div className="flex items-baseline justify-between gap-4">
            <dt className="text-h3">Deadline</dt>
            <dd className="text-right">
              <p className="text-body">{deadlineDate}</p>
              {deadlineNote && <p className="text-small text-neutral">{deadlineNote}</p>}
            </dd>
          </div>
        </dl>
        {status === "IN_PROCESS" && revisions && (
          <div className="flex items-baseline justify-between gap-4 border-t border-primary-500 px-6 py-5">
            <span className="text-h3">Revisions</span>
            <span className="text-body">{revisions.used}/{revisions.limit} used</span>
          </div>
        )}
        {actions && (
          <div className={`flex flex-wrap justify-center gap-6 px-6 pb-4 ${status === "IN_PROCESS" && revisions ? "" : "border-t border-primary-500 pt-4"}`}>
            {actions}
          </div>
        )}
      </WhiteCard>
    </>
  );
}
