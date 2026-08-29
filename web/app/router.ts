// Hand-rolled, path-based router (docs/spa-migration.md §3.7, decision 3 in
// §7). No router dependency: App.svelte holds the current path in `$state`,
// this module supplies the pure matching logic plus the two DOM listeners
// (popstate, link interception) that keep it in sync with the URL.
import type { Component } from "svelte";
import AccountIndex from "./routes/account/Index.svelte";
import DashboardIndex from "./routes/dashboard/Index.svelte";
import DeleteDataIndex from "./routes/delete_data/Index.svelte";
import ExportsIndex from "./routes/exports/Index.svelte";
import Budgets from "./routes/expenses/Budgets.svelte";
import ExpensesEdit from "./routes/expenses/Edit.svelte";
import ExpensesIndex from "./routes/expenses/Index.svelte";
import ExpensesNew from "./routes/expenses/New.svelte";
import ExpensesShow from "./routes/expenses/Show.svelte";
import Stats from "./routes/expenses/Stats.svelte";
import RecurrentExpensesArchived from "./routes/recurrent_expenses/Archived.svelte";
import RecurrentExpensesEdit from "./routes/recurrent_expenses/Edit.svelte";
import RecurrentExpensesIndex from "./routes/recurrent_expenses/Index.svelte";
import RecurrentExpensesNew from "./routes/recurrent_expenses/New.svelte";
import RecurrentExpensesShow from "./routes/recurrent_expenses/Show.svelte";
import LoginIndex from "./routes/login/Index.svelte";
import RegisterIndex from "./routes/register/Index.svelte";

// The SPA lives at "/" since Phase 7 moved it off "/app/*" (§7 decision 2).
// Every route pattern below is relative to this constant rather than the
// literal, so isUnderBasePath/navigate/toRoutePath keep working unchanged
// with an empty prefix.
export const BASE_PATH = "";

export interface RouteDef {
  /** Pattern relative to BASE_PATH, e.g. "/recurrent-expenses/:id/edit". */
  path: string;
  // `any` rather than a shared props shape: each route declares whatever
  // props it actually needs (an ":id" param, an optional "search" string),
  // and there is no one type all of them satisfy — the dashboard takes none,
  // Show requires "id". App.svelte spreads matchRoute's params plus `search` at
  // the call site, which is where a real mismatch would surface instead.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- see above
  component: Component<any>;
}

export interface RouteMatch {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- see RouteDef above
  component: Component<any>;
  params: Record<string, string>;
}

// The match table. Phase 2 onward adds entries here as each resource lands.
// Order matters: a literal segment ("archived", "new") must come before the
// ":id" pattern that would otherwise swallow it as a param.
export const routes: RouteDef[] = [
  { path: "/", component: DashboardIndex },
  { path: "/expenses", component: ExpensesIndex },
  { path: "/expenses/new", component: ExpensesNew },
  { path: "/expenses/stats", component: Stats },
  { path: "/expenses/budgets", component: Budgets },
  { path: "/expenses/:id/edit", component: ExpensesEdit },
  { path: "/expenses/:id", component: ExpensesShow },
  { path: "/recurrent-expenses", component: RecurrentExpensesIndex },
  {
    path: "/recurrent-expenses/archived",
    component: RecurrentExpensesArchived,
  },
  { path: "/recurrent-expenses/new", component: RecurrentExpensesNew },
  { path: "/recurrent-expenses/:id/edit", component: RecurrentExpensesEdit },
  { path: "/recurrent-expenses/:id", component: RecurrentExpensesShow },
  { path: "/account", component: AccountIndex },
  { path: "/account/delete-data", component: DeleteDataIndex },
  { path: "/account/exports", component: ExportsIndex },
  { path: "/login", component: LoginIndex },
  { path: "/register", component: RegisterIndex },
];

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

// Segment-boundary prefix test, kept for the Phase 7 BASE_PATH of "": every
// same-origin path is under an empty base, including "/api/...". That is the
// intent — the catch-all serves the shell for everything the API and /static
// do not claim — but it means this can no longer be the check that keeps the
// client router off a non-route link. onLinkClick's `download`, `target` and
// rel="external" tests are what do that now (see routes/exports/Index.svelte).
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
  // Compare against pathname+search, not pathname alone — full can carry a
  // query string (List.svelte's filter/page-size links), and pathname never
  // does, so that comparison could never dedupe a query-only change.
  const current = `${window.location.pathname}${window.location.search}`;
  if (full === current) return;

  window.history.pushState({}, "", full);
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export function onPopState(handler: (path: string) => void): () => void {
  const listener = () => handler(toRoutePath(window.location.pathname));
  window.addEventListener("popstate", listener);

  return () => window.removeEventListener("popstate", listener);
}

// An internal <a> is same-origin, under BASE_PATH, carries no override
// (`target`, `download`, or rel="external"), and was not already handled (a
// modifier click opens a new tab, a prevented default means another handler
// claimed it).
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
  // rel="external" means the server owns this URL: let the browser make a real
  // request so it can act on whatever comes back, a redirect included. The
  // export link needs this — see routes/exports/Index.svelte.
  if (anchor.relList.contains("external")) return false;
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
