import { Controller } from "@hotwired/stimulus";
import {
  Chart,
  LineController,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  Legend,
} from "chart.js";

Chart.register(
  LineController,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  Legend,
);

interface TrendDataset {
  label: string;
  data: number[];
}

interface TrendChartData {
  labels: string[];
  datasets: TrendDataset[];
}

const COLORS = ["#2d6eb0", "#3aab6d", "#e07b39", "#b03060"];

export default class extends Controller {
  static values = { data: Object };
  static targets = ["canvas"];

  declare readonly dataValue: TrendChartData;
  declare readonly canvasTarget: HTMLCanvasElement;

  private chart: Chart | null = null;

  connect() {
    const raw = this.dataValue;
    if (!raw?.labels?.length) return;

    const datasets = raw.datasets.map((ds, i) => ({
      label: ds.label,
      data: ds.data,
      borderColor: COLORS[i % COLORS.length],
      backgroundColor: COLORS[i % COLORS.length],
      tension: 0.3,
      pointRadius: 2,
      fill: false,
    }));

    this.chart = new Chart(this.canvasTarget, {
      type: "line",
      data: { labels: raw.labels, datasets },
      options: {
        responsive: true,
        plugins: {
          legend: { position: "bottom" },
        },
        scales: {
          y: { beginAtZero: true },
        },
      },
    });
  }

  disconnect() {
    this.chart?.destroy();
    this.chart = null;
  }
}
