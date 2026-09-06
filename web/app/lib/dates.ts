// Every date the app renders passes through here. Read §3.6 of
// docs/spa-migration.md before changing anything in this file: the rules it
// encodes are the ones a mistake hides from review, from the default TZ=UTC
// test run, and from everybody except the user.
//
// The app stores two kinds of value, both as epoch seconds in an INTEGER
// column, so neither the type system nor the driver can tell them apart:
//
//   Calendar date — expenses.date, recurrent_expenses.last_copy_created_at.
//     Epoch seconds at UTC midnight of that date. "the 21st", a label on a
//     calendar. It has no time and no zone, so it is formatted with UTC getters
//     in every zone.
//
//     A billed date is a calendar date the app only ever *shows* as a month:
//     the expense form picks `YYYY-MM` and stores day 01, and formatMonthUTC
//     renders "Sep 2026". The stored value is still a full calendar date —
//     rows written before the month picker carry a real day — so the month
//     helpers read the day rather than assuming it is 01.
//   Instant — created_at, updated_at. A real moment, formatted with local
//     getters because it means something different to a viewer elsewhere.
//
// The split below is the whole point of the file: formatDateUTC is for the
// first kind, formatDate and formatDateTime for the second. Formatting a
// calendar date with local getters shows the previous day to every user west
// of UTC, and nothing in the code looks wrong.
//
// The same split decides how a `YYYY-MM-DD` the user typed is turned into a
// query bound. Bounding a calendar-date column means UTC midnight
// (calendarDateToUnix); bounding an *instant* column means midnight where the
// viewer is (localDayStart / localDayEnd), because "the 3rd" is a span of real
// time and which span depends on the zone. Using the UTC pair for an instant
// is the production bug those two exist to prevent: an expense created at
// 20:00 on the 3rd in UTC-5 is 01:00 on the 4th in UTC, so a search for the
// 3rd dropped a row the listing itself labelled "Sep 3".

const MONTHS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

const CALENDAR_DATE_RE = /^(\d{4})-(\d{2})-(\d{2})$/;
const CALENDAR_MONTH_RE = /^(\d{4})-(\d{2})$/;
const SECONDS_PER_DAY = 86400;

/**
 * Epoch seconds, or an already-built Date. Seconds rather than milliseconds:
 * that is what the database stores, what the API sends, and the only unit this
 * module accepts as a number.
 */
export type UnixOrDate = number | Date;

function toDate(value: UnixOrDate): Date {
  const date = value instanceof Date ? value : new Date(value * 1000);

  // Without this every formatter answers an undefined value with the string
  // "undefined NaN, NaN" and unixToCalendarDate with "0NaN-NaN-NaN", both of
  // which render into the page as if they were dates. A missing or non-numeric
  // field is a bug in the caller; surface it the way calendarDateToUnix
  // surfaces an impossible date.
  if (Number.isNaN(date.getTime())) {
    throw new RangeError(`not a date: ${String(value)}`);
  }

  return date;
}

/**
 * Formats a calendar date — `Aug 22, 2026`. UTC getters only: the value is a
 * label, not a moment, and a local getter would move it a day.
 */
export function formatDateUTC(value: UnixOrDate): string {
  const date = toDate(value);

  return `${MONTHS[date.getUTCMonth()]} ${date.getUTCDate()}, ${date.getUTCFullYear()}`;
}

/**
 * Formats a calendar date as its month — `Aug 2026`. UTC getters for the same
 * reason as formatDateUTC: a local getter would move a UTC-midnight value back
 * a day for every viewer west of UTC, which on the 1st of a month is also a
 * different month.
 */
export function formatMonthUTC(value: UnixOrDate): string {
  const date = toDate(value);

  return `${MONTHS[date.getUTCMonth()]} ${date.getUTCFullYear()}`;
}

/**
 * Formats an instant's visible text in the viewer's zone — `Aug 22, 2026`.
 * Never call this on a calendar date.
 */
export function formatDate(value: UnixOrDate): string {
  const date = toDate(value);

  return `${MONTHS[date.getMonth()]} ${date.getDate()}, ${date.getFullYear()}`;
}

/**
 * Formats an instant in full — `Aug 22, 2026 1:05:09 PM`, local. This is the
 * `title` tooltip an instant carries beside its formatDate text; the pair comes
 * from localDateController and both halves must survive the port.
 */
export function formatDateTime(value: UnixOrDate): string {
  const date = toDate(value);
  const hours = date.getHours();
  const period = hours >= 12 ? "PM" : "AM";
  const hours12 = hours % 12 || 12;
  const minutes = String(date.getMinutes()).padStart(2, "0");
  const seconds = String(date.getSeconds()).padStart(2, "0");

  return `${formatDate(date)} ${hours12}:${minutes}:${seconds} ${period}`;
}

/**
 * Converts a stored calendar date to the `YYYY-MM-DD` string client state
 * should hold it as (§3.6 rule 2). A string cannot be shifted by a zone by
 * accident; an epoch can.
 */
export function unixToCalendarDate(value: UnixOrDate): string {
  const date = toDate(value);
  const year = String(date.getUTCFullYear()).padStart(4, "0");
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  const day = String(date.getUTCDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}

/**
 * Converts a `YYYY-MM-DD` calendar date to the epoch seconds the API stores —
 * UTC midnight of that date, in every zone.
 *
 * The string is parsed by hand rather than with `new Date(str)` on purpose:
 * `new Date("2026-08-22")` is UTC midnight but `new Date("2026-08-22T00:00:00")`
 * is *local* midnight, two nearly identical strings a day apart in output
 * (§3.6 rule 4). Nothing here should depend on remembering which is which.
 */
export function calendarDateToUnix(calendarDate: string): number {
  const parts = CALENDAR_DATE_RE.exec(calendarDate);
  if (!parts) {
    throw new RangeError(`not a YYYY-MM-DD calendar date: ${calendarDate}`);
  }

  const year = Number(parts[1]);
  const month = Number(parts[2]);
  const day = Number(parts[3]);

  // Date.UTC maps a year of 0-99 to 1900+year, so Date.UTC(26, ...) is 1926 and
  // the round-trip below would reject "0026-08-22" as impossible. setUTCFullYear
  // takes the year literally.
  const date = new Date(0);
  date.setUTCFullYear(year, month - 1, day);
  const unix = date.getTime() / 1000;

  // Date rolls a nonsense date over instead of rejecting it: month 13 becomes
  // January of the next year, February 31st becomes March. Round-trip
  // the result so an impossible date is an error rather than a silently
  // different one.
  if (unixToCalendarDate(unix) !== calendarDate) {
    throw new RangeError(`not a real calendar date: ${calendarDate}`);
  }

  return unix;
}

/**
 * The epoch seconds at which a `YYYY-MM-DD` calendar day begins *where the
 * viewer is* — the inclusive start bound for filtering an instant column.
 *
 * Local getters are correct here for the same reason they are wrong in
 * calendarDateToUnix: the value being compared against is a real moment, not a
 * label. The server never learns the zone, matching how named ranges already
 * work (§3.6, "Retiring tz_offset on the API side").
 */
export function localDayStart(calendarDate: string): number {
  return localMidnight(calendarDate, 0);
}

/**
 * The epoch seconds at which a `YYYY-MM-DD` calendar day *ends* locally, i.e.
 * the start of the next local day — the exclusive end bound.
 *
 * Built by stepping the calendar day rather than adding 86400 seconds, so a
 * DST transition inside the day does not leave the bound an hour short or an
 * hour long.
 */
export function localDayEnd(calendarDate: string): number {
  return localMidnight(calendarDate, 1);
}

function localMidnight(calendarDate: string, dayOffset: number): number {
  const parts = CALENDAR_DATE_RE.exec(calendarDate);
  if (!parts) {
    throw new RangeError(`not a YYYY-MM-DD calendar date: ${calendarDate}`);
  }

  const year = Number(parts[1]);
  const month = Number(parts[2]);
  const day = Number(parts[3]);

  // Seeded from a fixed, unambiguous local date rather than new Date(0): the
  // epoch is a *UTC* instant, so in a zone behind UTC its local fields are
  // already the previous year, and setFullYear on 12-31 would then carry.
  const date = new Date(2000, 0, 1, 0, 0, 0, 0);
  date.setFullYear(year, month - 1, day);
  // setHours after the date fields, not in the constructor: a DST transition
  // between the seed date and the target date otherwise leaves the clock an
  // hour off midnight. In the rare zone where local midnight does not exist,
  // this lands on the first instant the day does have, which is the correct
  // start of that day.
  date.setHours(0, 0, 0, 0);

  // Date rolls a nonsense date over instead of rejecting it (month 13 becomes
  // January, February 31st becomes March), so verify before stepping the day.
  if (
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day
  ) {
    throw new RangeError(`not a real calendar date: ${calendarDate}`);
  }

  if (dayOffset !== 0) {
    date.setDate(date.getDate() + dayOffset);
    date.setHours(0, 0, 0, 0);
  }

  return Math.floor(date.getTime() / 1000);
}

/**
 * Converts a stored calendar date to the `YYYY-MM` string the billed-month
 * input holds it as. The day is discarded, which is the whole point: a row
 * written before the month picker carries a real day, and seeding the input
 * with it would be a value `<input type="month">` refuses to display.
 */
export function unixToCalendarMonth(value: UnixOrDate): string {
  const date = toDate(value);
  const year = String(date.getUTCFullYear()).padStart(4, "0");
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");

  return `${year}-${month}`;
}

/**
 * Converts a `YYYY-MM` billed month to the epoch seconds the API stores — UTC
 * midnight of the 1st, in every zone. Day 01 is a convention, not a meaning:
 * nothing reads the day back, and `date` stays a calendar-date column so the
 * preset month ranges keep working unchanged.
 */
export function calendarMonthToUnix(calendarMonth: string): number {
  const parts = CALENDAR_MONTH_RE.exec(calendarMonth);
  if (!parts) {
    throw new RangeError(`not a YYYY-MM calendar month: ${calendarMonth}`);
  }

  // Delegated so the year-0-99 trap and the rollover round-trip are handled in
  // one place; month 13 is rejected there rather than sliding into January.
  return calendarDateToUnix(`${parts[1]}-${parts[2]}-01`);
}

/**
 * The calendar month the viewer's own clock is showing, as `YYYY-MM` — what
 * the billed-month input defaults to for a new expense. Local getters for the
 * same reason as todayCalendarDate: the question is which month it is where
 * the user is.
 */
export function todayCalendarMonth(now: Date = new Date()): string {
  return todayCalendarDate(now).slice(0, 7);
}

/**
 * The calendar date the viewer's own clock is showing, as `YYYY-MM-DD`. Local
 * getters are correct here and only here: the question is what day it is where
 * the user is, which is exactly what a `<input type="date">` defaults to.
 */
export function todayCalendarDate(now: Date = new Date()): string {
  const year = String(now.getFullYear()).padStart(4, "0");
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}

/**
 * Moves a calendar date by whole days, staying in `YYYY-MM-DD`. Safe across
 * DST because the arithmetic happens in UTC, where every day is 86400 seconds.
 */
export function addDays(calendarDate: string, days: number): string {
  return unixToCalendarDate(
    calendarDateToUnix(calendarDate) + days * SECONDS_PER_DAY,
  );
}
