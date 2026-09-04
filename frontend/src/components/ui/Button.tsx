import type { ButtonHTMLAttributes, ReactNode } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  icon?: ReactNode;
  variant?: "dark" | "light";
}

export function Button({
  children,
  icon,
  className = "",
  variant = "light",
  ...props
}: ButtonProps) {
  const variantClass = {
    light: "rounded-lg bg-white border px-5 py-1 text-small text-primary-500",
    dark: "rounded-lg bg-primary-500 px-5 py-1 text-small text-white",
  };
  return (
    <button {...props} className={`${variantClass[variant]} ${className}`}>
      {icon && <span>{icon}</span>}
      {children}
    </button>
  );
}
