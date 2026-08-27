"use client";

import { useState, useRef, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import {
  Bell,
  BellRing,
  BellOff,
  Check,
  CheckCheck,
  AlertCircle,
  AlertTriangle,
  Info,
  X,
  ExternalLink,
  Loader2,
} from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { cn } from "@/lib/utils";
import { useMounted } from "@/lib/useMounted";
import { notificationApi, Notification } from "@/lib/api/notification.api";
import { usePushNotifications } from "@/hooks/usePushNotifications";

// Helper to get icon based on notification type
const getNotificationIcon = (type: Notification["type"]) => {
  switch (type) {
    case "warning":
      return <AlertTriangle className="h-4 w-4 text-amber-500" />;
    case "error":
      return <AlertCircle className="h-4 w-4 text-red-500" />;
    case "success":
      return <Check className="h-4 w-4 text-green-500" />;
    default:
      return <Info className="h-4 w-4 text-blue-500" />;
  }
};

// Helper to format time ago
const formatTimeAgo = (dateString: string): string => {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffMins < 1) return "Baru saja";
  if (diffMins < 60) return `${diffMins}m lalu`;
  if (diffHours < 24) return `${diffHours}j lalu`;
  if (diffDays < 7) return `${diffDays}h lalu`;
  return date.toLocaleDateString("id-ID", { day: "numeric", month: "short" });
};

// Single Notification Item Component
function NotificationItem({
  notification,
  basePath,
  onRead,
  onClose,
}: {
  notification: Notification;
  basePath: string;
  onRead: (id: string) => void;
  onClose: () => void;
}) {
  const router = useRouter();

  const handleClick = () => {
    if (!notification.is_read) {
      onRead(notification.id);
    }
    if (notification.link) {
      const finalLink = notification.link.startsWith("/") 
        ? `${basePath}${notification.link}` 
        : notification.link;
      router.push(finalLink);
    }
    onClose();
  };

  return (
    <div
      onClick={handleClick}
      className={cn(
        "group flex gap-3 p-3 hover:bg-muted/50 cursor-pointer transition-colors border-b border-border last:border-b-0",
        !notification.is_read && "bg-primary/5"
      )}
    >
      <div className="flex-shrink-0 mt-0.5">
        {getNotificationIcon(notification.type)}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <p className={cn(
            "text-sm truncate",
            !notification.is_read ? "font-medium" : "text-muted-foreground"
          )}>
            {notification.title}
          </p>
          {!notification.is_read && (
            <span className="flex-shrink-0 w-2 h-2 rounded-full bg-primary mt-1.5" />
          )}
        </div>
        <p className="text-xs text-muted-foreground line-clamp-2 mt-0.5">
          {notification.message}
        </p>
        <div className="flex items-center gap-2 mt-1.5">
          <span className="text-[10px] text-muted-foreground/70">
            {formatTimeAgo(notification.created_at)}
          </span>
          {notification.link && (
            <ExternalLink className="h-3 w-3 text-muted-foreground/50 group-hover:text-primary transition-colors" />
          )}
        </div>
      </div>
    </div>
  );
}

// Main Notification Bell Component
export function NotificationBell() {
  const router = useRouter();
  const params = useParams();
  const user = useAuthStore((state) => state.user);
  const mounted = useMounted();
  const queryClient = useQueryClient();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const push = usePushNotifications();

  const plantCode = (params?.plantCode as string) || user?.plant_id || "global";
  const basePath = `/cimory/${plantCode}/dashboard/${mounted ? user?.id : 'overview'}`;

  // Fetch notifications
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["notifications"],
    queryFn: () => notificationApi.getAll(),
    enabled: mounted && !!user,
    refetchInterval: 60000, // Refetch every minute
  });

  const notifications = data?.data?.items || [];
  const unreadCount = notifications.filter((n) => !n.is_read).length;

  // Mark as read mutation
  const markAsReadMutation = useMutation({
    mutationFn: (id: string) => notificationApi.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  // Mark all as read mutation
  const markAllAsReadMutation = useMutation({
    mutationFn: () => notificationApi.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  // Close on escape key
  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleEscape);
    }
    return () => {
      document.removeEventListener("keydown", handleEscape);
    };
  }, [isOpen]);

  const handleMarkAsRead = (id: string) => {
    markAsReadMutation.mutate(id);
  };

  if (!mounted || !user) {
    return (
      <button
        type="button"
        disabled
        aria-label="Memuat notifikasi"
        className="relative flex h-9 w-9 items-center justify-center rounded-full"
      >
        <Bell aria-hidden="true" className="h-4 w-4" />
      </button>
    );
  }

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Bell Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "flex h-9 w-9 min-h-[36px] min-w-[36px] items-center justify-center rounded-full hover:bg-muted transition-colors relative focus:outline-none focus-visible:ring-2 focus-visible:ring-primary",
          isOpen && "bg-muted"
        )}
        aria-label={`Pusat Notifikasi${unreadCount > 0 ? `, ${unreadCount} belum dibaca` : ""}`}
        title="Notifikasi"
      >
        <Bell className="h-4 w-4" />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white ring-2 ring-background">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </button>

      {/* Dropdown Panel */}
      {isOpen && (
        <div className="
          notification-panel-mobile fixed left-3 right-3 z-[60] flex flex-col overflow-hidden
          sm:absolute sm:left-auto sm:right-[-10px] sm:top-full sm:mt-2 sm:w-96
          origin-top sm:origin-top-right 
          rounded-xl bg-card border border-border shadow-2xl 
          animate-in fade-in-0 zoom-in-95 slide-in-from-top-2 duration-200
        "
        role="dialog"
        aria-modal="true"
        aria-label="Daftar Notifikasi"
        >
          {/* Header */}
          <div className="flex items-center justify-between p-3 border-b border-border">
            <h3 className="font-semibold text-sm">Notifikasi</h3>
            <div className="flex items-center gap-1">
              {push.state === "denied" ? (
                <span
                  className="p-1.5 text-muted-foreground/50"
                  title="Notifikasi push diblokir di pengaturan browser"
                >
                  <BellOff className="h-3.5 w-3.5" />
                </span>
              ) : push.state !== "unsupported" ? (
                <button
                  onClick={() => (push.isSubscribed ? push.unsubscribe() : push.subscribe())}
                  disabled={push.isBusy}
                  className="flex items-center gap-1 px-2 py-1 text-xs text-primary hover:bg-primary/10 rounded-md transition-colors disabled:opacity-50 focus-visible:ring-2 focus-visible:ring-primary"
                  title={push.isSubscribed ? "Matikan notifikasi push di perangkat ini" : "Aktifkan notifikasi push di perangkat ini"}
                  aria-label={push.isSubscribed ? "Matikan notifikasi push" : "Aktifkan notifikasi push"}
                >
                  {push.isBusy ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : push.isSubscribed ? (
                    <BellRing className="h-3 w-3" />
                  ) : (
                    <Bell className="h-3 w-3" />
                  )}
                </button>
              ) : null}
              {unreadCount > 0 && (
                <button
                  onClick={() => markAllAsReadMutation.mutate()}
                  disabled={markAllAsReadMutation.isPending}
                  className="flex items-center gap-1 px-2 py-1 text-xs text-primary hover:bg-primary/10 rounded-md transition-colors disabled:opacity-50 focus-visible:ring-2 focus-visible:ring-primary"
                  aria-label="Tandai semua notifikasi sebagai dibaca"
                >
                  {markAllAsReadMutation.isPending ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <CheckCheck className="h-3 w-3" />
                  )}
                  Tandai semua dibaca
                </button>
              )}
              <button
                onClick={() => setIsOpen(false)}
                className="p-1 hover:bg-muted rounded-md transition-colors focus-visible:ring-2 focus-visible:ring-primary"
                aria-label="Tutup panel notifikasi"
                title="Tutup"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>

          {/* Notification List */}
          <div className="min-h-0 flex-1 overflow-y-auto sm:max-h-96">
            {isLoading || isFetching ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : notifications.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
                <Bell className="h-10 w-10 text-muted-foreground/30 mb-2" />
                <p className="text-sm text-muted-foreground">Tidak ada notifikasi</p>
              </div>
            ) : (
              notifications.map((notification) => (
                <NotificationItem
                  key={notification.id}
                  notification={notification}
                  basePath={basePath}
                  onRead={handleMarkAsRead}
                  onClose={() => setIsOpen(false)}
                />
              ))
            )}
          </div>

          {/* Footer */}
          {notifications.length > 0 && (
            <div className="p-2 border-t border-border">
              <button 
                onClick={() => {
                  setIsOpen(false);
                  router.push(`${basePath}/notifications`);
                }}
                className="w-full text-center text-xs text-primary hover:bg-primary/10 py-1.5 rounded-md transition-colors"
              >
                Lihat semua notifikasi
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default NotificationBell;
