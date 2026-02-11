import "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";
import DateController from "./controllers/dateController";

window.Stimulus = Application.start();
window.Stimulus.register("date", DateController);
