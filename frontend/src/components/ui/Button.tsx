import type { ButtonHTMLAttributes, ReactNode } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  icon?: ReactNode;
  variant?: "dark" | "light" | "accent-500";
}

export function Button({
  children,
  icon,
  className = "",
  variant = "light",
  ...props
}: ButtonProps) {
  const variantClass = {
    light: "rounded-lg bg-white border px-5 py-1 text-button text-primary-500",
    dark: "rounded-lg bg-primary-500 px-5 py-1 text-button text-white",
    "accent-500": "rounded-lg bg-accent-500 px-5 py-1 text-button text-white",
  };
  return (
    <button {...props} className={`${variantClass[variant]} ${className}`}>
      {icon && <span>{icon}</span>}
      {children}
    </button>
  );
}
