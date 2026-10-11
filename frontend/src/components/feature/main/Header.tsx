"use client";

import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { Suspense, useEffect, useRef, useState } from "react";

import { routes } from "@/lib/routes";

import { Button } from "@/components/ui/Button";
import { CustomerSearchBar } from "./CustomerSearchBar";
import { SettingPopup } from "./SettingPopup";

interface HeaderProps {
  usertype: string;
}

export default function Header({ usertype = "customer" }: HeaderProps) {
  const [isSettingPopupOpen, setIsSettingPopupOpen] = useState(false);
  const pathname = usePathname();

  const navItems =
    usertype == "customer"
      ? [
          {
            name: "Home",
            page: "Home",
            path: routes.home,
            icon: "/icons/home.svg",
          },
          {
            name: "Order",
            page: "Order History",
            path: routes.order.history,
            icon: "/icons/order.svg",
          },
          {
            name: "Notification",
            page: "Notification",
            path: routes.notification,
            icon: "/icons/notification.svg",
          },
        ]
      : [
          {
            name: "Order",
            page: "Home",
            path: routes.home,
            icon: "/icons/order.svg",
          },
          {
            name: "Artist Profile",
            page: "Artist Profile",
            path: routes.artist.profile,
            icon: "/icons/artist_profile.svg",
          },
          {
            name: "Notification",
            page: "Notification",
            path: routes.notification,
            icon: "/icons/notification.svg",
          },
        ];
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsSettingPopupOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // ฟังก์ชันจัดการการคลิกปุ่ม
  const handleNavClick = (e: React.MouseEvent, isActive: boolean) => {
    // ถ้าคลิกปุ่มของหน้าปัจจุบันอยู่แล้ว ให้ทำการ Refresh หน้าเดิม
    if (isActive) {
      e.preventDefault(); // ป้องกันไม่ให้ Link นำทางไป URL ใหม่
      window.location.reload(); // รีเฟรชหน้าปัจจุบัน
    }
  };

  return (
    <header className="sticky top-0 z-50 w-full bg-white border-b border-neutral">
      <div className="mx-auto flex h-16 max-w-[1400px] items-center gap-3 px-3 sm:h-20 sm:gap-4 sm:px-6">
        {/* ด้านซ้าย: Logo ARTMISSION */}
        <div className="flex shrink-0 items-center">
          <Link href="/" className="flex items-center gap-2 sm:gap-2.5">
            <Image
              src="/icons/A_logo.svg"
              alt="Artmission Logo"
              width={50}
              height={50}
              className="h-9 w-9 object-contain sm:h-[50px] sm:w-[50px]"
            />
            <div className="hidden flex-col leading-none font-extrabold text-accent-500 tracking-wider text-lg min-[480px]:flex">
              {usertype == "artist" ? (
                <>
                  <span>RTMISSION</span> <span className="mt-1 ">RTIST</span>
                </>
              ) : (
                <span className="text-[28px] sm:text-[36px]">RTMISSION</span>
              )}
            </div>
          </Link>
        </div>

        {usertype == "customer" && (
          <div className="min-w-0 flex-1 basis-0 px-1 sm:px-2 md:max-w-80 md:flex-none md:basis-80 md:px-0">
            <Suspense fallback={<div className="h-10 w-full" />}>
              <CustomerSearchBar />
            </Suspense>
          </div>
        )}

        {/* ด้านขวา: เมนู และ รูปโปรไฟล์ User */}
        <div className="relative ml-auto flex shrink-0 items-center gap-1 sm:gap-2 md:gap-4">
          <nav className="hidden items-center gap-2 md:flex">
            {navItems.map((item) => {
              const isActive = pathname === item.path;

              return (
                <Link
                  key={item.name}
                  href={item.path}
                  onClick={(e) => handleNavClick(e, isActive)}
                >
                  <Button
                    variant={isActive ? "accent-300" : "light"}
                    icon={
                      <Image
                        src={item.icon}
                        alt={item.name}
                        width={24}
                        height={24}
                        className="object-contain"
                      />
                    }
                    className={`cursor-pointer transition-all flex items-center gap-1 ${
                      !isActive
                        ? "border-none text-gray-700 hover:bg-gray-100"
                        : ""
                    }`}
                  >
                    {item.name}
                  </Button>
                </Link>
              );
            })}
          </nav>
          {/* Setting Button */}
          <div>
            <Button
              onClick={() => {
                setIsSettingPopupOpen(!isSettingPopupOpen);
              }}
              variant={pathname === routes.settings ? "accent-300" : "light"}
              icon={
                <Image
                  src="/icons/setting.svg"
                  alt="Setting"
                  width={24}
                  height={24}
                  className="object-contain"
                />
              }
              className={`cursor-pointer transition-all flex items-center gap-1 px-2 sm:px-5 ${
                pathname != routes.settings
                  ? "border-none text-gray-700 hover:bg-gray-100"
                  : ""
              }`}
              aria-label="Setting"
            >
              <span className="sr-only sm:not-sr-only sm:inline">Setting</span>
            </Button>
            <div ref={dropdownRef}>
              <SettingPopup isActive={isSettingPopupOpen} />
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
