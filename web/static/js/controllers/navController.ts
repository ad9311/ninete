import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static targets = ["dropdown"];
  declare readonly dropdownTarget: HTMLElement;

  private handleOutsideClick = (e: MouseEvent) => {
    if (!this.element.contains(e.target as Node)) {
      this.close();
    }
  };

  toggle() {
    const isOpen = this.dropdownTarget.classList.toggle("open");
    if (isOpen) {
      document.addEventListener("click", this.handleOutsideClick);
    } else {
      document.removeEventListener("click", this.handleOutsideClick);
    }
  }

  close() {
    this.dropdownTarget.classList.remove("open");
    document.removeEventListener("click", this.handleOutsideClick);
  }
}
