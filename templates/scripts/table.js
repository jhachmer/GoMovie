function sortTable(columnIndex) {
    const table = document.getElementById("moviesTable");
    const rows = Array.from(table.tBodies[0].rows);
    const isAscending = table.getAttribute("data-sort-asc") === "true";

    rows.sort((a, b) => {
        const cellA = a.cells[columnIndex].textContent.trim();
        const cellB = b.cells[columnIndex].textContent.trim();

        const parseValue = (value) => {
            if (value.includes('%')) {
                return parseFloat(value.replace('%', ''));
            }
            if (value.includes('/')) {
                const parts = value.split('/').map(num => parseFloat(num));
                return parts[0] / parts[1];
            }
            const num = parseFloat(value);
            return isNaN(num) ? value : num;
        };

        const valueA = parseValue(cellA);
        const valueB = parseValue(cellB);

        if (typeof valueA === "number" && typeof valueB === "number") {
            return isAscending ? valueA - valueB : valueB - valueA;
        }

        return isAscending
            ? String(valueA).localeCompare(String(valueB))
            : String(valueB).localeCompare(String(valueA));
    });

    rows.forEach(row => table.tBodies[0].appendChild(row));
    table.setAttribute("data-sort-asc", !isAscending);
}
