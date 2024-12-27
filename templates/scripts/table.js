window.onload = function() {
    const ratingCells = document.querySelectorAll('.ratings');

    ratingCells.forEach(function(cell) {
        let rawText = cell.textContent || cell.innerText;

        let ratings = rawText.replace(/[\[\]]/g, '')
                            .split('Rotten Tomatoes')
                            .map(function(part, index) {
                                if (index === 0) {
                                    return part.trim().replace('Internet Movie Database', '<span class="imdb">IMDb</span>');
                                } else {
                                    return '<span class="rotten-tomatoes"> Rotten Tomatoes</span>' + part.trim().replace('Metacritic', ', <span class="metacritic">Metacritic</span>');
                                }
                            }).join(', ');

        cell.innerHTML = ratings;
    });
};


function sortTable(columnIndex) {
    const table = document.getElementById("moviesTable");
    const rows = Array.from(table.tBodies[0].rows);
    const isAscending = table.getAttribute("data-sort-asc") === "true";

    rows.sort((a, b) => {
        const cellA = a.cells[columnIndex].textContent.trim();
        const cellB = b.cells[columnIndex].textContent.trim();

        if (columnIndex === 2) {
            const numA = parseInt(cellA, 10);
            const numB = parseInt(cellB, 10);

            if (isNaN(numA) || isNaN(numB)) {
                return 0;
            }

            return isAscending ? numA - numB : numB - numA;
        }

        if (!isNaN(cellA) && !isNaN(cellB)) {
            return isAscending ? cellA - cellB : cellB - cellA;
        }
        return isAscending
            ? cellA.localeCompare(cellB)
            : cellB.localeCompare(cellA);
    });

    rows.forEach(row => table.tBodies[0].appendChild(row));
    table.setAttribute("data-sort-asc", !isAscending);
}
