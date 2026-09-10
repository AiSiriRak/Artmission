"use client";

import MainLayout from "@/components/feature/main/MainLayout";

export default function HomePage() {
  return (
    <MainLayout page={"Home"} usertype={"customer"}>
      <div>(Home Page)</div>
    </MainLayout>
  );
}
