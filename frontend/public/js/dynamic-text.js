(function () {
    var DAYS = {
        "week": 7,
        "month": 30,
        "year": 365
    };

    function init() {
        var typeSelect = document.getElementById("type");
        var timeSelect = document.getElementById("time");
        var feedback = document.getElementById("time-feedback");

        if (!typeSelect || !timeSelect || !feedback) return;

        function update() {
            var typeText = typeSelect.options[typeSelect.selectedIndex].text;
            var days = DAYS[timeSelect.value] || 7;

            feedback.textContent = "Loading...";

            fetch("/api/report/absences?days=" + days)
                .then(function (res) { return res.json(); })
                .then(function (data) {
                    feedback.textContent = "Total " + typeText.toLowerCase() + ": " + data.total + " in " + days + " days";
                })
                .catch(function () {
                    feedback.textContent = "Total " + typeText.toLowerCase() + " in " + days + " days";
                });
        }

        update();
        typeSelect.addEventListener("change", update);
        timeSelect.addEventListener("change", update);
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", init);
    } else {
        init();
    }
})();
