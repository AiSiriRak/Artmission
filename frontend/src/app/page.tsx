"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { UserAccount } from "@/lib/api/types";
import { getAccount } from "@/lib/api/users";
import { routes } from "@/lib/routes";

import OrderArtistPage from "@/components/feature/homepage/OrderArtistPage";
import HomePage from "@/components/feature/homepage/HomePage";
import { Loading } from "@/components/ui/Loading";

export default function Home() {
  const router = useRouter();
  const [user, setUser] = useState<UserAccount | null>(null);

  useEffect(() => {
    async function loadUser() {
      try {
        const accountdata = await getAccount();
        setUser(accountdata);
      } catch (error: any) {
        if (error.status === 401) {
          router.replace(routes.login);
        }
      }
    }
    loadUser();
  }, [router]);

  if (!user) {
    return <Loading />;
  }
  return user.role == "customer" ? <HomePage /> : <OrderArtistPage />;
}
