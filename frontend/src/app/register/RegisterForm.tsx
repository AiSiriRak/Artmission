"use client";

import Image from "next/image";
import Link from "next/link";
import { RegisterActions } from "@/components/feature/register/RegisterActions";
import {
  RegisterStepContent,
  RegisterStepIndicator,
} from "@/components/feature/register/RegisterSteps";
import { RegisterSuccessScreen } from "@/components/feature/register/RegisterSuccessScreen";
import { useRegisterForm } from "@/components/feature/register/useRegisterForm";
import { WhiteCard } from "@/components/ui/WhiteCard";

export function RegisterForm() {
  const registerForm = useRegisterForm();

  if (registerForm.accountCreated) {
    return <RegisterSuccessScreen onLogIn={registerForm.goLogIn} />;
  }

  return (
    <main className="relative flex min-h-screen items-center justify-center bg-white px-6 py-12">
      <Image
        src="/icons/artmission_logo.svg"
        alt="Artmission logo"
        width={176}
        height={30}
        priority
        className="absolute left-24 top-6 h-auto w-36 sm:w-52"
      />

      <WhiteCard className="!max-w-lg">
        <div className="mb-8 text-center">
          <h1 className="text-h1 text-primary-500">Create Account</h1>
        </div>

        <RegisterStepIndicator step={registerForm.step} />

        <form
          className="relative flex w-full flex-col gap-5"
          onSubmit={registerForm.handleSubmit}
        >
          <RegisterStepContent
            step={registerForm.step}
            values={registerForm.values}
            fieldErrors={registerForm.fieldErrors}
            onSelectRole={registerForm.selectRole}
            onEmailChange={registerForm.changeEmail}
            onUsernameChange={registerForm.changeUsername}
            onPasswordChange={registerForm.changePassword}
            onConfirmPasswordChange={registerForm.changeConfirmPassword}
            onBankChange={registerForm.changeBank}
            onAccountHolderNameChange={registerForm.changeAccountHolderName}
            onAccountNumberChange={registerForm.changeAccountNumber}
            onAcceptedTermsChange={registerForm.changeAcceptedTerms}
          />

          <RegisterActions
            step={registerForm.step}
            submitError={registerForm.submitError}
            isSubmitting={registerForm.isSubmitting}
            onBack={registerForm.goBack}
            onNext={registerForm.goNext}
          />
        </form>

        <p className="mt-8 text-center text-subtle text-primary-500">
          Already have an account?{" "}
          <Link href="/login" className="text-accent-500 hover:underline">
            Log in
          </Link>
        </p>
      </WhiteCard>
    </main>
  );
}
