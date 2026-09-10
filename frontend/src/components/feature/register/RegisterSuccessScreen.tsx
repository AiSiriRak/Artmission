import Image from "next/image";
import { Button } from "@/components/ui/Button";

interface RegisterSuccessScreenProps {
  onLogIn: () => void;
}

export function RegisterSuccessScreen({ onLogIn }: RegisterSuccessScreenProps) {
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

      <section
        aria-labelledby="account-created-title"
        className="w-full max-w-[430px] rounded-lg bg-white px-12 py-10 text-center shadow-card"
      >
        <div
          className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-success text-white"
          role="status"
          aria-live="polite"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            className="h-8 w-8"
            aria-hidden="true"
          >
            <path
              d="M5 12.5L10 17.5L19 7.5"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>

        <div className="mt-5">
          <h1 id="account-created-title" className="text-h2 text-primary-500">
            Account created!
          </h1>
          <p className="mx-auto mt-8 max-w-[280px] text-small text-primary-500">
            Your account has been created successfully. Please log in to continue.
          </p>
        </div>

        <div className="mt-7 flex justify-end">
          <Button
            type="button"
            variant="dark"
            onClick={onLogIn}
            className="flex h-[34px] w-[114px] items-center justify-center gap-2 rounded-lg px-0 py-0 text-subtle text-white"
            icon={
              <Image
                src="/icons/login-icon.svg"
                alt=""
                width={16}
                height={16}
                aria-hidden="true"
              />
            }
          >
            Log in
          </Button>
        </div>
      </section>
    </main>
  );
}
