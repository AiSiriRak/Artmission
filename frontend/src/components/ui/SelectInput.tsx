import type { SelectHTMLAttributes } from "react";

interface SelectInputProps extends SelectHTMLAttributes<HTMLSelectElement> {
  options: {
    value: string;
    label: string;
  }[];
}

export function SelectInput({
  options,
  className = "",
  ...props
}: SelectInputProps) {
  return (
    <select
      {...props}
      className={`
    mt-1
    w-full
    rounded
    border-2
    px-3
    py-2
    pr-10
    text-small
    text-primary-500
    appearance-none
    bg-white
    bg-[url('/icons/downarrow.svg')]
    bg-no-repeat
    bg-[right_12px_center]
    bg-[length:12px]
    ${className}
  `}
    >
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}
