"use client";

import { useState, useEffect } from "react";
import { Download, CheckCircle2, Laptop, Smartphone, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

interface StandaloneNavigator extends Navigator {
  standalone?: boolean;
}

export function InstallAppButton() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [isStandalone, setIsStandalone] = useState(false);
  const [showGuideModal, setShowGuideModal] = useState(false);

  useEffect(() => {
    const checkStandalone = () => {
      const isStandaloneMode =
        window.matchMedia("(display-mode: standalone)").matches ||
        (window.navigator as StandaloneNavigator).standalone === true;
      setIsStandalone(isStandaloneMode);
    };

    checkStandalone();

    const handleBeforeInstallPrompt = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e as BeforeInstallPromptEvent);
    };

    window.addEventListener("beforeinstallprompt", handleBeforeInstallPrompt);

    return () => {
      window.removeEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
    };
  }, []);

  const handleInstallClick = async () => {
    if (deferredPrompt) {
      try {
        deferredPrompt.prompt();
        const { outcome } = await deferredPrompt.userChoice;
        if (outcome === "accepted") {
          setDeferredPrompt(null);
          toast.success("Aplikasi berhasil dipasang!");
        }
      } catch {
        setShowGuideModal(true);
      }
    } else {
      setShowGuideModal(true);
    }
  };

  if (isStandalone) {
    return (
      <div className="flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
        <CheckCircle2 className="h-3.5 w-3.5" />
        <span className="hidden sm:inline">App Mode</span>
      </div>
    );
  }

  return (
    <>
      <button
        onClick={handleInstallClick}
        className={cn(
          "flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-semibold transition-all",
          "bg-primary/10 text-primary hover:bg-primary hover:text-primary-foreground border border-primary/30 shadow-sm"
        )}
        title="Install Aplikasi ke Perangkat"
      >
        <Download className="h-3.5 w-3.5" />
        <span className="hidden sm:inline-block">Install App</span>
      </button>

      {showGuideModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="relative w-full max-w-md rounded-2xl border border-border bg-card p-5 shadow-xl space-y-4 text-left">
            <button
              onClick={() => setShowGuideModal(false)}
              className="absolute right-4 top-4 text-muted-foreground hover:text-foreground"
            >
              <X className="h-5 w-5" />
            </button>

            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
                <Download className="h-5 w-5" />
              </div>
              <div>
                <h3 className="font-bold text-base text-foreground">Install Aplikasi (PWA)</h3>
                <p className="text-xs text-muted-foreground">Petunjuk pemasangan ke perangkat</p>
              </div>
            </div>

            <div className="space-y-3 text-xs leading-relaxed text-muted-foreground">
              <div className="p-3 rounded-xl bg-muted/50 border border-border/50 space-y-1.5">
                <div className="flex items-center gap-2 font-semibold text-foreground">
                  <Laptop className="h-4 w-4 text-primary" />
                  <span>Google Chrome / MS Edge (Desktop)</span>
                </div>
                <p>
                  Klik ikon titik tiga <strong>(⋮)</strong> di pojok kanan atas browser ➔ pilih <strong>Simpan dan Bagikan / Install App</strong> ➔ klik <strong>Install</strong>.
                </p>
              </div>

              <div className="p-3 rounded-xl bg-muted/50 border border-border/50 space-y-1.5">
                <div className="flex items-center gap-2 font-semibold text-foreground">
                  <Smartphone className="h-4 w-4 text-primary" />
                  <span>Android (Chrome)</span>
                </div>
                <p>
                  Buka menu browser <strong>(⋮)</strong> ➔ pilih <strong>Tambahkan ke Layar Utama (Add to Home screen)</strong> atau <strong>Install Aplikasi</strong>.
                </p>
              </div>

              <div className="p-3 rounded-xl bg-muted/50 border border-border/50 space-y-1.5">
                <div className="flex items-center gap-2 font-semibold text-foreground">
                  <Smartphone className="h-4 w-4 text-primary" />
                  <span>iOS Safari (iPhone / iPad)</span>
                </div>
                <p>
                  Tekan tombol Bagikan <strong>(Share)</strong> di bagian bawah ➔ gulir ke bawah dan pilih <strong>Tambahkan ke Layar Utama (Add to Home Screen)</strong>.
                </p>
              </div>
            </div>

            <Button
              onClick={() => setShowGuideModal(false)}
              className="w-full rounded-xl"
            >
              Mengerti
            </Button>
          </div>
        </div>
      )}
    </>
  );
}
