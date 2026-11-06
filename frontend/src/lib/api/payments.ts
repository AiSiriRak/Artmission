import { apiFetch, ApiError } from "./client";

export interface PaymentQR {
  success: boolean;
  message: string;
  payment_id: string;
  order_id: string;
  qr_payload: string;
  qr_expire: string;
}

export interface SlipVerification {
  success: boolean;
  message: string;
  order_id: string;
  reason?: string;
  payment_id?: string;
  status?: string;
}

export async function loadPaymentQR(orderId: string): Promise<PaymentQR> {
  const path = `/orders/${encodeURIComponent(orderId)}/payment/qr`;
  let result: PaymentQR;
  try {
    result = await apiFetch<PaymentQR>(path, { cache: "no-store" });
  } catch (error) {
    if (!(error instanceof ApiError) || ![404, 410].includes(error.status)) throw error;
    result = await apiFetch<PaymentQR>(path, { method: "POST" });
  }
  if (result?.success !== true || !result.qr_payload?.trim() || !Number.isFinite(Date.parse(result.qr_expire)) || result.order_id !== orderId) {
    throw new Error("The payment QR response was invalid. Please try again.");
  }
  return result;
}

export async function verifyPaymentSlip(orderId: string, file: File): Promise<SlipVerification> {
  const form = new FormData();
  form.append("file", file);
  const result = await apiFetch<SlipVerification>(`/orders/${encodeURIComponent(orderId)}/payment/verify`, {
    method: "POST",
    body: form,
  });
  if (typeof result?.success !== "boolean" || result.order_id !== orderId ||
      (result.success ? result.status !== "HELD" || !result.payment_id : !result.message?.trim())) {
    throw new Error("The verification response was invalid. Please check your order status before resubmitting.");
  }
  return result;
}
