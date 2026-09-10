"use client";

import { Loading } from "@/components/ui/Loading";
import { UserAccount } from "@/lib/api/types";
import { getAccount } from "@/lib/api/users";
import { useEffect, useState } from "react";

import OrderArtistPage from "@/components/feature/homepage/OrderArtistPage";
import HomePage from "@/components/feature/homepage/HomePage";
import MainLayout from "@/components/feature/main/MainLayout";

export default function Home() {
  const [user, setUser] = useState<UserAccount | null>(null);

  useEffect(() => {
    async function loadUser() {
      const accountdata = await getAccount();

      setUser(accountdata);
    }

    loadUser();
  }, []);

  if (!user) {
    return <Loading />;
  }
  return (
    <MainLayout page={"Home"} usertype={"customer"}>
      <div>(Notification)</div>
    </MainLayout>
  );
}
