"use client";

import { Loading } from "@/components/ui/Loading";
import { UserAccount } from "@/lib/api/types";
import { getAccount } from "@/lib/api/users";
import { useEffect, useState } from "react";

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
    <MainLayout page={"Notification"} usertype={user.role}>
      <div>(Notification)</div>
    </MainLayout>
  );
}
