"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";

export interface SSEMessage<T = unknown> {
  type: string;
  data: T;
  timestamp?: string;
}

export interface SSEOptions {
  /** Event types to listen for */
  eventTypes?: string[];
  /** Reconnect interval in ms (default: 5000) */
  reconnectInterval?: number;
  /** Maximum reconnection attempts (default: 10) */
  maxRetries?: number;
  /** Callback when connection opens */
  onOpen?: () => void;
  /** Callback when connection closes */
  onClose?: () => void;
  /** Callback on error */
  onError?: (error: Event) => void;
}

interface UseSSEState<T> {
  data: T | null;
  isConnected: boolean;
  error: Error | null;
  lastMessage: SSEMessage<T> | null;
}

/**
 * Hook for Server-Sent Events (SSE) connection.
 * Automatically handles reconnection and cleanup.
 *
 * @param endpoint - SSE endpoint URL (without leading slash)
 * @param options - SSE configuration options
 * @returns SSE state and controls
 *
 * @example
 * const { data, isConnected, sendMessage } = useSSE<MyData>('/api/events');
 */
export function useSSE<T = unknown>(
  endpoint: string,
  options: SSEOptions = {}
): UseSSEState<T> & {
  sendMessage: (event: string, data: unknown) => void;
  reconnect: () => void;
} {
  const { token } = useAuthStore();
  const mounted = useMounted();
  const {
    reconnectInterval = 5000,
    maxRetries = 10,
    onOpen,
    onClose,
    onError,
  } = options;

  const [data, setData] = useState<T | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [lastMessage, setLastMessage] = useState<SSEMessage<T> | null>(null);

  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const retryCountRef = useRef(0);
  const reconnectAttemptRef = useRef(0);

  const baseURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
  const fullUrl = `${baseURL}/${endpoint}`;

  const connect = useCallback(() => {
    if (!mounted || eventSourceRef.current) return;

    try {
      // Build URL with auth token
      const url = new URL(fullUrl);
      if (token) {
        url.searchParams.set("token", token);
      }

      const eventSource = new EventSource(url.toString());
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        setIsConnected(true);
        setError(null);
        retryCountRef.current = 0;
        reconnectAttemptRef.current = 0;
        onOpen?.();
      };

      eventSource.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as SSEMessage<T>;
          setLastMessage(message);
          setData(message.data as T);
        } catch {
          // Plain text message
          setData(event.data as unknown as T);
        }
      };

      eventSource.onerror = (err) => {
        setIsConnected(false);
        const errorEvent = new Event("error");
        onError?.(errorEvent);

        eventSource.close();
        eventSourceRef.current = null;

        // Reconnect logic
        if (retryCountRef.current < maxRetries) {
          reconnectTimeoutRef.current = setTimeout(() => {
            retryCountRef.current++;
            reconnectAttemptRef.current++;
            connect();
          }, reconnectInterval);
        } else {
          setError(new Error("Max reconnection attempts reached"));
        }
      };

      // Add custom event listeners
      if (options.eventTypes) {
        options.eventTypes.forEach((eventType) => {
          eventSource.addEventListener(eventType, (e: MessageEvent) => {
            try {
              const message = JSON.parse(e.data) as SSEMessage<T>;
              setLastMessage(message);
              setData(message.data as T);
            } catch {
              setData(e.data as unknown as T);
            }
          });
        });
      }
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to connect"));
    }
  }, [mounted, token, fullUrl, options.eventTypes, maxRetries, reconnectInterval, onOpen, onError]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setIsConnected(false);
    onClose?.();
  }, [onClose]);

  const reconnect = useCallback(() => {
    disconnect();
    retryCountRef.current = 0;
    connect();
  }, [connect, disconnect]);

  // Note: We can't actually send messages via EventSource (it's receive-only)
  // For sending, you'd need a separate HTTP POST call
  const sendMessage = useCallback((_event: string, _data: unknown) => {
    // This is a placeholder - SSE is receive-only
    // For sending data, use fetch/axios instead
    console.warn("SSE sendMessage: SSE is receive-only. Use HTTP POST for sending data.");
  }, []);

  // Connect on mount, disconnect on unmount
  useEffect(() => {
    if (mounted) {
      connect();
    }
    return () => {
      disconnect();
    };
  }, [mounted, connect, disconnect]);

  return {
    data,
    isConnected,
    error,
    lastMessage,
    sendMessage,
    reconnect,
  };
}

/**
 * Hook for dashboard real-time statistics via SSE
 */
export interface DashboardStats {
  total_inspections: number;
  open_issues: number;
  closed_issues: number;
  total_pics: number;
  issues_by_month: Array<{ month: string; count: number }>;
  recent_inspections: Array<{
    id: string;
    name: string;
    status: string;
    created_at: string;
  }>;
}

export function useDashboardSSE(): UseSSEState<DashboardStats> & { refresh: () => void } {
  const mounted = useMounted();
  const { token } = useAuthStore();

  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [lastUpdate, setLastUpdate] = useState<SSEMessage<DashboardStats> | null>(null);

  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (!mounted || !token) return;

    const baseURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
    const url = new URL(`${baseURL}/events/dashboard`);
    url.searchParams.set("token", token);

    const connect = () => {
      const eventSource = new EventSource(url.toString());

      eventSource.onopen = () => {
        setIsConnected(true);
        setError(null);
      };

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          setStats(data);
          setLastUpdate({ type: "stats", data, timestamp: new Date().toISOString() });
        } catch (err) {
          console.error("Failed to parse SSE message:", err);
        }
      };

      eventSource.addEventListener("stats", (e: MessageEvent) => {
        try {
          const data = JSON.parse(e.data) as DashboardStats;
          setStats(data);
          setLastUpdate({ type: "stats", data, timestamp: new Date().toISOString() });
        } catch (err) {
          console.error("Failed to parse stats event:", err);
        }
      });

      eventSource.onerror = () => {
        setIsConnected(false);
        eventSource.close();
        eventSourceRef.current = null;

        // Reconnect after 5 seconds
        reconnectTimeoutRef.current = setTimeout(connect, 5000);
      };

      eventSourceRef.current = eventSource;
    };

    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    };
  }, [mounted, token]);

  const refresh = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    // Trigger reconnection
    setTimeout(() => {
      if (mounted && token) {
        const baseURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
        const url = new URL(`${baseURL}/events/dashboard`);
        url.searchParams.set("token", token);
        const eventSource = new EventSource(url.toString());

        eventSource.onmessage = (e) => {
          try {
            setStats(JSON.parse(e.data));
          } catch {}
        };

        eventSource.onerror = () => {
          eventSource.close();
        };

        eventSourceRef.current = eventSource;
      }
    }, 100);
  }, [mounted, token]);

  return {
    data: stats,
    isConnected,
    error,
    lastMessage: lastUpdate,
    refresh,
  };
}

/**
 * Hook for real-time notifications via SSE
 */
export function useNotificationsSSE(): UseSSEState<unknown[]> & {
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
} {
  const mounted = useMounted();
  const { token } = useAuthStore();
  const [notifications, setNotifications] = useState<unknown[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [lastMessage, setLastMessage] = useState<SSEMessage<unknown[]> | null>(null);

  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (!mounted || !token) return;

    const baseURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
    const url = new URL(`${baseURL}/events/notifications`);
    url.searchParams.set("token", token);

    const connect = () => {
      const eventSource = new EventSource(url.toString());

      eventSource.onopen = () => {
        setIsConnected(true);
        setError(null);
      };

      eventSource.addEventListener("notification", (e: MessageEvent) => {
        try {
          const data = JSON.parse(e.data);
          setNotifications((prev) => [data, ...prev.slice(0, 49)]); // Keep last 50
          setLastMessage({ type: "notification", data, timestamp: new Date().toISOString() });
        } catch (err) {
          console.error("Failed to parse notification:", err);
        }
      });

      eventSource.addEventListener("clear", () => {
        setNotifications([]);
        setLastMessage({ type: "clear", data: [] as unknown[], timestamp: new Date().toISOString() });
      });

      eventSource.onerror = () => {
        setIsConnected(false);
        eventSource.close();
        eventSourceRef.current = null;

        // Reconnect after 5 seconds
        reconnectTimeoutRef.current = setTimeout(connect, 5000);
      };

      eventSourceRef.current = eventSource;
    };

    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    };
  }, [mounted, token]);

  const markAsRead = useCallback(async (id: string) => {
    // Send HTTP request to mark as read
    const baseURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
    try {
      await fetch(`${baseURL}/notifications/${id}/read`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      });
    } catch (err) {
      console.error("Failed to mark notification as read:", err);
    }
  }, [token]);

  const markAllAsRead = useCallback(async () => {
    const baseURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
    try {
      await fetch(`${baseURL}/notifications/read-all`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      });
    } catch (err) {
      console.error("Failed to mark all notifications as read:", err);
    }
  }, [token]);

  return {
    data: notifications,
    isConnected,
    error,
    lastMessage,
    markAsRead,
    markAllAsRead,
  };
}
