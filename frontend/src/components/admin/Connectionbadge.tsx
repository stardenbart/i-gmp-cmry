import { cn } from "@/lib/utils";
import { Wifi, WifiOff } from "lucide-react";

export function ConnectionBadge({ isConnected }: { isConnected: boolean }) {
  return (
    <div
      className={cn(
        "flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium transition-colors",
        isConnected
          ? "bg-green-500/10 text-green-500"
          : "bg-muted text-muted-foreground"
      )}
    >
      {isConnected ? (
        <>
          <Wifi className="h-3 w-3" />
          <span>Real-time</span>
        </>
      ) : (
        <>
          <WifiOff className="h-3 w-3" />
          <span>Polling</span>
        </>
      )}
    </div>
  );
}