import type { Artwork } from "@/lib/api/types";
import ArtworkCard from "./ArtworkCard";

interface ArtworkCollectionProps {
  artworks: Artwork[];
  isCustomerMode: boolean;
  showEditControls: boolean;
  isEditing: boolean;
  onAdd: () => void;
  onOpen: (artwork: Artwork) => void;
}

export default function ArtworkCollection({
  artworks,
  isCustomerMode,
  showEditControls,
  isEditing,
  onAdd,
  onOpen,
}: ArtworkCollectionProps) {
  return (
    <div className="max-w-5xl mx-auto px-8 pb-10">
      {isCustomerMode ? (
        <div className="flex items-center gap-2 bg-secondary-300 border-l-4 border-accent-500 px-4 py-2.5 mb-6 rounded-r-md text-h3 font-bold text-gray-900">
          <span>Artworks</span>
        </div>
      ) : (
        <>
          <hr className="border-gray-200 mb-10" />
          <div className="flex justify-between items-center mb-6">
            <div className="flex items-center gap-3">
              <h2 className="text-h3 font-bold text-gray-900">Your Artwork</h2>
              <span className="text-sm font-normal text-gray-400">
                {artworks.length} samples
              </span>
            </div>
          </div>
        </>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {showEditControls && (
          <div
            onClick={onAdd}
            className="border-2 border-dashed border-accent-300 rounded-2xl flex flex-col items-center justify-center text-accent-500 cursor-pointer hover:bg-accent-50 transition-colors h-full min-h-[250px]"
          >
            <span className="font-bold mb-3">Add your artwork</span>
            <div className="w-12 h-12 border border-accent-300 rounded-md flex items-center justify-center text-2xl">
              +
            </div>
          </div>
        )}

        {artworks.map((art, idx) => (
          <ArtworkCard
            key={art.id || art.name || idx}
            artwork={art}
            showEditControls={showEditControls}
            onClick={() => {
              if (!isEditing) onOpen(art);
            }}
          />
        ))}
      </div>
    </div>
  );
}
