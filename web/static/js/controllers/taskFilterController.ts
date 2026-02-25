import { Controller } from "@hotwired/stimulus";
import * as Turbo from "@hotwired/turbo";

export default class extends Controller {
  static targets = ["done", "priority"];
  declare readonly doneTarget: HTMLSelectElement;
  declare readonly priorityTarget: HTMLSelectElement;

  apply() {
    const url = new URL(window.location.href);
    const params = url.searchParams;

    params.set("page", "1");

    this.setOrDelete(params, "done", this.doneTarget.value);
    this.setOrDelete(params, "priority", this.priorityTarget.value);

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
