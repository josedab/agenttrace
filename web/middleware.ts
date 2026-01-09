import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { auth } from "@/lib/auth";

// Routes that don't require authentication
const publicRoutes = [
  "/",
  "/sign-in",
  "/sign-up",
  "/forgot-password",
  "/reset-password",
  "/error",
  "/terms",
  "/privacy",
];

// Routes that should redirect to dashboard if authenticated
const authRoutes = [
  "/sign-in",
  "/sign-up",
  "/forgot-password",
  "/reset-password",
];

export default auth((req) => {
  const { nextUrl } = req;
  const isAuthenticated = !!req.auth;
  const pathname = nextUrl.pathname;

  // Check if the current path matches any public routes
  const isPublicRoute = publicRoutes.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  // Check if the current path is an auth route
  const isAuthRoute = authRoutes.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  // API routes should be handled separately
  if (pathname.startsWith("/api")) {
    return NextResponse.next();
  }

  // Redirect authenticated users away from auth routes
  if (isAuthenticated && isAuthRoute) {
    return NextResponse.redirect(new URL("/dashboard", nextUrl.origin));
  }

  // Redirect unauthenticated users to sign-in
  if (!isAuthenticated && !isPublicRoute) {
    const callbackUrl = encodeURIComponent(pathname + nextUrl.search);
    return NextResponse.redirect(
      new URL(`/sign-in?callbackUrl=${callbackUrl}`, nextUrl.origin)
    );
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (images, etc.)
     */
    "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
  ],
};
