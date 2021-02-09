const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const fuzziness = data.fuzzy && data.fuzzy === 'on' ? 1 : 0;
    const response = fetch(`/search?q=${data.query}&fuzziness=${fuzziness}&page[size]=5000`).then((response) => {
      response.json().then((results) => {
        Controller.updateResultView(results);
      });
    });
  },

  updateResultView: (results) => {
    // total
    const totalDiv = document.getElementById("total");
    const totalResults = results.meta.totalResults;
    totalDiv.textContent = totalResults ? `${results.meta.totalResults} resutls` : 'No results';

    // table
    const table = document.getElementById("table-body");
    const rows = [];
    for (let hit of results.data) {
      rows.push(`<tr>[${hit.title}] ${hit.lineNumber}.&ensp;${hit.line}</tr>`);
    }
    table.innerHTML = rows;
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
