const ctx = document.getElementById("pizzaChart");

const chart = new Chart(ctx, {
  type: "pie",
  data: {
    labels: [],
    datasets: [{
      data: []
    }]
  }
});

async function loadChart() {
  const res = await fetch("/overview");
  const data = await res.json();

  const departments = data.departments;

  chart.data.labels = departments.map(d => d.name);
  chart.data.datasets[0].data = departments.map(d => d.count);

  chart.update();
}

loadChart();