"use client";

import { useState } from "react";
import Image from "next/image";

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
  onSave: (data: UpdateAccountInput) => Promise<string | null>;
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

  const [oldpassword, setOldPassword] = useState("");
  const [newpassword, setNewPassword] = useState("");
  const [confirmpassword, setConfimPassword] = useState("");

  const [isEditingPassword, setIsEditingPassword] = useState(false);

  const [usernameError, setUsernameError] = useState("");
  const [oldpasswordError, setOldpasswordError] = useState("");
  const [newpasswordError, setNewpasswordError] = useState("");
  const [confirmpasswordError, setConfirmpasswordError] = useState("");

  const handleSave = async () => {
    let hasError = false;

    // Validate username
    if (!username.trim()) {
      setUsernameError("Please enter your name.");
      hasError = true;
    } else if (username.length < 3) {
      setUsernameError("Name must be at least 3 characters.");
      hasError = true;
    } else if (username.length > 20) {
      setUsernameError("Name must be 20 characters or less.");
      hasError = true;
    } else {
      setUsernameError("");
    }

    // Validate password
    // Old password
    if (isEditingPassword) {
      setOldpasswordError("");
      setNewpasswordError("");
      setConfirmpasswordError("");
      // Old password
      if (!oldpassword.trim()) {
        setOldpasswordError("Please enter your old password.");
        hasError = true;
      }
      // New password
      else if (!newpassword.trim()) {
        setNewpasswordError("Please enter a password.");
        hasError = true;
      } else if (newpassword.length < 8) {
        setNewpasswordError("Password must be at least 8 characters.");
        hasError = true;
      } else if (newpassword.length > 16) {
        setNewpasswordError("Password must be 16 characters or less.");
        hasError = true;
      }
      // Confirm password
      else if (confirmpassword != newpassword) {
        setNewpasswordError("Passwords do not match.");
        hasError = true;
      } else {
        setOldpasswordError("");
        setNewpasswordError("");
        setConfirmpasswordError("");
      }
    }
    if (hasError) return;
    const updatedAccount: UpdateAccountInput = isEditingPassword
      ? {
          username: username.trim(),
          old_password: oldpassword,
          new_password: newpassword,
        }
      : {
          username: username.trim(),
        };

    const error = await onSave(updatedAccount);

    if (error == "Old password is incorrect.") {
      setOldpasswordError(error);
      return;
    }
  };

  const handleCancel = () => {
    setUserame(user.username);
    setIsEditingPassword(false);
    setEmail(user.email);
    setOldPassword("");
    setNewPassword("");
    setConfimPassword("");

    setUsernameError("");
    setOldpasswordError("");
    setNewpasswordError("");
    setConfirmpasswordError("");

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
            <span className="flex mt-1 text-small text-error justify-center">
              {usernameError}
            </span>
          </div>
          {/* Email */}
          <div>
            <label className="text-small text-primary-500">Email</label>
            <TextInput value={email} disabled={true} onChange={() => {}} />{" "}
          </div>
          {/* Password */}
          {!isEditingPassword ? (
            <div>
              {/* Default */}
              <div>
                <label className="text-small text-primary-500">Password</label>
                <div className="relative  items-center">
                  <TextInput
                    value="•••••••••"
                    onChange={() => {}}
                    disabled={true}
                  />
                  <button
                    onClick={() => setIsEditingPassword(true)}
                    type="button"
                    className="absolute flex rounded-full right-2 top-1/2 h-8 w-8 justify-center brightness-90 items-center bg-white -translate-y-1/2 hover:brightness-80"
                  >
                    <Image
                      src="/icons/edit.svg"
                      alt="Search"
                      width={20}
                      height={20}
                    />
                  </button>
                </div>
                <span className="flex mt-1 text-small text-error justify-center">
                  {oldpasswordError}
                </span>
              </div>
            </div>
          ) : (
            <div>
              {/* Old */}
              <div>
                <label className="text-small text-primary-500">
                  Old password
                </label>
                <TextInput value={oldpassword} onChange={setOldPassword} />
                <span className="flex mt-1 text-small text-error justify-center">
                  {oldpasswordError}
                </span>
              </div>
              {/* New */}
              <div>
                <label className="text-small text-primary-500">
                  New password
                </label>
                <TextInput value={newpassword} onChange={setNewPassword} />
                <span className="flex mt-1 text-small text-error justify-center">
                  {newpasswordError}
                </span>
              </div>
              {/* Confirm New */}
              <div>
                <label className="text-small text-primary-500">
                  Confirm new password
                </label>
                <TextInput
                  value={confirmpassword}
                  onChange={setConfimPassword}
                />
                <span className="flex mt-1 text-small text-error justify-center">
                  {confirmpasswordError}
                </span>
              </div>
            </div>
          )}

          {/* Buttons */}
          <div className="flex w-full justify-between gap-2 pt-2">
            <Button
              variant="light"
              icon={
                <Image
                  src="/icons/cancel.svg"
                  alt={""}
                  width={24}
                  height={24}
                />
              }
              onClick={handleCancel}
            >
              Cancel
            </Button>
            <Button
              variant="dark"
              icon={
                <Image src="/icons/done.svg" alt={""} width={24} height={24} />
              }
              onClick={handleSave}
            >
              Done
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
              <Button
                variant="light"
                icon={
                  <Image
                    src="/icons/edit.svg"
                    alt={""}
                    width={24}
                    height={24}
                  />
                }
                onClick={onEdit}
              >
                Edit
              </Button>
            </div>
          )}
        </div>
      )}
    </WhiteCard>
  );
}
