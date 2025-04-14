async function formReq(event, method, url, swap) {
  event.preventDefault();
  const formData = new FormData(event.target);
  try {
    const response = await fetch(url, {
      method: method,
      body: formData,
    });
    if (!response.ok) {
      throw new Error(`Server error: ${response.status} ${response.statusText}`);
    }

    const contentType = response.headers.get("Content-Type");
    if (contentType && contentType.includes("text/html") && swap) {
      const html = await response.text();
      const documentContainer = document.querySelector(".addpost");
      documentContainer.insertAdjacentHTML(swap, html);
      event.target.reset();
    }
  } catch (error) {
    console.error("Error:", error);
  }
}
async function btnReq(method, url) {
  console.log(method);
  try {
    const response = await fetch(url, {
      method: method,
    });
    if (!response.ok) {
      throw new Error("request failed");
    }
    const data = response;
    console.log("Success:", data);
  } catch (error) {
    console.error("Error:", error);
  }
}
function theme() {
  const currentTheme = localStorage.getItem("theme") === "dark" ? "light" : "dark";
  document.documentElement.classList.toggle("dark", currentTheme === "dark");
  document.documentElement.classList.toggle("light", currentTheme === "light");
  localStorage.setItem("theme", currentTheme);
}
const savedTheme = localStorage.getItem("theme") || (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
document.documentElement.classList.add(savedTheme);

document.addEventListener("DOMContentLoaded", function () {
  const path = window.location.pathname;
  const button = document.getElementById("doc-toggle-button");
  if (path.startsWith("/doc/")) {
    button.classList.remove("hidden");
  }
});
