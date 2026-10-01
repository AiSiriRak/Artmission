import CustomerArtistProfile from "@/components/feature/artistprofile/CustomerArtistProfile";

interface ArtistPageProps {
  params: Promise<{ artistId: string }>;
}

export default async function ArtistPage({ params }: ArtistPageProps) {
  const { artistId } = await params;

  return (
    <CustomerArtistProfile
      key={artistId}
      artistId={artistId}
    />
  );
}