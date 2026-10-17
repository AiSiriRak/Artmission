"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import ArtistOrderDetails from "@/components/feature/orderdetails/artist/ArtistOrderDetails";
import CustomerOrderDetail from "@/components/feature/orderdetails/customer/CustomerOrderDetail";
import { Loading } from "@/components/ui/Loading";
import { getAccount } from "@/lib/api/users";
import { isApiError } from "@/lib/api/error";
import { routes } from "@/lib/routes";

export default function OrderDetailsPage() {
  const { orderId } = useParams<{ orderId: string }>();
  const router = useRouter();
  const [account, setAccount] = useState<{ orderId: string; role: "artist" | "customer" } | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    getAccount().then((value) => {
      if (!active) return;
      if (value.role !== "artist" && value.role !== "customer") {
        setError("บัญชีนี้ไม่สามารถเปิดออร์เดอร์ได้");
        return;
      }
      setError("");
      setAccount({ orderId, role: value.role });
    }).catch((error: unknown) => {
      if (!active) return;
      if (isApiError(error) && error.status === 401) router.replace(routes.login);
      setError(error instanceof Error ? error.message : "โหลดข้อมูลบัญชีไม่สำเร็จ");
    });
    return () => { active = false; };
  }, [orderId, router]);

  if (error) return <p role="alert" className="p-12 text-center text-error">{error}</p>;
  if (account?.orderId !== orderId) return <Loading />;
  return account.role === "artist"
    ? <ArtistOrderDetails key={orderId} />
    : <CustomerOrderDetail key={orderId} orderId={orderId} />;
}
