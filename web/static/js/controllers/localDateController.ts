import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static values = { unix: Number, datetime: Boolean };
  declare readonly unixValue: number;
  declare readonly datetimeValue: boolean;

  connect() {
    if (!this.unixValue) return;

    const date = new Date(this.unixValue * 1000);

    if (this.datetimeValue) {
      this.element.textContent = formatDate(date);
      (this.element as HTMLElement).title = formatDateTime(date);
    } else {
      this.element.textContent = formatDateUTC(date);
    }
  }
}

const MONTHS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

function formatDate(date: Date): string {
  const month = MONTHS[date.getMonth()];
  const day = date.getDate();
  const year = date.getFullYear();

  return `${month} ${day}, ${year}`;
}

function formatDateUTC(date: Date): string {
  const month = MONTHS[date.getUTCMonth()];
  const day = date.getUTCDate();
  const year = date.getUTCFullYear();

  return `${month} ${day}, ${year}`;
}

function formatDateTime(date: Date): string {
  const datePart = formatDate(date);
  const h = date.getHours();
  const period = h >= 12 ? "PM" : "AM";
  const hours12 = h % 12 || 12;
  const minutes = String(date.getMinutes()).padStart(2, "0");
  const seconds = String(date.getSeconds()).padStart(2, "0");

  return `${datePart} ${hours12}:${minutes}:${seconds} ${period}`;
}
