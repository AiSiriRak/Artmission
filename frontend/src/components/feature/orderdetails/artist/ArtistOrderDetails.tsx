"use client";

import { useState } from "react";
import type { OrderStatus } from "@/lib/api/types";
import OrderDetails from "../shared/OrderDetails";
import OrderDescription from "../shared/OrderDescription";
import OrderInfo from "../shared/OrderInfo";
import UserInfo from "../shared/UserInfo";
import ArtistOrderActions from "./ArtistOrderActions";
import SubmitDraft from "./SubmitDraft";
import PostponeNotice from "./PostponeNotice";
import ConfirmOrderPopup from "./popups/ConfirmOrderPopup";
import CancelOrderPopup from "./popups/CancelOrderPopup";
import PostponeOrderPopup, { type PostponeRequest } from "./popups/PostponeOrderPopup";

const mockStatus: OrderStatus = "IN_PROCESS";
const mockRequestDetail =
  "Turn your favorite memories of your companion into a timeless piece of fine digital art. This custom pet portrait combines the precise detail of realism with the raw, evocative charm of traditional pencil sketching. Every stroke is tailored to highlight what makes your pet one-of-a-kind—from the subtle glint of life in their eyes and the exact contour of their snout to the natural flow, depth, and softness of their fur. " +
  "The composition bridges two distinct worlds: an intricately rendered focal point across the face that captures lifelike texture, which then fades organically into loose, expressive graphite-style line work toward the chest and shoulders. Set against a warm, neutral textured canvas, the final result feels less like a sterile photo copy and more like a handcrafted, intimate study found in an artist’s sketchbook. Delivered as a 300 DPI, print-ready digital file, this piece is fully prepared for high-end giclée printing, canvas mounting, or framing as an enduring tribute to your pet's place in your life.";

export default function ArtistOrderDetails() {
  // UI preview state only. TODO: Replace with API state after integration.
  const [status, setStatus] = useState<OrderStatus>(mockStatus);
  const [popup, setPopup] = useState<"reject" | "accept" | "cancel" | "postpone" | null>(null);
  const [postponeRequest, setPostponeRequest] = useState<PostponeRequest | null>(null);
  const closePopup = () => setPopup(null);

  return (
    <>
      <OrderDetails
        usertype="artist"
        description={
          <>
            {postponeRequest && <PostponeNotice deadline={postponeRequest.deadline} />}
            <OrderDescription
              orderNumber="1024"
              artworkName="Cutie Catto"
              categoryName="Pet Portrait"
              categories={["Digital Art"]}
              styles={["Realism", "Sketch"]}
              requestDetail={mockRequestDetail}
            />
          </>
        }
        info={
          <OrderInfo
            status={status}
            priceSatang={300000}
            orderDate="02 Sep 2026"
            deadlineDate="10 Sep 2026"
            deadlineNote="Due in 7 days"
            revisions={{ used: 0, limit: 2 }}
            actions={status !== "SUCCESS" && status !== "CANCEL" ? (
              <ArtistOrderActions
                status={status}
                onReject={() => setPopup("reject")}
                onAccept={() => setPopup("accept")}
                onCancel={() => setPopup("cancel")}
                onPostpone={() => setPopup("postpone")}
              />
            ) : undefined}
          />
        }
        userInfo={
          <UserInfo title="Customer Info" name="Username" email="useremail@gmail.com" />
        }
      >
        {status === "IN_PROCESS" && <SubmitDraft />}
      </OrderDetails>

      {(popup === "reject" || popup === "accept") && (
        <ConfirmOrderPopup
          action={popup}
          onClose={closePopup}
          onConfirm={() => {
            // TODO: Call confirm API, then use the status returned by Backend.
            setStatus(popup === "accept" ? "NOT_PAID" : "CANCEL");
            closePopup();
          }}
        />
      )}
      {popup === "cancel" && (
        <CancelOrderPopup
          orderSummary="Order #1024 · Cutie Catto · ฿3,000"
          onClose={closePopup}
          onConfirm={() => {
            // TODO: Send cancellation reason/description to the API.
            setStatus("CANCEL");
            setPostponeRequest(null);
            closePopup();
          }}
        />
      )}
      {popup === "postpone" && (
        <PostponeOrderPopup
          orderSummary="Order #1024 · Cutie Catto · ฿3,000"
          currentDeadline="2026-09-10"
          currentDeadlineLabel="10 Sep 2026"
          onClose={closePopup}
          onConfirm={(request) => {
            // TODO: Send request to the API; show notice only after it succeeds.
            setPostponeRequest(request);
            closePopup();
          }}
        />
      )}
    </>
  );
}
