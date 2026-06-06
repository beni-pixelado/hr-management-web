const ctx = document.getElementById("pizzaChart");

const chart = new Chart(ctx, {
  type: "pie",
  data: {
    labels: [],
    datasets: [{
      data: [],
      backgroundColor: [
        "#5c6bc0","#f59e0b","#ec489a","#3f51b5",
        "#10b981","#ef4444","#06b6d4","#8b5cf6",
        "#f97316","#84cc16","#6366f1","#14b8a6"
      ]
    }]
  },
  options: {
    plugins: {
      legend: { position: "bottom" }
    }
  }
});

async function loadChart() {
  try {
    const res = await fetch("/api/overview");
    if (!res.ok) throw new Error("Failed to fetch overview data");
    const data = await res.json();

    const departments = data.departments || [];

    chart.data.labels = departments.map(d => d.name);
    chart.data.datasets[0].data = departments.map(d => d.count);
    chart.update();
  } catch (err) {
    console.error("Chart load error:", err);
  }
}

loadChart();