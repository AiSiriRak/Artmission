"use client";

import { useRouter } from "next/navigation";
import type { MouseEvent, SubmitEvent } from "react";
import { useState } from "react";
import { register } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";
import type {
  RegisterStep,
  RegisterValidationField,
  RegisterValidationErrors,
  Role,
} from "./registerConfig";
import {
  hasValidationErrors,
  registerValidationMessages,
  validateRegisterStepFields,
} from "./registerConfig";

function isDuplicateEmailError(error: unknown) {
  if (!(error instanceof ApiError) || error.status !== 409) {
    return false;
  }

  return error.message.toLowerCase().includes("email");
}

export function useRegisterForm() {
  const router = useRouter();
  const [step, setStep] = useState<RegisterStep>(1);
  const [role, setRole] = useState<Role>("");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [bank, setBank] = useState("");
  const [accountHolderName, setAccountHolderName] = useState("");
  const [accountNumber, setAccountNumber] = useState("");
  const [acceptedTerms, setAcceptedTerms] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<RegisterValidationErrors>({});
  const [submitError, setSubmitError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [accountCreated, setAccountCreated] = useState(false);

  const values = {
    role,
    email,
    username,
    password,
    confirmPassword,
    bank,
    accountHolderName,
    accountNumber,
    acceptedTerms,
  };

  function clearErrors() {
    setFieldErrors({});
    setSubmitError("");
  }

  function clearFieldError(field: RegisterValidationField) {
    setSubmitError("");
    setFieldErrors((currentErrors) => {
      if (!currentErrors[field]) {
        return currentErrors;
      }

      const nextErrors = { ...currentErrors };
      delete nextErrors[field];
      return nextErrors;
    });
  }

  function validateCurrentStep() {
    return validateRegisterStepFields(step, values);
  }

  function selectRole(nextRole: Role) {
    setRole(nextRole);
    clearFieldError("role");
  }

  function changeEmail(value: string) {
    setEmail(value);
    clearFieldError("email");
  }

  function changeUsername(value: string) {
    setUsername(value);
    clearFieldError("username");
  }

  function changePassword(value: string) {
    setPassword(value);
    clearFieldError("password");
    clearFieldError("confirmPassword");
  }

  function changeConfirmPassword(value: string) {
    setConfirmPassword(value);
    clearFieldError("confirmPassword");
  }

  function changeBank(value: string) {
    setBank(value);
    clearFieldError("bank");
  }

  function changeAccountHolderName(value: string) {
    setAccountHolderName(value);
    clearFieldError("accountHolderName");
  }

  function changeAccountNumber(value: string) {
    setAccountNumber(value);
    clearFieldError("accountNumber");
  }

  function changeAcceptedTerms(checked: boolean) {
    setAcceptedTerms(checked);
    clearFieldError("acceptedTerms");
  }

  function goNext(event: MouseEvent<HTMLButtonElement>) {
    event.preventDefault();

    const errors = validateCurrentStep();

    if (hasValidationErrors(errors)) {
      setFieldErrors(errors);
      return;
    }

    clearErrors();
    setStep((currentStep) => (currentStep === 1 ? 2 : 3));
  }

  function goBack(event: MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    clearErrors();

    if (step === 1) {
      router.push("/login");
      return;
    }

    setStep((currentStep) => (currentStep === 3 ? 2 : 1));
  }

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();

    if (isSubmitting) {
      return;
    }

    const errors = validateCurrentStep();

    if (hasValidationErrors(errors)) {
      setFieldErrors(errors);
      return;
    }

    setSubmitError("");
    setIsSubmitting(true);

    try {
      await register({
        username: username.trim(),
        email: email.trim(),
        password,
        role: role === "artist" ? "artist" : "customer",
        bank_account: {
          bank_name: bank.trim(),
          account_holder_name: accountHolderName.trim(),
          account_number: accountNumber.trim(),
        },
        ...(role === "artist"
          ? { artist: { description: `${username.trim()} artist profile` } }
          : {}),
      });

      clearErrors();
      setAccountCreated(true);
    } catch (error) {
      if (isDuplicateEmailError(error)) {
        setStep(2);
        setSubmitError("");
        setFieldErrors({
          email: registerValidationMessages.email.duplicate,
        });
        return;
      }

      setSubmitError(
        error instanceof Error ? error.message : "Failed to create account",
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  function goLogIn() {
    router.push("/login");
  }

  return {
    step,
    values,
    fieldErrors,
    submitError,
    isSubmitting,
    accountCreated,
    selectRole,
    changeEmail,
    changeUsername,
    changePassword,
    changeConfirmPassword,
    changeBank,
    changeAccountHolderName,
    changeAccountNumber,
    changeAcceptedTerms,
    goNext,
    goBack,
    handleSubmit,
    goLogIn,
  };
}
