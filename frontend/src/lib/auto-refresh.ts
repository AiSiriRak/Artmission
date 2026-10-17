// Check for changes while the page is visible, and when the user returns.
export function startAutoRefresh(refresh: () => void, intervalMs = 3000) {
  function refreshIfVisible() {
    if (document.visibilityState === "visible") refresh();
  }

  const interval = window.setInterval(refreshIfVisible, intervalMs);
  window.addEventListener("focus", refreshIfVisible);
  window.addEventListener("online", refreshIfVisible);
  document.addEventListener("visibilitychange", refreshIfVisible);

  return () => {
    window.clearInterval(interval);
    window.removeEventListener("focus", refreshIfVisible);
    window.removeEventListener("online", refreshIfVisible);
    document.removeEventListener("visibilitychange", refreshIfVisible);
  };
}
