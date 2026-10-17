import type { InputHTMLAttributes, TextareaHTMLAttributes } from "react";

type CombinedProps = InputHTMLAttributes<HTMLInputElement> &
  TextareaHTMLAttributes<HTMLTextAreaElement>;

interface TextInputProps extends Omit<CombinedProps, "onChange"> {
  value: string;
  onChange: (value: string) => void;
  multiline?: boolean; // 👈 เพิ่ม prop สำหรับเปิดใช้งานหลายบรรทัด
}

export function TextInput({
  value,
  onChange,
  type = "text",
  className = "",
  multiline = false,
  ...props
}: TextInputProps) {
  const commonClasses = `w-full rounded border-2 border-primary-500 px-3 py-2 text-small
    disabled:cursor-not-allowed 
    disabled:bg-neutral-400 
    disabled:text-neutral 
    ${className}`;

  if (multiline) {
    return (
      <textarea
        {...props}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={`resize-none ${commonClasses}`}
      />
    );
  }
  
  return (
    <input
      {...props}
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={commonClasses}
    />
  );
}
