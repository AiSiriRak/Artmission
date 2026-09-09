"use client";

import { useState } from "react";

import { BankAccount, UpdateBankAccountInput } from "@/lib/api/types";
import { BANK_LABELS } from "@/lib/types";

import { WhiteCard } from "@/components/ui/WhiteCard";
import { TextInput } from "@/components/ui/TextInput";
import { Button } from "@/components/ui/Button";
import { SelectInput } from "@/components/ui/SelectInput";

interface BankAccountCard {
  bankAccount: BankAccount;
  isEditing: boolean;
  disabled: boolean;
  onEdit: () => void;
  onCancel: () => void;
  onSave: (updatedBankAccount: UpdateBankAccountInput) => void;
}

export function BankAccountCard({
  bankAccount,
  isEditing,
  disabled,
  onEdit,
  onCancel,
  onSave,
}: BankAccountCard) {
  const [bankName, setBankName] = useState(bankAccount.bank_name);

  const [accountHolder, setAccountHolder] = useState(
    bankAccount.account_holder_name,
  );
  const [accountNumber, setAccountNumber] = useState("");

  const [accountHolderError, setAccountHolderError] = useState("");
  const [accountNumberError, setAccountNumberError] = useState("");

  const bankOptions = Object.entries(BANK_LABELS).map(([value, label]) => ({
    value,
    label,
  }));

  const handleSave = () => {
    let hasError = false;

    // Validate account holder
    if (!accountHolder.trim()) {
      setAccountHolderError("Please enter the account holder's name.");
      hasError = true;
    } else {
      setAccountHolderError("");
    }

    // Validate account number
    if (!accountNumber.trim()) {
      setAccountNumberError("Please enter your bank account number.");
      hasError = true;
    } else if (!/^\d{8,16}$/.test(accountNumber.trim())) {
      setAccountNumberError("Please enter a valid bank account number.");
      hasError = true;
    } else {
      setAccountNumberError("");
    }

    if (hasError) return;

    const updatedBank: UpdateBankAccountInput = {
      bank_name: bankName,
      account_holder_name: accountHolder,
      account_number: accountNumber,
    };
    onSave(updatedBank);
  };

  const handleCancel = () => {
    setBankName(bankAccount.bank_name);
    setAccountHolder(bankAccount.account_holder_name);
    setAccountNumber("");

    setAccountHolderError("");
    setAccountNumberError("");
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
            <TextInput
              value={accountHolder}
              onChange={setAccountHolder}
              placeholder="Enter account holder name"
            />{" "}
            <span className="flex mt-1 text-small text-error justify-center">
              {accountHolderError}
            </span>
          </div>
          {/* Account number */}
          <div>
            <label className="text-small text-primary-500">
              Account holder number
            </label>
            <TextInput
              value={accountNumber}
              onChange={setAccountNumber}
              placeholder="Enter account number"
            />{" "}
            <span className="flex mt-1 text-small text-error justify-center">
              {accountNumberError}
            </span>
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
              {BANK_LABELS[bankAccount.bank_name as keyof typeof BANK_LABELS]}
            </p>
          </div>
          <div className="space-y-3">
            <label className="text-small text-primary-500">
              Account holder name
            </label>

            <p className="mt-1 text-body indent-2 text-primary-500">
              {bankAccount.account_holder_name}
            </p>
          </div>
          <div className="space-y-3">
            <label className="text-small text-primary-500">
              Account holder number
            </label>

            <p className="mt-1 text-body indent-2 text-primary-500">
              {bankAccount.account_last4}
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
