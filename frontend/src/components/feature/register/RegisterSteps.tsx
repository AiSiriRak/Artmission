"use client";

import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { TextInput } from "@/components/ui/TextInput";
import {
  fieldErrorIds,
  registerBanks,
  registerSteps,
  type RegisterFormValues,
  type RegisterStep,
  type RegisterValidationErrors,
  type RegisterValidationField,
  type Role,
} from "./registerConfig";

interface FormConstraintMessageProps {
  children: React.ReactNode;
  className?: string;
}

function FormConstraintMessage({
  children,
  className = "",
}: FormConstraintMessageProps) {
  return (
    <p
      className={`mt-1 flex items-center justify-center gap-2 text-caption text-accent-500 ${className}`}
    >
      <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-accent-500 text-subtle font-bold text-white">
        i
      </span>
      <span>{children}</span>
    </p>
  );
}

interface FormErrorMessageProps {
  children?: React.ReactNode;
  className?: string;
  id?: string;
}

function FormErrorMessage({
  children,
  className = "",
  id,
}: FormErrorMessageProps) {
  if (!children) {
    return null;
  }

  return (
    <p
      id={id}
      role="alert"
      aria-live="polite"
      className={`text-center text-caption leading-5 text-error ${className}`}
    >
      {children}
    </p>
  );
}

interface RegisterFieldErrorProps {
  field: RegisterValidationField;
  fieldErrors: RegisterValidationErrors;
}

function RegisterFieldError({ field, fieldErrors }: RegisterFieldErrorProps) {
  return (
    <FormErrorMessage
      id={fieldErrorIds[field]}
      className="pointer-events-none absolute left-0 top-full z-10 mt-1 w-full"
    >
      {fieldErrors[field]}
    </FormErrorMessage>
  );
}

interface PasswordInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  showPasswordLabel?: string;
  hidePasswordLabel?: string;
}

function PasswordInput({
  value,
  onChange,
  placeholder = "Enter your password",
  showPasswordLabel = "Show password",
  hidePasswordLabel = "Hide password",
}: PasswordInputProps) {
  const [showPassword, setShowPassword] = useState(false);

  return (
    <div className="relative mt-1 [&_input]:mt-0 [&_input]:pr-12">
      <TextInput
        type={showPassword ? "text" : "password"}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
      />

      <button
        type="button"
        onClick={() => setShowPassword(!showPassword)}
        className="absolute right-3 top-1/2 flex h-5 w-5 -translate-y-1/2 items-center justify-center"
        aria-label={showPassword ? hidePasswordLabel : showPasswordLabel}
      >
        <Image
          src={showPassword ? "/icons/eye-on.svg" : "/icons/eye-off.svg"}
          alt=""
          width={20}
          height={20}
          className="block"
        />
      </button>
    </div>
  );
}

interface RegisterStepIndicatorProps {
  step: RegisterStep;
}

export function RegisterStepIndicator({ step }: RegisterStepIndicatorProps) {
  return (
    <div className="mx-auto mb-10 flex w-full max-w-xs items-center">
      {registerSteps.map((stepNumber, index) => {
        const isActive = stepNumber === step;
        const isComplete = stepNumber < step;
        const isFuture = stepNumber > step;

        return (
          <div key={stepNumber} className="flex flex-1 items-center last:flex-none">
            <div
              className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-full border-4 ${
                isComplete
                  ? "border-primary-500 bg-primary-500"
                  : isActive
                    ? "border-primary-500 bg-white"
                    : "border-neutral-400 bg-white"
              }`}
              aria-label={`Step ${stepNumber}`}
            >
              {isComplete && (
                <svg
                  className="h-6 w-6 text-white"
                  viewBox="0 0 24 24"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                  aria-hidden="true"
                >
                  <path
                    d="M5 12.5L10 17.5L19 7.5"
                    stroke="currentColor"
                    strokeWidth="3"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              )}
              {isActive && <span className="h-3 w-3 rounded-full bg-primary-500" />}
              {isFuture && <span className="sr-only">Future step</span>}
            </div>

            {index < registerSteps.length - 1 && (
              <div
                className={`h-1 flex-1 ${
                  stepNumber < step ? "bg-primary-500" : "bg-neutral-400"
                }`}
              />
            )}
          </div>
        );
      })}
    </div>
  );
}

interface RegisterStepContentProps {
  step: RegisterStep;
  values: RegisterFormValues;
  fieldErrors: RegisterValidationErrors;
  onSelectRole: (role: Role) => void;
  onEmailChange: (value: string) => void;
  onUsernameChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
  onConfirmPasswordChange: (value: string) => void;
  onBankChange: (value: string) => void;
  onAccountHolderNameChange: (value: string) => void;
  onAccountNumberChange: (value: string) => void;
  onAcceptedTermsChange: (checked: boolean) => void;
}

export function RegisterStepContent({
  step,
  values,
  fieldErrors,
  onSelectRole,
  onEmailChange,
  onUsernameChange,
  onPasswordChange,
  onConfirmPasswordChange,
  onBankChange,
  onAccountHolderNameChange,
  onAccountNumberChange,
  onAcceptedTermsChange,
}: RegisterStepContentProps) {
  if (step === 1) {
    return (
      <RegisterRoleStep
        role={values.role}
        fieldErrors={fieldErrors}
        onSelectRole={onSelectRole}
      />
    );
  }

  if (step === 2) {
    return (
      <RegisterAccountInformationStep
        values={values}
        fieldErrors={fieldErrors}
        onEmailChange={onEmailChange}
        onUsernameChange={onUsernameChange}
        onPasswordChange={onPasswordChange}
        onConfirmPasswordChange={onConfirmPasswordChange}
      />
    );
  }

  return (
    <RegisterBankAccountStep
      values={values}
      fieldErrors={fieldErrors}
      onBankChange={onBankChange}
      onAccountHolderNameChange={onAccountHolderNameChange}
      onAccountNumberChange={onAccountNumberChange}
      onAcceptedTermsChange={onAcceptedTermsChange}
    />
  );
}

interface RegisterRoleStepProps {
  role: Role;
  fieldErrors: RegisterValidationErrors;
  onSelectRole: (role: Role) => void;
}

function RegisterRoleStep({
  role,
  fieldErrors,
  onSelectRole,
}: RegisterRoleStepProps) {
  return (
    <section className="flex h-[320px] w-full flex-col items-center gap-10 text-center">
      <div>
        <h2 className="text-body text-primary-500">Choose your role</h2>
        <FormConstraintMessage>Your role cannot be changed later</FormConstraintMessage>
      </div>

      <div className="relative w-full">
        <div className="grid w-full grid-cols-1 gap-4 sm:grid-cols-2">
          <button
            type="button"
            onClick={() => onSelectRole("artist")}
            aria-pressed={role === "artist"}
            className={`rounded-2xl border px-4 py-6 text-center transition ${
              role === "artist"
                ? "border-primary-500 bg-primary-500 text-white"
                : "border-primary-500 bg-white text-primary-500 hover:bg-secondary-200"
            }`}
          >
            <p className="text-h3">Artist</p>
            <p className="mt-3 text-small">Sell artwork</p>
          </button>

          <button
            type="button"
            onClick={() => onSelectRole("customer")}
            aria-pressed={role === "customer"}
            className={`rounded-2xl border px-4 py-6 text-center transition ${
              role === "customer"
                ? "border-primary-500 bg-primary-500 text-white"
                : "border-primary-500 bg-white text-primary-500 hover:bg-secondary-200"
            }`}
          >
            <p className="text-h3">Customer</p>
            <p className="mt-3 text-small">Buy artwork</p>
          </button>
        </div>
        <RegisterFieldError field="role" fieldErrors={fieldErrors} />
      </div>
    </section>
  );
}

interface RegisterAccountInformationStepProps {
  values: RegisterFormValues;
  fieldErrors: RegisterValidationErrors;
  onEmailChange: (value: string) => void;
  onUsernameChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
  onConfirmPasswordChange: (value: string) => void;
}

function RegisterAccountInformationStep({
  values,
  fieldErrors,
  onEmailChange,
  onUsernameChange,
  onPasswordChange,
  onConfirmPasswordChange,
}: RegisterAccountInformationStepProps) {
  return (
    <section className="flex h-[320px] w-full flex-col gap-6">
      <h2 className="text-center text-body text-primary-500">
        Account information
      </h2>

      <div className="relative">
        <label className="text-caption text-primary-500">
          Email
          <div className="mt-1 [&_input]:mt-0">
            <TextInput
              type="email"
              value={values.email}
              onChange={onEmailChange}
              placeholder="Enter your email"
            />
          </div>
        </label>
        <RegisterFieldError field="email" fieldErrors={fieldErrors} />
      </div>

      <div className="relative">
        <label className="text-caption text-primary-500">
          Username
          <div className="mt-1 [&_input]:mt-0">
            <TextInput
              value={values.username}
              onChange={onUsernameChange}
              placeholder="Enter your username"
            />
          </div>
        </label>
        <RegisterFieldError field="username" fieldErrors={fieldErrors} />
      </div>

      <div className="relative">
        <label className="text-caption text-primary-500">
          Password
          <PasswordInput
            value={values.password}
            onChange={onPasswordChange}
            placeholder="Enter your password"
          />
        </label>
        <RegisterFieldError field="password" fieldErrors={fieldErrors} />
      </div>

      <div className="relative">
        <label className="text-caption text-primary-500">
          Confirm password
          <PasswordInput
            value={values.confirmPassword}
            onChange={onConfirmPasswordChange}
            placeholder="Confirm your password"
            showPasswordLabel="Show confirm password"
            hidePasswordLabel="Hide confirm password"
          />
        </label>
        <RegisterFieldError field="confirmPassword" fieldErrors={fieldErrors} />
      </div>
    </section>
  );
}

interface RegisterBankAccountStepProps {
  values: RegisterFormValues;
  fieldErrors: RegisterValidationErrors;
  onBankChange: (value: string) => void;
  onAccountHolderNameChange: (value: string) => void;
  onAccountNumberChange: (value: string) => void;
  onAcceptedTermsChange: (checked: boolean) => void;
}

function RegisterBankAccountStep({
  values,
  fieldErrors,
  onBankChange,
  onAccountHolderNameChange,
  onAccountNumberChange,
  onAcceptedTermsChange,
}: RegisterBankAccountStepProps) {
  return (
    <section className="flex h-[320px] w-full flex-col gap-6">
      <h2 className="text-center text-body text-primary-500">Bank account</h2>

      <div className="relative">
        <label htmlFor="bank" className="text-caption text-primary-500">
          Bank
        </label>
        <div className="relative mt-1">
          <select
            id="bank"
            value={values.bank}
            onChange={(event) => onBankChange(event.target.value)}
            aria-invalid={Boolean(fieldErrors.bank) || undefined}
            aria-describedby={fieldErrors.bank ? fieldErrorIds.bank : undefined}
            className="w-full appearance-none rounded border px-3 py-2 pr-10 text-small"
          >
            <option value="">Select bank</option>
            {registerBanks.map((bankName) => (
              <option key={bankName} value={bankName}>
                {bankName}
              </option>
            ))}
          </select>
          <svg
            aria-hidden="true"
            viewBox="0 0 20 20"
            className="pointer-events-none absolute right-3 top-1/2 h-5 w-5 -translate-y-1/2 text-primary-500"
          >
            <path
              d="M5 7.5L10 12.5L15 7.5"
              fill="none"
              stroke="currentColor"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
            />
          </svg>
        </div>
        <RegisterFieldError field="bank" fieldErrors={fieldErrors} />
      </div>

      <div className="relative">
        <label className="text-caption text-primary-500">
          Account holder name
          <div className="mt-1 [&_input]:mt-0">
            <TextInput
              value={values.accountHolderName}
              onChange={onAccountHolderNameChange}
              placeholder="Enter account holder name"
            />
          </div>
        </label>
        <RegisterFieldError field="accountHolderName" fieldErrors={fieldErrors} />
      </div>

      <div className="relative">
        <label className="text-caption text-primary-500">
          Account number
          <div className="mt-1 [&_input]:mt-0">
            <TextInput
              value={values.accountNumber}
              onChange={onAccountNumberChange}
              placeholder="Enter account number"
            />
          </div>
        </label>
        <RegisterFieldError field="accountNumber" fieldErrors={fieldErrors} />
      </div>

      <div className="relative mt-6">
        <label className="flex items-start gap-3 text-subtle text-primary-500">
          <input
            type="checkbox"
            checked={values.acceptedTerms}
            onChange={(event) => onAcceptedTermsChange(event.target.checked)}
            aria-invalid={Boolean(fieldErrors.acceptedTerms) || undefined}
            aria-describedby={
              fieldErrors.acceptedTerms ? fieldErrorIds.acceptedTerms : undefined
            }
            className="peer sr-only"
          />
          <span
            aria-hidden="true"
            className={`mt-1 flex h-4 w-4 shrink-0 items-center justify-center rounded border peer-focus-visible:outline peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-accent-500 ${
              values.acceptedTerms
                ? "border-accent-500 bg-accent-500"
                : "border-primary-500 bg-white"
            }`}
          >
            {values.acceptedTerms && (
              <svg viewBox="0 0 16 16" fill="none" className="h-3 w-3 text-white">
                <path
                  d="M3.5 8L6.5 11L12.5 5"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            )}
          </span>
          <span>
            I agree to the{" "}
            <Link href="#" className="text-accent-500 hover:underline">
              Terms of Service
            </Link>{" "}
            and{" "}
            <Link href="#" className="text-accent-500 hover:underline">
              Privacy Policy
            </Link>
            .
          </span>
        </label>
        <RegisterFieldError field="acceptedTerms" fieldErrors={fieldErrors} />
      </div>
    </section>
  );
}
