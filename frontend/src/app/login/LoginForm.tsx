"use client";

import Image from "next/image";
import Link from "next/link";
import type { SubmitEvent } from "react";
import { useState } from "react";
import { Button } from "@/components/ui/Button";
import { TextInput } from "@/components/ui/TextInput";
import { WhiteCard } from "@/components/ui/WhiteCard";

export function LoginForm() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-secondary-200 px-6 py-6">
      <WhiteCard>
        <div className="mb-8 text-center">
          <h1 className="mt-2 text-h1 text-primary-500">Login</h1>
          <p className="mt-3 whitespace-nowrap text-h2 text-accent-300">
            Welcome to Artmission
          </p>
        </div>

        <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
          <label className="text-caption text-primary-500">
            Email
            <TextInput
              type="email"
              value={email}
              onChange={setEmail}
              placeholder="Enter your email"
            />
          </label>

          <label className="text-caption text-primary-500">
            Password
            <div className="relative mt-1 [&_input]:mt-0 [&_input]:pr-12">
              <TextInput
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={setPassword}
                placeholder="Enter your password"
              />

              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2"
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                <Image
                  src={showPassword ? "/icons/eye-on.svg" : "/icons/eye-off.svg"}
                  alt=""
                  width={20}
                  height={20}
                  className="mx-auto"
                />
              </button>
            </div>
          </label>

          <p className="mt-0.5 text-center text-subtle text-primary-500">
            Don&apos;t have an account?{" "}
            <Link href="/register" className="text-accent-500 hover:underline">
              Create Account
            </Link>
          </p>

          <Button
            type="submit"
            variant="accent-500"
            className="mt-4 w-full py-3 text-white"
          >
            Log in
          </Button>
        </form>

      </WhiteCard>
    </main>
  );
}
