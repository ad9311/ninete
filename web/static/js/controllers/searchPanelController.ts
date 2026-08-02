import { Controller } from "@hotwired/stimulus";

const STORAGE_KEY = "search-panel-open";

// Keeps the search <details> open across Turbo visits. The server only renders
// it open when a search is active, so without this any control inside the panel
// that triggers a visit — the date-field toggle in particular — would collapse
// the panel out from under the user.
export default class extends Controller<HTMLDetailsElement> {
  connect() {
    if (this.read() === "true") {
      this.element.open = true;
    }
  }

  toggle() {
    this.write(String(this.element.open));
  }

  private read(): string | null {
    try {
      return window.sessionStorage.getItem(this.storageKey());
    } catch {
      return null;
    }
  }

  private write(value: string) {
    try {
      window.sessionStorage.setItem(this.storageKey(), value);
    } catch {
      // Storage can be unavailable (private mode, blocked cookies); the panel
      // simply falls back to the server-rendered state.
    }
  }

  // Scope the key per listing so one page's panel does not open another's.
  private storageKey(): string {
    return `${STORAGE_KEY}:${window.location.pathname}`;
  }
}
