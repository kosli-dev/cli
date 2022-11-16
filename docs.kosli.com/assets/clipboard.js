(function () {
  document.querySelectorAll("div.command").forEach(code => {
    let tag = document.createElement("div");
    let text = document.createTextNode("copied!")
    tag.appendChild(text);
    code.appendChild(tag);
    tag.classList.add("copiedText");
    code.addEventListener("click", function (event) {
      if (navigator.clipboard) {
        let commandText = code.querySelector("code")
        navigator.clipboard.writeText(commandText.textContent);
        tag.classList.add("visible");
        setTimeout(() => {
          tag.classList.remove("visible");        
        }, 1500);
      }
    });
  });
})();
