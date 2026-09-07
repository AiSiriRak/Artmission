// Bank Account Card

"use client";

import { useState } from "react";

import { User } from "@/types/user";

import { WhiteCard } from "@/components/ui/WhiteCard";
import { TextInput } from "@/components/ui/TextInput";
import { Button } from "@/components/ui/Button";
import { SelectInput } from "@/components/ui/SelectInput";

interface BankAccountCard {
  user: User;
  bankOptions: {
    value: string;
    label: string;
  }[];
  isEditing: boolean;
  disabled: boolean;
  onEdit: () => void;
  onCancel: () => void;
  onSave: (updatedUser: User) => void;
}

export function BankAccountCard({
  user,
  bankOptions,
  isEditing,
  disabled,
  onEdit,
  onCancel,
  onSave,
}: BankAccountCard) {
  const [bankName, setBankName] = useState(user.bank.name);
  const [accountHolder, setAccountHolder] = useState(user.bank.accountHolder);
  const [accountNumber, setAccountNumber] = useState(user.bank.accountNumber);

  const handleSave = () => {
    const updatedUser: User = {
      ...user,
      bank: {
        name: bankName,
        accountHolder,
        accountNumber,
      },
    };

    onSave(updatedUser);
  };

  const handleCancel = () => {
    setBankName(user.bank.name);
    setAccountHolder(user.bank.accountHolder);
    setAccountNumber(user.bank.accountNumber);

    onCancel();
  };

  return (
    <WhiteCard>
      <h2 className="text-center text-h3 text-primary-500">
        User Bank Account Info
      </h2>
      {isEditing ? (
        // === EDIT MODE ===
        <div className="mt-6 space-y-4.5">
          {/* Bank */}
          <div>
            <label className="text-small text-primary-500">Name</label>
            <SelectInput
              value={bankName}
              onChange={(e) => setBankName(e.target.value)}
              options={bankOptions}
            />
          </div>
          {/* Account holder */}
          <div>
            <label className="text-small text-primary-500">
              Account holder name
            </label>
            <TextInput value={accountHolder} onChange={setAccountHolder} />
          </div>
          {/* Account number */}
          <div>
            <label className="text-small text-primary-500">
              Account holder number
            </label>
            <TextInput value={accountNumber} onChange={setAccountNumber} />
          </div>

          {/* Buttons */}

          <div className="flex w-full justify-between gap-2 pt-2">
            <Button variant="light" icon={<></>} onClick={handleCancel}>
              Cancel
            </Button>
            <Button variant="dark" icon={<></>} onClick={handleSave}>
              Save
            </Button>
          </div>
        </div>
      ) : (
        // === DISPLAY MODE ===
        <div className="mt-6 space-y-8">
          {/* Name */}
          <div className="space-y-3">
            <label className="text-small text-primary-500">Bank</label>

            <p className="mt-1 text-body indent-2 text-primary-500">
              {user.bank.name}
            </p>
          </div>
          <div className="space-y-3">
            <label className="text-small text-primary-500">
              Account holder name
            </label>

            <p className="mt-1 text-body indent-2 text-primary-500">
              {user.bank.accountHolder}
            </p>
          </div>
          <div className="space-y-3">
            <label className="text-small text-primary-500">
              Account holder number
            </label>

            <p className="mt-1 text-body indent-2 text-primary-500">
              {user.bank.accountNumber}
            </p>
          </div>

          {/* Edit Button */}
          {!isEditing && !disabled && (
            <div className="flex justify-end pt-2">
              <Button variant="light" icon={<></>} onClick={onEdit}>
                Edit
              </Button>
            </div>
          )}
        </div>
      )}
    </WhiteCard>
  );
}
