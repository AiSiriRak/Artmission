// Personal Information Card

"use client";

import { useState } from "react";

import { WhiteCard } from "@/components/ui/WhiteCard";
import { Button } from "@/components/ui/Button";

interface DeleteAccountCard {
  onDelete: () => void;
  disabled: boolean;
}

export function DeleteAccountCard({ disabled, onDelete }: DeleteAccountCard) {
  return (
    <WhiteCard>
      <h2 className="text-center text-h3 text-error">Delete Account</h2>
      <div className="mt-6 space-y-5">
        <Button
          variant="red"
          disabled={disabled}
          className="w-full"
          onClick={onDelete}
        >
          Delete Account
        </Button>
      </div>
    </WhiteCard>
  );
}
