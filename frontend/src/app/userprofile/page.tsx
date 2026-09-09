"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import {
  deleteAccount,
  getAccount,
  updateAccount,
  getBankAccount,
  updateBankAccount,
} from "@/lib/api/users";
import { BankAccount, UserAccount } from "@/lib/api/types";

import { PersonalInfoCard } from "@/components/feature/userprofile/PersonalInfoCard";
import { ProfileHeader } from "@/components/feature/userprofile/ProfileHeader";
import { BankAccountCard } from "@/components/feature/userprofile/BankAccountCard";
import { DeleteAccountCard } from "@/components/feature/userprofile/DeleteAccountCard";
import { DeleteAccountPopup } from "@/components/feature/userprofile/DeleteAccountPopup";
import { Loading } from "@/components/ui/Loading";

const bankOptions = [
  { value: "ธนาคารกรุงเทพ (BBL)", label: "ธนาคารกรุงเทพ (BBL)" },
  { value: "ธนาคารกสิกรไทย (KBANK)", label: "ธนาคารกสิกรไทย (KBANK)" },
  { value: "ธนาคารกรุงไทย (KTB)", label: "ธนาคารกรุงไทย (KTB)" },
  { value: "ธนาคารไทยพาณิชย์ (SCB)", label: "ธนาคารไทยพาณิชย์ (SCB)" },
  { value: "ธนาคารกรุงศรีอยุธยา (BAY)", label: "ธนาคารกรุงศรีอยุธยา (BAY)" },
];

export default function HomePage() {
  const router = useRouter();
  const [editingSection, setEditingSection] = useState<
    "personal" | "bank" | null
  >(null);
  const [isDeletePopupOpen, setIsDeletePopupOpen] = useState(false);
  const [profile, setProfile] = useState<File | null>(null);
  const [user, setUser] = useState<UserAccount | null>(null);
  const [bank, setBank] = useState<BankAccount | null>(null);
  const [deleteStatus, setDeleteStatus] = useState<
    "success" | "fail" | "default"
  >("default");

  useEffect(() => {
    async function loadUser() {
      const accountdata = await getAccount();
      const bankdata = await getBankAccount();

      setBank(bankdata);
      setUser(accountdata);
    }

    loadUser();
  }, []);

  if (!user || !bank) {
    return <Loading />;
  }

  const handleConfirmDelete = async () => {
    try {
      deleteAccount();
      console.log("Account deleted successfully");
      setDeleteStatus("success");
      setEditingSection(null);
    } catch (error) {
      console.error("Failed to delete bank account:", error);
    }
  };

  return (
    <main className="min-h-screen">
      <div className="mx-auto flex max-w-5xl flex-col items-center">
        <>
          <ProfileHeader
            username={user.username}
            onUpload={function (imageUrl: File): void {
              setProfile(imageUrl);
            }}
          />
          <PersonalInfoCard
            user={user}
            isEditing={editingSection === "personal"}
            disabled={editingSection !== null && editingSection !== "personal"}
            onEdit={() => setEditingSection("personal")}
            onCancel={() => setEditingSection(null)}
            onSave={async (updatedAccount) => {
              try {
                const savedUser = await updateAccount(updatedAccount);

                setUser(savedUser);
                setEditingSection(null);
              } catch (error) {
                console.error("Failed to update user:", error);
              }
            }}
          />
          <BankAccountCard
            bankAccount={bank}
            bankOptions={bankOptions}
            isEditing={editingSection === "bank"}
            disabled={editingSection !== null && editingSection !== "bank"}
            onEdit={() => setEditingSection("bank")}
            onCancel={() => setEditingSection(null)}
            onSave={async (updatedBankAccount) => {
              try {
                const savedBank = await updateBankAccount(updatedBankAccount);

                setBank(savedBank);
                setEditingSection(null);
              } catch (error) {
                console.error("Failed to update bank account:", error);
              }
            }}
          />
          <DeleteAccountCard
            disabled={editingSection !== null}
            onDelete={() => setIsDeletePopupOpen(true)}
          />
          <DeleteAccountPopup
            isOpen={isDeletePopupOpen}
            onCancel={
              deleteStatus == "success"
                ? () => router.push("/")
                : () => {
                    setIsDeletePopupOpen(false);
                    setDeleteStatus("default");
                  }
            }
            onConfirm={handleConfirmDelete}
            status={deleteStatus}
          />
        </>
      </div>
    </main>
  );
}
