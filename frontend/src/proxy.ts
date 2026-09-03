import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export default function proxy(request: NextRequest) {
  // user-id is a non-sensitive hint cookie (see lib/utils.ts) — the actual
  // session token is an httpOnly cookie this edge proxy can't (and doesn't
  // need to) read. This is a UX redirect shortcut only, not the real access
  // check: every API call is independently re-authenticated server-side by
  // the backend regardless of what happens here.
  const userId = request.cookies.get('user-id')?.value;
  const isLoggedInHint = Boolean(userId);
  const { pathname } = request.nextUrl;

  // Allow MinIO bucket images to bypass auth middleware
  if (pathname.startsWith('/monitoring-audit-bucket')) {
    return NextResponse.next();
  }

  const isAuthRoute = pathname === '/login' || pathname === '/forgot-password';

  if (!isLoggedInHint && !isAuthRoute) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  const plantCode = request.cookies.get('plant-code')?.value || 'global';

  if (isLoggedInHint && isAuthRoute) {
    return NextResponse.redirect(new URL(`/cimory/${plantCode}/dashboard/${userId || 'overview'}`, request.url));
  }

  // Optional: Redirect root to dashboard if logged in
  if (isLoggedInHint && pathname === '/') {
    return NextResponse.redirect(new URL(`/cimory/${plantCode}/dashboard/${userId || 'overview'}`, request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths EXCEPT:
     * - api (API routes / rewrite proxy)
     * - monitoring-audit-bucket (MinIO proxied images)
     * - _next/static (static files, JS/CSS chunks)
     * - _next/image (Next.js image optimizer)
     * - All image/media extensions (.png, .jpg, .jpeg, .webp, .svg, .ico, .gif)
     * - manifest.json (PWA manifest)
     * - serwist, sw.js, workbox-* (PWA service worker)
     */
    '/((?!api|monitoring-audit-bucket|serwist|_next/static|_next/image|.*\\.png|.*\\.jpg|.*\\.jpeg|.*\\.webp|.*\\.svg|.*\\.ico|.*\\.gif|manifest\\.json|sw\\.js|workbox-.*).*)',
  ],
};
