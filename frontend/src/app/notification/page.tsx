"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { UserAccount } from "@/lib/api/types";
import { getAccount } from "@/lib/api/users";
import { routes } from "@/lib/routes";
import { isApiError } from "@/lib/api/error";

import { Loading } from "@/components/ui/Loading";
import MainLayout from "@/components/feature/main/MainLayout";

export default function Home() {
  const router = useRouter();
  const [user, setUser] = useState<UserAccount | null>(null);

  useEffect(() => {
    async function loadUser() {
      try {
        const accountdata = await getAccount();
        setUser(accountdata);
      } catch (error: unknown) {
        if (isApiError(error) && error.status === 401) {
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
