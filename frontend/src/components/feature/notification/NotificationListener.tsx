"use client";

import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { routes } from "@/lib/routes";
import { getNotifications } from "@/lib/api/notification";

const POLL_INTERVAL_MS = 10000;

export function NotificationListener() {
  const knownIds = useRef<Set<string> | null>(null);
  const router = useRouter();

  useEffect(() => {
    const checkNewNotifications = async () => {
      try {
        const notifications = await getNotifications();
        const known = knownIds.current;

        if (!known) {
          knownIds.current = new Set(notifications?.items?.map((n) => n.id));
          return;
        }

        notifications?.items?.forEach((n) => {
          if (known.has(n.id)) return;

          known.add(n.id);

          // Send custom event to trigger app
          window.dispatchEvent(new Event("notification-updated"));

          toast.custom((t) => (
            <div
              onClick={() => {
                toast.dismiss(t);
                router.replace(routes.notification);
              }}
              className="relative flex w-full cursor-pointer flex-col rounded-xl bg-secondary-200 p-3 font-sans shadow-card transition-all hover:brightness-95"
            >
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  toast.dismiss(t);
                }}
                className="absolute right-3 top-3 flex aspect-square w-5 items-center justify-center rounded-full bg-secondary-200 hover:brightness-90"
              >
                ✕
              </button>

              <p className="line-clamp-1 text-body font-semibold">{n.title}</p>
              <p className="line-clamp-3 text-small opacity-80">{n.message}</p>
            </div>
          ));
        });
      } catch (error) {
        console.error("Failed to fetch notifications:", error);
      }
    };

    checkNewNotifications();
    const interval = setInterval(checkNewNotifications, POLL_INTERVAL_MS);

    return () => clearInterval(interval);
  }, [router]);

  return null;
}
