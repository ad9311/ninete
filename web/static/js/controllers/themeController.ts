import { Controller } from "@hotwired/stimulus";

type Theme = "light" | "dark" | "auto";

const STORAGE_KEY = "theme";

export default class extends Controller {
  static targets = ["select"];
  declare readonly selectTarget: HTMLSelectElement;

  private media = window.matchMedia("(prefers-color-scheme: dark)");

  private handleSystemChange = () => {
    if (this.preference() === "auto") {
      this.apply("auto");
    }
  };

  connect() {
    this.selectTarget.value = this.preference();
    this.apply(this.preference());
    this.media.addEventListener("change", this.handleSystemChange);
  }

  disconnect() {
    this.media.removeEventListener("change", this.handleSystemChange);
  }

  change() {
    const value = this.selectTarget.value as Theme;
    localStorage.setItem(STORAGE_KEY, value);
    this.apply(value);
  }

  private preference(): Theme {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "light" || stored === "dark" || stored === "auto") {
      return stored;
    }
    return "auto";
  }

  private apply(theme: Theme) {
    const dark = theme === "dark" || (theme === "auto" && this.media.matches);
    document.documentElement.className = dark ? "theme-dark" : "theme-light";
  }
}
