"use client";

import Link from "next/link";
import { Button } from "@/components/ui/Button";

interface HeaderProps {
  activeMenu?: "Order" | "Artist Profile" | "Notification" | "Setting";
  orderPath?: string;
  artistProfilePath?: string;
  notificationPath?: string;
  settingPath?: string;
}

export default function Header({
  activeMenu = "Artist Profile",
  orderPath = "/order",
  artistProfilePath = "/artist-profile",
  notificationPath = "/notification",
  settingPath = "/setting",
}: HeaderProps) {
  // กำหนดรายการเมนู
  const navItems = [
    { name: "Order", path: orderPath, icon: "/icons/order.svg" },
    { name: "Artist Profile", path: artistProfilePath, icon: "/icons/artist_profile.svg" },
    { name: "Notification", path: notificationPath, icon: "/icons/notification.svg" },
    { name: "Setting", path: settingPath, icon: "/icons/setting.svg" },
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
                {/* 1. รูปไอคอนตัว A (ขาตั้งวาดรูป) */}
                <img 
                src="/icons/A_logo.svg" 
                alt="Artmission Logo" 
                className="w-9 h-10 object-contain" 
                />
                
                {/* 2. ข้อความ 2 บรรทัดเรียงลงมา */}
                <div className="flex flex-col leading-none font-extrabold text-accent-500 tracking-wider text-lg">
                <span>RTMISSION</span>
                <span className="mt-1">RTIST</span>
                </div>
            </Link>
        </div>

        {/* ด้านขวา: เมนู และ รูปโปรไฟล์ User */}
        <div className="flex items-center gap-4">
          <nav className="hidden md:flex items-center gap-2">
            {navItems.map((item) => {
              const isActive = activeMenu === item.name;

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
                      !isActive ? "border-none text-gray-700 hover:bg-gray-100" : ""
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