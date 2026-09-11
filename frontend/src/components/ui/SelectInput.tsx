"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";

interface SelectInput {
  options: {
    value: string;
    label: string;
  }[];
  value?: string;
  onChange?: (value: string) => void;
}

export function SelectInput({ options, value, onChange }: SelectInput) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const selectedOption = options.find((option) => option.value === value);

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
    <div className="relative" ref={dropdownRef}>
      {/* Dropdown button */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="mt-1 w-full rounded border-2 px-3 py-2 text-small text-primary-500 text-left bg-white hover:brightness-90"
      >
        <div className="flex justify-between items-center">
          <span>{selectedOption?.label ?? "Select"}</span>

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
              className={`
                flex w-full cursor-pointer items-center
                px-2 py-2
                text-left
                text-subtle
                text-primary-500
                hover:bg-neutral-100
                rounded
              `}
            >
              {option.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
