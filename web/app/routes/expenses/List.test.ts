// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// The first test of a route component rather than of the toolchain
// (toolchain/Probe.test.ts is that). It covers the parts of the search panel
// that are wholly client-side: the "Single day" box, which has no query
// parameter of its own, so the only proof it works is the URL the form
// navigates to; and the created-date bounds, which are validated and resolved
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

function singleDayBox(): HTMLInputElement {
  return screen.getByLabelText("Single day") as HTMLInputElement;
}

// The accessible name is the whole label's text, so it picks up the visible
// "From"/"To" abbreviation after the sr-only wording: match the prefix.
function fromField(): HTMLInputElement {
  return screen.getByLabelText(/^Created from date/) as HTMLInputElement;
}

function toField(): HTMLInputElement {
  return screen.getByLabelText(/^Created to date/) as HTMLInputElement;
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

describe("the single-day search box", () => {
  it("is unchecked and leaves the To field editable by default", () => {
    renderList();

    expect(singleDayBox().checked).toBe(false);
    expect(toField().disabled).toBe(false);
  });

  it("disables the To field and mirrors From into it when checked", async () => {
    renderList();

    await fireEvent.input(fromField(), { target: { value: "2026-09-03" } });
    await fireEvent.click(singleDayBox());

    expect(toField().disabled).toBe(true);
    expect(toField().value).toBe("2026-09-03");
  });

  it("keeps mirroring while From changes", async () => {
    renderList();

    await fireEvent.click(singleDayBox());
    await fireEvent.input(fromField(), { target: { value: "2026-09-04" } });

    expect(toField().value).toBe("2026-09-04");
  });

  it("searches one day, sending the same date as both bounds", async () => {
    renderList();

    await fireEvent.input(fromField(), { target: { value: "2026-09-03" } });
    await fireEvent.click(singleDayBox());
    await submit();

    const href = navigate.mock.calls[0][0] as string;
    const params = new URLSearchParams(href.split("?")[1]);
    expect(params.get("date_from")).toBe("2026-09-03");
    expect(params.get("date_to")).toBe("2026-09-03");
  });

  // Checking the box mirrors From over To, so a date typed only into To would
  // otherwise be silently discarded.
  it("folds a To-only date back into From when checked", async () => {
    renderList();

    await fireEvent.input(toField(), { target: { value: "2026-09-03" } });
    await fireEvent.click(singleDayBox());

    expect(fromField().value).toBe("2026-09-03");
    expect(toField().value).toBe("2026-09-03");
  });

  it("leaves the mirrored date in place when unchecked, ready to widen", async () => {
    renderList();

    await fireEvent.input(fromField(), { target: { value: "2026-09-03" } });
    await fireEvent.click(singleDayBox());
    await fireEvent.click(singleDayBox());

    expect(toField().disabled).toBe(false);
    expect(toField().value).toBe("2026-09-03");
  });
});

describe("the single-day box derived from the URL", () => {
  it("is checked when a search bounds one day", () => {
    renderList("?date_from=2026-09-03&date_to=2026-09-03");

    expect(singleDayBox().checked).toBe(true);
    expect(toField().disabled).toBe(true);
  });

  it("is unchecked for a real range", () => {
    renderList("?date_from=2026-09-03&date_to=2026-09-05");

    expect(singleDayBox().checked).toBe(false);
    expect(toField().disabled).toBe(false);
  });

  // Both empty are also "equal", which must not read as a one-day search.
  it("is unchecked when neither bound is set", () => {
    renderList("?q=coffee");

    expect(singleDayBox().checked).toBe(false);
  });
});

describe("the created-date bounds", () => {
  // Reproductions, not invariant guards: a half-filled pair used to complete
  // itself from the bound that was given, so a From-only search quietly became
  // a one-day search while the box above sat unchecked.
  it("rejects a From-only search rather than narrowing it to that day", async () => {
    renderList("?date_from=2026-08-01");

    expect(await screen.findByText(/Fill in both dates/)).toBeTruthy();
    expect(vi.mocked(get)).not.toHaveBeenCalled();
    expect(singleDayBox().checked).toBe(false);
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
