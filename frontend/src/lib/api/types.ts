import type { components, paths } from "./v1";

// Utility type for Request Body from paths
type RequestBody<
  P extends keyof paths,
  M extends keyof paths[P]
> = paths[P][M] extends { requestBody?: { content: infer C } }
  ? C[keyof C]
  : never;

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
export type ViewOrdersOutput = components["schemas"]["ViewOrdersOutputBody"];
export type OrderQuery = NonNullable<
  paths["/orders"]["get"]["parameters"]["query"]
>;
export type OrderStatus = NonNullable<OrderQuery["status"]>[number];
export type OrderSortField = OrderQuery["sort"];
export type OrderSortOrder = OrderQuery["order"];

// Artist
export type ArtistProfile = components["schemas"]["ArtistProfileView"];
export type ArtistReference = components["schemas"]["ArtistReferenceView"];
export type ArtistReview = components["schemas"]["ArtistReviewView"];
export type ArtistArtwork = components["schemas"]["ArtistArtworkView"];
export type GetArtistArtworksOutput = components["schemas"]["GetArtistArtworksOutputBody"];

// Type from multipart/form-data from path
export type UpdateArtistInput = RequestBody<"/artists/me", "put">;

// Artwork
export type Artwork = components["schemas"]["ArtworkView"];
export type ArtworkSample = components["schemas"]["ArtworkSampleView"];
export type CreateArtworkInput = components["schemas"]["CreateArtworkInputBody"];