import React from "react";

import Header from "./Header";
import Footer from "./Footer";

interface MainLayoutProps {
  children: React.ReactNode;
  showHeader?: boolean;
  showFooter?: boolean;
  usertype: string;
}

export default function MainLayout({
  children,
  showHeader = true,
  showFooter = true,
  usertype = "customer",
}: MainLayoutProps) {
  return (
    <div className="min-h-screen flex flex-col">
      {/* ถ้า showHeader เป็น true จะแสดง Header*/}
      {showHeader && <Header usertype={usertype} />}

      {/* เนื้อหาหลักของเพจ (flex-grow จะช่วยดันให้ Footer ตกลงไปอยู่ด้านล่างสุดเสมอ) */}
      <main className="flex-grow">{children}</main>

      {/* ถ้า showFooter เป็น true จะแสดง Footer */}
      {showFooter && <Footer />}
    </div>
  );
}
