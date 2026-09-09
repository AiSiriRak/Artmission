import type { components } from "./v1";

// User
export type UserAccount = components["schemas"]["AccountView"];
export type UpdateAccountInput =
  components["schemas"]["UpdateAccountInputBody"];

// Bank Account
export type BankAccount = components["schemas"]["BankAccountView"];
export type UpdateBankAccountInput =
  components["schemas"]["UpdateBankAccountInputBody"];

// Order
export type Order = components["schemas"]["OrderView"];

// Artist
export type ArtistProfile = components["schemas"]["ArtistProfileView"];
export interface UpdateArtistInput {
  description?: string;
  // profileImage?: string; 
}

// ==========================================
// Artwork (หาก Backend มีเพิ่ม Schema ในอนาคต ค่อยมาเปลี่ยนเป็นแบบด้านบน)
// ==========================================

export interface Artwork {
  id: string | number;
  title?: string;
  imageUrl: string;
  price?: number;
  category?: string;
  style?: string;
}

export interface CreateArtworkInput {
  title?: string;
  imageUrl: string;
  price?: number;
  category?: string;
  style?: string;
}