"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import Image from "next/image";

import { cn } from "@/lib/cn";

interface SelectInputProps {
  options: {
    value: string;
    label: string;
  }[];
  value?: string;
  onChange?: (value: string) => void;
  /** Customize the closed-trigger label. Defaults to the selected option label. */
  formatSelectedLabel?: (label: string) => ReactNode;
  className?: string;
  buttonClassName?: string;
}

export function SelectInput({
  options,
  value,
  onChange,
  formatSelectedLabel,
  className,
  buttonClassName,
}: SelectInputProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const selectedOption = options.find((option) => option.value === value);
  const selectedLabel = selectedOption?.label ?? "Select";

  function selectOption(option: string) {
    onChange?.(option);
    setIsOpen(false);
  }

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  return (
    <div className={cn("relative", className)} ref={dropdownRef}>
      {/* Dropdown button */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "w-full rounded border-2 px-3 py-2 text-primary-500 text-small text-left bg-white hover:brightness-90",
          buttonClassName,
        )}
      >
        <div className="flex justify-between items-center gap-2">
          <span className="truncate">
            {formatSelectedLabel
              ? formatSelectedLabel(selectedLabel)
              : selectedLabel}
          </span>

          <Image src="/icons/downarrow.svg" width={12} height={12} alt="" />
        </div>
      </button>

      {/* Dropdown menu */}
      {isOpen && (
        <div className="absolute z-50 w-full rounded-lg bg-white p-3 shadow-card">
          {options.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => selectOption(option.value)}
              className="flex w-full cursor-pointer items-center px-2 py-2 rounded text-left text-subtle text-primary-500 hover:bg-neutral-100"
            >
              {option.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
