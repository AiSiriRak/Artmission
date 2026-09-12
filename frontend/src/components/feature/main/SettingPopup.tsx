"use client";

import Image from "next/image";
import { useRouter } from "next/navigation";

import { WhiteCard } from "@/components/ui/WhiteCard";
import { routes } from "@/lib/routes";
import { logout } from "@/lib/api/auth";

interface SettingPopup {
  isActive: boolean;
}
export function SettingPopup({ isActive }: SettingPopup) {
  const router = useRouter();
  return (
    isActive && (
      <WhiteCard
        className="absolute right-0 z-[60] mt-2"
        margin="m-0"
        padding="p-2"
      >
        <button
          type="button"
          onClick={() => router.replace(routes.settings)}
          className={`
                flex w-max cursor-pointer items-center
                px-2 py-2
                text-left
                text-subtle
                text-primary-500
                hover:bg-neutral-100
                rounded
                gap-3
              `}
        >
          {" "}
          <Image src="/icons/profile.svg" alt={""} width={24} height={24} />
          Profile
        </button>
        <button
          type="button"
          onClick={async () => {
            await logout();
            router.replace(routes.login);
          }}
          className={`
                flex w-max cursor-pointer items-center
                px-2 py-2
                text-left
                text-subtle
                text-primary-500
                hover:bg-neutral-100
                rounded
                gap-3
              `}
        >
          <Image src="/icons/logout.svg" alt={""} width={24} height={24} />
          Logout
        </button>
      </WhiteCard>
    )
  );
}