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
//   Instant — created_at, updated_at. A real moment, formatted with local
//     getters because it means something different to a viewer elsewhere.
//
// The split below is the whole point of the file: formatDateUTC is for the
// first kind, formatDate and formatDateTime for the second. Formatting a
// calendar date with local getters shows the previous day to every user west
// of UTC, and nothing in the code looks wrong.

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
const SECONDS_PER_DAY = 86400;

/**
 * Epoch seconds, or an already-built Date. Seconds rather than milliseconds:
 * that is what the database stores, what the API sends, and the only unit this
 * module accepts as a number.
 */
export type UnixOrDate = number | Date;

function toDate(value: UnixOrDate): Date {
  return value instanceof Date ? value : new Date(value * 1000);
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
  const unix = Date.UTC(year, month - 1, day) / 1000;

  // Date.UTC rolls a nonsense date over instead of rejecting it: month 13
  // becomes January of the next year, February 31st becomes March. Round-trip
  // the result so an impossible date is an error rather than a silently
  // different one.
  if (unixToCalendarDate(unix) !== calendarDate) {
    throw new RangeError(`not a real calendar date: ${calendarDate}`);
  }

  return unix;
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
