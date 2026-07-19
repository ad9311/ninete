import { Controller } from "@hotwired/stimulus";

const STORAGE_KEY = "expense-quick-mode";

export default class extends Controller {
  static targets = ["regular", "quick", "switch"];

  declare readonly regularTarget: HTMLElement;
  declare readonly quickTarget: HTMLElement;
  declare readonly switchTarget: HTMLInputElement;

  connect() {
    const force = this.element.getAttribute("data-quick-force") === "1";
    const on = force || this.stored();
    this.switchTarget.checked = on;
    this.apply(on);
  }

  toggle() {
    const on = this.switchTarget.checked;
    localStorage.setItem(STORAGE_KEY, on ? "1" : "0");
    this.apply(on);
  }

  private stored(): boolean {
    return localStorage.getItem(STORAGE_KEY) === "1";
  }

  private apply(on: boolean) {
    this.regularTarget.hidden = on;
    this.quickTarget.hidden = !on;
  }
}
