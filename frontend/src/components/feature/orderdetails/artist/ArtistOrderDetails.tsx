import type { OrderStatus } from "@/lib/api/types";
import OrderDetails from "../shared/OrderDetails";
import OrderDescription from "../shared/OrderDescription";
import OrderInfo from "../shared/OrderInfo";
import UserInfo from "../shared/UserInfo";
import ArtistOrderActions from "./ArtistOrderActions";
import SubmitDraft from "./SubmitDraft";

const mockStatus: OrderStatus = "IN_PROCESS";
const mockRequestDetail =
  "Turn your favorite memories of your companion into a timeless piece of fine digital art. This custom pet portrait combines the precise detail of realism with the raw, evocative charm of traditional pencil sketching. Every stroke is tailored to highlight what makes your pet one-of-a-kind—from the subtle glint of life in their eyes and the exact contour of their snout to the natural flow, depth, and softness of their fur. " +
  "The composition bridges two distinct worlds: an intricately rendered focal point across the face that captures lifelike texture, which then fades organically into loose, expressive graphite-style line work toward the chest and shoulders. Set against a warm, neutral textured canvas, the final result feels less like a sterile photo copy and more like a handcrafted, intimate study found in an artist’s sketchbook. Delivered as a 300 DPI, print-ready digital file, this piece is fully prepared for high-end giclée printing, canvas mounting, or framing as an enduring tribute to your pet's place in your life.";

function renderArtistActions(status: OrderStatus) {
  if (status === "SUCCESS" || status === "CANCEL") return null;

  return <ArtistOrderActions status={status} />;
}

function renderArtistDraft(status: OrderStatus) {
  if (status !== "IN_PROCESS") return null;

  return <SubmitDraft />;
}

export default function ArtistOrderDetails() {
  return (
    <OrderDetails
      usertype="artist"
      description={
        <OrderDescription
          orderNumber="1024"
          artworkName="Cutie Catto"
          categoryName="Pet Portrait"
          categories={["Digital Art"]}
          styles={["Realism", "Sketch"]}
          requestDetail={mockRequestDetail}
        />
      }
      info={
        <OrderInfo
          status={mockStatus}
          priceSatang={300000}
          orderDate="02 Sep 2026"
          deadlineDate="10 Sep 2026"
          deadlineNote="Due in 7 days"
          revisions={{ used: 0, limit: 2 }}
          actions={renderArtistActions(mockStatus)}
        />
      }
      userInfo={
        <UserInfo
          title="Customer Info"
          name="Username"
          email="useremail@gmail.com"
        />
      }
    >
      {renderArtistDraft(mockStatus)}
    </OrderDetails>
  );
}
