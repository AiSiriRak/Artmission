import { ApiError, apiFetch } from "./client";
import { loadPaymentQR, verifyPaymentSlip } from "./payments";

jest.mock("./client", () => {
  const original = jest.requireActual<typeof import("./client")>("./client");
  return { ...original, apiFetch: jest.fn() };
});

const fetchMock = jest.mocked(apiFetch);
const qr = {
  success: true,
  message: "Payment QR retrieved.",
  payment_id: "payment-1",
  order_id: "order-1",
  qr_payload: "000201-test",
  qr_expire: "2026-11-02T10:00:00Z",
};

beforeEach(() => fetchMock.mockReset());

it("retrieves an existing QR without generating another", async () => {
  fetchMock.mockResolvedValueOnce(qr);
  await expect(loadPaymentQR("order-1")).resolves.toEqual(qr);
  expect(fetchMock).toHaveBeenCalledTimes(1);
  expect(fetchMock).toHaveBeenCalledWith("/orders/order-1/payment/qr", { cache: "no-store" });
});

it("generates a QR only when the existing QR is not found", async () => {
  fetchMock.mockRejectedValueOnce(new ApiError("not found", 404, null)).mockResolvedValueOnce(qr);
  await expect(loadPaymentQR("order-1")).resolves.toEqual(qr);
  expect(fetchMock).toHaveBeenLastCalledWith("/orders/order-1/payment/qr", { method: "POST" });
});

it("requests generation for an expired QR and preserves a backend expiry rejection", async () => {
  fetchMock.mockRejectedValueOnce(new ApiError("expired", 410, null))
    .mockRejectedValueOnce(new ApiError("expired", 410, null));
  await expect(loadPaymentQR("order-1")).rejects.toMatchObject({ status: 410 });
  expect(fetchMock).toHaveBeenCalledTimes(2);
  expect(fetchMock).toHaveBeenLastCalledWith("/orders/order-1/payment/qr", { method: "POST" });
});

it("uses the new QR if generation after expiry succeeds", async () => {
  fetchMock.mockRejectedValueOnce(new ApiError("expired", 410, null)).mockResolvedValueOnce(qr);
  await expect(loadPaymentQR("order-1")).resolves.toEqual(qr);
});

it("does not attempt generation when the customer cannot access the order", async () => {
  fetchMock.mockRejectedValueOnce(new ApiError("forbidden", 403, null));
  await expect(loadPaymentQR("order-1")).rejects.toMatchObject({ status: 403 });
  expect(fetchMock).toHaveBeenCalledTimes(1);
});

it("rejects a QR for a different order or with an invalid expiry", async () => {
  fetchMock.mockResolvedValueOnce({ ...qr, order_id: "other" });
  await expect(loadPaymentQR("order-1")).rejects.toThrow("response was invalid");
  fetchMock.mockResolvedValueOnce({ ...qr, qr_expire: "invalid" });
  await expect(loadPaymentQR("order-1")).rejects.toThrow("response was invalid");
});

it("submits the slip in the backend's multipart file field", async () => {
  const file = new File(["test"], "slip.png", { type: "image/png" });
  const result = { success: false, message: "Wrong receiver", order_id: "order-1" };
  fetchMock.mockResolvedValueOnce(result);
  await expect(verifyPaymentSlip("order-1", file)).resolves.toEqual(result);
  const [path, options] = fetchMock.mock.calls[0];
  expect(path).toBe("/orders/order-1/payment/verify");
  expect(options?.method).toBe("POST");
  expect((options?.body as FormData).get("file")).toBe(file);
});

it("accepts the confirmed HELD payment and rejects mismatched success responses", async () => {
  const file = new File(["test"], "slip.png", { type: "image/png" });
  const result = { success: true, message: "Verified", order_id: "order-1", payment_id: "payment-1", status: "HELD" };
  fetchMock.mockResolvedValueOnce(result);
  await expect(verifyPaymentSlip("order-1", file)).resolves.toEqual(result);
  fetchMock.mockResolvedValueOnce({ ...result, order_id: "other" });
  await expect(verifyPaymentSlip("order-1", file)).rejects.toThrow("verification response was invalid");
  fetchMock.mockResolvedValueOnce({ ...result, status: "PENDING" });
  await expect(verifyPaymentSlip("order-1", file)).rejects.toThrow("verification response was invalid");
});
