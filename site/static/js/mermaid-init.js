/* Mermaid re-init on theme change */
(function () {
  document.addEventListener("click", (e) => {
    if (e.target.closest && e.target.closest("[data-theme-toggle]")) {
      // wait a tick so the data-theme attribute updates first
      setTimeout(() => {
        if (!window.mermaid) return;
        const dark = document.documentElement.dataset.theme === "dark"
          || (document.documentElement.dataset.theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches);
        // re-render
        document.querySelectorAll(".mermaid").forEach((el) => {
          const src = el.dataset.src || el.textContent;
          el.dataset.src = src;
          el.removeAttribute("data-processed");
          el.innerHTML = src;
        });
        window.mermaid.initialize({
          startOnLoad: false,
          theme: dark ? "dark" : "neutral",
          securityLevel: "loose",
          themeVariables: { fontFamily: 'ui-monospace, "JetBrains Mono", monospace', fontSize: "14px" },
        });
        window.mermaid.run();
      }, 80);
    }
  });
})();
