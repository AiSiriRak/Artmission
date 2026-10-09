"use client";

import type { ReactNode } from "react";
import { useRouter } from "next/navigation";

import MainLayout from "@/components/feature/main/MainLayout";
import { Button } from "@/components/ui/Button";

interface OrderDetailsProps {
  usertype: "artist" | "customer";
  children?: ReactNode;
}

export default function OrderDetails({
  usertype,
  children,
}: OrderDetailsProps) {
  const router = useRouter();

  return (
    <MainLayout usertype={usertype}>
      <div className="mx-auto max-w-[1680px] px-6 py-10 sm:px-12 lg:px-20">
        <Button
          type="button"
          variant="light"
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

        <h1 className="mt-10 mb-8 text-h2 font-bold text-primary-500">
          Order Detail
        </h1>

        {children}
      </div>
    </MainLayout>
  );
}