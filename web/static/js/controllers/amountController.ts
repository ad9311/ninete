import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static targets = ["local", "value"];
  declare readonly localTarget: HTMLInputElement;
  declare readonly valueTarget: HTMLInputElement;

  connect() {
    this.hydrateLocalAmount();
  }

  sync() {
    this.valueTarget.value = toCentsString(this.localTarget.value);
  }

  prepare() {
    this.sync();
  }

  private hydrateLocalAmount() {
    if (this.localTarget.value) {
      this.sync();
      return;
    }

    const rawValue = this.valueTarget.value?.trim();
    if (!rawValue) {
      return;
    }

    const cents = Number(rawValue);
    if (!Number.isFinite(cents)) {
      return;
    }

    this.localTarget.value = (cents / 100).toFixed(2);
  }
}

function toCentsString(localAmount: string): string {
  const normalized = localAmount.replaceAll(",", "").trim();
  if (!normalized) {
    return "";
  }

  const parsed = Number(normalized);
  if (!Number.isFinite(parsed)) {
    return "";
  }

  return String(Math.round(parsed * 100));
}
