"use client";

import { useState } from "react";
import Image from "next/image";

import { Button } from "@/components/ui/Button";
import { WhiteCardBorder } from "@/components/ui/WhiteCardBorder";

interface Wallet {
  balance_satang: number;
  currency: string;
}
export function Wallet({ balance_satang, currency }: Wallet) {
  const [isCashout, setIsCashout] = useState(false);
  const formatPrice = (balance_satang: number) => (balance_satang / 100).toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  return (
    <WhiteCardBorder>
      <div className="flex gap-3 w-full justify-center items-center fill-accent-500">
        {/* Head */}
        <Image src="/icons/wallet.svg" alt="" width={40} height={40} />
        <h2 className="text-center text-h2 text-primary-500">Your Wallet</h2>
      </div>
      <div className="mt-6 space-y-8 flex-1">
        {/* Balance text */}
        <p className="text-center text-h3 text-primary-500 font-semibold leading-none">
          balance : {formatPrice(balance_satang)} {currency}
        </p>
        {/* ashout button */}
        <div className="flex flex-col w-full justify-center">
          <Button
            variant="accent-500"
            onClick={() => {
              setIsCashout(!isCashout);
              console.log(
                `Cash out : ${formatPrice(balance_satang)} ${currency}`,
              );
            }}
          >
            Cash out
          </Button>
          {isCashout && (
            <span className="flex mt-1 text-small text-neutral justify-center">
              ( Not available yet )
            </span>
          )}
        </div>
      </div>
    </WhiteCardBorder>
  );
}
