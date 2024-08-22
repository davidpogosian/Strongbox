function logout() {
    Cookies.remove("auth-session");
    window.location = "/";
}

async function traverseFileTree(item, formData, path = '') {
    let currentUrlPath = window.location.pathname.substring(7); // trim the "/files/"

    if (item.isFile) {
        // If it's a file, add it to the FormData
        return new Promise((resolve) => {
            item.file((file) => {
                if (currentUrlPath !== "") {
                    currentUrlPath = currentUrlPath + "/";
                }
                if (path !== "") {
                    path = path + "/";
                }
                const fullPath = `${currentUrlPath}${path}${file.name}`;
                console.log(`Adding file: ${fullPath}`);
                formData.append('files[]', file, fullPath);
                formData.append('filePaths[]', fullPath);
                resolve(); // Resolve promise once the file is added
            });
        });
    } else if (item.isDirectory) {
        // If it's a directory, read its contents recursively
        return new Promise((resolve) => {
            const dirReader = item.createReader();
            dirReader.readEntries(async (entries) => {
                let newPath = item.name;
                if (path !== "") {
                    newPath = path + "/" + newPath;
                }
                const promises = entries.map(entry => traverseFileTree(entry, formData, newPath));
                await Promise.all(promises);
                resolve(); // Resolve promise once all entries are processed
            });
        });
    }
}

async function overwriteFiles(formData) {
    try {
        const response = await fetch("/api/upload?overwrite=true", {
            method: "POST",
            body: formData,
        });

        if (!response.ok) {
            throw new Error("File overwrite failed");
        }
    } catch (error) {
        console.error('Error uploading files:', error);
        alert('An error occurred during file upload');
    }
}

async function uploadFiles(formData) {
    try {
        const response = await fetch('/api/upload', {
            method: 'POST',
            body: formData,
        });

        if (response.ok) {
            alert('Files and folders uploaded successfully');
        } else {
            if (response.status === 409) {
                const confirmOverwrite = confirm("At least one of the files already exists. Do you want to overwrite it?");
                if (confirmOverwrite) {
                    overwriteFiles(formData);
                } else {
                    console.debug("upload cancelled by user");
                }
            } else {
                alert('Failed to upload files');
            }
        }
    } catch (error) {
        console.error('Error uploading files:', error);
        alert('An error occurred during file upload');
    }
}

async function downloadFile(s3Key) {
    try {
        const response = await fetch(`/api/download?key=${encodeURIComponent(s3Key)}`, { method: "GET" });

        if (response.ok) {
            const data = await response.json(); // Parse JSON response
            const downloadUrl = data.download_url;

            // Create a link element and trigger the download
            const a = document.createElement("a");
            a.style.display = "none";
            a.href = downloadUrl;
            a.download = s3Key.split("/").pop(); // Set the default download file name
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
        } else {
            alert("Failed to get download URL");
        }
    } catch (error) {
        console.error("Error downloading file:", error);
        alert("An error occurred while trying to download the file.");
    }
}

async function copyDownloadUrl(s3Key) {
    try {
        // Fetch the download URL from the server
        const response = await fetch(`/api/download?key=${encodeURIComponent(s3Key)}`, { method: "GET" });

        if (response.ok) {
            const data = await response.json(); // Parse JSON response
            const downloadUrl = data.download_url;

            // Copy the download URL to the clipboard
            await navigator.clipboard.writeText(downloadUrl);

            // Optionally provide feedback to the user
            alert("Download URL copied to clipboard!");
        } else {
            alert("Failed to get download URL");
        }
    } catch (error) {
        console.error("Error copying URL:", error);
        alert("An error occurred while trying to copy the URL.");
    }
}

async function destroy(s3Key) {
    const parts = s3Key.split('/');
    const name = parts[parts.length - 1];
    const confirmDestroy = confirm(`Are you sure you want to destroy '${name}'`);
    if (!confirmDestroy) {
        console.debug("Destroy cancelled by user");
        return;
    }

    try {
        const response = await fetch(`/api/destroy?key=${encodeURIComponent(s3Key)}`, { method: "DELETE" });

        if (response.ok) {
            alert("Delete successful");
        } else {
            throw new Error("Unable to destroy");
        }
    } catch (error) {
        console.error("Error destroying:", error);
        alert("Destroy failed.");
    }

    location.reload();
}

async function downloadFolder(s3Key) {
    try {
        // Make a GET request to the /api/downloadFolder endpoint with the s3Key
        const response = await fetch(`/api/downloadFolder?key=${encodeURIComponent(s3Key)}`, { method: "GET" });

        if (response.ok) {
            const data = await response.json(); // Parse JSON response
            const downloadUrl = data.download_url;

            // Create a link element and trigger the download
            const a = document.createElement("a");
            a.style.display = "none";
            a.href = downloadUrl;
            a.download = s3Key.split("/").pop() + ".zip"; // Set the default download file name
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            setTimeout(() => {
                location.reload(); // Reload the page after 3 seconds (adjust as needed)
            }, 500);
        } else {
            alert("Failed to get download URL");
        }
    } catch (error) {
        console.error("Error downloading folder:", error);
        alert("An error occurred while trying to download the folder.");
    }
}

async function copyDownloadFolderUrl(s3Key) {
    try {
        // Fetch the download URL from the server
        const response = await fetch(`/api/downloadFolder?key=${encodeURIComponent(s3Key)}`, { method: "GET" });

        if (response.ok) {
            const data = await response.json(); // Parse JSON response
            const downloadUrl = data.download_url;

            // Copy the download URL to the clipboard
            await navigator.clipboard.writeText(downloadUrl);

            // Optionally provide feedback to the user
            alert("Download URL copied to clipboard!");
        } else {
            alert("Failed to get download URL");
        }
    } catch (error) {
        console.error("Error copying download folder URL:", error);
        alert("An error occurred while trying to copy the download folder URL.");
    }

    location.reload();
}

document.addEventListener("DOMContentLoaded", (event) => {
    const dropZone = document.getElementById("drop_zone");

    // Prevent default behavior (Prevent file from being opened)
    dropZone.addEventListener("dragover", (event) => {
        event.preventDefault();
        dropZone.classList.add("dragover");
    });

    dropZone.addEventListener("dragleave", () => {
        dropZone.classList.remove("dragover");
    });

    dropZone.addEventListener('drop', async (event) => {
        event.preventDefault();
        dropZone.classList.remove('dragover');

        const items = event.dataTransfer.items;

        // Create a FormData object to store the files
        const formData = new FormData();

        // Collect all promises for processing files and directories
        const promises = [];
        for (let i = 0; i < items.length; i++) {
            const item = items[i].webkitGetAsEntry();
            if (item) {
                promises.push(traverseFileTree(item, formData));
            }
        }

        // Wait for all promises to resolve
        await Promise.all(promises);

        // Send the formData to the server
        await uploadFiles(formData);

        location.reload();
    });
});
