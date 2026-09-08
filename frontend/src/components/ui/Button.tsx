import type { ButtonHTMLAttributes, ReactNode } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  icon?: ReactNode;
  variant?: "dark" | "light" | "accent-500" | "accent-300" | "transparent" | "error" | "red" | "disable";
}

export function Button({
  children,
  icon,
  className = "",
  variant = "light",
  disabled,
  ...props
}: ButtonProps) {
  const variantClass = {
    light: "bg-white border-2 text-primary-500",
    dark: "bg-primary-500 text-white",
    "accent-500": "bg-accent-500 text-white",
    red: "bg-error border-primary-500 border-2 text-white",
    disable: "bg-neutral-400 border-2 text-white",
    "accent-300": "bg-accent-300 text-primary-500",
    transparent: "bg-transparent border text-primary-500",
    error: "bg-error text-white"
  };

  const currentVariant = disabled ? "disable" : variant;

  return (
    <button
      {...props}
      disabled={disabled}
      className={`${variantClass[currentVariant]} flex items-center justify-center gap-2 px-5 py-2 rounded-lg text-button transition-colors cursor-pointer hover:brightness-90 ${className}`}
    >
      {icon && <span>{icon}</span>}
      {children}
    </button>
  );
}