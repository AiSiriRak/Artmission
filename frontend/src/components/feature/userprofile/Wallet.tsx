"use client";

import Image from "next/image";

import { Button } from "@/components/ui/Button";

interface Wallet {
  balance_satang: number;
  currency: string;
}
export function Wallet({ balance_satang, currency }: Wallet) {
  return (
    <div className="w-full max-w-md bg-secondary-300 shadow-card rounded-lg m-10 p-16">
      <div className="flex gap-3 w-full justify-center items-center fill-accent-500">
        {/* Head */}
        <Image src="/icons/wallet.svg" alt="" width={40} height={40} />
        <h2 className="text-center text-h2 text-primary-500">Your Wallet</h2>
      </div>
      <div className="mt-6 space-y-8 flex-1">
        {/* Balance text */}
        <p className="text-center text-h3 text-primary-500 font-semibold leading-none">
          balance : {(balance_satang / 100).toLocaleString("en-US")} {currency}
        </p>
        {/* ashout button */}
        <div className="flex w-full justify-center">
          <Button
            variant="accent-500"
            onClick={() => {
              console.log(
                `Cash out : ${(balance_satang / 100).toLocaleString("en-US")} ${currency}`,
              );
            }}
          >
            Cash out
          </Button>
        </div>
      </div>
    </div>
  );
}
