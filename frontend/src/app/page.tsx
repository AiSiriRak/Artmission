import MainLayout from "@/components/feature/main/MainLayout";
import Image from "next/image";

export default function Home() {
  return (
    <MainLayout page={"Home"} usertype={"customer"}>
      <div>(Home Page)</div>
    </MainLayout>
  );
}
