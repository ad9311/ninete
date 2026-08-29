// @vitest-environment jsdom
import { cleanup, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App.svelte";
import { PROGRESS_DELAY_MS, reset as resetPending } from "./lib/pending";
import { BASE_PATH } from "./router";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  resetPending();
  document.head.innerHTML = "";
});

function setMeta(name: string, content: string): void {
  const meta = document.createElement("meta");
  meta.setAttribute("name", name);
  meta.setAttribute("content", content);
  document.head.appendChild(meta);
}

// What GetAPIDashboard answers for the root route, which renders as
// "$123.45" and "+23% vs last month ($100.00)".
const DASHBOARD_BODY = {
  data: {
    this_month_total: 12345,
    last_month_total: 10000,
    month_change_sign: "+",
    month_change_pct: 23,
    top_categories: [
      { name: "Rent", total: 9000 },
      { name: "Food", total: 3345 },
    ],
  },
};

// A fresh Response per call, dispatched on the URL. A Response body can be read
// only once, and since Phase 4 the root route fetches /api/dashboard alongside
// Header.svelte's /api/session — one shared instance leaves whichever request
// parses second throwing on a consumed body, which the component swallows into
// its error branch.
function mockSession(body: unknown, status = 200): void {
  vi.stubGlobal(
    "fetch",
    vi.fn((input: RequestInfo | URL) => {
      const isDashboard = String(input).startsWith("/api/dashboard");
      const payload = isDashboard ? DASHBOARD_BODY : body;

      return Promise.resolve(
        new Response(JSON.stringify(payload), {
          status: isDashboard ? 200 : status,
          headers: { "Content-Type": "application/json" },
        }),
      );
    }),
  );
}

beforeEach(() => {
  setMeta("version", "v-test");
  setMeta("csrf-token", "test-token");
  window.history.replaceState({}, "", `${BASE_PATH}/`);

  // jsdom has no matchMedia implementation; Header.svelte's theme handling
  // reads it unconditionally on mount.
  vi.stubGlobal(
    "matchMedia",
    vi.fn().mockReturnValue({
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }),
  );
});

describe("App", () => {
  it("renders the chrome, the dashboard route and the signed-in nav", async () => {
    mockSession({ id: 1, username: "ada", email: "ada@example.com" });

    render(App);

    expect(
      screen.getByRole("link", { name: "NINETE" }).getAttribute("href"),
    ).toBe("/");
    expect(screen.getByText("v-test")).toBeTruthy();
    expect(screen.getByText("This month's spending")).toBeTruthy();
    // The heading alone renders in the error branch too, so assert on the
    // fetched figures — otherwise a broken /api/dashboard call passes here.
    expect(await screen.findByText("$123.45")).toBeTruthy();
    expect(screen.getByText(/\+23% vs last month/)).toBeTruthy();

    // GetAPISession resolves asynchronously (Header.svelte's $effect); the
    // username-carrying nav only appears once that promise settles.
    expect(await screen.findByText(/ada/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "Logout" })).toBeTruthy();
  });

  it("shows a not-found state for a route missing from the match table", async () => {
    window.history.replaceState({}, "", `${BASE_PATH}/does-not-exist`);
    mockSession({ id: 1, username: "ada", email: "ada@example.com" });

    render(App);

    expect(screen.getByText("Not found.")).toBeTruthy();
    // The dashboard root route did not render alongside it.
    expect(screen.queryByText("This month's spending")).toBeNull();
  });

  // Regression: Phase 1 left `pending` as a router-owned $state nothing ever
  // assigned, so the backdrop never rendered once — including for the four
  // Phase 2 routes that fetch on mount. Driving it from lib/api.ts is what
  // makes this pass; it fails if App.svelte stops subscribing.
  it("covers the page while a request is in flight", async () => {
    let settle: (value: Response) => void = () => {};
    const inFlight = new Promise<Response>((resolve) => {
      settle = resolve;
    });
    vi.stubGlobal("fetch", vi.fn().mockReturnValue(inFlight));

    const { container } = render(App);

    expect(container.querySelector(".route-progress-bar")).toBeNull();

    // Header.svelte's session fetch is still open. Outlast the anti-flash
    // delay and the backdrop has to appear without any route touching a flag.
    await new Promise((resolve) => setTimeout(resolve, PROGRESS_DELAY_MS + 10));

    expect(container.querySelector(".route-progress-bar")).toBeTruthy();

    settle(
      new Response(JSON.stringify({ id: 1, username: "ada", email: "a@b.c" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    await screen.findByText(/ada/);

    expect(container.querySelector(".route-progress-bar")).toBeNull();
  });

  it("leaves the nav out when the session request fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("network error")),
    );

    render(App);

    // Let the rejected promise's .catch in Header.svelte run before asserting
    // the negative — there is no element to await appearing here.
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(screen.queryByRole("button", { name: "Logout" })).toBeNull();
  });
});
