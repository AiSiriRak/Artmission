"use client";

import Image from "next/image";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/Button";
import { TextInput } from "@/components/ui/TextInput";
import {
  getCategories,
  getStyles,
  type CatalogReference,
} from "@/lib/api/artworks";

export type ArtistSearchFilters = {
  minPriceThb: string;
  maxPriceThb: string;
  category: string | null;
  styles: string[];
  minRating: number | null;
};

export const EMPTY_FILTERS: ArtistSearchFilters = {
  minPriceThb: "0",
  maxPriceThb: "",
  category: null,
  styles: [],
  minRating: null,
};

const RATING_OPTIONS: { value: number | null; label: string }[] = [
  { value: 5, label: "5★" },
  { value: 4, label: "4★" },
  { value: 3, label: "3★" },
  { value: 2, label: "2★" },
  { value: 1, label: "1★" },
  { value: null, label: "Any Rating" },
];

/**
 * Controlled decimal draft for THB prices.
 * Prefer type="text" + inputMode="decimal" over type="number":
 * browsers can display invalid number input without updating React state.
 */
function sanitizePriceThb(raw: string): string {
  const cleaned = raw.replace(/[^\d.]/g, "");
  if (!cleaned) return "";

  const [whole = "", ...rest] = cleaned.split(".");
  if (rest.length === 0) return whole;

  const fraction = rest.join("").slice(0, 2);
  return `${whole}.${fraction}`;
}

function parsePriceThb(raw: string): number | null {
  const trimmed = raw.trim();
  if (!trimmed || trimmed === ".") return null;
  const value = Number(trimmed);
  return Number.isFinite(value) ? value : null;
}

/** True when both bounds are set and max is strictly less than min. */
function isPriceRangeInvalid(minRaw: string, maxRaw: string): boolean {
  const min = parsePriceThb(minRaw);
  const max = parsePriceThb(maxRaw);
  if (min === null || max === null) return false;
  return max < min;
}

interface FilterDrawerProps {
  open: boolean;
  value: ArtistSearchFilters;
  onChange: (next: ArtistSearchFilters) => void;
  onClose: () => void;
  onReset: () => void;
  onApply: () => void;
}

export function FilterDrawer({
  open,
  value,
  onChange,
  onClose,
  onReset,
  onApply,
}: FilterDrawerProps) {
  const [categories, setCategories] = useState<CatalogReference[]>([]);
  const [styles, setStyles] = useState<CatalogReference[]>([]);
  const [catalogError, setCatalogError] = useState(false);
  const [catalogLoading, setCatalogLoading] = useState(false);

  const catalogLoadedRef = useRef(false);

  useEffect(() => {
    if (!open || catalogLoadedRef.current) return;

    let cancelled = false;

    async function loadCatalog() {
      setCatalogLoading(true);
      setCatalogError(false);
      try {
        const [nextCategories, nextStyles] = await Promise.all([
          getCategories(),
          getStyles(),
        ]);
        if (!cancelled) {
          setCategories(nextCategories);
          setStyles(nextStyles);
          catalogLoadedRef.current = true;
        }
      } catch {
        if (!cancelled) {
          setCatalogError(true);
        }
      } finally {
        if (!cancelled) {
          setCatalogLoading(false);
        }
      }
    }

    void loadCatalog();
    return () => {
      cancelled = true;
    };
  }, [open]);

  if (!open) return null;

  const priceRangeInvalid = isPriceRangeInvalid(
    value.minPriceThb,
    value.maxPriceThb,
  );

  function toggleStyle(style: string) {
    const nextStyles = value.styles.includes(style)
      ? value.styles.filter((item) => item !== style)
      : [...value.styles, style];
    onChange({ ...value, styles: nextStyles });
  }

  return (
    <div className="fixed inset-0 z-60 flex">
      <button
        type="button"
        aria-label="Close filters backdrop"
        className="absolute inset-0 bg-black/20"
        onClick={onClose}
      />

      <aside className="relative z-10 flex h-full w-full max-w-145 flex-col bg-white shadow-card">
        <div className="flex items-center justify-between border-b border-neutral-400 px-8 py-5">
          <h2 className="text-h2 text-primary-500">Filters</h2>
          <button
            type="button"
            onClick={onClose}
            className="flex size-11 items-center justify-center rounded-full bg-secondary-200 hover:brightness-95"
            aria-label="Close filters"
          >
            <Image src="/icons/close_round.svg" alt="" width={28} height={28} />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-8">
          <section className="space-y-3 border-b border-neutral-400 py-6">
            <p className="text-body text-primary-500">Price range</p>
            <div className="grid grid-cols-2 gap-3">
              <label className="space-y-1">
                <span className="text-small text-neutral">Minimum</span>
                <TextInput
                  type="text"
                  inputMode="decimal"
                  autoComplete="off"
                  value={value.minPriceThb}
                  onChange={(minPriceThb) =>
                    onChange({
                      ...value,
                      minPriceThb: sanitizePriceThb(minPriceThb),
                    })
                  }
                  placeholder="0"
                />
              </label>
              <label className="space-y-1">
                <span className="text-small text-neutral">Maximum</span>
                <TextInput
                  type="text"
                  inputMode="decimal"
                  autoComplete="off"
                  value={value.maxPriceThb}
                  onChange={(maxPriceThb) =>
                    onChange({
                      ...value,
                      maxPriceThb: sanitizePriceThb(maxPriceThb),
                    })
                  }
                  placeholder="Maximum Price"
                />
              </label>
            </div>
            {priceRangeInvalid && (
              <p className="text-small text-error">
                Maximum price must be greater than or equal to minimum price.
              </p>
            )}
          </section>

          <section className="space-y-3 border-b border-neutral-400 py-6">
            <p className="text-body text-primary-500">Category</p>
            {catalogLoading && (
              <p className="text-small text-neutral">Loading categories…</p>
            )}
            {catalogError && (
              <p className="text-small text-error">
                Couldn’t load categories. Try again later.
              </p>
            )}
            <div className="flex flex-wrap gap-2">
              {categories.map((category) => {
                const selected = value.category === category.label;
                return (
                  <button
                    key={category.id}
                    type="button"
                    onClick={() =>
                      onChange({
                        ...value,
                        category: selected ? null : category.label,
                      })
                    }
                    className={`rounded-full border px-4 py-2 text-button transition-colors ${
                      selected
                        ? "border-accent-500 bg-accent-500 text-white"
                        : "border-neutral bg-white text-primary-500 hover:bg-neutral-200"
                    }`}
                  >
                    {category.label}
                  </button>
                );
              })}
            </div>
          </section>

          <section className="space-y-3 border-b border-neutral-400 py-6">
            <p className="text-body text-primary-500">Style</p>
            {catalogLoading && (
              <p className="text-small text-neutral">Loading styles…</p>
            )}
            {catalogError && (
              <p className="text-small text-error">
                Couldn’t load styles. Try again later.
              </p>
            )}
            <div className="grid grid-cols-2 gap-x-3 gap-y-3 sm:grid-cols-3">
              {styles.map((style) => {
                const checked = value.styles.includes(style.label);
                return (
                  <label
                    key={style.id}
                    className="flex cursor-pointer items-center gap-2 text-small text-primary-500"
                  >
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggleStyle(style.label)}
                      className="peer sr-only"
                    />
                    <span
                      className={`flex size-5 shrink-0 items-center justify-center rounded border-2 ${
                        checked
                          ? "border-accent-500 bg-accent-500"
                          : "border-neutral bg-white"
                      }`}
                    >
                      {checked && (
                        <Image
                          src="/icons/check.svg"
                          alt=""
                          width={12}
                          height={12}
                        />
                      )}
                    </span>
                    <span className="leading-tight">{style.label}</span>
                  </label>
                );
              })}
            </div>
          </section>

          <section className="space-y-3 py-6">
            <p className="text-body text-primary-500">Minimum Rating</p>
            <div className="grid grid-cols-2 gap-y-2">
              {RATING_OPTIONS.map((option) => {
                const selected = value.minRating === option.value;
                return (
                  <label
                    key={option.label}
                    className="flex cursor-pointer items-center gap-3 text-small text-primary-500"
                  >
                    <input
                      type="radio"
                      name="min-rating"
                      checked={selected}
                      onChange={() =>
                        onChange({ ...value, minRating: option.value })
                      }
                      className="sr-only"
                    />
                    <span
                      className={`flex size-5 shrink-0 items-center justify-center rounded-full border-2 ${
                        selected ? "border-accent-500" : "border-neutral"
                      }`}
                    >
                      {selected && (
                        <span className="size-2.5 rounded-full bg-accent-500" />
                      )}
                    </span>
                    <span>{option.label}</span>
                  </label>
                );
              })}
            </div>
          </section>
        </div>

        <div className="flex justify-end gap-3 border-t border-neutral-400 px-8 py-5">
          <Button
            type="button"
            variant="light"
            className="min-w-30.75"
            onClick={onReset}
          >
            Reset all
          </Button>
          <Button
            type="button"
            variant="accent-500"
            className="min-w-30.75"
            disabled={priceRangeInvalid}
            onClick={onApply}
          >
            Show results
          </Button>
        </div>
      </aside>
    </div>
  );
}
