"use client"
import CustomerOrderDetail from "@/components/feature/orderdetails/customer/CustomerOrderDetail";
import { Loading } from "@/components/ui/Loading";
import { getOrderById, type OrderDetail } from "@/lib/api/orders";
import { getAccount } from "@/lib/api/users";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import ArtistOrderDetails from "@/components/feature/orderdetails/artist/ArtistOrderDetails";

export default function OrderDetailsPage() {
  const { orderId } = useParams<{ orderId: string }>();
  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [role, setRole] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    Promise.all([getOrderById(orderId), getAccount()])
      .then(([fetched, account]) => {
        if (cancelled) return;
        setRole(account.role);
        setOrder(fetched);
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load order.");
        }
      });

    return () => {
      cancelled = true;
    };
  }, [orderId]);

  if (error) {
    return <p role="alert" className="p-12 text-center text-error">{error}</p>;
  }
  if (order && role === "customer") {
    return <CustomerOrderDetail key={order.id} order={order} />;
  }
  if (order) {
    return <ArtistOrderDetails />;
  }
  return <Loading />;
}