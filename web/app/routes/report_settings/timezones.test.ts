import { afterEach, describe, expect, it, vi } from "vitest";
import { browserTimezone, timezoneOptions } from "./timezones";

type IntlWithSupported = typeof Intl & {
  supportedValuesOf?: (key: string) => string[];
};

const intl = Intl as IntlWithSupported;

afterEach(() => {
  vi.restoreAllMocks();
});

describe("timezoneOptions", () => {
  it("returns the browser's own IANA list", () => {
    vi.spyOn(intl, "supportedValuesOf").mockReturnValue([
      "America/Bogota",
      "UTC",
    ]);

    expect(timezoneOptions("UTC")).toEqual(["America/Bogota", "UTC"]);
  });

  it("keeps a saved zone the runtime does not list", () => {
    // A zone dropped from the browser's copy of the database, or one saved
    // from another device, must still appear — otherwise the select silently
    // rewrites the user's setting on the next save.
    vi.spyOn(intl, "supportedValuesOf").mockReturnValue(["UTC"]);

    expect(timezoneOptions("America/Bogota")).toEqual([
      "America/Bogota",
      "UTC",
    ]);
  });

  it("falls back to UTC when the runtime has no zone list", () => {
    vi.spyOn(intl, "supportedValuesOf").mockImplementation(() => {
      throw new Error("unsupported");
    });

    expect(timezoneOptions("")).toEqual(["UTC"]);
  });
});

describe("browserTimezone", () => {
  it("reads the resolved zone", () => {
    expect(browserTimezone()).toBe(
      Intl.DateTimeFormat().resolvedOptions().timeZone,
    );
  });
});
