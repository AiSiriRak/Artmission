import CustomerArtistProfile from "@/components/feature/artistprofile/CustomerArtistProfile"; //import Component

interface ArtistPageProps { //for dynamic route parameter
  params: Promise<{ artistId: string }>; //passed URL into params
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