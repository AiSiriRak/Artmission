"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { OrderHistory, OrderSortField, OrderSortOrder } from "@/lib/api/types";
import { getOrderHistory } from "@/lib/api/orders";
import { routes } from "@/lib/routes";
import { OrderStatus } from "@/lib/api/types";

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
  const [sortOrder, setSortOrder] = useState<string | null>("deadline-asc");
  const selectedSortOrder = sortOrder_list.find(
    (option) => option.value === sortOrder,
  );
  const [status, setStatus] = useState<OrderStatus[]>(
    status_list.map((option) => option.value as OrderStatus),
  );

  useEffect(() => {
    async function loadUser() {
      try {
        const orderdata = await getOrderHistory({
          status: status,
          sort: selectedSortOrder?.sort,
          order: selectedSortOrder?.order,
        });
        setOrder(orderdata);
      } catch (error: any) {
        if (error.status === 401) {
          router.replace(routes.login);
        }
      }
    }
    loadUser();
  }, [status, sortOrder, router]);

  if (!order) {
    return <Loading />;
  }

  return (
    <MainLayout page={"Order History"} usertype={"customer"}>
      <div className="min-h-screen">
        <div className="flex items-center justify-between m-10">
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
                value={sortOrder || "deadline-asc"}
                onChange={(e) => {
                  setSortOrder(e.target.value);
                }}
                options={sortOrder_list}
              />
            </div>
          </div>
        </div>
        {/* Order List */}
        <div className="relative">
          <div className="px-10 grid grid-cols-[repeat(auto-fit,minmax(280px,320px))] gap-12 space-y-6  item justify-start mb-20">
            {order.orders ? (
              order.orders.map((option) => (
                <OrderCard
                  key={option.id}
                  order={option}
                  status={status_list}
                />
              ))
            ) : (
              <></>
            )}
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
