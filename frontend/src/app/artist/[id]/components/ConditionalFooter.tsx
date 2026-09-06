"use client";

import { usePathname } from "next/navigation";
import Footer from "@/app/artist/[id]/components/Footer"; // Path ตามที่คุณ import ไว้

export default function ConditionalFooter() {
  const pathname = usePathname();

  // ระบุ path ของหน้าที่ "ไม่ต้องการ" ให้แสดง Footer
  const hideFooterRoutes = ["/register", "/login"];

  // เช็คว่าหน้าปัจจุบันตรงกับเส้นทางที่ต้องการซ่อนหรือไม่
  const isHidden = hideFooterRoutes.some((route) => pathname.startsWith(route));

  if (isHidden) {
    return null; // ถ้าตรง ให้ซ่อน Footer ทันที
  }

  return <Footer />;
}