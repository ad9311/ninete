import { describe, expect, it } from "vitest";
import {
  addDays,
  calendarDateToUnix,
  formatDate,
  formatDateTime,
  formatDateUTC,
  todayCalendarDate,
  unixToCalendarDate,
} from "./dates";

// A calendar date the app would store: UTC midnight, epoch seconds.
function utcMidnight(year: number, month: number, day: number): number {
  return Date.UTC(year, month - 1, day) / 1000;
}

// Dates chosen to break a wrong implementation rather than to be
// representative: the ends of a month, both DST transitions of each zone the
// suite runs in, and a year boundary (§3.6).
const CALENDAR_CASES = [
  { date: "2026-01-01", text: "Jan 1, 2026" }, // year start
  { date: "2026-02-28", text: "Feb 28, 2026" }, // short month end
  { date: "2028-02-29", text: "Feb 29, 2028" }, // leap day
  { date: "2026-03-08", text: "Mar 8, 2026" }, // US spring forward
  { date: "2026-04-05", text: "Apr 5, 2026" }, // NZ DST ends
  { date: "2026-08-01", text: "Aug 1, 2026" }, // month start
  { date: "2026-08-31", text: "Aug 31, 2026" }, // month end
  { date: "2026-09-27", text: "Sep 27, 2026" }, // NZ DST starts
  { date: "2026-11-01", text: "Nov 1, 2026" }, // US fall back
  { date: "2026-12-31", text: "Dec 31, 2026" }, // year end
];

describe("the test zone", () => {
  // If this fails, vitest.config.mts stopped applying TZ and every test below
  // has quietly become a no-op: under UTC they all pass whether the formatters
  // use UTC getters or local ones. That is the entire reason for the config.
  it("is not UTC", () => {
    expect(new Date().getTimezoneOffset()).not.toBe(0);
  });
});

describe("formatDateUTC", () => {
  it.each(CALENDAR_CASES)("renders $date as $text", ({ date, text }) => {
    expect(formatDateUTC(calendarDateToUnix(date))).toBe(text);
  });

  it("does not shift a date stored at UTC midnight", () => {
    // The failing case for a local getter: in any zone ahead of UTC this reads
    // as the 22nd, in any zone behind it as the 21st.
    expect(formatDateUTC(utcMidnight(2026, 8, 22))).toBe("Aug 22, 2026");
  });

  it("accepts a Date as well as epoch seconds", () => {
    const unix = utcMidnight(2026, 8, 22);

    expect(formatDateUTC(new Date(unix * 1000))).toBe(formatDateUTC(unix));
  });
});

describe("a value that is not a date", () => {
  // Every one of these rendered as "undefined NaN, NaN" or "0NaN-NaN-NaN"
  // before the guard — strings that reach the page looking like content.
  it.each([
    { name: "NaN", value: NaN },
    { name: "an invalid Date", value: new Date("nope") },
  ])("is rejected rather than formatted: $name", ({ value }) => {
    expect(() => formatDateUTC(value)).toThrow(RangeError);
    expect(() => formatDate(value)).toThrow(RangeError);
    expect(() => formatDateTime(value)).toThrow(RangeError);
    expect(() => unixToCalendarDate(value)).toThrow(RangeError);
  });

  it("still accepts the epoch itself", () => {
    expect(unixToCalendarDate(0)).toBe("1970-01-01");
  });
});

describe("formatDate", () => {
  // Built from local components and read back with local getters, so the
  // expectation holds in every zone — which is the property an instant has and
  // a calendar date does not.
  it("renders an instant in the viewer's zone", () => {
    expect(formatDate(new Date(2026, 7, 22, 13, 5, 9))).toBe("Aug 22, 2026");
  });

  it("differs from formatDateUTC for an instant near local midnight", () => {
    // Which side of midnight lands on a different UTC day depends on the sign
    // of the zone: west of UTC it is late evening, east of it early morning.
    // Picking the wrong one makes this assertion pass by accident.
    const westOfUTC = new Date(2026, 7, 22).getTimezoneOffset() > 0;
    const nearMidnight = westOfUTC
      ? new Date(2026, 7, 22, 23, 30, 0)
      : new Date(2026, 7, 22, 0, 30, 0);

    expect(formatDate(nearMidnight)).toBe("Aug 22, 2026");
    expect(formatDateUTC(nearMidnight)).not.toBe(formatDate(nearMidnight));
  });
});

describe("formatDateTime", () => {
  it.each([
    { hour: 0, minute: 0, second: 0, text: "Aug 22, 2026 12:00:00 AM" },
    { hour: 9, minute: 7, second: 4, text: "Aug 22, 2026 9:07:04 AM" },
    { hour: 12, minute: 0, second: 0, text: "Aug 22, 2026 12:00:00 PM" },
    { hour: 13, minute: 5, second: 9, text: "Aug 22, 2026 1:05:09 PM" },
    { hour: 23, minute: 59, second: 59, text: "Aug 22, 2026 11:59:59 PM" },
  ])(
    "renders $hour:$minute:$second as $text",
    ({ hour, minute, second, text }) => {
      expect(formatDateTime(new Date(2026, 7, 22, hour, minute, second))).toBe(
        text,
      );
    },
  );
});

describe("calendarDateToUnix", () => {
  it.each(CALENDAR_CASES)("puts $date at UTC midnight", ({ date }) => {
    const unix = calendarDateToUnix(date);

    expect(unix % 86400).toBe(0);
    expect(new Date(unix * 1000).toISOString()).toBe(`${date}T00:00:00.000Z`);
  });

  it("rejects a date that is not YYYY-MM-DD", () => {
    expect(() => calendarDateToUnix("22-08-2026")).toThrow(RangeError);
    expect(() => calendarDateToUnix("2026-8-22")).toThrow(RangeError);
    expect(() => calendarDateToUnix("2026-08-22T00:00:00Z")).toThrow(
      RangeError,
    );
    expect(() => calendarDateToUnix("")).toThrow(RangeError);
  });

  it("accepts a year below 100 rather than reading it as 19xx", () => {
    // Date.UTC(26, 7, 22) is 1926, which would fail the round-trip check and
    // reject a well-formed date as impossible.
    expect(
      new Date(calendarDateToUnix("0026-08-22") * 1000).toISOString(),
    ).toBe("0026-08-22T00:00:00.000Z");
  });

  it("rejects a date that does not exist instead of rolling it over", () => {
    // Date.UTC would answer these with March 3rd, January 1st of 2027 and
    // October 1st respectively.
    expect(() => calendarDateToUnix("2026-02-31")).toThrow(RangeError);
    expect(() => calendarDateToUnix("2026-13-01")).toThrow(RangeError);
    expect(() => calendarDateToUnix("2026-09-31")).toThrow(RangeError);
  });
});

describe("unixToCalendarDate", () => {
  it.each(CALENDAR_CASES)("round-trips $date", ({ date }) => {
    expect(unixToCalendarDate(calendarDateToUnix(date))).toBe(date);
  });

  it("reads the UTC day of an instant, not the local one", () => {
    // 23:30 UTC on the 21st. A local implementation answers the 22nd in
    // Auckland and the 21st in Los Angeles; the stored value means the 21st.
    expect(unixToCalendarDate(Date.UTC(2026, 7, 21, 23, 30) / 1000)).toBe(
      "2026-08-21",
    );
  });
});

describe("todayCalendarDate", () => {
  it("reads the viewer's own calendar day", () => {
    const now = new Date(2026, 7, 22, 23, 30, 0);

    expect(todayCalendarDate(now)).toBe("2026-08-22");
  });

  it("returns a well-formed date for the real clock", () => {
    expect(todayCalendarDate()).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});

describe("addDays", () => {
  it.each([
    { from: "2026-08-31", days: 1, to: "2026-09-01" },
    { from: "2026-09-01", days: -1, to: "2026-08-31" },
    { from: "2026-12-31", days: 1, to: "2027-01-01" },
    { from: "2028-02-28", days: 1, to: "2028-02-29" },
    { from: "2026-02-28", days: 1, to: "2026-03-01" },
    { from: "2026-08-22", days: 0, to: "2026-08-22" },
  ])("moves $from by $days days to $to", ({ from, days, to }) => {
    expect(addDays(from, days)).toBe(to);
  });

  // The round-trip addDays(addDays(d, 1), -1) === d holds for a local-getter
  // implementation too, so it proves nothing. Name the day on each side of the
  // transition instead: a local implementation lands on the transition day
  // itself, or repeats it.
  it.each([
    { before: "2026-03-07", day: "2026-03-08", after: "2026-03-09" }, // US spring forward
    { before: "2026-04-04", day: "2026-04-05", after: "2026-04-06" }, // NZ DST ends
    { before: "2026-09-26", day: "2026-09-27", after: "2026-09-28" }, // NZ DST starts
    { before: "2026-10-31", day: "2026-11-01", after: "2026-11-02" }, // US fall back
  ])(
    "crosses the DST transition at $day without losing a day",
    ({ before, day, after }) => {
      expect(addDays(before, 1)).toBe(day);
      expect(addDays(day, 1)).toBe(after);
      expect(addDays(after, -1)).toBe(day);
      expect(addDays(day, -1)).toBe(before);
    },
  );
});
