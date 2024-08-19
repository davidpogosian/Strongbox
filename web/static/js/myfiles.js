function logout() {
    Cookies.remove("auth-session");
    window.location = "/";
}

document.addEventListener("DOMContentLoaded", (event) => {
    const logoutButton = document.getElementById("logout_button");
    logoutButton.onclick = logout;
});
