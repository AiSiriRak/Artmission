"use client"

import MainLayout from "@/components/feature/main/MainLayout";
import { Button } from "@/components/ui/Button";
import { OrderRequestSection } from "@/components/feature/order-placement/OrderRequestSection";

export default function OrderPlacementPage() {
  return (
    <MainLayout usertype="customer">
        
        <div className="p-20">
            <Button
              variant="light"
              /*onClick={onBack}*/
              className="text-button mb-10 flex items-center gap-2 !border"
            >
              <span>←</span> Back
            </Button>

            <h2 className="text-h2 mb-10 font-bold text-gray-900">Order Detail</h2>
            <OrderRequestSection />

        </div>
        </MainLayout>
  );
}