export default function PostponeNotice({ deadline }: { deadline: string }) {
  const dateLabel = new Date(deadline + "T00:00:00").toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" });
  return (
    <div role="status" className="mb-6 rounded-3xl bg-status-pending-pale p-6">
      <h3 className="text-body"><span aria-hidden="true">⌛ </span>Postpone Request Pending</h3>
      <p className="mt-2 text-small">You have requested to extend the deadline to {dateLabel}. Waiting for the customer&apos;s response (Auto declines in 48 hours).</p>
    </div>
  );
}
