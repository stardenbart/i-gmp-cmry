"use client";

import { useState, useEffect } from "react";
import { Download } from "lucide-react";
import { cn } from "@/lib/utils";

export function InstallAppButton() {
  const [deferredPrompt, setDeferredPrompt] = useState<any>(null);
  const [isInstallable, setIsInstallable] = useState(false);

  useEffect(() => {
    const handleBeforeInstallPrompt = (e: any) => {
      // Prevent the mini-infobar from appearing on mobile
      e.preventDefault();
      // Stash the event so it can be triggered later.
      setDeferredPrompt(e);
      // Update UI notify the user they can install the PWA
      setIsInstallable(true);
    };

    window.addEventListener("beforeinstallprompt", handleBeforeInstallPrompt);

    return () => {
      window.removeEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
    };
  }, []);

  const handleInstallClick = async () => {
    if (!deferredPrompt) return;
    
    // Show the install prompt
    deferredPrompt.prompt();
    
    // Wait for the user to respond to the prompt
    const { outcome } = await deferredPrompt.userChoice;
    
    // We no longer need the prompt. Clear it up.
    if (outcome === 'accepted') {
      setDeferredPrompt(null);
      setIsInstallable(false);
    }
  };

  if (!isInstallable) {
    return null; // Hide if not installable (already installed, or not supported)
  }

  return (
    <button
      onClick={handleInstallClick}
      className={cn(
        "flex items-center gap-2 rounded-full px-3 py-1.5 text-sm font-medium transition-all",
        "bg-primary/10 text-primary hover:bg-primary hover:text-primary-foreground shadow-sm"
      )}
      title="Install Aplikasi ke Perangkat"
    >
      <Download className="h-4 w-4" />
      <span className="hidden sm:inline-block">Install App</span>
    </button>
  );
}
