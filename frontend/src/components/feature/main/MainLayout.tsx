import React from "react";
import { Toaster } from "sonner";

import Header from "./Header";
import Footer from "./Footer";
import { NotificationListener } from "../notification/NotificationListener";

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
      <NotificationListener />
      <Toaster
        position="top-right"
        closeButton
        expand
        toastOptions={{
          unstyled: true,
          classNames: {
            toast:
              "rounded-xl shadow-card bg-secondary-200 font-sans relative flex flex-col relative p-3",
            title: "text-body font-semibold line-clamp-1",
            description: "text-small opacity-80 line-clamp-3",
            icon: "hidden",
            closeButton:
              "absolute  bg-secondary-200 rounded-full top-3 right-3 w-5 aspect-[1/1] items-center justify-center flex hover:brightness-90",
          },
        }}
      />
    </div>
  );
}
