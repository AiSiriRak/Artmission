"use client";

import Image from "next/image";
import { useRef, useState } from "react";

interface ProfileHeader {
  username: string;
  profileImageUrl: string | null;
  onUpload: (imagefile: File) => void;
}

export function ProfileHeader({
  username,
  profileImageUrl,
  onUpload,
}: ProfileHeader) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];

    if (!file) return;

    onUpload(file);
  };

  return (
    <div className="mt-12 flex flex-col items-center">
      <div className="relative">
        <Image
          src={profileImageUrl ?? "/icons/emptyprofile.svg"}
          alt=""
          width={150}
          height={150}
          className="mx-auto h-[150px] w-[150px] object-cover rounded-full"
        />
        <button
          className="absolute bottom-0 right-0 rounded-full bg-[#787878] p-2 hover:brightness-90"
          onClick={handleClick}
        >
          <Image
            src="/icons/camera.svg"
            alt=""
            width={20}
            height={20}
            className="mx-auto"
          />
        </button>{" "}
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileChange}
          className="hidden"
        />
      </div>

      <h1 className="mt-4 text-h2 text-primary-500">{username}</h1>
    </div>
  );
}
