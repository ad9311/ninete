import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { PROGRESS_DELAY_MS, begin, end, reset, subscribe } from "./pending";

function record(): boolean[] {
  const seen: boolean[] = [];
  subscribe((visible) => seen.push(visible));

  return seen;
}

describe("pending", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    reset();
  });

  afterEach(() => {
    reset();
    vi.useRealTimers();
  });

  it("replays the current value to a new subscriber", () => {
    expect(record()).toEqual([false]);
  });

  it("shows nothing for a request that settles inside the delay", () => {
    const seen = record();

    begin();
    vi.advanceTimersByTime(PROGRESS_DELAY_MS - 1);
    end();
    vi.advanceTimersByTime(PROGRESS_DELAY_MS);

    expect(seen).toEqual([false]);
  });

  it("shows the backdrop once a request outlives the delay", () => {
    const seen = record();

    begin();
    vi.advanceTimersByTime(PROGRESS_DELAY_MS);

    expect(seen).toEqual([false, true]);

    end();

    expect(seen).toEqual([false, true, false]);
  });

  it("stays up until the last of several overlapping requests settles", () => {
    const seen = record();

    begin();
    begin();
    vi.advanceTimersByTime(PROGRESS_DELAY_MS);
    end();

    expect(seen).toEqual([false, true]);

    end();

    expect(seen).toEqual([false, true, false]);
  });

  it("re-arms the delay after everything settles", () => {
    const seen = record();

    begin();
    vi.advanceTimersByTime(PROGRESS_DELAY_MS);
    end();
    begin();

    expect(seen).toEqual([false, true, false]);

    vi.advanceTimersByTime(PROGRESS_DELAY_MS);

    expect(seen).toEqual([false, true, false, true]);
  });

  it("stops notifying an unsubscribed listener", () => {
    const seen: boolean[] = [];
    const off = subscribe((visible) => seen.push(visible));
    off();

    begin();
    vi.advanceTimersByTime(PROGRESS_DELAY_MS);

    expect(seen).toEqual([false]);
  });
});
