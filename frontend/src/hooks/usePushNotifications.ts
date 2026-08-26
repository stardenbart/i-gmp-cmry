"use client";

import { useCallback, useEffect, useState } from "react";
import { notificationApi } from "@/lib/api/notification.api";

// Converts the base64url VAPID public key into the Uint8Array PushManager expects.
function urlBase64ToUint8Array(base64String: string): Uint8Array<ArrayBuffer> {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const rawData = window.atob(base64);
  const output = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; i++) {
    output[i] = rawData.charCodeAt(i);
  }
  return output;
}

export type PushPermissionState = "unsupported" | "default" | "denied" | "granted";

/**
 * Manages the browser Web Push subscription lifecycle: checks support,
 * surfaces current permission, and lets the user opt in/out. Subscribing
 * MUST be triggered by a user gesture (button click) — browsers reject a
 * silent `Notification.requestPermission()` call on page load.
 */
export function usePushNotifications() {
  const [state, setState] = useState<PushPermissionState>("default");
  const [isSubscribed, setIsSubscribed] = useState(false);
  const [isBusy, setIsBusy] = useState(false);

  const vapidPublicKey = process.env.NEXT_PUBLIC_VAPID_PUBLIC_KEY;
  const supported =
    typeof window !== "undefined" &&
    "serviceWorker" in navigator &&
    "PushManager" in window &&
    !!vapidPublicKey;

  const refresh = useCallback(async () => {
    if (!supported) {
      setState("unsupported");
      return;
    }
    setState(Notification.permission as PushPermissionState);
    try {
      const registration = await navigator.serviceWorker.ready;
      const existing = await registration.pushManager.getSubscription();
      setIsSubscribed(!!existing);
    } catch {
      setIsSubscribed(false);
    }
  }, [supported]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const subscribe = useCallback(async () => {
    if (!supported || !vapidPublicKey) return false;
    setIsBusy(true);
    try {
      const permission = await Notification.requestPermission();
      setState(permission as PushPermissionState);
      if (permission !== "granted") return false;

      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidPublicKey),
      });

      await notificationApi.subscribePush(subscription.toJSON() as PushSubscriptionJSON);
      setIsSubscribed(true);
      return true;
    } catch (err) {
      console.error("[usePushNotifications] subscribe failed:", err);
      return false;
    } finally {
      setIsBusy(false);
    }
  }, [supported, vapidPublicKey]);

  const unsubscribe = useCallback(async () => {
    setIsBusy(true);
    try {
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.getSubscription();
      if (subscription) {
        await notificationApi.unsubscribePush(subscription.endpoint);
        await subscription.unsubscribe();
      }
      setIsSubscribed(false);
      return true;
    } catch (err) {
      console.error("[usePushNotifications] unsubscribe failed:", err);
      return false;
    } finally {
      setIsBusy(false);
    }
  }, []);

  return { state, isSubscribed, isBusy, subscribe, unsubscribe };
}
