import Link from "next/link";
import { Button } from "@/components/ui/Button";
import type { ArtistProfile } from "@/lib/api/types";
import { WhiteCard } from "@/components/ui/WhiteCard";

interface ArtistMiniProfileProps {
  artist: Pick<
    ArtistProfile,
    "artist_id" | "artist_name" | "profile_url" | "review_score"
  >;
}

export function ArtistMiniProfile({ artist }: ArtistMiniProfileProps) {
  return (
    <WhiteCard
      margin=""
      padding="p-6 sm:p-8"
      roundsize="rounded-3xl"
      className="max-w-none border border-primary-500 bg-secondary-200 shadow-none"
    >
      <div className="flex items-center gap-4">
        <div className="w-20 h-20 rounded-full overflow-hidden">
          <img
            src={artist.profile_url || "/default-avatar.png"}
            alt="Artist Avatar"
            className="w-full h-full object-cover"
          />
        </div>

        <div className="flex flex-row gap-4 items-start">
          <div className="flex flex-col gap-2 items-start">
            <h3 className="text-body font-semibold">Profile Name</h3>

            <span className="text-body">{artist.artist_name}</span>
            <div className="flex items-center gap-1">
              <span
                role="img"
                aria-label="star icon"
                className="inline-block w-6 h-6 bg-accent-500"
                style={{
                  maskImage: "url('/icons/star.svg')",
                  maskRepeat: "no-repeat",
                  maskPosition: "center",
                  maskSize: "contain",
                }}
              />
              <span className="text-body">
                {artist.review_score !== null
                  ? `${artist.review_score.toFixed(1)}/5`
                  : "-/5"}
              </span>
            </div>
          </div>
          <div className="ml-auto">
            <Link
              href={`/artists/${artist.artist_id}`}
              className="items-start ml-auto"
            >
              <Button
                variant="light"
                className="text-button flex items-center gap-2 !border"
              >
                <img
                  src="/icons/external-link-icon.svg"
                  alt="external link icon"
                  className="w-4 h-4"
                />{" "}
                View
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </WhiteCard>
  );
}
