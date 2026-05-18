import "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";

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
import MacroSelectController from "./controllers/macroSelectController";
import NavController from "./controllers/navController";
import SortController from "./controllers/sortController";
import ChartController from "./controllers/chartController";
import MacroDateController from "./controllers/macroDateController";
import DashboardDateController from "./controllers/dashboardDateController";
import LocalDateController from "./controllers/localDateController";
import MacroCalcController from "./controllers/macroCalcController";
import MacroTrendController from "./controllers/macroTrendController";
import MoodChartController from "./controllers/moodChartController";

window.Stimulus = Application.start();
window.Stimulus.register("date", DateController);
window.Stimulus.register("macro-date", MacroDateController);
window.Stimulus.register("dashboard-date", DashboardDateController);
window.Stimulus.register("amount", AmountController);
window.Stimulus.register("filter", FilterController);
window.Stimulus.register("macro-select", MacroSelectController);
window.Stimulus.register("nav", NavController);
window.Stimulus.register("sort", SortController);
window.Stimulus.register("chart", ChartController);
window.Stimulus.register("local-date", LocalDateController);
window.Stimulus.register("macro-calc", MacroCalcController);
window.Stimulus.register("macro-trend", MacroTrendController);
window.Stimulus.register("mood-chart", MoodChartController);
