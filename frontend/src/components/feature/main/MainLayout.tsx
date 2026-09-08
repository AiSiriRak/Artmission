import React from "react";
import Header from "./Header";
import Footer from "./Footer";

// ดึง Type ของ Props มาจาก Header เพื่อให้ TypeScript รู้ว่าเราสามารถส่งอะไรไปให้ Header ได้บ้าง
interface HeaderProps {
  activeMenu?: "Order" | "Artist Profile" | "Notification" | "Setting";
  orderPath?: string;
  artistProfilePath?: string;
  notificationPath?: string;
  settingPath?: string;
}

interface MainLayoutProps {
  children: React.ReactNode; // เนื้อหาของหน้าเว็บ
  showHeader?: boolean;      // กำหนดว่าจะแสดง Header ไหม (ค่าเริ่มต้น: true)
  showFooter?: boolean;      // กำหนดว่าจะแสดง Footer ไหม (ค่าเริ่มต้น: true)
  headerProps?: HeaderProps; // สำหรับส่งค่าไปเปลี่ยนข้อมูลใน Header
}

export default function MainLayout({
  children,
  showHeader = true,
  showFooter = true,
  headerProps,
}: MainLayoutProps) {
  return (
    <div className="min-h-screen flex flex-col">
      {/* ถ้า showHeader เป็น true จะแสดง Header พร้อมแนบ Props (ถ้ามี) ไปให้ */}
      {showHeader && <Header {...headerProps} />}

      {/* เนื้อหาหลักของเพจ (flex-grow จะช่วยดันให้ Footer ตกลงไปอยู่ด้านล่างสุดเสมอ) */}
      <main className="flex-grow">{children}</main>

      {/* ถ้า showFooter เป็น true จะแสดง Footer */}
      {showFooter && <Footer />}
    </div>
  );
}