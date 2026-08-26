import { describe, expect, it } from "vitest";
import { joinTagNames, parseTagsInput } from "./tags";

describe("parseTagsInput", () => {
  it("splits on semicolons and trims each entry", () => {
    expect(parseTagsInput("media; Bills ;  groceries")).toEqual([
      "media",
      "Bills",
      "groceries",
    ]);
  });

  it("drops empty segments", () => {
    expect(parseTagsInput("media;;bills;")).toEqual(["media", "bills"]);
  });

  it("returns an empty array for blank input", () => {
    expect(parseTagsInput("")).toEqual([]);
    expect(parseTagsInput("   ")).toEqual([]);
  });
});

describe("joinTagNames", () => {
  it("joins with a semicolon and a space", () => {
    expect(joinTagNames(["media", "bills"])).toBe("media; bills");
  });

  it("returns an empty string for no tags", () => {
    expect(joinTagNames([])).toBe("");
  });
});
