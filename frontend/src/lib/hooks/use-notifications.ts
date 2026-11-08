"use client";

import { useCallback, useEffect, useState } from "react";
import { getUnreadCount } from "@/lib/api/notification";

const POLL_INTERVAL_MS = 10000;

export function useUnreadNotifications(enabled = true) {
  const [unreadCount, setUnreadCount] = useState(0);

  const fetchUnreadCount = useCallback(async () => {
    try {
      const response = await getUnreadCount();
      setUnreadCount(response.count);
    } catch (error) {
      console.error("Failed to fetch unread notifications:", error);
    }
  }, []);

  useEffect(() => {
    if (!enabled) return;

    let isMounted = true;

    getUnreadCount()
      .then((response) => {
        if (isMounted) {
          setUnreadCount(response.count);
        }
      })
      .catch((error) => {
        console.error("Failed to fetch unread notifications:", error);
      });

    const refreshIfVisible = () => {
      if (document.visibilityState === "visible") {
        fetchUnreadCount();
      }
    };

    const interval = setInterval(refreshIfVisible, POLL_INTERVAL_MS);
    document.addEventListener("visibilitychange", refreshIfVisible);

    return () => {
      isMounted = false;
      clearInterval(interval);
      document.removeEventListener("visibilitychange", refreshIfVisible);
    };
  }, [enabled, fetchUnreadCount]);

  return { unreadCount, refreshUnreadCount: fetchUnreadCount };
}
