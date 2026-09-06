import { describe, expect, it } from "vitest";
import {
  addDays,
  calendarDateToUnix,
  calendarMonthToUnix,
  formatDate,
  formatDateTime,
  formatDateUTC,
  formatMonthUTC,
  lastCalendarMonth,
  localDayEnd,
  localDayStart,
  todayCalendarDate,
  todayCalendarMonth,
  unixToCalendarDate,
  unixToCalendarMonth,
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

// Bounding an instant column by a day the user typed. These run in both
// configured zones and assert with *local* getters, so they say the same thing
// east and west of UTC — which is the point: the server is never told a zone,
// and a fixed server-side zone would be wrong for every request made from
// anywhere else (docs/spa-migration.md §3.6).
describe("localDayStart and localDayEnd", () => {
  // Reads a bound back as the local wall clock it lands on.
  function local(unix: number) {
    const d = new Date(unix * 1000);

    return {
      date: `${String(d.getFullYear()).padStart(4, "0")}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`,
      hours: d.getHours(),
      minutes: d.getMinutes(),
      seconds: d.getSeconds(),
    };
  }

  it.each([
    "2026-01-01", // year start
    "2026-02-28", // short month end
    "2028-02-29", // leap day
    "2026-08-31", // month end
    "2026-09-03", // the day from the production report
    "2026-12-31", // year end
  ])("starts %s at local midnight", (date) => {
    expect(local(localDayStart(date))).toEqual({
      date,
      hours: 0,
      minutes: 0,
      seconds: 0,
    });
  });

  it.each([
    { day: "2026-01-01", next: "2026-01-02" },
    { day: "2026-02-28", next: "2026-03-01" }, // non-leap month end
    { day: "2028-02-28", next: "2028-02-29" }, // leap year
    { day: "2026-08-31", next: "2026-09-01" },
    { day: "2026-12-31", next: "2027-01-01" }, // year boundary
  ])("ends $day at the start of $next", ({ day, next }) => {
    expect(local(localDayEnd(day))).toEqual({
      date: next,
      hours: 0,
      minutes: 0,
      seconds: 0,
    });
  });

  // The reported bug, as a test: an expense entered at 20:00 on the 3rd is
  // 01:00 on the 4th in UTC, so UTC-midnight bounds excluded it while the
  // listing still labelled it "Sep 3".
  it("contains an instant late on its own local day", () => {
    const start = localDayStart("2026-09-03");
    const end = localDayEnd("2026-09-03");
    const evening = new Date(2026, 8, 3, 20, 0, 0).getTime() / 1000;

    expect(evening).toBeGreaterThanOrEqual(start);
    expect(evening).toBeLessThan(end);
  });

  it("excludes the first instant of the next local day", () => {
    const end = localDayEnd("2026-09-03");
    const nextMidnight = new Date(2026, 8, 4, 0, 0, 0).getTime() / 1000;

    expect(end).toBe(nextMidnight);
  });

  // A DST day is 23 or 25 hours long. Adding a flat 86400 seconds would put
  // the end bound an hour off the next midnight, hiding or over-including an
  // hour of results — which is why localDayEnd steps the calendar day.
  it.each([
    "2026-03-08", // US spring forward
    "2026-11-01", // US fall back
    "2026-04-05", // NZ DST ends
    "2026-09-27", // NZ DST starts
  ])("spans exactly one calendar day across the transition at %s", (date) => {
    const span = localDayEnd(date) - localDayStart(date);

    // Whatever the zone does on this date, the bound lands on midnight.
    expect(local(localDayEnd(date)).hours).toBe(0);
    expect([23, 24, 25].map((h) => h * 3600)).toContain(span);
  });

  it.each(["2026-13-01", "2026-02-31", "2026-9-3", "2026-09", "", "nope"])(
    "rejects %s",
    (value) => {
      expect(() => localDayStart(value)).toThrow(RangeError);
      expect(() => localDayEnd(value)).toThrow(RangeError);
    },
  );

  it("takes a two-digit year literally rather than as 19xx", () => {
    expect(local(localDayStart("0026-08-22")).date).toBe("0026-08-22");
  });
});

// The billed date is stored as a calendar date and shown as a month, so every
// helper below has the same failure mode as formatDateUTC: a local getter moves
// a UTC-midnight value back a day, and on the 1st that is the previous month.
describe("formatMonthUTC", () => {
  it.each([
    { date: "2026-01-01", text: "Jan 2026" }, // year start, month start
    { date: "2026-03-08", text: "Mar 2026" }, // US spring forward
    { date: "2026-08-31", text: "Aug 2026" }, // month end
    { date: "2026-09-27", text: "Sep 2026" }, // NZ DST starts
    { date: "2026-12-31", text: "Dec 2026" }, // year end
  ])("renders $date as $text", ({ date, text }) => {
    expect(formatMonthUTC(calendarDateToUnix(date))).toBe(text);
  });

  it("does not shift the 1st of a month into the previous one", () => {
    // The failing case for a local getter: west of UTC this reads as the last
    // day of August, which is also a different month.
    expect(formatMonthUTC(utcMidnight(2026, 9, 1))).toBe("Sep 2026");
  });

  it("throws on a value that is not a date", () => {
    expect(() => formatMonthUTC(undefined as unknown as number)).toThrow(
      RangeError,
    );
  });
});

describe("unixToCalendarMonth", () => {
  it("drops the day a pre-month-picker row still carries", () => {
    expect(unixToCalendarMonth(utcMidnight(2026, 9, 3))).toBe("2026-09");
  });

  it.each([
    { date: "2026-01-01", month: "2026-01" },
    { date: "2026-08-31", month: "2026-08" },
    { date: "2026-12-31", month: "2026-12" },
  ])("reads $date as $month", ({ date, month }) => {
    expect(unixToCalendarMonth(calendarDateToUnix(date))).toBe(month);
  });
});

describe("calendarMonthToUnix", () => {
  it("returns UTC midnight of the 1st", () => {
    expect(calendarMonthToUnix("2026-09")).toBe(utcMidnight(2026, 9, 1));
  });

  it("round-trips through unixToCalendarMonth", () => {
    expect(unixToCalendarMonth(calendarMonthToUnix("2026-09"))).toBe("2026-09");
  });

  it("takes a two-digit year literally rather than as 19xx", () => {
    expect(unixToCalendarDate(calendarMonthToUnix("0026-08"))).toBe(
      "0026-08-01",
    );
  });

  it.each(["2026-13", "2026-00", "2026-9", "2026", "2026-09-01", "", "nope"])(
    "rejects %s",
    (value) => {
      expect(() => calendarMonthToUnix(value)).toThrow(RangeError);
    },
  );
});

describe("todayCalendarMonth", () => {
  it("reads the viewer's own calendar month", () => {
    // Late on the last day of the month locally: a UTC getter east of the
    // viewer would already have rolled into the next month.
    const now = new Date(2026, 7, 31, 23, 30, 0);

    expect(todayCalendarMonth(now)).toBe("2026-08");
  });

  it("returns a well-formed month for the real clock", () => {
    expect(todayCalendarMonth()).toMatch(/^\d{4}-\d{2}$/);
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

describe("lastCalendarMonth", () => {
  it("steps back one month", () => {
    expect(lastCalendarMonth(new Date(2026, 8, 6, 12))).toBe("2026-08");
  });

  it("rolls back over the year boundary", () => {
    expect(lastCalendarMonth(new Date(2026, 0, 1, 12))).toBe("2025-12");
  });

  it("reads the local month, not the UTC one", () => {
    // Both suite zones are far enough from UTC that a local noon and a UTC
    // instant can land in different months at a boundary; noon keeps the two
    // in step, which is what makes the assertion above meaningful.
    expect(lastCalendarMonth(new Date(2026, 2, 31, 12))).toBe("2026-02");
  });
});
