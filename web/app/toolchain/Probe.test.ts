// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it } from "vitest";

import Probe from "./Probe.svelte";

// A canary for the test toolchain rather than a test of the app: it fails when
// vitest can no longer compile a `.svelte` file, when it resolves Svelte's
// server build instead of the client one, or when the jsdom opt-in stops
// working. Phase 1 writes the first real component against this setup, so the
// breakage should surface here and not there. See docs/spa-migration.md §5,
// Phase 0.6.
//
// Cleanup is explicit because vitest runs without `globals`, and
// @testing-library/svelte only registers its own afterEach hook when a global
// one exists.
afterEach(cleanup);

describe("the component toolchain", () => {
  it("compiles a component, mounts it and re-renders on state change", async () => {
    render(Probe, { label: "clicks" });

    const button = screen.getByRole("button");
    expect(button.textContent).toBe("clicks: 0");

    await fireEvent.click(button);
    expect(button.textContent).toBe("clicks: 1");
  });
});
