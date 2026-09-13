"use client";

import { useRef, useState, useEffect } from "react";

interface ProfileImageProps {
  imageUrl?: string | null;
  isEditing?: boolean;
  onImageChange?: (newImageUrl: string, file?: File) => void;
  className?: string;
  defaultAvatar?: string;
}

export default function ProfileImage({
  imageUrl,
  isEditing = false,
  onImageChange,
  className = "w-32 h-32 md:w-40 md:h-40",
  defaultAvatar = "/default-avatar.png",
}: ProfileImageProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  // เก็บสถานะว่าโหลดรูปพังหรือไม่
  const [imgError, setImgError] = useState(false);
  const [prevImageUrl, setPrevImageUrl] = useState(imageUrl);

  if (imageUrl !== prevImageUrl) {
    setPrevImageUrl(imageUrl);
    setImgError(false);
  }

  const getValidImageUrl = (url?: string | null) => {
    if (!url || url.trim() === "" || url === "null" || url === "undefined") {
      return defaultAvatar;
    }
    return url;
  };

  // คำนวณรูปที่จะแสดงผลตรงนี้เลย (Derived State)
  const currentSrc = imgError ? defaultAvatar : getValidImageUrl(imageUrl);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file && onImageChange) {
      if (currentSrc.startsWith("blob:")) {
        URL.revokeObjectURL(currentSrc);
      }
      const newUrl = URL.createObjectURL(file);
      onImageChange(newUrl, file);
    }
  };

  return (
    <div className="relative inline-block">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={currentSrc}
        alt="Profile"
        onError={() => {
          setImgError(true);
        }}
        className={`rounded-full object-cover shadow-sm bg-white border border-neutral ${className}`}
      />
      {isEditing && (
        <button
          type="button"
          onClick={() => fileInputRef.current?.click()}
          className="absolute bottom-0 right-2 md:bottom-2 md:right-2 bg-white border border-gray-200 rounded-full w-9 h-9 flex items-center justify-center shadow-md hover:bg-gray-50 cursor-pointer transition-transform hover:scale-105"
          title="Change Profile Picture"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            strokeWidth={1.8}
            stroke="currentColor"
            className="w-5 h-5 text-gray-700"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574v9.176A2.25 2.25 0 0 0 4.5 21h15a2.25 2.25 0 0 0 2.25-2.25V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316A2.192 2.192 0 0 0 15.15 3.75h-6.3c-.643 0-1.229.3-1.611.809l-.412.616ZM12 17.25a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9Z"
            />
          </svg>
        </button>
      )}
      <input
        type="file"
        ref={fileInputRef}
        accept="image/*"
        className="hidden"
        onChange={handleFileChange}
      />
    </div>
  );
}
