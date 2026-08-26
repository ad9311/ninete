// @vitest-environment jsdom
import { cleanup, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App.svelte";
import { BASE_PATH } from "./router";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  document.head.innerHTML = "";
});

function setMeta(name: string, content: string): void {
  const meta = document.createElement("meta");
  meta.setAttribute("name", name);
  meta.setAttribute("content", content);
  document.head.appendChild(meta);
}

function mockSession(body: unknown, status = 200): void {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(JSON.stringify(body), {
        status,
        headers: { "Content-Type": "application/json" },
      }),
    ),
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
  it("renders the chrome, the placeholder route and the signed-in nav", async () => {
    mockSession({ id: 1, username: "ada", email: "ada@example.com" });

    render(App);

    expect(
      screen.getByRole("link", { name: "NINETE" }).getAttribute("href"),
    ).toBe(BASE_PATH);
    expect(screen.getByText("v-test")).toBeTruthy();
    expect(screen.getByText(/SPA is under construction/)).toBeTruthy();

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
    // The placeholder root route did not render alongside it.
    expect(screen.queryByText(/SPA is under construction/)).toBeNull();
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
