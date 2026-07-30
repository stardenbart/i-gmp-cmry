import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export default function proxy(request: NextRequest) {
  const token = request.cookies.get('auth-token')?.value;
  const userId = request.cookies.get('user-id')?.value;
  const { pathname } = request.nextUrl;

  // Allow MinIO bucket images to bypass auth middleware
  if (pathname.startsWith('/monitoring-audit-bucket')) {
    return NextResponse.next();
  }

  const isAuthRoute = pathname === '/login' || pathname === '/forgot-password';

  if (!token && !isAuthRoute) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  if (token && isAuthRoute) {
    return NextResponse.redirect(new URL(`/cimory/dashboard/${userId || 'overview'}`, request.url));
  }
  
  // Optional: Redirect root to dashboard if logged in
  if (token && pathname === '/') {
    return NextResponse.redirect(new URL(`/cimory/dashboard/${userId || 'overview'}`, request.url));
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
     * - sw.js, workbox-* (PWA service worker)
     */
    '/((?!api|monitoring-audit-bucket|_next/static|_next/image|.*\\.png|.*\\.jpg|.*\\.jpeg|.*\\.webp|.*\\.svg|.*\\.ico|.*\\.gif|manifest\\.json|sw\\.js|workbox-.*).*)',
  ],
};
