const mockArtistData = {
  name: "Artist Name",
  bio: "Artist Bio or Description",
  avatarUrl: "/default-avatar.png",
  review_score: 4.5,
};

export function ArtistMiniProfile() {
  return (
    <div className="flex items-center gap-4">
      <div className="w-40 h-40 rounded-full overflow-hidden">
        <img
          src={mockArtistData.avatarUrl}
          alt="Artist Avatar"
          className="w-full h-full object-cover"
        />
      </div>
      <div className="flex flex-col gap-2">
        <h3 className="text-h3">Artist Profile</h3>
        <span className="text-body">{mockArtistData.name}</span>
        <div className="flex items-center gap-1">
          <span className="text-h3">
            {mockArtistData.review_score}/5
          </span>
        </div>
      </div>
    </div>
  );
}
