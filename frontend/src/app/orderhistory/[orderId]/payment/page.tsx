"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import MainLayout from "@/components/feature/main/MainLayout";
import CustomerPayment from "@/components/feature/payment/CustomerPayment";
import { Loading } from "@/components/ui/Loading";
import { getOrderById, type OrderDetail } from "@/lib/api/orders";
import { getAccount } from "@/lib/api/users";
import { routes } from "@/lib/routes";

export default function CustomerPaymentPage() {
  const { orderId } = useParams<{ orderId: string }>();
  const [result, setResult] = useState<{ orderId: string; order?: OrderDetail; error?: string } | null>(null);

  useEffect(() => {
    let active = true;
    Promise.all([getOrderById(orderId, { cache: "no-store" }), getAccount()])
      .then(([order, account]) => {
        if (!active) return;
        if (account.role !== "customer") {
          setResult({ orderId, error: "This payment page is only available to customers." });
        } else if (order.status !== "NOT_PAID") {
          setResult({ orderId, error: "This order is no longer awaiting payment." });
        } else {
          setResult({ orderId, order });
        }
      })
      .catch((error: unknown) => {
        if (active) setResult({ orderId, error: error instanceof Error ? error.message : "Could not load payment details." });
      });
    return () => { active = false; };
  }, [orderId]);

  if (!result || result.orderId !== orderId) return <Loading />;
  if (result.order) return <CustomerPayment key={orderId} order={result.order} />;
  return (
    <MainLayout usertype="customer">
      <div className="mx-auto max-w-3xl space-y-6 px-6 py-12">
        <h1 className="text-h2">Payment</h1>
        <p role="alert" className="text-body">{result.error}</p>
        <Link href={routes.order.detail(orderId)} className="underline">Back to order</Link>
      </div>
    </MainLayout>
  );
}
