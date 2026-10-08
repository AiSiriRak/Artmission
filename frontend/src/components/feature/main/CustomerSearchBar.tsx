"use client";

import Image from "next/image";
import { useRouter, useSearchParams } from "next/navigation";
import { SubmitEvent, useState } from "react";

import { TextInput } from "@/components/ui/TextInput";
import { routes } from "@/lib/routes";

export function CustomerSearchBar() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [searchValue, setSearchValue] = useState(searchParams.get("q") ?? "");

  function submitSearch(event?: SubmitEvent) {
    event?.preventDefault();
    const q = searchValue.trim();
    const href = q ? `${routes.home}?q=${encodeURIComponent(q)}` : routes.home;
    router.push(href);
  }

  return (
    <form
      onSubmit={submitSearch}
      className="relative flex h-10 w-full min-w-0 max-w-80 items-center"
    >
      <TextInput
        value={searchValue}
        onChange={setSearchValue}
        placeholder="Search"
        className="h-full pr-12"
      />
      <button
        type="submit"
        className="absolute right-2 top-1/2 flex h-6 w-6 -translate-y-1/2 items-center justify-center rounded-full bg-white hover:brightness-90"
        aria-label="Search artists"
      >
        <Image src="/icons/search.svg" alt="Search" width={10} height={10} />
      </button>
    </form>
  );
}
