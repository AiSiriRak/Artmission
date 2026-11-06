import { paymentDisplayDeadline, remainingSeconds } from "./payment-countdown";

const now = Date.parse("2026-11-06T12:00:00Z");

it("caps a long-lived QR at five minutes", () => {
  const expiry = new Date(now + 24 * 60 * 60 * 1000).toISOString();
  const deadline = paymentDisplayDeadline(expiry, now);
  expect(remainingSeconds(deadline, now)).toBe(300);
  expect(remainingSeconds(deadline, now + 60_000)).toBe(240);
  expect(remainingSeconds(deadline, now + 300_000)).toBe(0);
});

it("never extends a QR with less than five minutes left", () => {
  const expiry = new Date(now + 90_000).toISOString();
  expect(remainingSeconds(paymentDisplayDeadline(expiry, now), now)).toBe(90);
  expect(remainingSeconds(paymentDisplayDeadline(expiry, now + 60_000), now + 60_000)).toBe(30);
});

it("starts a fresh display window when reusing a still-valid QR", () => {
  const expiry = new Date(now + 600_000).toISOString();
  const later = now + 300_000;
  expect(remainingSeconds(paymentDisplayDeadline(expiry, later), later)).toBe(300);
});

it("keeps a backend-expired QR at zero", () => {
  const expiry = new Date(now - 1000).toISOString();
  expect(remainingSeconds(paymentDisplayDeadline(expiry, now), now)).toBe(0);
});
