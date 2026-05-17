import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static values = {
    baseKcal: Number,
    baseProtein: Number,
    baseCarbs: Number,
    baseFat: Number,
    baseAmount: Number,
    foodId: Number,
  };

  static targets = ["amount", "kcal", "protein", "carbs", "fat", "useLink"];

  declare readonly baseKcalValue: number;
  declare readonly baseProteinValue: number;
  declare readonly baseCarbsValue: number;
  declare readonly baseFatValue: number;
  declare readonly baseAmountValue: number;
  declare readonly foodIdValue: number;

  declare readonly amountTarget: HTMLInputElement;
  declare readonly kcalTarget: HTMLElement;
  declare readonly proteinTarget: HTMLElement;
  declare readonly carbsTarget: HTMLElement;
  declare readonly fatTarget: HTMLElement;
  declare readonly hasUseLinkTarget: boolean;
  declare readonly useLinkTarget: HTMLAnchorElement;

  connect() {
    this.calculate();
  }

  calculate() {
    const actual = parseFloat(this.amountTarget.value) || 0;
    const base = this.baseAmountValue;
    if (base <= 0) return;

    const scale = actual / base;
    const kcal = round(this.baseKcalValue * scale);
    const protein = round(this.baseProteinValue * scale);
    const carbs = round(this.baseCarbsValue * scale);
    const fat = round(this.baseFatValue * scale);

    setValue(this.kcalTarget, kcal);
    setValue(this.proteinTarget, protein);
    setValue(this.carbsTarget, carbs);
    setValue(this.fatTarget, fat);

    if (this.hasUseLinkTarget) {
      const params = new URLSearchParams({
        from_food: this.foodIdValue.toString(),
        amount: actual.toString(),
      });
      this.useLinkTarget.href = `/macros/new?${params.toString()}`;
    }
  }

  use(event: Event) {
    event.preventDefault();
    if (this.hasUseLinkTarget) {
      window.Turbo.visit(this.useLinkTarget.href);
    }
  }
}

function round(n: number): number {
  return Math.round(n * 100) / 100;
}

function setValue(el: HTMLElement, val: number) {
  if (el instanceof HTMLInputElement) {
    el.value = val.toString();
  } else {
    el.textContent = val.toString();
  }
}
