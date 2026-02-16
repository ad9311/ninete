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

    const parsed = new Date(localDate);
    if (Number.isNaN(parsed.getTime())) {
      this.valueTarget.value = "";
      return;
    }

    this.valueTarget.value = toRFC3339WithOffset(parsed);
  }

  private hydrateLocalDate() {
    const setCurrentLocalDate = () => {
      this.localTarget.value = toLocalDateTimeInput(new Date());
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
        this.localTarget.value = toLocalDateTimeInput(date);

        return;
      }

      setCurrentLocalDate();

      return;
    }

    const parsed = new Date(rawValue);
    if (!Number.isNaN(parsed.getTime())) {
      this.localTarget.value = toLocalDateTimeInput(parsed);

      return;
    }

    setCurrentLocalDate();
  }
}

function toRFC3339WithOffset(date: Date): string {
  const offset = -date.getTimezoneOffset();
  const sign = offset >= 0 ? "+" : "-";
  const absOffset = Math.abs(offset);
  const hours = String(Math.floor(absOffset / 60)).padStart(2, "0");
  const mins = String(absOffset % 60).padStart(2, "0");

  return date.toISOString().slice(0, 19) + `${sign}${hours}:${mins}`;
}

function toLocalDateTimeInput(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  const seconds = String(date.getSeconds()).padStart(2, "0");

  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}`;
}
