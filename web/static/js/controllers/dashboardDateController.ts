import { Controller } from "@hotwired/stimulus";
import * as Turbo from "@hotwired/turbo";

export default class extends Controller {
  connect() {
    const params = new URLSearchParams(window.location.search);

    if (!params.has("date")) {
      params.set("date", localDate());
      Turbo.visit(`/dashboard?${params.toString()}`, { action: "replace" });
    }
  }
}

function localDate(): string {
  const d = new Date();
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}
