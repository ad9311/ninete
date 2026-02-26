import "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";
import DateController from "./controllers/dateController";
import AmountController from "./controllers/amountController";
import FilterController from "./controllers/filterController";
import TaskFilterController from "./controllers/taskFilterController";
import SortController from "./controllers/sortController";

window.Stimulus = Application.start();
window.Stimulus.register("date", DateController);
window.Stimulus.register("amount", AmountController);
window.Stimulus.register("filter", FilterController);
window.Stimulus.register("task-filter", TaskFilterController);
window.Stimulus.register("sort", SortController);
