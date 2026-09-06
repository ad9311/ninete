// The timezone list the settings select offers.
//
// `Intl.supportedValuesOf("timeZone")` is the browser's own copy of the IANA
// database, so there is no hand-maintained list to drift from the zone data Go
// resolves the name against on the other side. It is not in every runtime the
// suite runs on — and TypeScript's lib does not declare it below ES2022 — so
// the lookup is guarded and falls back to the value already saved plus UTC,
// which keeps the select from silently dropping the user's own zone.

const FALLBACK_ZONES = ["UTC"];

export function timezoneOptions(current: string): string[] {
  const supported = supportedTimezones();
  const zones = supported.length > 0 ? supported : FALLBACK_ZONES;

  if (current !== "" && !zones.includes(current)) {
    return [current, ...zones];
  }

  return zones;
}

function supportedTimezones(): string[] {
  const intl = Intl as typeof Intl & {
    supportedValuesOf?: (key: string) => string[];
  };

  if (typeof intl.supportedValuesOf !== "function") {
    return [];
  }

  try {
    return intl.supportedValuesOf("timeZone");
  } catch {
    return [];
  }
}

/** The browser's own zone, seeding the form before anything has been saved. */
export function browserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}
