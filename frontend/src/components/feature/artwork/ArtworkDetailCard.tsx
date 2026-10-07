const mockArtwork = {
  name: "The Starry Night",
  description:
    "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
  category: "Painting",
  styles: ["Abstract", "Modern"],
  images: ["/default-avatar.png", "/default-avatar.png", "/default-avatar.png"],
};

export default function ArtworkDetailCard() {
  return (
    
    <div className="border border-primary-500 rounded-3xl p-10 bg-secondary-200 shadow-sm">
      <div className="relative">
          <div className="p-4 flex flex-col flex-1 gap-6">
            {/* Todo: replace with actual artwork name load from backend */}

            {/* Artwork Info section*/}
            <div className="flex flex-col gap-2">
              <h1 className="text-h1">{mockArtwork.name}</h1>
              <div className="flex flex-wrap gap-2">
                <p className="px-4 py-1.5 bg-accent-200 text-primary-400 rounded-full text-sm font-semibold shadow-sm">
                  {mockArtwork.category}
                </p>
                {mockArtwork.styles.map((style, index) => (
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
              <h3 className="text-h3">Description</h3>
              <p className="text-body">{mockArtwork.description}</p>
            </div>

            {/* Artwork Sample Gallery section */}
            <h3 className="text-h3">Artwork Samples</h3>
            <div className="flex justify-center items-center gap-2">
              <div className="flex flex-wrap justify-center gap-4">
                {mockArtwork.images.map((imgUrl, index) => (
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
