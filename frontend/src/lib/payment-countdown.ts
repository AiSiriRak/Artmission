const DISPLAY_WINDOW_MS = 5 * 60 * 1000;

export function paymentDisplayDeadline(qrExpire: string, now: number): number {
  return Math.min(now + DISPLAY_WINDOW_MS, Date.parse(qrExpire));
}

export function remainingSeconds(deadline: number, now: number): number {
  return Math.max(0, Math.ceil((deadline - now) / 1000));
}
