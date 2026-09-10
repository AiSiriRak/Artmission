"use client";

import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { SubmitEvent } from "react";
import { useState } from "react";
import { Button } from "@/components/ui/Button";
import { TextInput } from "@/components/ui/TextInput";
import { WhiteCard } from "@/components/ui/WhiteCard";
import { login } from "@/lib/api/auth";
import { setAccessToken } from "@/lib/token";

const INVALID_LOGIN_MESSAGE =
  "Cannot login because the email or password is incorrect.";
const showPasswordLabel = "Show password";
const hidePasswordLabel = "Hide password";

export function LoginForm() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const canSubmit =
    email.trim().length > 0 && password.trim().length > 0 && !isSubmitting;

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!canSubmit) {
      return;
    }

    setErrorMessage("");
    setIsSubmitting(true);

    try {
      const result = await login({
        email: email.trim(),
        password,
      });

      setAccessToken(result.access_token);
      router.push("/");
    } catch {
      setErrorMessage(INVALID_LOGIN_MESSAGE);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="relative flex min-h-screen items-center justify-center bg-white px-6 py-6">
      <Image
        src="/icons/artmission_logo.svg"
        alt="Artmission logo"
        width={176}
        height={30}
        priority
        className="absolute left-24 top-6 h-auto w-36 sm:w-52"
      />

      <WhiteCard>
        <div className="mb-8 text-center">
          <h1 className="mt-2 text-h1 text-primary-500">Login</h1>
          <p className="mt-3 whitespace-nowrap text-h2 text-accent-300">
            Welcome to Artmission
          </p>
        </div>

        <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
          <label className="text-small text-primary-500">
            Email
            <TextInput
              type="email"
              value={email}
              onChange={(value) => {
                setEmail(value);
                setErrorMessage("");
              }}
              placeholder="Enter your email"
              className="mt-1"
            />
          </label>

          <label className="text-small text-primary-500">
            Password
            <div className="relative mt-1 [&_input]:mt-0 [&_input]:pr-12">
              <TextInput
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(value) => {
                  setPassword(value);
                  setErrorMessage("");
                }}
                placeholder="Enter your password"
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
          </label>

          <p className="mt-1 text-center text-subtle text-primary-500">
            Don&apos;t have an account?{" "}
            <Link href="/register" className="text-accent-500 hover:underline">
              Create Account
            </Link>
          </p>

          <p
            role={errorMessage ? "alert" : undefined}
            className="flex h-10 items-end justify-center text-center text-caption text-error"
          >
            {errorMessage}
          </p>

          <Button
            type="submit"
            variant="accent-500"
            disabled={!canSubmit}
            aria-disabled={!canSubmit}
            className={`mt-0 w-full border-2 border-transparent py-3 text-white transition ${
              canSubmit ? "opacity-100" : "opacity-50"
            }`}
          >
            {isSubmitting ? "Logging in..." : "Log in"}
          </Button>
        </form>

      </WhiteCard>
    </main>
  );
}
