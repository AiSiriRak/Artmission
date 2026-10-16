import type { components, paths } from "./v1";

// Utility type for Request Body from paths
type RequestBody<
  P extends keyof paths,
  M extends keyof paths[P],
> = paths[P][M] extends { requestBody?: { content: infer C } }
  ? C[keyof C]
  : never;

// Auth
export type LoginInput = components["schemas"]["LoginInputBody"];
export type RegisterInput = components["schemas"]["RegisterInputBody"];
export type AuthResult = components["schemas"]["AuthResultBody"];

// User / Account
export type UserAccount = components["schemas"]["AccountView"];
export type UpdateAccountInput =
  components["schemas"]["UpdateAccountInputBody"];

// Bank Account
export type BankAccount = components["schemas"]["BankAccountView"];
export type UpdateBankAccountInput =
  components["schemas"]["UpdateBankAccountInputBody"];

// Order
export type OrderHistory = components["schemas"]["ViewOrdersOutputBody"];
export type Order = components["schemas"]["OrderSummaryView"];
export type OrderQuery = NonNullable<
  paths["/orders"]["get"]["parameters"]["query"]
>;
export type OrderStatus = NonNullable<OrderQuery["status"]>[number];
export type OrderSortField = OrderQuery["sort"];
export type OrderSortOrder = OrderQuery["order"];
export type CreateOrderInput = components["schemas"]["CreateOrderInputBody"];
export type CancelOrderInput = components["schemas"]["CancelOrderInputBody"];
export type CancelOrderResult = components["schemas"]["CancelOrderOutputBody"];
export type ConfirmDeliverableInput =
  components["schemas"]["ConfirmDeliverableInputBody"];
export type ConfirmDeliverableResult =
  components["schemas"]["ConfirmDeliverableOutputBody"];
export type DeliverableDecision = "WAIT" | "APPROVED" | "REJECTED";
// The generated schema types these enums as plain strings, so narrow them here.
export type OrderDeliverable = Omit<
  components["schemas"]["DeliverableView"],
  "decision"
> & { decision: DeliverableDecision };
export type DelayRequestDecision = "WAIT" | "APPROVED" | "REJECTED";
export type DelayRequest = Omit<
  components["schemas"]["DelayRequestView"],
  "customer_decision"
> & { customer_decision: DelayRequestDecision };
export type ConfirmDelayRequestInput =
  components["schemas"]["ConfirmDelayRequestInputBody"];
export type ConfirmDelayRequestResult =
  components["schemas"]["ConfirmDelayRequestOutputBody"];
// The backend returns null when there is no active delay request.
export type OrderDetail = Omit<
  components["schemas"]["OrderDetailView"],
  "status" | "deliverables" | "active_delay_request"
> & {
  status: OrderStatus;
  deliverables: OrderDeliverable[] | null;
  active_delay_request: DelayRequest | null;
};

// Artist
export type ArtistProfile = components["schemas"]["ArtistProfileView"];
export type ArtistReference = components["schemas"]["ArtistReferenceView"];
export type ArtistReview = components["schemas"]["ArtistReviewView"];
export type ArtistArtwork = components["schemas"]["ArtistArtworkView"];
export type GetArtistArtworksOutput =
  components["schemas"]["GetArtistArtworksOutputBody"];

// Type from multipart/form-data from path
export type UpdateArtistInput = RequestBody<"/artists/me", "put">;

// Artwork
export type Artwork = components["schemas"]["ArtworkView"];
export type ArtworkSample = components["schemas"]["ArtworkSampleView"];
export type SearchArtwork = components["schemas"]["SearchArtworkView"];
export type SearchArtworksOutput =
  components["schemas"]["SearchArtworksOutputBody"];
export type SearchArtworksQuery = NonNullable<
  paths["/artworks"]["get"]["parameters"]["query"]
>;
export type SearchArtworksSort = NonNullable<SearchArtworksQuery["sort"]>;

// Type from multipart/form-data from path
export type CreateArtworkInput = RequestBody<"/artworks", "post">;
export type UpdateArtworkInput = RequestBody<"/artworks/{artwork_id}", "put">;
