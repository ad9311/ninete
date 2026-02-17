import "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";
import DateController from "./controllers/dateController";
import AmountController from "./controllers/amountController";
import FilterController from "./controllers/filterController";

window.Stimulus = Application.start();
window.Stimulus.register("date", DateController);
window.Stimulus.register("amount", AmountController);
window.Stimulus.register("filter", FilterController);
