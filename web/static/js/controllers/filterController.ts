import { Controller } from "@hotwired/stimulus";
import * as Turbo from "@hotwired/turbo";

const SEARCH_KEYS = ["q", "tag", "date_from", "date_to"];

export default class extends Controller {
  static targets = [
    "categoryId",
    "dateRange",
    "search",
    "tag",
    "dateFrom",
    "dateTo",
  ];
  declare readonly hasCategoryIdTarget: boolean;
  declare readonly categoryIdTarget: HTMLSelectElement;
  declare readonly hasDateRangeTarget: boolean;
  declare readonly dateRangeTarget: HTMLSelectElement;
  declare readonly hasSearchTarget: boolean;
  declare readonly searchTarget: HTMLInputElement;
  declare readonly hasTagTarget: boolean;
  declare readonly tagTarget: HTMLInputElement;
  declare readonly hasDateFromTarget: boolean;
  declare readonly dateFromTarget: HTMLInputElement;
  declare readonly hasDateToTarget: boolean;
  declare readonly dateToTarget: HTMLInputElement;

  apply() {
    const url = new URL(window.location.href);
    const params = url.searchParams;

    params.set("page", "1");

    if (this.hasCategoryIdTarget) {
      this.setOrDelete(params, "category_id", this.categoryIdTarget.value);
    }

    if (this.hasDateRangeTarget) {
      this.setOrDelete(params, "date_range", this.dateRangeTarget.value);
    }

    // The search form is a separate controller instance, so carry its current
    // values over rather than dropping text the user typed but never submitted.
    this.syncSearchInputs(params);

    Turbo.visit(url.toString());
  }

  applySearch(event: Event) {
    event.preventDefault();

    const url = new URL(window.location.href);
    const params = url.searchParams;

    params.set("page", "1");

    if (this.hasSearchTarget) {
      this.setOrDelete(params, "q", this.searchTarget.value.trim());
    }

    if (this.hasTagTarget) {
      this.setOrDelete(params, "tag", this.tagTarget.value.trim());
    }

    if (this.hasDateFromTarget) {
      this.setOrDelete(params, "date_from", this.dateFromTarget.value.trim());
    }

    if (this.hasDateToTarget) {
      this.setOrDelete(params, "date_to", this.dateToTarget.value.trim());
    }

    Turbo.visit(url.toString());
  }

  clearSearch() {
    const url = new URL(window.location.href);
    const params = url.searchParams;
    const hadSearch = SEARCH_KEYS.some((key) => params.has(key));

    params.set("page", "1");
    SEARCH_KEYS.forEach((key) => params.delete(key));

    // An active search forces date_range=all_time server-side, and sort and
    // pagination links bake it into the URL. Dropping it here keeps Clear from
    // leaving behind an unbounded, unfiltered listing.
    if (hadSearch && params.get("date_range") === "all_time") {
      params.delete("date_range");
    }

    Turbo.visit(url.toString());
  }

  private syncSearchInputs(params: URLSearchParams) {
    const form = this.element
      .closest("section")
      ?.querySelector("form.search-bar");
    if (!(form instanceof HTMLFormElement)) {
      return;
    }

    SEARCH_KEYS.forEach((key) => {
      const input = form.elements.namedItem(key);
      if (input instanceof HTMLInputElement) {
        this.setOrDelete(params, key, input.value.trim());
      }
    });
  }

  private setOrDelete(params: URLSearchParams, key: string, value: string) {
    if (value) {
      params.set(key, value);
    } else {
      params.delete(key);
    }
  }
}
