import { Controller } from "@hotwired/stimulus";

// Delay before the overlay appears. Most responses on a local SQLite app come
// back well inside this window, so the overlay never flashes for them.
const SHOW_DELAY_MS = 250;

// Safety net: if a request never resolves (dropped connection, a Turbo event we
// do not observe), the overlay hides itself rather than trapping the page.
const MAX_VISIBLE_MS = 15000;

// Announced by the live region while the overlay is up. It is written on show
// and cleared on hide on purpose: a live region announces content *changes*, so
// text that is always present is not reliably read out when the element merely
// becomes visible.
const LABEL = "Loading";

// turbo:visit covers link clicks and programmatic visits; turbo:submit-start
// covers form submissions. Deliberately NOT turbo:before-fetch-request: Turbo
// prefetches links on hover, and that fires before-fetch-request without ever
// producing a matching turbo:load — so the overlay would appear under the
// cursor before the click landed and then swallow it, locking the page.
const SHOW_EVENTS = ["turbo:visit", "turbo:submit-start"];

// Any one of these means the navigation that started the current cycle is over.
// Turbo emits several of them per navigation (submit-end, then before-render,
// render, load); the first one ends the cycle and the rest are no-ops.
//
// turbo:before-cache is NOT in this list — see onBeforeCache below.
const HIDE_EVENTS = [
  "turbo:before-render",
  "turbo:render",
  "turbo:load",
  "turbo:frame-render",
  "turbo:submit-end",
  "turbo:fetch-request-error",
];

// Global loading overlay. Every Turbo navigation — form submits, filter and sort
// links, pagination, plain navigation — emits one of the events above, so one
// document-level listener covers them all. Requests that opt out of Turbo
// (`data-turbo="false"`, such as the export download) fire none of these events
// and are deliberately left alone.
export default class extends Controller<HTMLElement> {
  static targets = ["label"];

  declare readonly labelTarget: HTMLElement;

  // True from the moment a navigation starts until the first event that ends
  // it. Deliberately a flag and not a counter: Turbo serializes visits, and a
  // single navigation fires several hide events, so a reference count cannot
  // survive past the first one anyway.
  private active = false;
  private showTimer?: ReturnType<typeof setTimeout>;
  private hideTimer?: ReturnType<typeof setTimeout>;
  private readonly onShow = () => this.begin();
  private readonly onHide = () => this.end();
  // Turbo caches the current page before leaving it; a visible overlay would
  // otherwise be baked into the snapshot and shown again as the restore
  // preview. This only unmounts the element — it must not end the cycle,
  // because on a visit that has a cached snapshot Turbo fires before-cache at
  // the *start* of the visit (Visit#loadCachedSnapshot), long before the
  // response lands. Ending the cycle there cancelled the pending show timer and
  // left every revisit and back/forward navigation with no overlay at all.
  private readonly onBeforeCache = () => this.unmount();

  connect() {
    SHOW_EVENTS.forEach((name) => document.addEventListener(name, this.onShow));
    HIDE_EVENTS.forEach((name) => document.addEventListener(name, this.onHide));
    document.addEventListener("turbo:before-cache", this.onBeforeCache);
  }

  disconnect() {
    SHOW_EVENTS.forEach((name) =>
      document.removeEventListener(name, this.onShow),
    );
    HIDE_EVENTS.forEach((name) =>
      document.removeEventListener(name, this.onHide),
    );
    document.removeEventListener("turbo:before-cache", this.onBeforeCache);
    this.clearTimers();
  }

  private begin() {
    if (this.active) {
      return;
    }
    this.active = true;
    this.showTimer = setTimeout(() => {
      this.showTimer = undefined;
      this.show();
    }, SHOW_DELAY_MS);
  }

  private end() {
    if (!this.active) {
      return;
    }
    this.active = false;
    this.clearTimers();
    this.unmount();
  }

  private show() {
    this.element.classList.add("is-active");
    this.element.setAttribute("aria-hidden", "false");
    this.labelTarget.textContent = LABEL;
    document.body.setAttribute("aria-busy", "true");
    this.hideTimer = setTimeout(() => {
      this.hideTimer = undefined;
      this.end();
    }, MAX_VISIBLE_MS);
  }

  private unmount() {
    this.element.classList.remove("is-active");
    this.element.setAttribute("aria-hidden", "true");
    this.labelTarget.textContent = "";
    document.body.removeAttribute("aria-busy");
  }

  private clearTimers() {
    if (this.showTimer) {
      clearTimeout(this.showTimer);
      this.showTimer = undefined;
    }
    if (this.hideTimer) {
      clearTimeout(this.hideTimer);
      this.hideTimer = undefined;
    }
  }
}
