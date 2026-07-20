import { Controller } from "@hotwired/stimulus";

// Submits the containing form when the element's value changes, replacing
// inline `onchange="this.form.submit()"` handlers that CSP now blocks.
export default class extends Controller {
  submit(event: Event) {
    const el = event.target as HTMLInputElement | HTMLSelectElement;
    el.form?.requestSubmit();
  }
}
