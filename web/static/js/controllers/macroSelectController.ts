import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static targets = [
    "checkbox",
    "kcalTotal",
    "proteinTotal",
    "carbsTotal",
    "fatTotal",
  ];

  declare readonly checkboxTargets: HTMLInputElement[];
  declare readonly kcalTotalTarget: HTMLElement;
  declare readonly proteinTotalTarget: HTMLElement;
  declare readonly carbsTotalTarget: HTMLElement;
  declare readonly fatTotalTarget: HTMLElement;

  update() {
    const checked = this.checkboxTargets.filter((cb) => cb.checked);

    if (checked.length === 0) {
      this.kcalTotalTarget.textContent =
        this.kcalTotalTarget.dataset.base ?? "";
      this.proteinTotalTarget.textContent =
        this.proteinTotalTarget.dataset.base ?? "";
      this.carbsTotalTarget.textContent =
        this.carbsTotalTarget.dataset.base ?? "";
      this.fatTotalTarget.textContent = this.fatTotalTarget.dataset.base ?? "";
      return;
    }

    const sum = (key: string) =>
      checked.reduce((acc, cb) => acc + parseFloat(cb.dataset[key] ?? "0"), 0);

    this.kcalTotalTarget.textContent = round(sum("kcal")).toString();
    this.proteinTotalTarget.textContent = round(sum("protein")).toString();
    this.carbsTotalTarget.textContent = round(sum("carbs")).toString();
    this.fatTotalTarget.textContent = round(sum("fat")).toString();
  }
}

function round(n: number): number {
  return Math.round(n * 100) / 100;
}
