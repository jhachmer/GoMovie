document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('menu-search-bar').addEventListener('submit', function (event) {
        event.preventDefault();

        const searchInput = document.getElementById('search-input').value.trim();

        const imdbIdPattern = /^tt\d{7,8}$/;
        console.log('Test JS Loaded');

        if (imdbIdPattern.test(searchInput)) {
            window.location.href = `/films/${searchInput}`;
        } else {
            window.alert("invalid imdb id");
            // window.location.href = `/search?q=${encodeURIComponent(searchInput)}`;
        }
    });
});
