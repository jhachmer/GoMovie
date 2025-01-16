document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('update-button').addEventListener('click', async () => {
        const currentUrl = window.location.href;
        const url = currentUrl
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
