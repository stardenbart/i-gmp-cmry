/// <reference lib="esnext" />
/// <reference lib="webworker" />
import { defaultCache } from "@serwist/turbopack/worker";
import type { PrecacheEntry, SerwistGlobalConfig } from "serwist";
import { Serwist } from "serwist";

declare global {
  interface WorkerGlobalScope extends SerwistGlobalConfig {
    __SW_MANIFEST: (PrecacheEntry | string)[] | undefined;
  }
}

declare const self: ServiceWorkerGlobalScope;

const serwist = new Serwist({
  precacheEntries: self.__SW_MANIFEST,
  skipWaiting: true,
  clientsClaim: true,
  navigationPreload: true,
  runtimeCaching: defaultCache,
});

serwist.addEventListeners();

// ── Web Push ─────────────────────────────────────────────────────────────
// Handles a push message sent by the backend (see pkg/webpush on the Go
// side) and shows it as a native OS/lock-screen notification — this is what
// makes a notification appear even when the app/tab is closed.
interface PushPayload {
  title: string;
  body: string;
  link?: string;
  tag?: string;
}

self.addEventListener("push", (event: PushEvent) => {
  let payload: PushPayload = { title: "Monitoring Audit", body: "Anda memiliki notifikasi baru." };
  try {
    if (event.data) payload = { ...payload, ...event.data.json() };
  } catch {
    // Fall back to the default payload above if the push body isn't valid JSON.
  }

  event.waitUntil(
    self.registration.showNotification(payload.title, {
      body: payload.body,
      icon: "/Logo_plant_New.png",
      badge: "/Logo_plant_New.png",
      tag: payload.tag,
      data: { link: payload.link || "/" },
    })
  );
});

// Focus an already-open tab on the target page if one exists, otherwise open
// a new one — standard "click a notification" behavior.
self.addEventListener("notificationclick", (event: NotificationEvent) => {
  event.notification.close();
  const link = (event.notification.data as { link?: string } | undefined)?.link || "/";

  event.waitUntil(
    (async () => {
      const allClients = await self.clients.matchAll({ type: "window", includeUncontrolled: true });
      for (const client of allClients) {
        if (client.url.includes(link) && "focus" in client) {
          return client.focus();
        }
      }
      return self.clients.openWindow(link);
    })()
  );
});
