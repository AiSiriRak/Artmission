"use client";

import type { ReactNode } from "react";
import { useRouter } from "next/navigation";
import MainLayout from "@/components/feature/main/MainLayout";
import { Button } from "@/components/ui/Button";

interface OrderDetailsProps {
  usertype: "artist" | "customer";
  description: ReactNode;
  info: ReactNode;
  userInfo?: ReactNode;
  notice?: ReactNode;
  children?: ReactNode;
}

export default function OrderDetails({
  usertype,
  description,
  info,
  userInfo,
  notice,
  children,
}: OrderDetailsProps) {
  const router = useRouter();

  return (
    <MainLayout usertype={usertype}>
      <div className="mx-auto max-w-[1920px] px-6 py-10 sm:px-12 lg:px-40 text-primary-500">
        <Button
          type="button"
          className="!border w-24"
          onClick={() => router.back()}
          icon={
            <img
              src="/icons/Arrow_left.svg"
              alt=""
              className="w-4 h-4 object-contain"
            />
          }
        >
          Back
        </Button>

        <h1 className="mt-6 mb-8 text-h2">Order Detail</h1>

        {notice}

        <div className="grid grid-cols-1 items-start gap-12 lg:grid-cols-[minmax(0,1.85fr)_minmax(0,1fr)]">
          <section className="min-w-0">
            {description}
            {children}
          </section>

          <section className="min-w-0">
            {info}
            {userInfo && <div className="mt-10">{userInfo}</div>}
          </section>
        </div>
      </div>
    </MainLayout>
  );
}
