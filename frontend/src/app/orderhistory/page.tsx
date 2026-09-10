"use client";

import { useEffect, useState } from "react";

import { OrderHistory } from "@/lib/api/types";
import { getOrderHistory } from "@/lib/api/orders";

import { OrderCard } from "@/components/feature/orderhistory/OrderCard";
import { SelectInput } from "@/components/ui/SelectInput";
import { CheckboxDropdown } from "@/components/ui/CheckboxDropdown";
import { Loading } from "@/components/ui/Loading";
import MainLayout from "@/components/feature/main/MainLayout";
import { TextInput } from "@/components/ui/TextInput";

const status_list = [
  { value: "PENDING", label: "PENDING" },
  { value: "NOT_PAID", label: "NOT PAID" },
  { value: "IN_PROCESS", label: "IN PROCESS" },
  { value: "SUCCESS", label: "SUCCESS" },
  { value: "CANCEL", label: "CANCEL" },
];
const sortOrder_list = [
  { value: "S2L", label: "Sooner to Later" },
  { value: "L2S", label: "Later to Sooner" },
];
export default function HomePage() {
  const [order, setOrder] = useState<OrderHistory | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [sortOrder, setSortOrder] = useState<string | null>(null);

  useEffect(() => {
    async function loadUser() {
      const orderData = await getOrderHistory();
      setOrder(orderData);
    }

    loadUser();
  }, []);

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
              <CheckboxDropdown options={status_list}></CheckboxDropdown>
            </div>
            <div className="w-48">
              <SelectInput
                value={sortOrder ? sortOrder : ""}
                onChange={(e) => {
                  setSortOrder(e.target.value);
                }}
                options={sortOrder_list}
              />
            </div>
          </div>
        </div>
        {/* Order List */}
        <div className="px-10 grid grid-cols-[repeat(auto-fit,minmax(280px,320px))] gap-6 space-y-12 item justify-between mb-20">
          {order.orders ? (
            order.orders.map((option) => (
              <OrderCard key={option.id} order={option} status={status_list} />
            ))
          ) : (
            <div />
          )}{" "}
          {order.orders ? (
            order.orders.map((option) => (
              <OrderCard key={option.id} order={option} status={status_list} />
            ))
          ) : (
            <div />
          )}
        </div>
      </div>
    </MainLayout>
  );
}
