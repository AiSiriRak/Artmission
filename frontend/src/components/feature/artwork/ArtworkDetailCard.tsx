import Image from "next/image";
import type { Artwork } from "@/lib/api/types";

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
        <div className="pt-0 p-4 flex flex-col flex-1 gap-6">
          {/* Artwork Info section*/}
          <div className="flex flex-col gap-2">
            <h1 className="text-h1">{artwork.name}</h1>
            <div className="flex flex-wrap gap-2">
              <p className="px-4 py-1.5 bg-accent-200 text-primary-400 rounded-full text-sm font-semibold shadow-sm">
                {artwork.category}
              </p>
              {(artwork.styles ?? []).map((style, index) => (
                <span
                  key={index}
                  className="px-4 py-1.5 bg-secondary-600 text-gray-900 rounded-full text-sm font-semibold shadow-sm"
                >
                  {style}
                </span>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <h3 className="text-h4">Description</h3>
            <p className="text-caption break-words">{artwork.description}</p>
          </div>

          {/* Artwork Sample Gallery section */}
          <h3 className="text-h4">Artwork Samples</h3>
          <div className="flex gap-4 overflow-x-auto pb-3">
            {(artwork.artwork_samples ?? []).map(
              ({ image_url: imgUrl }, index) => (
                <div
                  key={index}
                  className="relative h-48 w-fit flex-none overflow-hidden rounded-xl"
                >
                  <Image
                    src={imgUrl}
                    alt={`sample-${index}`}
                    width={400}
                    height={400}
                    className="h-48 w-auto max-w-none object-contain"
                  />
                </div>
              ),
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
