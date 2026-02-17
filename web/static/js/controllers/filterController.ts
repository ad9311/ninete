import { Controller } from "@hotwired/stimulus";
import * as Turbo from "@hotwired/turbo";

export default class extends Controller {
  static targets = ["categoryId", "dateRange"];
  declare readonly categoryIdTarget: HTMLSelectElement;
  declare readonly hasDateRangeTarget: boolean;
  declare readonly dateRangeTarget: HTMLSelectElement;

  apply() {
    const url = new URL(window.location.href);
    const params = url.searchParams;

    params.set("page", "1");

    this.setOrDelete(params, "category_id", this.categoryIdTarget.value);

    if (this.hasDateRangeTarget) {
      this.setOrDelete(params, "date_range", this.dateRangeTarget.value);
    }

    Turbo.visit(url.toString());
  }

  private setOrDelete(params: URLSearchParams, key: string, value: string) {
    if (value) {
      params.set(key, value);
    } else {
      params.delete(key);
    }
  }
}
