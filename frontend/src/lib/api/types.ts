import type { components, paths } from "./v1";

// Auth
export type LoginInput = components["schemas"]["LoginInputBody"];
export type RegisterInput = components["schemas"]["RegisterInputBody"];
export type AuthResult = components["schemas"]["AuthResultBody"];

// User / Account
export type UserAccount = components["schemas"]["AccountView"];
export type UpdateAccountInput = components["schemas"]["UpdateAccountInputBody"];

// Bank Account
export type BankAccount = components["schemas"]["BankAccountView"];
export type UpdateBankAccountInput = components["schemas"]["UpdateBankAccountInputBody"];

// Order
export type Order = components["schemas"]["OrderSummaryView"];

// Artist
export type ArtistProfile = components["schemas"]["ArtistProfileView"];
export type UpdateArtistInput = RequestBody<"/artists/me", "put">;
type RequestBody<
  P extends keyof paths,
  M extends keyof paths[P]
> = paths[P][M] extends { requestBody?: { content: infer C } }
  ? C[keyof C]
  : never;

// Artwork
export type Artwork = components["schemas"]["ArtworkView"];
export type CreateArtworkInput = components["schemas"]["CreateArtworkInputBody"];
export type UpdateArtworkInput = Partial<CreateArtworkInput>;
