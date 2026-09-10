"use client";

import Link from "next/link";
import Image from "next/image";

import { PageType } from "@/lib/types";
import { routes } from "@/lib/routes";

import { Button } from "@/components/ui/Button";
import { TextInput } from "@/components/ui/TextInput";

interface HeaderProps {
  page: PageType;
  usertype: string;
}

export default function Header({
  page: activeMenu = "Login",
  usertype = "customer",
}: HeaderProps) {
  // กำหนดรายการเมนู
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
          {
            name: "Setting",
            page: "Setting",
            path: routes.settings,
            icon: "/icons/setting.svg",
          },
        ]
      : [
          {
            name: "Order",
            page: "Order Artist",
            path: routes.order.artist,
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
          {
            name: "Setting",
            page: "Setting",
            path: routes.settings,
            icon: "/icons/setting.svg",
          },
        ];

  // ฟังก์ชันจัดการการคลิกปุ่ม
  const handleNavClick = (e: React.MouseEvent, isActive: boolean) => {
    // ถ้าคลิกปุ่มของหน้าปัจจุบันอยู่แล้ว ให้ทำการ Refresh หน้าเดิม
    if (isActive) {
      e.preventDefault(); // ป้องกันไม่ให้ Link นำทางไป URL ใหม่
      window.location.reload(); // รีเฟรชหน้าปัจจุบัน
    }
  };

  return (
    <header className="sticky top-0 z-50 w-full bg-white border-b border-primary-500">
      <div className="max-w-[1400px] mx-auto px-6 h-20 flex items-center justify-between">
        {/* ด้านซ้าย: Logo ARTMISSION */}
        <div className="flex-shrink-0 flex items-center">
          <Link href="/" className="flex items-center gap-2.5">
            <img
              src="/icons/A_logo.svg"
              alt="Artmission Logo"
              className="w-[50px] h-[50px] object-contain"
            />
            <div className="flex flex-col leading-none font-extrabold text-accent-500 tracking-wider text-lg">
              {usertype == "artist" ? (
                <>
                  <span>RTMISSION</span> <span className="mt-1 ">RTIST</span>
                </>
              ) : (
                <>
                  <span className="text-[36px]">RTMISSION</span>
                </>
              )}
            </div>
          </Link>
        </div>{" "}
        {/* Seach Bar (No Search Implemented) */}
        {usertype == "customer" && (
          <div className="relative w-80 h-10 items-center">
            <TextInput onChange={() => {}} className="absolute inset-0 pr-12" />
            <button
              type="button"
              className="absolute flex rounded-full right-2 top-1/2 h-6 w-6 justify-center items-center bg-white -translate-y-1/2 hover:brightness-90"
            >
              <Image
                src="/icons/search.svg"
                alt="Search"
                width={10}
                height={10}
              />
            </button>
          </div>
        )}
        {/* ด้านขวา: เมนู และ รูปโปรไฟล์ User */}
        <div className="flex items-center gap-4">
          <nav className="hidden md:flex items-center gap-2">
            {navItems.map((item) => {
              const isActive = activeMenu === item.page;

              return (
                <Link
                  key={item.name}
                  href={item.path}
                  onClick={(e) => handleNavClick(e, isActive)}
                >
                  <Button
                    variant={isActive ? "accent-300" : "light"}
                    icon={
                      <img
                        src={item.icon}
                        alt={item.name}
                        className="w-5 h-5 object-contain"
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
        </div>
      </div>
    </header>
  );
}
