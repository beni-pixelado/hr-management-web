(function () {
    var menu = document.getElementById("absence-menu");
    var markBtn = document.getElementById("absence-mark-btn");
    var currentRow = null;

    function hideMenu() {
        if (!menu) return;
        menu.classList.add("hidden");
        currentRow = null;
    }

    function showMenu(x, y) {
        if (!menu) return;
        menu.classList.remove("hidden");
        menu.style.left = "0px";
        menu.style.top = "0px";

        var rect = menu.getBoundingClientRect();
        var left = Math.min(x, window.innerWidth - rect.width - 8);
        var top = Math.min(y, window.innerHeight - rect.height - 8);

        menu.style.left = left + "px";
        menu.style.top = top + "px";
    }

    function init() {
        if (!menu || !markBtn) return;

        document.querySelectorAll("table.employees-table tbody tr").forEach(function (row) {
            row.addEventListener("contextmenu", function (e) {
                e.preventDefault();
                currentRow = row;
                showMenu(e.clientX, e.clientY);
            });
        });

        document.addEventListener("click", function (e) {
            if (menu.contains(e.target)) return;
            hideMenu();
        });

        window.addEventListener("scroll", hideMenu, true);
        window.addEventListener("resize", hideMenu);
        document.addEventListener("keydown", function (e) {
            if (e.key === "Escape") hideMenu();
        });

        markBtn.addEventListener("click", function () {
            if (!currentRow) return;

            var id = currentRow.getAttribute("data-id");
            fetch("/employees/" + id + "/absence", { method: "POST" })
                .then(function (res) {
                    if (!res.ok) throw new Error("Request failed");
                    return res.json();
                })
                .then(function () {
                    var cell = currentRow.querySelector(".absence-cell");
                    var badge = cell.querySelector(".absence-badge");
                    var next = parseInt(cell.getAttribute("data-absence") || "0", 10) + 1;
                    cell.setAttribute("data-absence", next);
                    if (badge) badge.textContent = next;
                    cell.classList.add("absence-flash");
                    setTimeout(function () {
                        cell.classList.remove("absence-flash");
                    }, 900);
                    hideMenu();
                })
                .catch(function () {
                    alert("Failed to mark absence.");
                    hideMenu();
                });
        });
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", init);
    } else {
        init();
    }
})();
