async function createLineChart(canvasId, endpoint, label) {
    const canvas = document.getElementById(canvasId);
    if (!canvas) {
        return null;
    }

    const existingChart = Chart.getChart(canvas);
    if (existingChart) {
        existingChart.destroy();
    }

    try {
        const response = await fetch(endpoint);

        if (!response.ok) {
            throw new Error("Failed to fetch chart data");
        }

        const data = await response.json();
        const employees = Array.isArray(data.employees) ? data.employees : [];

        const now = new Date();
        const startOfPreviousMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1);
        const endOfCurrentMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0);

        const pointsByDay = new Map(
            employees.map(e => [
                new Date(String(e.day || e.month || "")).toISOString().slice(0, 10),
                Number(e.count) || 0
            ])
        );

        const dataPoints = [];
        const cursor = new Date(startOfPreviousMonth);

        while (cursor <= endOfCurrentMonth) {
            const key = `${cursor.getFullYear()}-${String(cursor.getMonth() + 1).padStart(2, "0")}-${String(cursor.getDate()).padStart(2, "0")}`;
            const y = pointsByDay.get(key) || 0;

            dataPoints.push({
                x: new Date(cursor.getFullYear(), cursor.getMonth(), cursor.getDate()),
                y
            });

            cursor.setDate(cursor.getDate() + 1);
        }

        const maxCount = dataPoints.reduce((m, p) => Math.max(m, p.y), 0);
        const yMax = Math.max(10, Math.ceil(maxCount / 10) * 10);

        const chart = new Chart(canvas, {
            type: "line",
            data: {
                datasets: [{
                    label: label,
                    data: dataPoints,
                    borderColor: "#2563eb",
                    backgroundColor: "rgba(37, 99, 235, 0.2)",
                    tension: 0.4,
                    fill: true,
                    pointRadius: (context) => (context.parsed.y > 0 ? 3 : 0),
                    pointBackgroundColor: "#2563eb"
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true
                    },
                    tooltip: {
                        callbacks: {
                            title(items) {
                                const date = items[0]?.parsed?.x;
                                if (!(date instanceof Date)) {
                                    return "";
                                }
                                return date.toLocaleDateString("en-US", {
                                    day: "2-digit",
                                    month: "short"
                                });
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        type: "time",
                        time: {
                            unit: "day",
                            parser: "yyyy-MM-dd",
                            tooltipFormat: "dd/MM",
                            displayFormats: {
                                day: "dd"
                            }
                        },
                        min: startOfPreviousMonth,
                        max: endOfCurrentMonth,
                        ticks: {
                            autoSkip: false,
                            maxRotation: 0,
                            color: (context) => {
                                const date = new Date(context.tick.value);
                                return date.getMonth() === now.getMonth() ? "#334155" : "#94a3b8";
                            },
                            callback(value) {
                                const date = new Date(value);
                                const day = date.getDate();
                                const isWeeklyLabel = [1, 7, 14, 21, 28, 4].includes(day);

                                if (!isWeeklyLabel) {
                                    return "";
                                }

                                return String(day);
                            }
                        }
                    },
                    y: {
                        beginAtZero: true,
                        max: yMax,
                        ticks: {
                            stepSize: 1
                        }
                    }
                }
            }
        });

        return chart;
    } catch (err) {
        console.error("Chart load error:", err);
        return null;
    }
}

createLineChart(
    "lineEmployees",
    "/api/overview/employees",
    "Employees"
);