"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { notificationApi } from "@/lib/api/notification.api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import {
  Bell,
  Check,
  CheckCheck,
  Info,
  AlertTriangle,
  XCircle,
  CheckCircle2,
  Loader2,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { id as idLocale } from "date-fns/locale";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { useRouter, useParams } from "next/navigation";

// Helper to get icon based on notification type
const getNotificationIcon = (type: string) => {
  switch (type) {
    case "info":
      return <Info className="h-5 w-5 text-blue-500" />;
    case "warning":
      return <AlertTriangle className="h-5 w-5 text-yellow-500" />;
    case "error":
      return <XCircle className="h-5 w-5 text-red-500" />;
    case "success":
      return <CheckCircle2 className="h-5 w-5 text-green-500" />;
    default:
      return <Bell className="h-5 w-5 text-primary" />;
  }
};

export default function NotificationsPage() {
  const user = useAuthStore((state) => state.user);
  const mounted = useMounted();
  const queryClient = useQueryClient();
  const router = useRouter();
  const { plantCode } = useParams() as { plantCode?: string };

  const basePath = `/cimory/${plantCode || "all"}/dashboard/${mounted && user ? user.id : "overview"}`;

  const { data, isLoading } = useQuery({
    queryKey: ["notifications", "all"],
    queryFn: () => notificationApi.getAll(1, 100), // Get up to 100 on page for now
    enabled: mounted && !!user,
  });

  const markAsReadMutation = useMutation({
    mutationFn: (id: string) => notificationApi.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const markAllAsReadMutation = useMutation({
    mutationFn: () => notificationApi.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  if (!mounted) return null;

  const notifications = data?.data?.items || [];
  const unreadCount = notifications.filter((n) => !n.is_read).length;

  const handleNotificationClick = (notification: any) => {
    if (!notification.is_read) {
      markAsReadMutation.mutate(notification.id);
    }
    if (notification.link) {
      const finalLink = notification.link.startsWith("/")
        ? `${basePath}${notification.link}`
        : notification.link;
      router.push(finalLink);
    }
  };

  return (
    <div className="space-y-6 pb-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Semua Notifikasi</h1>
          <p className="text-muted-foreground mt-1">
            Riwayat lengkap pemberitahuan Anda.
          </p>
        </div>

        {unreadCount > 0 && (
          <button
            onClick={() => markAllAsReadMutation.mutate()}
            disabled={markAllAsReadMutation.isPending}
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary/10 text-primary hover:bg-primary/20 rounded-lg transition-colors font-medium text-sm disabled:opacity-50"
          >
            {markAllAsReadMutation.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <CheckCheck className="h-4 w-4" />
            )}
            Tandai semua sudah dibaca
          </button>
        )}
      </div>

      <div className="bg-card border border-border rounded-xl shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20">
            <Loader2 className="h-8 w-8 animate-spin text-primary mb-4" />
            <p className="text-muted-foreground text-sm">Memuat notifikasi...</p>
          </div>
        ) : notifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center px-4">
            <div className="h-16 w-16 bg-muted rounded-full flex items-center justify-center mb-4">
              <Bell className="h-8 w-8 text-muted-foreground/50" />
            </div>
            <h3 className="text-lg font-medium text-foreground mb-1">
              Tidak ada notifikasi
            </h3>
            <p className="text-muted-foreground text-sm">
              Anda belum memiliki riwayat pemberitahuan.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {notifications.map((notification) => (
              <div
                key={notification.id}
                onClick={() => handleNotificationClick(notification)}
                className={cn(
                  "p-4 sm:p-6 hover:bg-muted/50 transition-colors flex gap-4 sm:gap-6 group cursor-pointer",
                  !notification.is_read ? "bg-primary/5" : ""
                )}
              >
                <div
                  className={cn(
                    "flex-shrink-0 mt-1 h-10 w-10 sm:h-12 sm:w-12 rounded-full flex items-center justify-center border",
                    !notification.is_read
                      ? "bg-background border-primary/20 shadow-sm"
                      : "bg-muted border-border"
                  )}
                >
                  {getNotificationIcon(notification.type)}
                </div>
                
                <div className="flex-1 min-w-0">
                  <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-1 sm:gap-4">
                    <h4
                      className={cn(
                        "text-base sm:text-lg font-medium truncate",
                        !notification.is_read ? "text-foreground font-semibold" : "text-foreground/80"
                      )}
                    >
                      {notification.title}
                    </h4>
                    <span className="text-xs text-muted-foreground whitespace-nowrap flex-shrink-0">
                      {formatDistanceToNow(new Date(notification.created_at), {
                        addSuffix: true,
                        locale: idLocale,
                      })}
                    </span>
                  </div>
                  <p
                    className={cn(
                      "mt-1 text-sm leading-relaxed",
                      !notification.is_read ? "text-foreground/90" : "text-muted-foreground"
                    )}
                  >
                    {notification.message}
                  </p>
                </div>

                {!notification.is_read && (
                  <div className="flex-shrink-0 self-center hidden sm:block opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        markAsReadMutation.mutate(notification.id);
                      }}
                      className="p-2 hover:bg-primary/10 rounded-full text-primary transition-colors tooltip-trigger"
                      title="Tandai dibaca"
                    >
                      <Check className="h-5 w-5" />
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
