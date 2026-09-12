import type { Metadata } from "next";
import { RegisterForm } from "./RegisterForm";

export const metadata: Metadata = {
  title: "Register | Artmission",
  description: "Create your Artmission account.",
};

export default function RegisterPage() {
  return <RegisterForm />;
}
