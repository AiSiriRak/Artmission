export type ArtistData = {
  name: string;
  profileImage: string;
  description: string;
  style: string; // เก็บเป็น String คั่นด้วยลูกน้ำตอน Edit
  category: string;
  priceRange: string;
};

export type ArtworkData = {
  id: number;
  name: string;        
  coverImage: string;
  style: string;
  category: string;
  price: number;
  description: string;
  deadline: number;
  images?: string[];
};