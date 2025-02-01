// Function to sort the table columns
function sortTable(columnIndex) {
    const table = document.getElementById("moviesTable");
    const rows = Array.from(table.tBodies[0].rows);
    const isAscending = table.getAttribute("data-sort-asc") === "true";

    rows.sort((a, b) => {
        const cellA = a.cells[columnIndex].textContent.trim();
        const cellB = b.cells[columnIndex].textContent.trim();

        const parseValue = (value) => {
            if (value === "N/A") {
                return 0;
            }
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

// Checkbox Listener to filter out already watched movies
document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('filterNotWatched').addEventListener('change', function () {
        const showNotWatchedOnly = this.checked;
        const rows = document.querySelectorAll('#moviesTable tbody tr');

        rows.forEach(row => {
            if (row.classList.contains('not-watched')) {
                row.style.display = showNotWatchedOnly ? '' : 'table-row';
            } else {
                row.style.display = showNotWatchedOnly ? 'none' : 'table-row';
            }
        });
    });
});


document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll(".delete-button").forEach(button => {
        button.addEventListener("click", function () {
            const imdbID = this.getAttribute("data-imdbid");

            if (confirm("Are you sure you want to delete this movie?")) {
                fetch(`/films/${imdbID}`, {
                    method: "DELETE",
                    headers: {
                        "Content-Type": "application/json"
                    }
                })
                    .then(response => {
                        if (response.ok) {
                            this.closest("tr").remove(); // Remove row from table
                        } else {
                            return response.json().then(data => { throw new Error(data.message || "Delete failed"); });
                        }
                    })
                    .catch(error => alert("Error: " + error.message));
            }
        });
    });
});
