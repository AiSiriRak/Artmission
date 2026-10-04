"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { useRouter, useSearchParams } from "next/navigation";

import MainLayout from "@/components/feature/main/MainLayout";
import { SearchArtworkCard } from "@/components/feature/artistsearch/SearchArtworkCard";
import { SearchPagination } from "@/components/feature/artistsearch/SearchPagination";
import {
  EMPTY_FILTERS,
  FilterDrawer,
  type ArtistSearchFilters,
} from "@/components/feature/artistsearch/FilterDrawer";
import { Loading } from "@/components/ui/Loading";
import { Button } from "@/components/ui/Button";
import { SelectInput } from "@/components/ui/SelectInput";
import {
  searchArtists,
  type ArtistSearchItem,
  type ArtistSearchSort,
} from "@/lib/api/artists";
import { isApiError } from "@/lib/api/error";
import { routes } from "@/lib/routes";

const PAGE_SIZE = 12;

const SORT_OPTIONS: { value: ArtistSearchSort; label: string }[] = [
  { value: "price_asc", label: "Lowest Price" },
  { value: "price_desc", label: "Highest Price" },
  { value: "review_score_asc", label: "Lowest Review Score" },
  { value: "review_score_desc", label: "Highest Review Score" },
];

function thbToSatang(value: string): number | undefined {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const thb = Number(trimmed);
  if (Number.isNaN(thb)) return undefined;
  return Math.round(thb * 100);
}

export default function HomePage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const query = searchParams.get("q")?.trim() ?? "";

  const [sort, setSort] = useState<ArtistSearchSort>("price_asc");
  const [page, setPage] = useState(1);
  const [draftFilters, setDraftFilters] =
    useState<ArtistSearchFilters>(EMPTY_FILTERS);
  const [appliedFilters, setAppliedFilters] =
    useState<ArtistSearchFilters>(EMPTY_FILTERS);
  const [isFilterOpen, setIsFilterOpen] = useState(false);

  const [artworks, setArtworks] = useState<ArtistSearchItem[] | null>(null);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    setPage(1);
  }, [query]);

  useEffect(() => {
    let cancelled = false;

    async function loadResults() {
      setArtworks(null);
      try {
        const result = await searchArtists({
          search: query || undefined,
          sort,
          min_price_satang: thbToSatang(appliedFilters.minPriceThb),
          max_price_satang: thbToSatang(appliedFilters.maxPriceThb),
          category: appliedFilters.category ?? undefined,
          styles:
            appliedFilters.styles.length > 0
              ? appliedFilters.styles
              : undefined,
          min_rating: appliedFilters.minRating ?? undefined,
          limit: PAGE_SIZE,
          offset: (page - 1) * PAGE_SIZE,
        });
        if (!cancelled) {
          setArtworks(result.artworks);
          setTotal(result.total);
        }
      } catch (error: unknown) {
        if (isApiError(error) && error.status === 401) {
          router.replace(routes.login);
          return;
        }
        if (!cancelled) {
          setArtworks([]);
          setTotal(0);
        }
      }
    }

    loadResults();
    return () => {
      cancelled = true;
    };
  }, [query, sort, appliedFilters, page, router]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const title = query ? `Result matching “${query}”` : "Browse artists";
  const countLabel = `${total.toLocaleString()} results found`;

  return (
    <MainLayout usertype="customer">
      <div className="mx-auto min-h-screen w-full max-w-[1400px] px-4 py-6 sm:px-6 md:px-10 md:py-8">
        <div className="mb-2">
          <h1 className="text-h2 text-primary-500">{title}</h1>
          <p className="mt-1 text-body text-neutral sm:text-h3">{countLabel}</p>
        </div>

        <div className="mb-5 flex flex-wrap items-center gap-3 md:mb-6 md:gap-4">
          <Button
            type="button"
            variant="light"
            className="h-10 font-bold"
            icon={
              <Image src="/icons/filter.svg" alt="" width={24} height={24} />
            }
            onClick={() => {
              setDraftFilters(appliedFilters);
              setIsFilterOpen(true);
            }}
          >
            Filter
          </Button>
          <SelectInput
            className="w-full max-w-64 sm:w-64"
            buttonClassName="h-10 rounded-lg"
            value={sort}
            options={SORT_OPTIONS}
            formatSelectedLabel={(label) => (
              <>
                <span className="font-bold">Sort by :</span> {label}
              </>
            )}
            onChange={(next) => {
              setSort(next as ArtistSearchSort);
              setPage(1);
            }}
          />
        </div>

        {artworks === null && <Loading />}

        {artworks && artworks.length === 0 && (
          <div className="flex min-h-105 w-full flex-col items-center justify-center gap-8 rounded-[45px] border border-neutral px-6 py-16 text-center">
            <Image src="/icons/emptydoc.svg" alt="" width={96} height={96} />
            <div className="space-y-2">
              <p className="text-h1 text-neutral">No results found</p>
              <p className="max-w-xl text-body text-neutral">
                {query
                  ? `We couldn’t find any results matching “${query}”. Check the spelling, try another keyword, or adjust your filters.`
                  : "We couldn’t find any results. Try another keyword, or adjust your filters."}
              </p>
            </div>
          </div>
        )}

        {artworks && artworks.length > 0 && (
          <>
            <div className="mb-12 grid grid-cols-1 gap-x-6 gap-y-10 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 lg:gap-x-8 lg:gap-y-12">
              {artworks.map((item) => (
                <SearchArtworkCard key={item.artwork_id} item={item} />
              ))}
            </div>

            <SearchPagination
              page={page}
              totalPages={totalPages}
              onPageChange={setPage}
            />
          </>
        )}
      </div>

      <FilterDrawer
        open={isFilterOpen}
        value={draftFilters}
        onChange={setDraftFilters}
        onClose={() => setIsFilterOpen(false)}
        onReset={() => setDraftFilters(EMPTY_FILTERS)}
        onApply={() => {
          setAppliedFilters(draftFilters);
          setPage(1);
          setIsFilterOpen(false);
        }}
      />
    </MainLayout>
  );
}
