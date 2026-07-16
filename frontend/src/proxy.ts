import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export default function proxy(request: NextRequest) {
  // const token = request.cookies.get('auth-token')?.value;
  // const userId = request.cookies.get('user-id')?.value;
  // const { pathname } = request.nextUrl;

  // const isAuthRoute = pathname === '/login' || pathname === '/forgot-password';

  // if (!token && !isAuthRoute) {
  //   return NextResponse.redirect(new URL('/login', request.url));
  // }

  // if (token && isAuthRoute) {
  //   return NextResponse.redirect(new URL(`/cimory/dashboard/${userId || 'overview'}`, request.url));
  // }
  
  // // Optional: Redirect root to dashboard if logged in
  // if (token && pathname === '/') {
  //   return NextResponse.redirect(new URL(`/cimory/dashboard/${userId || 'overview'}`, request.url));
  // }

  // return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - manifest.json (PWA manifest)
     * - icon-*.png (PWA icons)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|manifest.json|icon-.*\\.png).*)',
  ],
};
