"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";

import { OrderHistory, OrderSortField, OrderSortOrder } from "@/lib/api/types";
import { getOrderHistory } from "@/lib/api/orders";
import { routes } from "@/lib/routes";
import { startAutoRefresh } from "@/lib/auto-refresh";
import { getAccount } from "@/lib/api/users";
import { OrderStatus } from "@/lib/api/types";
import { isApiError } from "@/lib/api/error";

import { OrderCard } from "@/components/feature/orderhistory/OrderCard";
import { SelectInput } from "@/components/ui/SelectInput";
import { CheckboxDropdown } from "@/components/ui/CheckboxDropdown";
import { Loading } from "@/components/ui/Loading";
import MainLayout from "@/components/feature/main/MainLayout";

export default function HomePage() {
  const router = useRouter();

  const status_list: { value: OrderStatus; label: string }[] = [
    { value: "PENDING", label: "PENDING" },
    { value: "NOT_PAID", label: "NOT PAID" },
    { value: "IN_PROCESS", label: "IN PROCESS" },
    { value: "SUCCESS", label: "SUCCESS" },
    { value: "CANCEL", label: "CANCEL" },
  ];

  type SortOrderOption = {
    value: string;
    sort: OrderSortField;
    order: OrderSortOrder;
    label: string;
  };

  const sortOrder_list: SortOrderOption[] = [
    {
      value: "update-asc",
      sort: "updated_at",
      order: "desc",
      label: "None",
    },
    {
      value: "deadline-asc",
      sort: "deadline",
      order: "asc",
      label: "Sooner to Later",
    },
    {
      value: "deadline-desc",
      sort: "deadline",
      order: "desc",
      label: "Later to Sooner",
    },
  ];

  const [order, setOrder] = useState<OrderHistory | null>(null);
  const [role, setRole] = useState<"artist" | "customer" | null>(null);
  const [accountError, setAccountError] = useState("");

  useEffect(() => {
    let active = true;
    getAccount().then((account) => {
      if (!active) return;
      if (account.role !== "artist" && account.role !== "customer") throw new Error("บัญชีนี้ไม่สามารถเปิดออร์เดอร์ได้");
      setRole(account.role);
    }).catch((error: unknown) => {
      if (!active) return;
      if (isApiError(error) && error.status === 401) router.replace(routes.login);
      setAccountError(error instanceof Error ? error.message : "โหลดบัญชีไม่สำเร็จ");
    });
    return () => { active = false; };
  }, [router]);
  const [sortOrder, setSortOrder] = useState<string | null>("update-asc");
  const selectedSortOrder = sortOrder_list.find(
    (option) => option.value === sortOrder,
  );
  const [status, setStatus] = useState<OrderStatus[]>(
    status_list.map((option) => option.value as OrderStatus),
  );

  useEffect(() => {
    let active = true;
    let loading = false;
    const controller = new AbortController();
    async function loadUser() {
      if (loading) return;
      loading = true;
      try {
        const orderdata = await getOrderHistory({
          status: status,
          sort: selectedSortOrder?.sort,
          order: selectedSortOrder?.order,
        }, { signal: AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]), cache: "no-store" });
        if (active) setOrder(orderdata);
      } catch (error: unknown) {
        if (active && isApiError(error) && error.status === 401) {
          router.replace(routes.login);
        }
      } finally {
        loading = false;
      }
    }
    void loadUser();
    const stopRefresh = startAutoRefresh(() => { void loadUser(); });
    return () => {
      active = false;
      controller.abort();
      stopRefresh();
    };
  }, [
    status,
    sortOrder,
    router,
    selectedSortOrder?.sort,
    selectedSortOrder?.order,
  ]);

  if (accountError) return <p role="alert" className="p-12 text-center text-error">{accountError}</p>;
  if (!order || !role) {
    return <Loading />;
  }
  return (
    <MainLayout usertype={role}>
      <div className="min-h-screen w-full px-4 py-6 sm:px-6 md:px-10 md:py-8">
        <div className="mb-5 flex items-center justify-between md:mb-6">
          <p className="truncate text-primary-500 text-h2 ">Recently Orders</p>
          <div className="flex space-x-8">
            <div className="w-48">
              <p className="text-small">Status</p>
              <CheckboxDropdown
                options={status_list}
                onChange={(selected) => {
                  setStatus(selected as OrderStatus[]);
                }}
              />
            </div>
            <div className="w-48">
              <p className="text-small">Deadline</p>
              <SelectInput
                buttonClassName="mt-1"
                value={sortOrder || "deadline-asc"}
                onChange={(value) => {
                  setSortOrder(value);
                }}
                options={sortOrder_list}
              />
            </div>
          </div>
        </div>
        {/* Order List */}
        {status.length != 0 && (
          <div className="relative">
            <div className="grid grid-cols-[repeat(auto-fit,minmax(280px,320px))] justify-start gap-12">
              {order.orders &&
                order.orders.length > 0 &&
                order.orders.map((option) => (
                  <OrderCard
                    key={option.id}
                    order={option}
                    canOpenDetails
                    status={status_list}
                  />
                ))}
            </div>
          </div>
        )}
        {(!order.orders || order.orders.length == 0 || status.length == 0) && (
          <div className="w-full relative flex flex-col items-center justify-center">
            <Image src="/icons/emptydoc.svg" alt={""} width={96} height={96} />
            <p className="text-h1 text-neutral">No orders yet</p>
            <p className="text-body text-neutral">
              Your commissioned artwork will appear here once you place an
              order.
            </p>
          </div>
        )}
      </div>
    </MainLayout>
  );
}
