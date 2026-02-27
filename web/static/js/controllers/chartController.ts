import { Controller } from "@hotwired/stimulus";
import {
  Chart,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Tooltip,
} from "chart.js";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip);

const PALETTE = [
  "#2d6eb0",
  "#e07b39",
  "#3aab6d",
  "#b03060",
  "#8a5cc4",
  "#c4a030",
  "#3399aa",
  "#a04040",
  "#5a8a30",
  "#7060b0",
];

function categoryColor(index: number): string {
  return PALETTE[index % PALETTE.length];
}

export default class extends Controller {
  static targets = ["canvas"];
  static values = { data: String };

  declare readonly canvasTarget: HTMLCanvasElement;
  declare readonly dataValue: string;

  private chart: Chart | null = null;

  connect() {
    const rows = JSON.parse(this.dataValue) as {
      name: string;
      total: number;
    }[];

    this.chart = new Chart(this.canvasTarget, {
      type: "bar",
      data: {
        labels: rows.map((r) => r.name),
        datasets: [
          {
            label: "Total",
            data: rows.map((r) => r.total),
            backgroundColor: rows.map((_, i) => categoryColor(i)),
          },
        ],
      },
      options: {
        indexAxis: "y",
        responsive: true,
        plugins: {
          tooltip: {
            callbacks: {
              label: (ctx) => {
                const raw = ctx.raw as number;
                const formatted = new Intl.NumberFormat("en-US", {
                  style: "currency",
                  currency: "USD",
                }).format(raw / 100);
                return ` ${formatted}`;
              },
            },
          },
        },
      },
    });
  }

  disconnect() {
    this.chart?.destroy();
    this.chart = null;
  }
}
