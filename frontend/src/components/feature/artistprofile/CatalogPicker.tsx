import { useState } from "react";
import {
  displayLabel,
  type CatalogOption,
} from "./artworkCatalog";

type CatalogPickerProps = {
  variant: "category" | "style";
  options: CatalogOption[];
  selectedIds: string[];
  error: string;
  onSelect: (id: string) => void;
  onRemove: (id: string) => void;
};

const pickerStyle = {
  category: {
    label: "Category:",
    labelClassName: "text-body font-bold text-primary-500 block mb-3",
    placeholder: "Search category...",
    optionClassName: "hover:bg-gray-100 cursor-pointer text-gray-800",
    chipClassName:
      "px-4 py-1.5 bg-accent-200 text-primary-500 rounded-full text-sm font-semibold flex items-center gap-2 shadow-sm",
    removeClassName: "text-gray-600 hover:text-black cursor-pointer leading-none",
    chipsClassName: "flex mt-3",
  },
  style: {
    label: "Style:",
    labelClassName: "text-body font-bold text-gray-900 block mb-3",
    placeholder: "Search style...",
    optionClassName: "hover:bg-gray-100 cursor-pointer text-primary-400",
    chipClassName:
      "px-4 py-1.5 bg-secondary-600 text-gray-900 rounded-full text-sm font-semibold flex items-center gap-2 shadow-sm",
    removeClassName: "text-gray-700 hover:text-black cursor-pointer leading-none",
    chipsClassName: "flex flex-wrap gap-2 mt-3",
  },
} as const;

export default function CatalogPicker({
  variant,
  options,
  selectedIds,
  error,
  onSelect,
  onRemove,
}: CatalogPickerProps) {
  const appearance = pickerStyle[variant];
  const [search, setSearch] = useState("");
  const [open, setOpen] = useState(false);
  const visibleOptions = options.filter((option) =>
    option.label.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div className="relative">
      <label className={appearance.labelClassName}>{appearance.label}</label>
      <input
        type="text"
        value={search}
        onChange={(event) => {
          setSearch(event.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => setTimeout(() => setOpen(false), 200)}
        placeholder={appearance.placeholder}
        className="border border-primary-500 p-2.5 w-full max-w-xs rounded-lg bg-white outline-none focus:border-black"
      />

      {open && (
        <div className="absolute z-10 w-full max-w-xs mt-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto">
          {visibleOptions.length > 0 ? (
            visibleOptions.map((option) => (
              <div
                key={option.label}
                onMouseDown={() => {
                  if (option.disabled) return;
                  onSelect(option.id);
                  setSearch("");
                  setOpen(false);
                }}
                aria-disabled={option.disabled}
                className={`px-4 py-2 text-sm ${
                  option.disabled
                    ? "text-gray-400 cursor-not-allowed"
                    : appearance.optionClassName
                }`}
              >
                {option.label}
              </div>
            ))
          ) : (
            <div className="px-4 py-2 text-sm text-gray-400">No results found</div>
          )}
        </div>
      )}

      {selectedIds.length > 0 && (
        <div className={appearance.chipsClassName}>
          {selectedIds.map((id) => (
            <span key={id} className={appearance.chipClassName}>
              {displayLabel(options, id)}
              <button
                onClick={() => onRemove(id)}
                className={appearance.removeClassName}
              >
                ✕
              </button>
            </span>
          ))}
        </div>
      )}

      <p className="text-red-500 text-sm mt-1 min-h-[20px]">{error || ""}</p>
    </div>
  );
}
