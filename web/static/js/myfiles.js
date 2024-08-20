function logout() {
    Cookies.remove("auth-session");
    window.location = "/";
}

async function handleFile(file) {
    console.log("File:", file.name);

    try {
        const result = await uploadFile(file);
        console.log('File upload successful:', result);
    } catch (error) {
        console.error('Error uploading file:', error);
    }
}

async function overwriteFile(file) {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch("/api/upload?overwrite=true", {
        method: "POST",
        body: formData,
    });

    if (!response.ok) {
        throw new Error("File overwrite failed");
    }

    return response.json();
}

async function uploadFile(file) {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch("/api/upload", {
        method: "POST",
        body: formData,
    });

    if (!response.ok) {
        if (response.status === 409) {
            const confirmOverwrite = confirm("The file already exists. Do you want to overwrite it?");
            if (confirmOverwrite) {
                return await overwriteFile(file);
            } else {
                throw new Error("File upload canceled by user");
            }
        }
        throw new Error("File upload failed");
    }

    return response.json();
}

document.addEventListener("DOMContentLoaded", (event) => {
    const logoutButton = document.getElementById("logout_button");
    logoutButton.onclick = logout;

    const dropZone = document.getElementById("drop_zone");

    // Prevent default behavior (Prevent file from being opened)
    dropZone.addEventListener("dragover", (event) => {
        event.preventDefault();
        dropZone.classList.add("dragover");
    });

    dropZone.addEventListener("dragleave", () => {
        dropZone.classList.remove("dragover");
    });

    dropZone.addEventListener("drop", (event) => {
        event.preventDefault();
        dropZone.classList.remove("dragover");

        // Get the files from the drop event
        const file = event.dataTransfer.files[0];
        if (file) {
            handleFile(file);
        }
    });
});
