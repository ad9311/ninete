import { describe, expect, it } from "vitest";
import {
  centsToInputValue,
  formatCurrency,
  inputValueToCents,
} from "./currency";

describe("formatCurrency", () => {
  it("formats cents with a thousands separator and two decimals", () => {
    expect(formatCurrency(123456)).toBe("$1,234.56");
  });

  it("formats zero", () => {
    expect(formatCurrency(0)).toBe("$0.00");
  });

  it("pads a single cent digit", () => {
    expect(formatCurrency(105)).toBe("$1.05");
  });
});

describe("centsToInputValue", () => {
  it("renders whole and fractional cents", () => {
    expect(centsToInputValue(1999)).toBe("19.99");
    expect(centsToInputValue(100)).toBe("1.00");
  });
});

describe("inputValueToCents", () => {
  it("parses a plain decimal", () => {
    expect(inputValueToCents("19.99")).toBe(1999);
  });

  it("strips thousands separators", () => {
    expect(inputValueToCents("1,234.50")).toBe(123450);
  });

  it("rounds sub-cent input", () => {
    expect(inputValueToCents("1.006")).toBe(101);
    expect(inputValueToCents("1.004")).toBe(100);
  });

  it("returns null for empty input", () => {
    expect(inputValueToCents("")).toBeNull();
    expect(inputValueToCents("   ")).toBeNull();
  });

  it("returns null for non-numeric input", () => {
    expect(inputValueToCents("abc")).toBeNull();
  });

  it("round-trips through centsToInputValue", () => {
    expect(inputValueToCents(centsToInputValue(250000))).toBe(250000);
  });
});
