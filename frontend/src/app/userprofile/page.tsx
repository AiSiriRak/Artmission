"use client";

import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { deleteUser, getUser, updateUser } from "@/services/userApi";
import { User } from "@/types/user";

import { PersonalInfoCard } from "@/components/feature/userprofile/PersonalInfoCard";
import { ProfileHeader } from "@/components/feature/userprofile/ProfileHeader";
import { BankAccountCard } from "@/components/feature/userprofile/BankAccountCard";
import { DeleteAccountCard } from "@/components/feature/userprofile/DeleteAccountCard";
import { DeleteAccountPopup } from "@/components/feature/userprofile/DeleteAccountPopup";

const id = 67;
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
  const [user, setUser] = useState<User | null>(null);
  const [deleteStatus, setDeleteStatus] = useState<
    "success" | "fail" | "default"
  >("default");

  useEffect(() => {
    async function loadUser() {
      const data = await getUser(67);
      setUser(data);
    }

    loadUser();
  }, []);

  if (!user) {
    return <div className="text-center">Loading...</div>;
  }

  const handleConfirmDelete = async () => {
    try {
      const result = await deleteUser(user.id, user);
      if (!result) {
        setDeleteStatus("fail");
        console.log("Cannot delete account");
        return;
      }
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
          <ProfileHeader />
          <PersonalInfoCard
            user={user}
            isEditing={editingSection === "personal"}
            disabled={editingSection !== null && editingSection !== "personal"}
            onEdit={() => setEditingSection("personal")}
            onCancel={() => setEditingSection(null)}
            onSave={async (updatedUser) => {
              try {
                const savedUser = await updateUser(user.id, updatedUser);

                setUser(savedUser);
                setEditingSection(null);
              } catch (error) {
                console.error("Failed to update user:", error);
              }
            }}
          />
          <BankAccountCard
            user={user}
            bankOptions={bankOptions}
            isEditing={editingSection === "bank"}
            disabled={editingSection !== null && editingSection !== "bank"}
            onEdit={() => setEditingSection("bank")}
            onCancel={() => setEditingSection(null)}
            onSave={async (updatedUser) => {
              try {
                const savedUser = await updateUser(user.id, updatedUser);

                setUser(savedUser);
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
