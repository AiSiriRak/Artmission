"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";

interface CheckboxDropdown {
  options: {
    value: string;
    label: string;
  }[];
}

export function CheckboxDropdown({ options }: CheckboxDropdown) {
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
          {options.map((option) => (
            <label
              key={option.value}
              className="flex cursor-pointer items-center gap-4 px-2 py-2"
            >
              <input
                type="checkbox"
                checked={selected.includes(option.value)}
                onChange={() => toggleOption(option.value)}
              />

              <span className="text-subtle text-primary-500">
                {option.value}
              </span>
            </label>
          ))}
        </div>
      )}
    </div>
  );
}
