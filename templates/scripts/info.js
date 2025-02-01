document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('update-button').addEventListener('click', async () => {
        const url = window.location.href;
        try {
            const response = await fetch(url, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({}),
            });

            if (response.ok) {
                alert('Film updated successfully!');
                location.reload()
            } else {
                alert(`Failed to update film. Status: ${response.status}`);
            }
        } catch (error) {
            console.error('Error updating film:', error);
            alert('An error occurred while updating the film.');
        }
    });
});

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('add-without-entry-button').addEventListener('click', async () => {
        const url = window.location.href;
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({}),
            });

            if (response.ok) {
                alert('Film added successfully!');
                location.reload()
            } else {
                alert(`Failed to add film. Status: ${response.status}`);
            }
        } catch (error) {
            console.error('Error adding film:', error);
            alert('An error occurred while adding the film.');
        }
    });
});

document.addEventListener("DOMContentLoaded", function () {
    const imdbID = window.location.href.substring(window.location.href.lastIndexOf('/') + 1);
    fetch(`/check/${imdbID}`)
        .then(response => response.json())
        .then(data => {
            if (data.exists) {
                document.getElementById("add-without-entry-button").style.display = "none";
            }
        })
        .catch(error => console.error("Error checking movie:", error));
});
