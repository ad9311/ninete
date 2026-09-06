// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// The first test of a route component rather than of the toolchain
// (toolchain/Probe.test.ts is that). It covers the parts of the search panel
// that are wholly client-side: the Range/Day mode, which has no query
// parameter of its own — a one-day search is date_from === date_to — so the
// only proof it works is the URL the form navigates to; and the created-date bounds, which are validated and resolved
// to epoch seconds here rather than by the API.
vi.mock("../../lib/api", async () => {
  const actual =
    await vi.importActual<typeof import("../../lib/api")>("../../lib/api");

  return {
    ...actual,
    get: vi.fn(() =>
      Promise.resolve({
        data: [],
        pagination: {
          current_page: 1,
          total_pages: 1,
          per_page: 25,
          total_count: 0,
          has_prev: false,
          has_next: false,
          sort_field: "date",
          sort_order: "DESC",
          category_id: 0,
        },
      }),
    ),
  };
});

vi.mock("../../lib/categories", () => ({
  fetchCategories: vi.fn(() => Promise.resolve([])),
}));

const navigate = vi.fn();
vi.mock("../../router", () => ({
  BASE_PATH: "",
  navigate: (href: string) => navigate(href),
}));

import { get } from "../../lib/api";
import List from "./List.svelte";

afterEach(cleanup);
beforeEach(() => {
  navigate.mockClear();
  vi.mocked(get).mockClear();
});

function modeRadio(label: "Range" | "Day"): HTMLInputElement {
  return screen.getByLabelText(label) as HTMLInputElement;
}

// The accessible name is the whole label's text, so it picks up the visible
// "From"/"To" abbreviation after the sr-only wording: match the prefix.
function fromField(): HTMLInputElement {
  return screen.getByLabelText(/^Created from date/) as HTMLInputElement;
}

// Day mode renames the single remaining field rather than keeping a "From"
// with nothing to pair with.
function dayField(): HTMLInputElement {
  return screen.getByLabelText(/^Created day/) as HTMLInputElement;
}

function toField(): HTMLInputElement {
  return screen.getByLabelText(/^Created to date/) as HTMLInputElement;
}

// Day mode removes the To field outright, so its absence needs a query that
// returns null instead of throwing.
function toFieldOrNull(): HTMLInputElement | null {
  return screen.queryByLabelText(/^Created to date/) as HTMLInputElement | null;
}

// The panel is a <details>, and whether it starts open depends on
// sessionStorage and on whether the URL carries a search. The fields are in the
// DOM either way, so force it open rather than asserting through that state —
// the disclosure is not what these tests are about.
function renderList(search = "") {
  const result = render(List, { search });
  const panel = result.container.querySelector("details");
  if (panel) panel.open = true;

  return result;
}

async function submit(): Promise<void> {
  await fireEvent.submit(screen.getByRole("search"));
}

describe("the Range/Day mode", () => {
  it("starts on Day, with one field and no To", () => {
    renderList();

    expect(modeRadio("Day").checked).toBe(true);
    expect(modeRadio("Range").checked).toBe(false);
    expect(dayField()).toBeTruthy();
    expect(toFieldOrNull()).toBeNull();
  });

  it("brings the To field back when Range is chosen", async () => {
    renderList();

    await fireEvent.input(dayField(), { target: { value: "2026-09-03" } });
    await fireEvent.click(modeRadio("Range"));

    expect(fromField().value).toBe("2026-09-03");
    expect(toField().disabled).toBe(false);
    // The mirrored date comes back with it, as the starting point for widening.
    expect(toField().value).toBe("2026-09-03");
  });

  it("drops the To field again on the way back to Day", async () => {
    renderList();

    await fireEvent.click(modeRadio("Range"));
    await fireEvent.input(fromField(), { target: { value: "2026-09-03" } });
    await fireEvent.click(modeRadio("Day"));

    expect(toFieldOrNull()).toBeNull();
    expect(dayField().value).toBe("2026-09-03");
  });

  it("searches one day, sending the same date as both bounds", async () => {
    renderList();

    await fireEvent.input(dayField(), { target: { value: "2026-09-03" } });
    await submit();

    const href = navigate.mock.calls[0][0] as string;
    const params = new URLSearchParams(href.split("?")[1]);
    expect(params.get("date_from")).toBe("2026-09-03");
    expect(params.get("date_to")).toBe("2026-09-03");
  });

  it("follows the day field, not the date the To field last held", async () => {
    renderList("?date_from=2026-09-03&date_to=2026-09-05");

    await fireEvent.click(modeRadio("Day"));
    await fireEvent.input(dayField(), { target: { value: "2026-09-09" } });
    await submit();

    const href = navigate.mock.calls[0][0] as string;
    const params = new URLSearchParams(href.split("?")[1]);
    expect(params.get("date_from")).toBe("2026-09-09");
    expect(params.get("date_to")).toBe("2026-09-09");
  });

  // Switching to Day mirrors From over To, so a date typed only into To would
  // otherwise be silently discarded along with the field.
  it("folds a To-only date back into the day field", async () => {
    renderList();

    await fireEvent.click(modeRadio("Range"));
    await fireEvent.input(toField(), { target: { value: "2026-09-03" } });
    await fireEvent.click(modeRadio("Day"));

    expect(dayField().value).toBe("2026-09-03");
  });

  // Day is only the default mode. It must not put bounds on a search nobody
  // asked to bound: an untouched panel searches whatever the range select says.
  it("sends no created bounds when the day field is left empty", async () => {
    renderList();

    await submit();

    const href = navigate.mock.calls[0][0] as string;
    const params = new URLSearchParams(href.split("?")[1]);
    expect(params.get("date_from")).toBeNull();
    expect(params.get("date_to")).toBeNull();
  });
});

describe("the mode derived from the URL", () => {
  it("opens on Day when a search bounds one day", () => {
    renderList("?date_from=2026-09-03&date_to=2026-09-03");

    expect(modeRadio("Day").checked).toBe(true);
    expect(toFieldOrNull()).toBeNull();
  });

  it("opens on Range for a real range", () => {
    renderList("?date_from=2026-09-03&date_to=2026-09-05");

    expect(modeRadio("Range").checked).toBe(true);
    expect(toField().disabled).toBe(false);
  });

  // Two empty bounds are equal, which is what makes Day the opening mode.
  it("opens on Day when neither bound is set", () => {
    renderList("?q=coffee");

    expect(modeRadio("Day").checked).toBe(true);
  });

  // A half-filled pair is neither: it opens on Range, where both fields are on
  // screen to be completed, and says so rather than narrowing to one day.
  it("opens on Range when only one bound is set", () => {
    renderList("?date_from=2026-08-01");

    expect(modeRadio("Range").checked).toBe(true);
    expect(toField().value).toBe("");
  });
});

describe("the created-date bounds", () => {
  // Reproductions, not invariant guards: a half-filled pair used to complete
  // itself from the bound that was given, so a From-only search quietly became
  // a one-day search while the panel still showed Range.
  it("rejects a From-only search rather than narrowing it to that day", async () => {
    renderList("?date_from=2026-08-01");

    expect(await screen.findByText(/Fill in both dates/)).toBeTruthy();
    expect(vi.mocked(get)).not.toHaveBeenCalled();
    expect(modeRadio("Range").checked).toBe(true);
  });

  it("rejects a To-only search the same way", async () => {
    renderList("?date_to=2026-08-01");

    expect(await screen.findByText(/Fill in both dates/)).toBeTruthy();
    expect(vi.mocked(get)).not.toHaveBeenCalled();
  });

  // The API answers an inverted pair with a message naming start and end,
  // which are not the fields on screen.
  it("names From and To when the range is inverted, without asking the API", async () => {
    renderList("?date_from=2026-09-05&date_to=2026-09-03");

    expect(
      await screen.findByText(/From date must be on or before the To date/),
    ).toBeTruthy();
    expect(vi.mocked(get)).not.toHaveBeenCalled();
  });

  it("sends both bounds for a valid range", () => {
    renderList("?date_from=2026-09-03&date_to=2026-09-05");

    const params = vi.mocked(get).mock.calls[0][1]?.params as Record<
      string,
      unknown
    >;
    expect(typeof params.created_start).toBe("number");
    expect(typeof params.created_end).toBe("number");
    expect(params.created_start).toBeLessThan(params.created_end as number);
  });
});
