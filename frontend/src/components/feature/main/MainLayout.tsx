import React from "react";

import Header from "./Header";
import Footer from "./Footer";
import { PageType, UserType } from "@/lib/types";

interface MainLayoutProps {
  children: React.ReactNode;
  showHeader?: boolean;
  showFooter?: boolean;
  page: PageType;
  usertype: UserType;
}

export default function MainLayout({
  children,
  showHeader = true,
  showFooter = true,
  page = "Home",
  usertype = "customer",
}: MainLayoutProps) {
  return (
    <div className="min-h-screen flex flex-col">
      {/* ถ้า showHeader เป็น true จะแสดง Header*/}
      {showHeader && <Header page={page} usertype={usertype} />}

      {/* เนื้อหาหลักของเพจ (flex-grow จะช่วยดันให้ Footer ตกลงไปอยู่ด้านล่างสุดเสมอ) */}
      <main className="flex-grow">{children}</main>

      {/* ถ้า showFooter เป็น true จะแสดง Footer */}
      {showFooter && <Footer />}
    </div>
  );
}
