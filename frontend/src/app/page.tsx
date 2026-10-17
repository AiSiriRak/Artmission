"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { UserAccount } from "@/lib/api/types";
import { getAccount } from "@/lib/api/users";
import { routes } from "@/lib/routes";
import { isApiError } from "@/lib/api/error";
import { clearAccessToken, getAccessToken } from "@/lib/token";
import { Button } from "@/components/ui/Button";

import OrderArtistPage from "@/components/feature/homepage/OrderArtistPage";
import HomePage from "@/components/feature/homepage/HomePage";
import { Loading } from "@/components/ui/Loading";

export default function Home() {
  const router = useRouter();
  const [user, setUser] = useState<UserAccount | null>(null);
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    if (!getAccessToken()) {
      router.replace(routes.login);
      return;
    }

    let active = true;
    const timeout = window.setTimeout(() => {
      if (!active) return;
      setErrorMessage("รอข้อมูลบัญชีนานเกินไป กรุณาตรวจสอบว่า Backend รันอยู่แล้วลองใหม่");
      active = false;
    }, 15000);

    async function loadUser() {
      try {
        const accountdata = await getAccount();
        if (!active) return;
        if (!accountdata) {
          throw new Error("API ไม่ได้ส่งข้อมูลบัญชีกลับมา");
        }
        setUser(accountdata);
      } catch (error: unknown) {
        if (!active) return;
        if (isApiError(error) && error.status === 401) {
          clearAccessToken();
          router.replace(routes.login);
          return;
        }
        setErrorMessage(
          "โหลดข้อมูลบัญชีไม่สำเร็จ กรุณาตรวจสอบ Backend แล้วลองใหม่",
        );
      } finally {
        window.clearTimeout(timeout);
      }
    }
    void loadUser();

    return () => {
      active = false;
      window.clearTimeout(timeout);
    };
  }, [router]);

  if (errorMessage) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center gap-6 px-6">
        <p role="alert" className="text-body text-error">{errorMessage}</p>
        <Button type="button" onClick={() => window.location.reload()}>
          ลองใหม่
        </Button>
      </main>
    );
  }

  if (!user) {
    return <Loading />;
  }

  if (user.role !== "customer") {
    return <OrderArtistPage />;
  }

  return (
    <Suspense fallback={<Loading />}>
      <HomePage />
    </Suspense>
  );
}
