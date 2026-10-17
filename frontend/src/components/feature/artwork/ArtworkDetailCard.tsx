import type { Artwork } from "@/lib/api/types";
import TagList from "@/components/feature/artistprofile/TagList";

export interface ArtworkDetailCardProps {
  artwork: Pick<
    Artwork,
    "name" | "category" | "styles" | "description" | "artwork_samples"
  >;
}

export default function ArtworkDetailCard({ artwork }: ArtworkDetailCardProps) {
  return (

    <div className="w-7/10 border border-primary-500 rounded-3xl p-10 bg-secondary-200 shadow-sm">
      <div className="">
          <div className="p-4 flex flex-col flex-1 gap-6">
            {/* Artwork Info section*/}
            <div className="flex flex-col gap-2">
              <h1 className="text-h1">{artwork.name}</h1>
              <div className="flex flex-wrap gap-2">
                <TagList items={artwork.category ? [artwork.category] : []} variant="category" />
                <TagList items={artwork.styles ?? []} variant="style" />
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <h3 className="text-h3">Description</h3>
              <p className="text-body break-words">{artwork.description}</p>
            </div>

            {/* Artwork Sample Gallery section */}
            <h3 className="text-h3">Artwork Samples</h3>
            <div className="flex justify-center items-center gap-2">
              <div className="flex flex-wrap justify-center gap-4">
                {(artwork.artwork_samples ?? []).map(({ image_url: imgUrl }, index) => (
                  <div
                    key={index}
                    className="w-96 h-96 bg-white border border-gray-200 rounded-xl relative p-2 flex items-center justify-center"
                  >
                    <img
                      src={imgUrl}
                      alt={`sample-${index}`}
                      className="max-w-full max-h-full object-contain"
                    />
                  </div>
                ))}
              </div>
            </div>
        </div>
      </div>
    </div>
  );
}
