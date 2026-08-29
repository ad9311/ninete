// @vitest-environment jsdom
import type { Component } from "svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  BASE_PATH,
  matchRoute,
  navigate,
  onLinkClick,
  onPopState,
  toRoutePath,
  type RouteDef,
} from "./router";

// Matching does not render, so a stand-in is enough here — the real
// components router.ts's own `routes` table uses are exercised through
// App.test.ts instead.
const stub = {} as Component<Record<string, string>>;

describe("toRoutePath", () => {
  it("strips BASE_PATH and a trailing slash", () => {
    expect(toRoutePath(`${BASE_PATH}/recurrent-expenses/`)).toBe(
      "/recurrent-expenses",
    );
  });

  it("maps the base path itself to the root route", () => {
    expect(toRoutePath(BASE_PATH)).toBe("/");
    expect(toRoutePath(`${BASE_PATH}/`)).toBe("/");
  });

  it("leaves a path with no BASE_PATH prefix untouched", () => {
    expect(toRoutePath("/login")).toBe("/login");
  });
});

describe("matchRoute", () => {
  const routes: RouteDef[] = [
    { path: "/", component: stub },
    { path: "/recurrent-expenses/:id/edit", component: stub },
  ];

  it("matches a static route", () => {
    expect(matchRoute(routes, "/")).toEqual({ component: stub, params: {} });
  });

  it("extracts named params", () => {
    const match = matchRoute(routes, "/recurrent-expenses/42/edit");
    expect(match?.params).toEqual({ id: "42" });
  });

  it("decodes a param", () => {
    const match = matchRoute(routes, "/recurrent-expenses/a%2Fb/edit");
    expect(match?.params.id).toBe("a/b");
  });

  // A malformed escape reaches this from the address bar, and matching runs
  // inside App.svelte's `$derived` — a throw there blanks the whole shell.
  it("falls back to the raw segment for a malformed escape", () => {
    const match = matchRoute(routes, "/recurrent-expenses/%E0%A4%A/edit");
    expect(match?.params.id).toBe("%E0%A4%A");
  });

  it("returns null for an unmatched path", () => {
    expect(matchRoute(routes, "/nope")).toBeNull();
  });

  it("requires the whole path to match, not a prefix of it", () => {
    expect(matchRoute(routes, "/recurrent-expenses/42/edit/extra")).toBeNull();
  });
});

describe("navigate", () => {
  beforeEach(() => {
    window.history.replaceState({}, "", `${BASE_PATH}/`);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("pushes a BASE_PATH-relative path and fires popstate", () => {
    const pushSpy = vi.spyOn(window.history, "pushState");
    const events: string[] = [];
    window.addEventListener("popstate", () => events.push("popstate"));

    navigate("/recurrent-expenses");

    expect(pushSpy).toHaveBeenCalledWith(
      {},
      "",
      `${BASE_PATH}/recurrent-expenses`,
    );
    expect(events).toEqual(["popstate"]);
  });

  it("is a no-op when already at the target path", () => {
    // Reaches the target through navigate() itself rather than a raw
    // history.replaceState call, since that is the codepath already proven
    // to update window.location.pathname by the test above.
    navigate("/recurrent-expenses");
    const pushSpy = vi.spyOn(window.history, "pushState");

    navigate("/recurrent-expenses");

    expect(pushSpy).not.toHaveBeenCalled();
  });
});

describe("onPopState", () => {
  it("reports the current route path on popstate, and stops after unsubscribing", () => {
    window.history.replaceState({}, "", `${BASE_PATH}/account`);
    const seen: string[] = [];
    const off = onPopState((path) => seen.push(path));

    window.dispatchEvent(new PopStateEvent("popstate"));
    expect(seen).toEqual(["/account"]);

    off();
    window.dispatchEvent(new PopStateEvent("popstate"));
    expect(seen).toEqual(["/account"]);
  });
});

describe("onLinkClick", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  function clickAnchor(href: string, attrs: Record<string, string> = {}): void {
    const anchor = document.createElement("a");
    anchor.href = href;
    for (const [key, value] of Object.entries(attrs)) {
      anchor.setAttribute(key, value);
    }
    document.body.appendChild(anchor);
    anchor.dispatchEvent(
      new MouseEvent("click", { bubbles: true, button: 0, cancelable: true }),
    );
  }

  it("intercepts an internal link under BASE_PATH", () => {
    window.history.replaceState({}, "", `${BASE_PATH}/`);
    const seen: string[] = [];
    const off = onLinkClick((path) => seen.push(path));

    clickAnchor(`${BASE_PATH}/account`);

    expect(seen).toEqual(["/account"]);
    expect(window.location.pathname).toBe(`${BASE_PATH}/account`);
    off();
  });

  // Genuine reproduction of the export-download bug's second half: with the
  // opt-out gone the client router swallows the link, so the browser never
  // makes the request that would either download the file or follow the
  // redirect to the login page.
  it('ignores a link with rel="external"', () => {
    const seen: string[] = [];
    const off = onLinkClick((path) => seen.push(path));

    clickAnchor(`${BASE_PATH}/exports/expenses.json`, { rel: "external" });

    expect(seen).toEqual([]);
    off();
  });

  it("ignores a link with target=_blank", () => {
    const seen: string[] = [];
    const off = onLinkClick((path) => seen.push(path));

    clickAnchor(`${BASE_PATH}/account`, { target: "_blank" });

    expect(seen).toEqual([]);
    off();
  });

  it("ignores an in-page fragment link", () => {
    window.history.replaceState({}, "", `${BASE_PATH}/account`);
    const seen: string[] = [];
    const off = onLinkClick((path) => seen.push(path));

    clickAnchor(`${BASE_PATH}/account#details`);

    expect(seen).toEqual([]);
    off();
  });

  it("ignores a download link", () => {
    const seen: string[] = [];
    const off = onLinkClick((path) => seen.push(path));

    clickAnchor(`${BASE_PATH}/account`, { download: "" });

    expect(seen).toEqual([]);
    off();
  });
});
