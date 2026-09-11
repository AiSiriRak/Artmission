"use client";

import { useEffect, useState } from "react";
import router from "next/router";

import { UserAccount } from "@/lib/api/types";
import { getAccount } from "@/lib/api/users";
import { routes } from "@/lib/routes";

import { Loading } from "@/components/ui/Loading";
import MainLayout from "@/components/feature/main/MainLayout";

export default function Home() {
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
  return (
    <MainLayout page={"Notification"} usertype={user.role}>
      <div>(Notification)</div>
    </MainLayout>
  );
}
