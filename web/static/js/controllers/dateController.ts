import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static targets = ["local", "value"];
  declare readonly localTarget: HTMLInputElement;
  declare readonly valueTarget: HTMLInputElement;

  connect() {
    this.hydrateLocalDate();
  }

  prepare() {
    const localDate = this.localTarget.value;

    if (!localDate) {
      this.valueTarget.value = "";
      return;
    }

    // Send as UTC midnight to preserve the calendar date regardless of timezone
    this.valueTarget.value = localDate + "T00:00:00Z";
  }

  private hydrateLocalDate() {
    const setCurrentLocalDate = () => {
      this.localTarget.value = toLocalDate(new Date());
    };

    if (this.localTarget.value) {
      return;
    }

    const rawValue = this.valueTarget.value?.trim();
    if (!rawValue) {
      setCurrentLocalDate();
      return;
    }

    const unix = Number(rawValue);
    if (!Number.isNaN(unix)) {
      if (unix <= 0) {
        setCurrentLocalDate();
        return;
      }

      const date = new Date(unix * 1000);
      if (!Number.isNaN(date.getTime())) {
        this.localTarget.value = toUTCDate(date);

        return;
      }

      setCurrentLocalDate();

      return;
    }

    const parsed = new Date(rawValue);
    if (!Number.isNaN(parsed.getTime())) {
      this.localTarget.value = toUTCDate(parsed);

      return;
    }

    setCurrentLocalDate();
  }
}

// Returns the local calendar date as YYYY-MM-DD (used for defaulting to today)
function toLocalDate(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}

// Returns the UTC calendar date as YYYY-MM-DD (used when hydrating a stored UTC timestamp)
function toUTCDate(date: Date): string {
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  const day = String(date.getUTCDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}
