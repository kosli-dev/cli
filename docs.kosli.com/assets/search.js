'use strict';

{{ $searchDataFile := printf "%s.search-data.json" .Language.Lang }}
{{ $searchData := resources.Get "search-data.json" | resources.ExecuteAsTemplate $searchDataFile . | resources.Minify | resources.Fingerprint }}
{{ $searchConfig := i18n "bookSearchConfig" | default "{}" }}

(function () {
  const searchDataURL = '{{ $searchData.RelPermalink }}';
  const indexConfig = Object.assign({{ $searchConfig }}, {
    doc: {
      id: 'id',
      field: ['title', 'content'],
      store: ['title', 'href', 'section']
    }
  });

  const input = document.querySelector('#book-search-input');
  const results = document.querySelector('#book-search-results');

  if (!input) {
    return
  }

  // const tabToResults2 = (event) => {
  //   let numResults = foundResults.childElementCount;
  //   if (event.key === "Tab" && numResults > 0) {
  //     event.preventDefault;
  //     const foundResults = document.querySelector('#book-search-results');
  //     console.log("tab");
  //     console.log(numResults)
  //     const firstResult = foundResults.firstChild.querySelector("a");
  //     firstResult.focus();
  //     numResults = 0
  //   }
  // }

  input.addEventListener('focus', init);
  input.addEventListener('keyup', search);
  input.addEventListener('keydown', tabToResults2);

  document.addEventListener('keypress', focusSearchFieldOnKeyPress);

  /**
   * @param {Event} event
   */
  function focusSearchFieldOnKeyPress(event) {
    if (input === document.activeElement) {
      return;
    }

    const characterPressed = String.fromCharCode(event.charCode);
    if (!isHotkey(characterPressed)) {
      return;
    }

    input.focus();
    event.preventDefault();
  }

  /**
   * @param {String} character
   * @returns {Boolean} 
   */
  function isHotkey(character) {
    const dataHotkeys = input.getAttribute('data-hotkeys') || '';
    return dataHotkeys.indexOf(character) >= 0;
  }

  function init() {
    input.removeEventListener('focus', init); // init once
    input.required = true;

    fetch(searchDataURL)
      .then(pages => pages.json())
      .then(pages => {
        window.bookSearchIndex = FlexSearch.create('balance', indexConfig);
        window.bookSearchIndex.add(pages);
      })
      .then(() => input.required = false)
      .then(search);
  }

  function search() {
    while (results.firstChild) {
      results.removeChild(results.firstChild);
    }

    if (!input.value) {
      return;
    }

    const searchHits = window.bookSearchIndex.search(input.value, 10);
    searchHits.forEach(function (page) {
      const li = element('<li><a href></a><small></small></li>');
      const a = li.querySelector('a'), small = li.querySelector('small');

      a.href = page.href;
      a.textContent = page.title;
      small.textContent = page.section;

      results.appendChild(li);
    });

    // tabToResults();

  }

  /**
   * @param {String} content
   * @returns {Node}
   */
  function element(content) {
    const div = document.createElement('div');
    div.innerHTML = content;
    return div.firstChild;
  }

  function tabToResults() {
    input.addEventListener("keydown", (event) => {
      if (event.key === "Tab") {
        event.preventDefault();
        const foundResults = document.querySelector('#book-search-results');
        let firstResult = foundResults.firstChild.querySelector("a");
        console.log(`After tab: ${firstResult}`);
        let numResults = foundResults.childElementCount;
        if (numResults > 0) {
          console.log(firstResult);
          firstResult.focus();
        } else {
          console.log("inside else");
        }
      }
    });
  }

  function tabToResults2(event) {
    if (event.key === "Tab") {
      event.preventDefault();
      const foundResults = document.querySelector('#book-search-results');
      let firstResult = foundResults.firstChild.querySelector("a");
      console.log(`After tab: ${firstResult}`);
      let numResults = foundResults.childElementCount;
      if (numResults > 0) {
        console.log(firstResult);
        firstResult.focus();
      } else {
        console.log("inside else");
      }
    }
  }

})();
