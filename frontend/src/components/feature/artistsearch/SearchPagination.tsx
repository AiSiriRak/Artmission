"use client";

import { Button } from "@/components/ui/Button";

type PageItem = number | "ellipsis";

/**
 * Pagination items (e.g. page 2 of 67):
 *   ‹  1  2  3  …  67  ›
 *
 * Middle cluster is previous / current / next (clamped at edges).
 * Last page is always reachable on the right when not already in the cluster.
 * First page is prepended with … when the cluster has moved away from 1.
 */
export function buildPageItems(page: number, totalPages: number): PageItem[] {
  if (totalPages <= 1) return [1];

  let start = page - 1;
  let end = page + 1;

  if (start < 1) {
    start = 1;
    end = Math.min(totalPages, 3);
  }
  if (end > totalPages) {
    end = totalPages;
    start = Math.max(1, totalPages - 2);
  }

  const windowPages: number[] = [];
  for (let p = start; p <= end; p++) {
    windowPages.push(p);
  }

  const items: PageItem[] = [];

  if (windowPages[0] > 1) {
    items.push(1);
    if (windowPages[0] > 2) {
      items.push("ellipsis");
    }
  }

  items.push(...windowPages);

  const lastInWindow = windowPages[windowPages.length - 1];
  if (lastInWindow < totalPages) {
    if (lastInWindow < totalPages - 1) {
      items.push("ellipsis");
    }
    items.push(totalPages);
  }

  return items;
}

interface SearchPaginationProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export function SearchPagination({
  page,
  totalPages,
  onPageChange,
}: SearchPaginationProps) {
  if (totalPages <= 1) return null;

  const items = buildPageItems(page, totalPages);

  return (
    <div className="mb-16 flex items-center justify-center gap-2">
      <Button
        type="button"
        variant="light"
        className="size-12 min-w-12 p-0"
        disabled={page <= 1}
        aria-label="Previous page"
        onClick={() => onPageChange(Math.max(1, page - 1))}
      >
        ‹
      </Button>

      {items.map((item, index) =>
        item === "ellipsis" ? (
          <span
            key={`ellipsis-${index}`}
            className="flex size-12 items-center justify-center text-small text-primary-500"
            aria-hidden
          >
            …
          </span>
        ) : (
          <Button
            key={item}
            type="button"
            variant={item === page ? "accent-300" : "light"}
            className="size-12 min-w-12 p-0"
            aria-current={item === page ? "page" : undefined}
            onClick={() => onPageChange(item)}
          >
            {item}
          </Button>
        ),
      )}

      <Button
        type="button"
        variant="light"
        className="size-12 min-w-12 p-0"
        disabled={page >= totalPages}
        aria-label="Next page"
        onClick={() => onPageChange(Math.min(totalPages, page + 1))}
      >
        ›
      </Button>
    </div>
  );
}
