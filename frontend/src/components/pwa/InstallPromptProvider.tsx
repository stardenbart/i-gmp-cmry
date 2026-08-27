"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

export interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

interface StandaloneNavigator extends Navigator {
  standalone?: boolean;
}

type InstallOutcome = "accepted" | "dismissed" | "unavailable";

interface InstallPromptContextValue {
  canPrompt: boolean;
  isStandalone: boolean;
  hasCheckedStandalone: boolean;
  requestInstall: () => Promise<InstallOutcome>;
}

const InstallPromptContext = createContext<InstallPromptContextValue | null>(null);

export function InstallPromptProvider({ children }: { children: React.ReactNode }) {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [isStandalone, setIsStandalone] = useState(false);
  const [hasCheckedStandalone, setHasCheckedStandalone] = useState(false);

  useEffect(() => {
    const displayMode = window.matchMedia("(display-mode: standalone)");
    const checkStandalone = () => {
      setIsStandalone(
        displayMode.matches ||
        (window.navigator as StandaloneNavigator).standalone === true
      );
      setHasCheckedStandalone(true);
    };
    const handleBeforeInstallPrompt = (event: Event) => {
      event.preventDefault();
      setDeferredPrompt(event as BeforeInstallPromptEvent);
    };
    const handleAppInstalled = () => {
      setDeferredPrompt(null);
      setIsStandalone(true);
    };

    checkStandalone();
    window.addEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
    window.addEventListener("appinstalled", handleAppInstalled);
    displayMode.addEventListener?.("change", checkStandalone);

    return () => {
      window.removeEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
      window.removeEventListener("appinstalled", handleAppInstalled);
      displayMode.removeEventListener?.("change", checkStandalone);
    };
  }, []);

  const requestInstall = useCallback(async (): Promise<InstallOutcome> => {
    if (!deferredPrompt) return "unavailable";

    const prompt = deferredPrompt;
    // A beforeinstallprompt event can only be used once, including when the
    // user dismisses it. Clear it before opening the browser UI.
    setDeferredPrompt(null);
    await prompt.prompt();
    const { outcome } = await prompt.userChoice;
    return outcome;
  }, [deferredPrompt]);

  const value = useMemo(
    () => ({
      canPrompt: deferredPrompt !== null,
      isStandalone,
      hasCheckedStandalone,
      requestInstall,
    }),
    [deferredPrompt, hasCheckedStandalone, isStandalone, requestInstall]
  );

  return (
    <InstallPromptContext.Provider value={value}>
      {children}
    </InstallPromptContext.Provider>
  );
}

export function useInstallPrompt() {
  const context = useContext(InstallPromptContext);
  if (!context) {
    throw new Error("useInstallPrompt must be used within InstallPromptProvider");
  }
  return context;
}
