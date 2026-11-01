"use client";

import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { getNotifications } from "@/lib/api/notification";

const POLL_INTERVAL_MS = 10000;

export function NotificationListener() {
  const knownIds = useRef<Set<string> | null>(null);

  useEffect(() => {
    const checkNewNotifications = async () => {
      try {
        const notifications = await getNotifications();
        const known = knownIds.current;

        if (!known) {
          knownIds.current = new Set(notifications.map((n) => n.id));
          return;
        }

        for (const n of notifications) {
          if (known.has(n.id)) continue;

          known.add(n.id);
          toast.info(n.title, {
            description: n.message,
            duration: 5000,
          });
        }
      } catch (error) {
        console.error("Failed to fetch notifications:", error);
      }
    };

    checkNewNotifications();
    const interval = setInterval(checkNewNotifications, POLL_INTERVAL_MS);

    return () => clearInterval(interval);
  }, []);

  return null;
}
