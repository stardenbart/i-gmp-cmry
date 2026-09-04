import type { NextRequest } from "next/server";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{ path: string[] }>;
};

const REQUEST_HEADERS_TO_REMOVE = [
  "connection",
  "content-length",
  "host",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
];

const RESPONSE_HEADERS_TO_REMOVE = new Set([
  "connection",
  "content-encoding",
  "content-length",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
]);

function backendBaseURL(): string {
  const fallback = process.env.NODE_ENV === "production"
    ? "http://backend:8080"
    : "http://127.0.0.1:8080";

  return (process.env.INTERNAL_BACKEND_URL || fallback).replace(/\/+$/, "");
}

function splitCombinedSetCookie(value: string): string[] {
  // A comma inside Expires (for example "Fri, 04 Sep 2026") is not a cookie
  // separator. A real separator is followed by another cookie-name=value.
  return value.split(/,(?=\s*[A-Za-z0-9!#$%&'*+.^_`|~-]+=)/g);
}

function getSetCookies(headers: Headers): string[] {
  const headersWithCookies = headers as Headers & {
    getSetCookie?: () => string[];
  };
  const cookies = headersWithCookies.getSetCookie?.() ?? [];
  if (cookies.length > 0) return cookies;

  const combined = headers.get("set-cookie");
  return combined ? splitCombinedSetCookie(combined) : [];
}

function buildRequestHeaders(request: NextRequest): Headers {
  const headers = new Headers(request.headers);
  for (const name of REQUEST_HEADERS_TO_REMOVE) headers.delete(name);

  const host = request.headers.get("host");
  if (host) headers.set("x-forwarded-host", host);
  headers.set("x-forwarded-proto", request.nextUrl.protocol.replace(":", ""));
  return headers;
}

function buildResponseHeaders(upstream: Response): Headers {
  const headers = new Headers();
  upstream.headers.forEach((value, name) => {
    const normalizedName = name.toLowerCase();
    if (normalizedName === "set-cookie" || RESPONSE_HEADERS_TO_REMOVE.has(normalizedName)) return;
    headers.append(name, value);
  });

  // Do not use headers.set("set-cookie", ...): login and refresh return three
  // distinct cookies (access, refresh, CSRF) and each one must survive.
  for (const cookie of getSetCookies(upstream.headers)) {
    headers.append("set-cookie", cookie);
  }
  return headers;
}

async function proxyAuthRequest(request: NextRequest, context: RouteContext): Promise<Response> {
  const { path } = await context.params;
  const encodedPath = path.map(encodeURIComponent).join("/");
  const target = new URL(`${backendBaseURL()}/api/v1/auth/${encodedPath}`);
  target.search = request.nextUrl.search;

  const method = request.method.toUpperCase();
  const body = method === "GET" || method === "HEAD"
    ? undefined
    : await request.arrayBuffer();

  try {
    const upstream = await fetch(target, {
      method,
      headers: buildRequestHeaders(request),
      body,
      cache: "no-store",
      redirect: "manual",
    });

    return new Response(upstream.body, {
      status: upstream.status,
      statusText: upstream.statusText,
      headers: buildResponseHeaders(upstream),
    });
  } catch {
    return Response.json(
      { code: "BAD_GATEWAY", message: "Layanan autentikasi tidak dapat dijangkau" },
      { status: 502 },
    );
  }
}

export const GET = proxyAuthRequest;
export const HEAD = proxyAuthRequest;
export const POST = proxyAuthRequest;
export const PUT = proxyAuthRequest;
export const PATCH = proxyAuthRequest;
export const DELETE = proxyAuthRequest;
export const OPTIONS = proxyAuthRequest;
