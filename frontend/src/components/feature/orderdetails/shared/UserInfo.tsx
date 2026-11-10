import Image from "next/image";

import { WhiteCard } from "@/components/ui/WhiteCard";

interface UserInfoProps {
  title: string;
  name: string;
  email: string;
  profileImageURL?: string;
}

export default function UserInfo({ title, name, email, profileImageURL }: UserInfoProps) {
  return (
    <section>
      <h2 className="mb-4 text-h3">{title}</h2>
      <WhiteCard
        margin=""
        padding="p-6 sm:p-8"
        roundsize="rounded-3xl"
        className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
      >
        <div className="flex flex-wrap items-center gap-4">
          <div className="relative h-20 w-20 shrink-0 overflow-hidden rounded-full">
            <Image
              src={profileImageURL || "/default-avatar.png"}
              alt="User Avatar"
              fill
              sizes="80px"
              className="object-cover"
            />
          </div>
          <dl className="grid min-w-0 flex-1 grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-4 gap-y-0.5">
            <dt className="text-h4">Name</dt>
            <dd className="break-words text-body">{name}</dd>
            <dt className="text-h4">Email</dt>
            <dd className="break-words text-body">{email}</dd>
          </dl>
        </div>
      </WhiteCard>
    </section>
  );
}
