import type { ButtonHTMLAttributes, ReactNode } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  icon?: ReactNode;
  variant?: "dark" | "light" | "accent-500" | "accent-300" | "transparent" | "error";
}

export function Button({
  children,
  icon,
  className = "",
  variant = "light",
  ...props
}: ButtonProps) {
  const variantClass = {
    light: "rounded-lg bg-white border px-5 py-1 text-button text-primary-500 hover:bg-neutral-200",
    dark: "rounded-lg bg-primary-500 px-5 py-1 text-button text-white hover:opacity-80",
    "accent-500": "rounded-lg bg-accent-500 px-5 py-1 text-button text-white hover:opacity-80",
    "accent-300": "rounded-lg bg-accent-300 px-5 py-1 text-button text-primary-500 hover:opacity-80",
    // เพิ่มแบบโปร่งใส (มีแค่ขอบ)
    "transparent": "rounded-lg bg-transparent border px-5 py-1 text-button text-primary-500 hover:bg-neutral-200",
    
    // เพิ่มแบบ Error (สีแดง)
    "error": "rounded-lg bg-error px-5 py-1 text-button text-white hover:opacity-80"
  };
  return (
    <button {...props} className={`cursor-pointer transition-colors flex items-center justify-center gap-2 ${variantClass[variant]} ${className}`}>
      {icon && <span>{icon}</span>}
      {children}
    </button>
  );
}
