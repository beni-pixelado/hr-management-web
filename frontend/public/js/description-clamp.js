(function () {
    const MAX_LINES = 3;

    function clampLine(el, lines) {
        const lineHeight = parseFloat(window.getComputedStyle(el).lineHeight);
        if (!lineHeight || !el.textContent.trim()) return;

        const maxHeight = lineHeight * lines;
        el.style.maxHeight = maxHeight + "px";
        el.style.overflow = "hidden";

        const fullText = el.textContent;
        if (el.scrollHeight <= maxHeight) {
            el.title = fullText;
            return;
        }

        let text = fullText;
        while (text.length > 0 && el.scrollHeight > maxHeight) {
            text = text.slice(0, -1);
            el.textContent = text;
        }

        el.textContent = text.replace(/\s+$/, "").slice(0, -3) + "…";
        el.title = fullText;
    }

    function init() {
        document.querySelectorAll(".description-text").forEach(function (el) {
            clampLine(el, MAX_LINES);
        });
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", init);
    } else {
        init();
    }
})();
