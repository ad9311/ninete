import { Controller } from "@hotwired/stimulus";

// Toggles a small tap-triggered popover listing accepted quick-add date
// formats. Closes on outside click or Escape.
export default class extends Controller {
  static targets = ["toggle", "panel"];

  declare readonly toggleTarget: HTMLButtonElement;
  declare readonly panelTarget: HTMLElement;

  private readonly onDocClick = (event: Event) => {
    if (!this.element.contains(event.target as Node)) {
      this.close();
    }
  };

  private readonly onKey = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      this.close();
    }
  };

  toggle(event: Event) {
    event.preventDefault();
    // Stop the click from reaching the document listener we add on open.
    event.stopPropagation();
    if (this.panelTarget.hidden) {
      this.open();
    } else {
      this.close();
    }
  }

  disconnect() {
    this.close();
  }

  private open() {
    this.panelTarget.hidden = false;
    this.toggleTarget.setAttribute("aria-expanded", "true");
    document.addEventListener("click", this.onDocClick);
    document.addEventListener("keydown", this.onKey);
  }

  private close() {
    this.panelTarget.hidden = true;
    this.toggleTarget.setAttribute("aria-expanded", "false");
    document.removeEventListener("click", this.onDocClick);
    document.removeEventListener("keydown", this.onKey);
  }
}
