// Hand-rolled, path-based router (docs/spa-migration.md §3.7, decision 3 in
// §7). No router dependency: App.svelte holds the current path in `$state`,
// this module supplies the pure matching logic plus the two DOM listeners
// (popstate, link interception) that keep it in sync with the URL.
import type { Component } from "svelte";
import Home from "./routes/Home.svelte";

// The SPA is staged under /app/* until Phase 7 moves it to "/" (§7 decision
// 2). Every route pattern below is relative to this — routes never mention it
// themselves, so the Phase 7 move is a one-line change here.
export const BASE_PATH = "/app";

export interface RouteDef {
  /** Pattern relative to BASE_PATH, e.g. "/recurrent-expenses/:id/edit". */
  path: string;
  component: Component<Record<string, string>>;
}

export interface RouteMatch {
  component: Component<Record<string, string>>;
  params: Record<string, string>;
}

// The match table. Phase 2 onward adds entries here as each resource lands;
// Phase 1 ships only the placeholder root so the mechanism exists before any
// real route needs it.
export const routes: RouteDef[] = [{ path: "/", component: Home }];

function compile(path: string): { regex: RegExp; keys: string[] } {
  const keys: string[] = [];
  const pattern = path
    .split("/")
    .map((segment) => {
      if (segment.startsWith(":")) {
        keys.push(segment.slice(1));
        return "([^/]+)";
      }
      return segment.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    })
    .join("/");

  return { regex: new RegExp(`^${pattern}$`), keys };
}

// decodeURIComponent throws URIError on a malformed escape ("%E0%A4%A"), and a
// URL is whatever the address bar was typed with. Matching runs inside a
// `$derived` in App.svelte, so an uncaught throw there takes down the whole
// shell instead of the one route; the raw segment is the honest fallback.
function decodeParam(value: string): string {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

/** Segment-boundary prefix test: "/apple" is not under a BASE_PATH of "/app". */
function isUnderBasePath(pathname: string): boolean {
  return pathname === BASE_PATH || pathname.startsWith(`${BASE_PATH}/`);
}

export function matchRoute(
  candidates: RouteDef[],
  path: string,
): RouteMatch | null {
  for (const route of candidates) {
    const { regex, keys } = compile(route.path);
    const found = regex.exec(path);
    if (!found) continue;

    const params: Record<string, string> = {};
    keys.forEach((key, i) => {
      params[key] = decodeParam(found[i + 1]);
    });

    return { component: route.component, params };
  }

  return null;
}

/** Strips BASE_PATH and a trailing slash, so route patterns never mention it. */
export function toRoutePath(pathname: string): string {
  let path = isUnderBasePath(pathname)
    ? pathname.slice(BASE_PATH.length)
    : pathname;

  if (path.length > 1 && path.endsWith("/")) {
    path = path.slice(0, -1);
  }

  return path === "" ? "/" : path;
}

/** Pushes a new history entry and notifies listeners registered with onPopState. */
export function navigate(path: string): void {
  const full = isUnderBasePath(path) ? path : `${BASE_PATH}${path}`;
  if (full === window.location.pathname) return;

  window.history.pushState({}, "", full);
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export function onPopState(handler: (path: string) => void): () => void {
  const listener = () => handler(toRoutePath(window.location.pathname));
  window.addEventListener("popstate", listener);

  return () => window.removeEventListener("popstate", listener);
}

// An internal <a> is same-origin, under BASE_PATH, has no target/download
// override, and was not already handled (a modifier click opens a new tab, a
// prevented default means another handler claimed it).
function isNavigableClick(
  event: MouseEvent,
  anchor: HTMLAnchorElement,
): boolean {
  if (event.defaultPrevented || event.button !== 0) return false;
  if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
    return false;
  }
  if (anchor.target && anchor.target !== "_self") return false;
  if (anchor.hasAttribute("download")) return false;
  if (anchor.origin !== window.location.origin) return false;
  // An in-page anchor ("#section", or any href landing on the current path with
  // a fragment) is the browser's job. Claiming it would preventDefault the
  // scroll and re-render the route that is already on screen.
  if (anchor.hash && anchor.pathname === window.location.pathname) return false;

  return isUnderBasePath(anchor.pathname);
}

export function onLinkClick(handler: (path: string) => void): () => void {
  const listener = (event: MouseEvent) => {
    const anchor = (event.target as Element).closest("a");
    if (!anchor || !isNavigableClick(event, anchor)) return;

    event.preventDefault();
    window.history.pushState({}, "", anchor.href);
    handler(toRoutePath(anchor.pathname));
  };

  document.addEventListener("click", listener);

  return () => document.removeEventListener("click", listener);
}
