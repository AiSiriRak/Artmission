"use client";

import { Button } from "@/components/ui/Button";
import { WhiteCardBorder } from "@/components/ui/WhiteCardBorder";

interface DeleteAccountCard {
  onDelete: () => void;
  disabled: boolean;
}

export function DeleteAccountCard({ disabled, onDelete }: DeleteAccountCard) {
  return (
    <WhiteCardBorder>
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
    </WhiteCardBorder>
  );
}
