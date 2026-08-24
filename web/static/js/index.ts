import { config as turboConfig } from "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";

// Turbo owns the loading indicator's timing: it shows .turbo-progress-bar for
// any visit or form submission still in flight after this delay, driven from
// inside its own visit lifecycle, so cached-snapshot previews, prefetches and
// aborted visits are all handled for us. 500 ms is Turbo's default and reads as
// sluggish here. layout.css restyles that element into the loading spinner.
turboConfig.drive.progressBarDelay = 250;

document.addEventListener("turbo:before-fetch-request", (event: Event) => {
  const detail = (event as CustomEvent).detail;
  const url = detail.url as URL;
  if (!url.searchParams.has("tz_offset")) {
    url.searchParams.set("tz_offset", String(new Date().getTimezoneOffset()));
  }
});
import DateController from "./controllers/dateController";
import AmountController from "./controllers/amountController";
import FilterController from "./controllers/filterController";
import NavController from "./controllers/navController";
import SortController from "./controllers/sortController";
import ChartController from "./controllers/chartController";
import LocalDateController from "./controllers/localDateController";
import ThemeController from "./controllers/themeController";
import QuickExpenseController from "./controllers/quickExpenseController";
import DateHelpController from "./controllers/dateHelpController";
import SearchPanelController from "./controllers/searchPanelController";
import { initIcons } from "./icons";

window.Stimulus = Application.start();
window.Stimulus.register("date", DateController);
window.Stimulus.register("amount", AmountController);
window.Stimulus.register("filter", FilterController);
window.Stimulus.register("nav", NavController);
window.Stimulus.register("sort", SortController);
window.Stimulus.register("chart", ChartController);
window.Stimulus.register("local-date", LocalDateController);
window.Stimulus.register("theme", ThemeController);
window.Stimulus.register("quick-expense", QuickExpenseController);
window.Stimulus.register("date-help", DateHelpController);
window.Stimulus.register("search-panel", SearchPanelController);

// turbo:load covers full-page visits; turbo:render also fires when Turbo
// re-renders a form response (including non-2xx error re-renders), which
// turbo:load does not — without it, freshly rendered <i data-lucide> icons
// stay unconverted and invisible.
document.addEventListener("turbo:load", initIcons);
document.addEventListener("turbo:render", initIcons);
