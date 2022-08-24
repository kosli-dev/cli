(function () {
  document.querySelectorAll("div.command").forEach(code => {
    code.addEventListener("click", function (event) {
      if (navigator.clipboard) {
        navigator.clipboard.writeText(code.textContent);
      }
    });
  });
})();
