import "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";
import DateController from "./controllers/dateController";
import AmountController from "./controllers/amountController";

window.Stimulus = Application.start();
window.Stimulus.register("date", DateController);
window.Stimulus.register("amount", AmountController);
