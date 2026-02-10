import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static targets = ["local", "value"];
  declare readonly localTarget: HTMLInputElement;
  declare readonly valueTarget: HTMLInputElement;

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
}

function toRFC3339WithOffset(date: Date): string {
  const offset = -date.getTimezoneOffset();
  const sign = offset >= 0 ? "+" : "-";
  const absOffset = Math.abs(offset);
  const hours = String(Math.floor(absOffset / 60)).padStart(2, "0");
  const mins = String(absOffset % 60).padStart(2, "0");

  return date.toISOString().slice(0, 19) + `${sign}${hours}:${mins}`;
}
