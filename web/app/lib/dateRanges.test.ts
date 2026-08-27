import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { computeDateRange } from "./dateRanges";

function utcMidnight(year: number, month: number, day: number): number {
  return Date.UTC(year, month - 1, day) / 1000;
}

// setSystemTime takes local wall-clock components under the vitest-pinned TZ,
// same as `new Date(year, month, day, ...)` — not epoch millis — so "now" is
// exactly what the browser's local getters would report at that moment.
function freezeLocal(
  year: number,
  month: number,
  day: number,
  hour = 12,
): void {
  vi.setSystemTime(new Date(year, month - 1, day, hour, 0, 0, 0));
}

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("computeDateRange", () => {
  it("returns null for all_time and an unknown key", () => {
    freezeLocal(2026, 8, 15);

    expect(computeDateRange("all_time")).toBeNull();
    expect(computeDateRange("not_a_range")).toBeNull();
  });

  describe("this_month", () => {
    it.each([
      {
        pinned: [2026, 8, 15] as const,
        start: [2026, 8, 1] as const,
        end: [2026, 9, 1] as const,
      },
      // The 1st of the month: the boundary itself must still resolve to the
      // current month, not the previous one.
      {
        pinned: [2026, 8, 1] as const,
        start: [2026, 8, 1] as const,
        end: [2026, 9, 1] as const,
      },
      // The last day of the month.
      {
        pinned: [2026, 8, 31] as const,
        start: [2026, 8, 1] as const,
        end: [2026, 9, 1] as const,
      },
      // A December pin must not carry the month into next year.
      {
        pinned: [2026, 12, 15] as const,
        start: [2026, 12, 1] as const,
        end: [2027, 1, 1] as const,
      },
    ])(
      "pinned at $pinned resolves to [$start, $end)",
      ({ pinned, start, end }) => {
        freezeLocal(pinned[0], pinned[1], pinned[2]);
        expect(computeDateRange("this_month")).toEqual({
          start: utcMidnight(start[0], start[1], start[2]),
          end: utcMidnight(end[0], end[1], end[2]),
        });
      },
    );
  });

  describe("next_month", () => {
    it("rolls over into the next year from December", () => {
      freezeLocal(2026, 12, 15);
      expect(computeDateRange("next_month")).toEqual({
        start: utcMidnight(2027, 1, 1),
        end: utcMidnight(2027, 2, 1),
      });
    });

    it("is the plain following month otherwise", () => {
      freezeLocal(2026, 8, 15);
      expect(computeDateRange("next_month")).toEqual({
        start: utcMidnight(2026, 9, 1),
        end: utcMidnight(2026, 10, 1),
      });
    });
  });

  describe("last_month", () => {
    it("rolls back into the previous year from January", () => {
      freezeLocal(2026, 1, 15);
      expect(computeDateRange("last_month")).toEqual({
        start: utcMidnight(2025, 12, 1),
        end: utcMidnight(2026, 1, 1),
      });
    });

    it("is the plain preceding month otherwise", () => {
      freezeLocal(2026, 8, 15);
      expect(computeDateRange("last_month")).toEqual({
        start: utcMidnight(2026, 7, 1),
        end: utcMidnight(2026, 8, 1),
      });
    });
  });

  describe("this_week", () => {
    // 2026-08-17 is a Monday; the week runs Mon 17 through the following Mon.
    it.each([
      { pinned: [2026, 8, 17] as const, label: "Monday" },
      { pinned: [2026, 8, 20] as const, label: "Thursday" },
      // Sunday is the end of the same Mon-started week, not the start of a new one.
      { pinned: [2026, 8, 23] as const, label: "Sunday" },
    ])(
      "pinned on $label stays in the week starting Mon Aug 17",
      ({ pinned }) => {
        freezeLocal(pinned[0], pinned[1], pinned[2]);
        expect(computeDateRange("this_week")).toEqual({
          start: utcMidnight(2026, 8, 17),
          end: utcMidnight(2026, 8, 24),
        });
      },
    );

    it("crosses a month boundary correctly", () => {
      // 2026-08-31 is a Monday.
      freezeLocal(2026, 8, 31);
      expect(computeDateRange("this_week")).toEqual({
        start: utcMidnight(2026, 8, 31),
        end: utcMidnight(2026, 9, 7),
      });
    });
  });

  describe("six_months", () => {
    it("spans five months back plus the current one", () => {
      freezeLocal(2026, 8, 15);
      expect(computeDateRange("six_months")).toEqual({
        start: utcMidnight(2026, 3, 1),
        end: utcMidnight(2026, 9, 1),
      });
    });

    it("crosses a year boundary", () => {
      freezeLocal(2026, 2, 15);
      expect(computeDateRange("six_months")).toEqual({
        start: utcMidnight(2025, 9, 1),
        end: utcMidnight(2026, 3, 1),
      });
    });
  });

  describe("this_year", () => {
    it("spans the calendar year", () => {
      freezeLocal(2026, 8, 15);
      expect(computeDateRange("this_year")).toEqual({
        start: utcMidnight(2026, 1, 1),
        end: utcMidnight(2027, 1, 1),
      });
    });

    it("still resolves to the current year on Dec 31", () => {
      freezeLocal(2026, 12, 31);
      expect(computeDateRange("this_year")).toEqual({
        start: utcMidnight(2026, 1, 1),
        end: utcMidnight(2027, 1, 1),
      });
    });
  });

  describe("a pin near local midnight", () => {
    // The case §3.6 names explicitly: the user's local calendar day and UTC's
    // disagree. West of UTC, local evening is already the next UTC day; east
    // of it, local early morning is still the previous UTC day. Either way,
    // the range must follow the *local* calendar day, not UTC's.
    it("resolves this_month by the local day, not the UTC day", () => {
      const westOfUTC = new Date(2026, 7, 31).getTimezoneOffset() > 0;
      if (westOfUTC) {
        freezeLocal(2026, 8, 31, 23);
      } else {
        freezeLocal(2026, 9, 1, 0);
      }

      expect(computeDateRange("this_month")).toEqual({
        start: utcMidnight(2026, westOfUTC ? 8 : 9, 1),
        end: utcMidnight(2026, westOfUTC ? 9 : 10, 1),
      });
    });
  });
});
