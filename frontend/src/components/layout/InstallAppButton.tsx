"use client";

import { useState, useEffect } from "react";
import { createPortal } from "react-dom";
import { Download, Laptop, Smartphone, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { useInstallPrompt } from "@/components/pwa/InstallPromptProvider";

export function InstallAppButton({ className }: { className?: string }) {
  const [showGuideModal, setShowGuideModal] = useState(false);
  const { canPrompt, isStandalone, hasCheckedStandalone, requestInstall } = useInstallPrompt();

  useEffect(() => {
    if (!showGuideModal) return;

    const previousOverflow = document.body.style.overflow;
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setShowGuideModal(false);
    };

    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", handleEscape);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", handleEscape);
    };
  }, [showGuideModal]);

  const handleInstallClick = async () => {
    if (!canPrompt) {
      setShowGuideModal(true);
      return;
    }

    try {
      const outcome = await requestInstall();
      if (outcome === "accepted") {
        toast.success("Permintaan instalasi diterima.");
      } else if (outcome === "unavailable") {
        setShowGuideModal(true);
      }
    } catch {
      setShowGuideModal(true);
    }
  };

  // Do not flash the install control while detecting PWA mode, and do not
  // occupy header space after the application has already been installed.
  if (!hasCheckedStandalone || isStandalone) return null;

  return (
    <>
      <button
        type="button"
        onClick={handleInstallClick}
        className={cn(
          "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-primary/30 bg-primary/10 text-primary shadow-sm transition-all",
          "hover:bg-primary hover:text-primary-foreground active:scale-95 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary",
          "xl:w-auto xl:gap-2 xl:rounded-full xl:px-3",
          className
        )}
        title="Install Aplikasi ke Perangkat"
        aria-label="Install aplikasi ke perangkat"
      >
        <Download className="h-4 w-4 shrink-0" />
        <span className="hidden whitespace-nowrap text-xs font-semibold xl:inline">Install App</span>
      </button>

      {showGuideModal && typeof document !== "undefined" && createPortal(
        <div
          className="fixed inset-0 z-[9999] flex items-start justify-center overflow-y-auto overscroll-contain bg-black/70 px-3 pb-[calc(0.75rem+var(--safe-area-bottom))] pt-[calc(0.75rem+var(--safe-area-top))] backdrop-blur-sm animate-in fade-in duration-200 sm:items-center sm:p-6"
          role="dialog"
          aria-modal="true"
          aria-labelledby="install-app-title"
          onClick={() => setShowGuideModal(false)}
        >
          <div
            className="install-guide-panel relative my-auto w-full max-w-md overflow-y-auto rounded-2xl border border-border bg-card p-4 text-left shadow-2xl sm:p-5"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              onClick={() => setShowGuideModal(false)}
              className="absolute right-3 top-3 flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
              aria-label="Tutup panduan instalasi"
            >
              <X className="h-5 w-5" />
            </button>

            <div className="flex items-center gap-3 pr-10">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <Download className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <h3 id="install-app-title" className="text-base font-bold text-foreground">Install Aplikasi</h3>
                <p className="text-xs text-muted-foreground">Petunjuk pemasangan ke perangkat</p>
              </div>
            </div>

            <div className="mt-4 space-y-3 text-xs leading-relaxed text-muted-foreground">
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
              className="mt-4 w-full rounded-xl"
            >
              Mengerti
            </Button>
          </div>
        </div>,
        document.body
      )}
    </>
  );
}
