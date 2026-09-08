"use client";

import { useState } from "react";

import { UserAccount, UpdateAccountInput } from "@/lib/api/types";

import { WhiteCard } from "@/components/ui/WhiteCard";
import { TextInput } from "@/components/ui/TextInput";
import { Button } from "@/components/ui/Button";

interface PersonalInfoCard {
  user: UserAccount;
  isEditing: boolean;
  disabled: boolean;
  onEdit: () => void;
  onCancel: () => void;
  onSave: (username: UpdateAccountInput) => void;
}

export function PersonalInfoCard({
  user,
  isEditing,
  disabled,
  onEdit,
  onCancel,
  onSave,
}: PersonalInfoCard) {
  const [username, setUserame] = useState(user.username);
  const [email, setEmail] = useState(user.email);
  // const [password, setPassword] = useState(user.password);

  const handleSave = () => {
    const updatedBank: UpdateAccountInput = {
      username: username,
    };
    onSave(updatedBank);
  };

  const handleCancel = () => {
    setUserame(user.username);
    setEmail(user.email);
    // setPassword(user.password);

    onCancel();
  };

  return (
    <WhiteCard>
      <h2 className="text-center text-h3 text-primary-500">
        User Personal Info
      </h2>

      {isEditing ? (
        // === EDIT MODE ===
        <div className="mt-6 space-y-4.5">
          {/* Name */}
          <div>
            <label className="text-small text-primary-500">Username</label>
            <TextInput
              value={username}
              onChange={setUserame}
              placeholder="Enter username"
            />
          </div>

          {/* Email */}
          <div>
            <label className="text-small text-primary-500">Email</label>
            <TextInput value={email} disabled={true} onChange={() => {}} />
          </div>

          {/* Password */}
          <div>
            <label className="text-small text-primary-500">Password</label>
            {/* <TextInput
              value={password}
              onChange={setPassword}
              disabled={true}
            /> */}
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
            <label className="text-small text-primary-500">Username</label>

            <p className="mt-1 text-body indent-2 text-primary-500">
              {user.username}
            </p>
          </div>

          {/* Email */}
          <div className="space-y-3">
            <label className="text-small text-primary-500">Email</label>

            <p className="mt-1 text-body indent-2 text-primary-500">
              {user.email}
            </p>
          </div>

          {/* Password */}
          <div className="space-y-3">
            <label className="text-small text-primary-500">Password</label>

            <p className="mt-1 text-body indent-2 text-primary-500">
              •••••••••
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
