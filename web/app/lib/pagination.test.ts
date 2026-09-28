import { describe, expect, it } from "vitest";
import {
  DEFAULT_PER_PAGE,
  PER_PAGE_CHOICES,
  pageRange,
  parsePage,
  parsePerPage,
} from "./pagination";

describe("pageRange", () => {
  it("returns nothing when there are no pages", () => {
    expect(pageRange(0, 1)).toEqual([]);
    expect(pageRange(-1, 1)).toEqual([]);
  });

  it("returns every page when there are fewer than the window", () => {
    expect(pageRange(3, 2)).toEqual([1, 2, 3]);
  });

  it("centres the window on the current page", () => {
    expect(pageRange(20, 10)).toEqual([8, 9, 10, 11, 12]);
  });

  it("clamps the window at both ends", () => {
    expect(pageRange(20, 1)).toEqual([1, 2, 3, 4, 5]);
    expect(pageRange(20, 20)).toEqual([16, 17, 18, 19, 20]);
  });
});

describe("parsePerPage", () => {
  it("accepts an offered value", () => {
    for (const choice of PER_PAGE_CHOICES) {
      expect(parsePerPage(new URLSearchParams(`per_page=${choice}`))).toBe(
        choice,
      );
    }
  });

  it("falls back to the default for anything else", () => {
    expect(parsePerPage(new URLSearchParams())).toBe(DEFAULT_PER_PAGE);
    expect(parsePerPage(new URLSearchParams("per_page=7"))).toBe(
      DEFAULT_PER_PAGE,
    );
    expect(parsePerPage(new URLSearchParams("per_page=all"))).toBe(
      DEFAULT_PER_PAGE,
    );
  });
});

describe("parsePage", () => {
  it("reads a page number", () => {
    expect(parsePage(new URLSearchParams("page=4"))).toBe(4);
  });

  it("clamps anything below one", () => {
    expect(parsePage(new URLSearchParams())).toBe(1);
    expect(parsePage(new URLSearchParams("page=0"))).toBe(1);
    expect(parsePage(new URLSearchParams("page=-3"))).toBe(1);
    expect(parsePage(new URLSearchParams("page=none"))).toBe(1);
  });
});
