import type { Metadata } from "next";
import { LoginForm } from "./LoginForm";

export const metadata: Metadata = {
  title: "Login | Artmission",
  description: "Log in to your Artmission account.",
};

export default function LoginPage() {
  return <LoginForm />;
}
