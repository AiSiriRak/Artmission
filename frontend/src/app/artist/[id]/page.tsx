import ProfileManager from "./components/ProfileManager";
import { ArtistData, ArtworkData } from "./types";
import Header from "./components/Header";

export default function ArtistProfilePage() {
  const artistInitialData: ArtistData = {
    name: "Artist's Name",
    profileImage: "",
    description: "Digital illustrator and visual storyteller. Bringing characters to life through rich color palettes, expressive lighting, and playful moods.\n🎨 Original prints & adoption available \n📩 Open for commissions & freelance work \n🔗 Explore the collection below: [Artmission Link]",
    style: "Pixel Art, Cartoon, Graphic",
    category: "Digital Art, Poster",
    priceRange: "500 - 2,000 THB",
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
    { 
      id: 2, 
      name: "Cute Cartoon", 
      coverImage: "https://placehold.co/300x200", 
      style: "Cartoon", 
      category: "categories", 
      price: 300,
      description: "Adorable digital cartoon illustrations of your pets or personal avatars with vibrant colors.",
      deadline: 7,
      images: ["https://placehold.co/300x200"]
    },
    { 
      id: 3, 
      name: "Vector Graphic", 
      coverImage: "https://placehold.co/300x200", 
      style: "Graphic", 
      category: "categories", 
      price: 300,
      description: "Clean, crisp, and scalable vector art for logos, stickers, or high-resolution printing.",
      deadline: 3,
      images: ["https://placehold.co/300x200"]
    },
    { 
      id: 4, 
      name: "3D Character", 
      coverImage: "https://placehold.co/300x200", 
      style: "3D", 
      category: "categories", 
      price: 300,
      description: "High-detail 3D character design ready for rendering, animation, or 3D printing.",
      deadline: 14,
      images: ["https://placehold.co/300x200"]
    },
    { 
      id: 5, 
      name: "Realistic Sketch", 
      coverImage: "https://placehold.co/300x200", 
      style: "Realism", 
      category: "categories", 
      price: 300,
      description: "Detailed graphite and digital charcoal sketch portrait with realistic shading.",
      deadline: 7,
      images: ["https://placehold.co/300x200"]
    },
  ];

  return (
    <div className="w-full bg-white text-black min-h-screen">

      <Header />

      {/* โยนทุกอย่างให้ Client Component เป็นคนตัดสินใจว่าจะโชว์ Layout ไหน */}
      <ProfileManager initialProfile={artistInitialData} initialArtworks={initialArtworks} />

      {/* ส่วน Review (Server Component) ห้ามแก้ */}
      <div className="max-w-5xl mx-auto px-8 pb-16">
        <div className="bg-white p-6 shadow-sm rounded-xl border border-gray-100 text-black">
          <h2 className="text-xl font-bold mb-4">Reviews</h2>
          {/* ใส่โค้ด Review ตรงนี้ได้เลยครับ */}
          <p className="text-gray-500 text-sm">ยังไม่มีรีวิว</p>
        </div>
      </div>
    </div>
  );
}