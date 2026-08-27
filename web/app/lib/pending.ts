// The router-level loading flag (§3.7 of docs/spa-migration.md), replacing
// Turbo's progress bar. Turbo drove its bar from navigation, but every visit it
// covered was an HTTP request, and the SPA's equivalent unit is the same: a
// route that fetches on mount, a form that submits. So the counter lives here
// and `lib/api.ts` drives it from `request()`, which means a new route or a new
// resource gets the feedback without wiring anything — the failure this
// replaces was a flag the router owned that no route ever set.
//
// No runes: §3.9 rule 3 keeps `lib/` component-free, and the tests run under
// the node environment. App.svelte holds the `$state` and subscribes.

// Parity with Turbo.config.drive.progressBarDelay in web/static/js/index.ts.
// Requests that settle inside it show nothing at all, which is the point: a
// warm local fetch would otherwise flash the backdrop on every click.
export const PROGRESS_DELAY_MS = 250;

type Listener = (visible: boolean) => void;

const listeners = new Set<Listener>();
let inFlight = 0;
let visible = false;
let timer: ReturnType<typeof setTimeout> | null = null;

function publish(next: boolean): void {
  if (next === visible) {
    return;
  }

  visible = next;
  for (const listener of listeners) {
    listener(visible);
  }
}

function clearTimer(): void {
  if (timer !== null) {
    clearTimeout(timer);
    timer = null;
  }
}

/** Marks one request in flight. Every call must be paired with `end()`. */
export function begin(): void {
  inFlight += 1;
  if (inFlight === 1 && !visible) {
    clearTimer();
    timer = setTimeout(() => {
      timer = null;
      // Re-checked rather than assumed: the last request can settle between
      // the timer firing and this line only in theory, but `end()` clearing
      // the timer is what normally prevents the flash, and a stale fire would
      // otherwise leave the backdrop up with nothing behind it.
      if (inFlight > 0) {
        publish(true);
      }
    }, PROGRESS_DELAY_MS);
  }
}

/**
 * Marks one request settled, successfully or not. Hiding is immediate — the
 * delay exists to suppress a flash, not to hold the backdrop open.
 */
export function end(): void {
  inFlight = Math.max(inFlight - 1, 0);
  if (inFlight === 0) {
    clearTimer();
    publish(false);
  }
}

/** Subscribes to visibility changes and replays the current value at once. */
export function subscribe(listener: Listener): () => void {
  listeners.add(listener);
  listener(visible);

  return () => {
    listeners.delete(listener);
  };
}

/** Test seam: drops every subscriber and any request the suite left counted. */
export function reset(): void {
  clearTimer();
  listeners.clear();
  inFlight = 0;
  visible = false;
}
