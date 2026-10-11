import type { ArtistProfile } from "@/lib/api/types";

interface ArtistMiniProfileProps {
  artist: Pick<ArtistProfile, "artist_name" | "profile_url" | "review_score">;
}

export function ArtistMiniProfile({ artist }: ArtistMiniProfileProps) {
  return (
    <div className="flex items-center gap-4">
      <div className="w-20 h-20 rounded-full overflow-hidden">
        <img
          src={artist.profile_url || "/default-avatar.png"}
          alt="Artist Avatar"
          className="w-full h-full object-cover"
        />
      </div>
      <div className="flex flex-col gap-2">
        <h3 className="text-h3">Artist Profile</h3>
        <span className="text-body">{artist.artist_name}</span>
        <div className="flex items-center gap-1">
          <span className="text-h3">
            {artist.review_score !== null
              ? `${artist.review_score.toFixed(1)}/5`
              : "No reviews"}
          </span>
        </div>
      </div>
    </div>
  );
}
