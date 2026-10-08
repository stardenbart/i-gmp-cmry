export function AppFooter() {
  const year = new Date().getFullYear();

  return (
    <footer className="pwa-footer-safe flex flex-wrap items-center justify-center gap-x-2 gap-y-1 border-t border-border bg-footer px-5 py-3.5 text-center text-[11.5px] text-muted-foreground">
      <span>© {year} PT Cisarua Mountain Dairy, Tbk — Plant Sentul</span>
      <span aria-hidden="true" className="text-border">|</span>
      <span>Powered by Digital Transformation</span>
    </footer>
  );
}
