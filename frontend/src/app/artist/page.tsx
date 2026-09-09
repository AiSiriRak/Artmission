import ProfileManager from "@/components/feature/artistprofile/ProfileManager";
import { ArtistData, ArtworkData } from "./types";
import MainLayout from "@/components/feature/main/MainLayout";
import { getArtistProfile } from "@/lib/api/artist";
import { ArtistProfile } from "@/lib/api/types"; 

export default async function ArtistProfilePage() {
  let username = "Artist's Name";
  let artistProfileData: Partial<ArtistProfile> = {};

  try {
    const artist = await getArtistProfile();
    if (artist) {
      artistProfileData = artist;
      username = artist.artist_name || username; 
    }
  } catch (error) {
    console.warn("Cannot fetch artist profile", error);
  }

  const categoryString = artistProfileData.categories?.map(c => c.label).join(", ");
  const styleString = artistProfileData.styles?.map(s => s.label).join(", ");

  // แปลงราคาจากหน่วย สตางค์ เป็น บาท (หาร 100)
  const minTHB = artistProfileData.min_price_satang ? artistProfileData.min_price_satang / 100 : null;
  const maxTHB = artistProfileData.max_price_satang ? artistProfileData.max_price_satang / 100 : null;
  const priceRangeString = minTHB && maxTHB ? `${minTHB} - ${maxTHB} THB` : undefined;

  const artistInitialData: ArtistData = {
    name: username,
    // Backend ยังไม่มีฟิลด์ profileImage ให้ส่งค่าว่างไปก่อน หรือใส่ Placeholder
    profileImage: "", 
    description:
      artistProfileData.description ||
      "Digital illustrator and visual storyteller. Bringing characters to life through rich color palettes, expressive lighting, and playful moods.\n🎨 Original prints & adoption available \n📩 Open for commissions & freelance work \n🔗 Explore the collection below: [Artmission Link]",
    style: styleString || "Pixel Art, Cartoon, Graphic",
    category: categoryString || "Digital Art, Poster",
    priceRange: priceRangeString || "500 - 2,000 THB",
  };

  const initialArtworks: ArtworkData[] = [
    { 
      id: 1, 
      name: "Pixel Art Portrait",
      coverImage: "https://placehold.co/300x200", 
      style: "Pixel Art", 
      category: "categories", 
      price: 300,
      description: "Turn your favorite memories into a nostalgic pixel art style. Perfect for game lovers and retro aesthetics.",
      deadline: 5,
      images: ["https://placehold.co/300x200"]
    },
  ];

  return (
    <MainLayout page="Artist Profile" usertype="artist">
      <div className="w-full bg-white text-black">
        <ProfileManager 
          initialProfile={artistInitialData} 
          initialArtworks={initialArtworks} 
        />
      </div>
    </MainLayout>
  );
}