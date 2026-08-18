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
        var timeLabel = document.getElementById("time-label");

        if (!typeSelect || !timeSelect || !feedback) return;

        function endpointFor(typeValue) {
            if (typeValue === "hired") return "/api/report/hired";
            if (typeValue === "fired") return "/api/report/fired";
            return "/api/report/absences";
        }

        function labelFor(typeValue) {
            if (typeValue === "hired") return "Count hired in the last";
            if (typeValue === "fired") return "Count fired in the last";
            return "Count absences in the last";
        }

        function update() {
            var typeValue = typeSelect.value;
            var typeText = typeSelect.options[typeSelect.selectedIndex].text;
            var days = DAYS[timeSelect.value] || 7;

            if (timeLabel) {
                timeLabel.textContent = labelFor(typeValue);
            }

            feedback.textContent = "Loading...";

            fetch(endpointFor(typeValue) + "?days=" + days)
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
