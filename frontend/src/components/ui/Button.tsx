import type { ButtonHTMLAttributes, ReactNode } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  icon?: ReactNode;
  variant?: "dark" | "light" | "accent-500" | "red" | "disable";
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
    light: " bg-white border-2  text-primary-500",
    dark: " bg-primary-500  text-white",
    "accent-500": " bg-accent-500   text-white",
    red: " bg-error border-primary-500 border-2   text-white",
    disable: " bg-neutral-400  border-2 text-white",
  };

  const currentVariant = disabled ? "disable" : variant;

  return (
    <button
      {...props}
      disabled={disabled}
      className={`${variantClass[currentVariant]} ${className} text-button px-5 py-2 rounded-lg hover:brightness-90`}
    >
      {icon && <span>{icon}</span>}
      {children}
    </button>
  );
}
