import type { InputHTMLAttributes } from "react";

interface TextInput extends Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "onChange"
> {
  onChange: (value: string) => void;
}

export function TextInput({
  value,
  onChange,
  type = "text",
  className = "",
  ...props
}: TextInput) {
  return (
    <input
      {...props}
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={`mt-1 w-full rounded border-2 border-primary-500 px-3 py-2 text-small
        disabled:cursor-not-allowed 
        disabled:bg-neutral-400 
        disabled:text-neutral 
        ${className}`}
    />
  );
}
