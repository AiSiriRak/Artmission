"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Loading } from "@/components/ui/Loading";
import { routes } from "@/lib/routes";

export default function OrderArtistPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace(routes.order.history);
  }, [router]);

  return <Loading />;
}
