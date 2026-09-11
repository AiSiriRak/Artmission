"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";

interface CheckboxDropdown {
  options: {
    value: string;
    label: string;
  }[];
  onChange?: (selected: string[]) => void;
}

export function CheckboxDropdown({ options, onChange }: CheckboxDropdown) {
  const [isOpen, setIsOpen] = useState(false);
  const [selected, setSelected] = useState<string[]>(
    options.map((option) => option.value),
  );

  const dropdownRef = useRef<HTMLDivElement>(null);

  function toggleOption(option: string) {
    setSelected((prev) =>
      prev.includes(option)
        ? prev.filter((item) => item !== option)
        : [...prev, option],
    );
  }

  useEffect(() => {
    onChange?.(selected);
  }, [selected, onChange]);

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
          {selected.length === options.length
            ? "All selected"
            : `${selected.length} selected`}
          <Image
            src={"/icons/downarrow.svg"}
            width={12}
            height={12}
            alt={""}
          ></Image>
        </div>
      </button>

      {/* Dropdown menu */}
      {isOpen && (
        <div className="absolute z-50 w-full rounded-lg bg-white p-3 shadow-card">
          {options.map((option) => {
            const isSelected = selected.includes(option.value);

            return (
              <button
                key={option.value}
                type="button"
                onClick={() => toggleOption(option.value)}
                className="flex w-full cursor-pointer items-center gap-4 px-2 py-2 text-left hover:bg-neutral-100 rounded"
              >
                {/* Custom checkbox */}
                <div
                  className={`
                    flex h-5 w-5 shrink-0 items-center justify-center
                    border-2 border-accent-500
                    ${isSelected ? "bg-accent-500" : "bg-white"}
                  `}
                >
                  {isSelected && (
                    <Image
                      src="/icons/check.svg"
                      width={15}
                      height={15}
                      alt=""
                    />
                  )}
                </div>

                <span className="text-subtle text-primary-500">
                  {option.label}
                </span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
