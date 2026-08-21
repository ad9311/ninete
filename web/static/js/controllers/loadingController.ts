import { Controller } from "@hotwired/stimulus";

// Delay before the overlay appears. Most responses on a local SQLite app come
// back well inside this window, so the overlay never flashes for them.
const SHOW_DELAY_MS = 250;

// Safety net: if a request never resolves (dropped connection, a Turbo event we
// do not observe), the overlay hides itself rather than trapping the page.
const MAX_VISIBLE_MS = 15000;

// turbo:visit covers link clicks and programmatic visits; turbo:submit-start
// covers form submissions. Deliberately NOT turbo:before-fetch-request: Turbo
// prefetches links on hover, and that fires before-fetch-request without ever
// producing a matching turbo:load — so the overlay would appear under the
// cursor before the click landed and then swallow it, locking the page.
const SHOW_EVENTS = ["turbo:visit", "turbo:submit-start"];

const HIDE_EVENTS = [
  "turbo:before-render",
  "turbo:render",
  "turbo:load",
  "turbo:frame-render",
  "turbo:submit-end",
  "turbo:fetch-request-error",
  // Turbo caches the current page before leaving it; a visible overlay would
  // otherwise be baked into the snapshot and shown again as the restore preview.
  "turbo:before-cache",
];

// Global loading overlay. Every Turbo navigation — form submits, filter and sort
// links, pagination, plain navigation — emits one of the events below, so one
// document-level listener covers them all. Requests that opt out of Turbo
// (`data-turbo="false"`, such as the export download) fire none of these events
// and are deliberately left alone.
export default class extends Controller<HTMLElement> {
  private pending = 0;
  private showTimer?: ReturnType<typeof setTimeout>;
  private hideTimer?: ReturnType<typeof setTimeout>;
  private readonly onShow = () => this.begin();
  private readonly onHide = () => this.end();

  connect() {
    SHOW_EVENTS.forEach((name) => document.addEventListener(name, this.onShow));
    HIDE_EVENTS.forEach((name) => document.addEventListener(name, this.onHide));
  }

  disconnect() {
    SHOW_EVENTS.forEach((name) =>
      document.removeEventListener(name, this.onShow),
    );
    HIDE_EVENTS.forEach((name) =>
      document.removeEventListener(name, this.onHide),
    );
    this.clearTimers();
  }

  private begin() {
    this.pending += 1;
    if (this.showTimer || this.isVisible()) {
      return;
    }
    this.showTimer = setTimeout(() => {
      this.showTimer = undefined;
      this.show();
    }, SHOW_DELAY_MS);
  }

  private end() {
    // A single request produces several hide events (submit-end, before-render,
    // render, load), so clamp instead of letting the counter go negative.
    this.pending = Math.max(0, this.pending - 1);
    if (this.pending > 0) {
      return;
    }
    this.clearTimers();
    this.hide();
  }

  private show() {
    this.element.classList.add("is-active");
    this.element.setAttribute("aria-hidden", "false");
    document.body.setAttribute("aria-busy", "true");
    this.hideTimer = setTimeout(() => {
      this.pending = 0;
      this.hideTimer = undefined;
      this.hide();
    }, MAX_VISIBLE_MS);
  }

  private hide() {
    this.element.classList.remove("is-active");
    this.element.setAttribute("aria-hidden", "true");
    document.body.removeAttribute("aria-busy");
  }

  private isVisible(): boolean {
    return this.element.classList.contains("is-active");
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
