import { Controller } from "@hotwired/stimulus";
import * as Turbo from "@hotwired/turbo";

export default class extends Controller {
  static targets = ["select"];
  declare readonly selectTarget: HTMLSelectElement;

  apply() {
    const url = new URL(window.location.href);
    const params = url.searchParams;
    const [field, order] = this.selectTarget.value.split(":");
    params.set("sort_field", field);
    params.set("sort_order", order);
    params.set("page", "1");
    Turbo.visit(url.toString());
  }
}
