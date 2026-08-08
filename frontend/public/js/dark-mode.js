(function () {
    function safeGetItem(key) {
        try {
            return localStorage.getItem(key);
        } catch (e) {
            return null;
        }
    }

    function safeSetItem(key, value) {
        try {
            localStorage.setItem(key, value);
        } catch (e) {
            // storage unavailable; keep running without persisting
        }
    }

    function init() {
        const urlTheme = new URLSearchParams(window.location.search).get("theme");
        const savedTheme = safeGetItem("theme");

        let theme = urlTheme === "dark" || urlTheme === "light"
            ? urlTheme
            : savedTheme === "dark" || savedTheme === "light"
                ? savedTheme
                : "light";

        safeSetItem("theme", theme);
        document.body.classList.toggle("dark", theme === "dark");

        const themeSelect = document.getElementById("theme");
        if (themeSelect) {
            themeSelect.value = theme;
            themeSelect.addEventListener("change", () => {
                theme = themeSelect.value === "dark" ? "dark" : "light";
                document.body.classList.toggle("dark", theme === "dark");
                safeSetItem("theme", theme);
            });
        }
    }

    if (document.body) {
        init();
    } else {
        document.addEventListener("DOMContentLoaded", init);
    }
})();
