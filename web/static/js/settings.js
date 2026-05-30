// settings.js - Quản lý Tuỳ chỉnh Giao diện (Appearance Settings)
document.addEventListener("DOMContentLoaded", function() {
    const settingsBtn = document.getElementById("settings-btn");
    const settingsPanel = document.getElementById("settings-panel");

    if (!settingsBtn || !settingsPanel) return;

    // Toggle settings panel mở/đóng
    settingsBtn.addEventListener("click", function(e) {
        e.stopPropagation();
        // Đóng account dropdown nếu đang mở
        const accountMenu = document.getElementById("account-dropdown-menu");
        if (accountMenu) accountMenu.classList.add("hidden");

        const isHidden = settingsPanel.classList.contains("hidden");
        if (isHidden) {
            settingsPanel.classList.remove("hidden");
            setTimeout(function() {
                settingsPanel.classList.remove("scale-95", "opacity-0");
                settingsPanel.classList.add("scale-100", "opacity-100");
            }, 10);
        } else {
            closeSettingsPanel();
        }
    });

    function closeSettingsPanel() {
        settingsPanel.classList.remove("scale-100", "opacity-100");
        settingsPanel.classList.add("scale-95", "opacity-0");
        setTimeout(function() {
            settingsPanel.classList.add("hidden");
        }, 150);
    }

    // Đóng khi click ra ngoài
    document.addEventListener("click", function(e) {
        if (!settingsPanel.contains(e.target) && e.target !== settingsBtn && !settingsBtn.contains(e.target)) {
            if (!settingsPanel.classList.contains("hidden")) {
                closeSettingsPanel();
            }
        }
    });

    // Đóng khi nhấn Escape
    document.addEventListener("keydown", function(e) {
        if (e.key === "Escape" && !settingsPanel.classList.contains("hidden")) {
            closeSettingsPanel();
        }
    });

    // Xử lý click vào các pill toggle
    var pills = settingsPanel.querySelectorAll(".setting-pill");
    pills.forEach(function(pill) {
        pill.addEventListener("click", function() {
            var group = this.closest(".setting-group");
            var setting = group.getAttribute("data-setting");
            var value = this.getAttribute("data-value");

            // Cập nhật trạng thái active trong nhóm
            group.querySelectorAll(".setting-pill").forEach(function(p) {
                p.classList.remove("active");
            });
            this.classList.add("active");

            // Lưu và áp dụng
            applySetting(setting, value);
        });
    });

    // Khởi tạo trạng thái pill từ localStorage
    initPillStates();
    updateLayoutWidth();

    // Lắng nghe thay đổi theme hệ thống (chế độ "Tự động")
    window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", function() {
        if (localStorage.getItem("axia-theme") === "auto") {
            applyTheme("auto");
        }
    });
});

// Cập nhật chiều rộng của main wrapper theo cấu hình hiện tại
function updateLayoutWidth() {
    var mainWrapper = document.querySelector('main');
    if (!mainWrapper) return;
    if (document.body.classList.contains('focus-active')) return;
    
    var isWide = document.body.classList.contains('content-width-wide');
    mainWrapper.style.maxWidth = isWide ? '1600px' : '1350px';
}

// Áp dụng một cài đặt cụ thể
function applySetting(setting, value) {
    var body = document.body;

    if (setting === "text-size") {
        body.classList.remove("text-size-sm", "text-size-md", "text-size-lg");
        body.classList.add("text-size-" + value);
        localStorage.setItem("axia-text-size", value);
    } else if (setting === "content-width") {
        body.classList.remove("content-width-standard", "content-width-wide");
        body.classList.add("content-width-" + value);
        localStorage.setItem("axia-content-width", value);
        updateLayoutWidth();
    } else if (setting === "theme") {
        applyTheme(value);
        localStorage.setItem("axia-theme", value);
    }
}

// Áp dụng theme lên body
function applyTheme(value) {
    var body = document.body;
    body.classList.remove("theme-light", "theme-dark", "theme-auto");

    if (value === "auto") {
        body.classList.add("theme-auto");
        var prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
        body.classList.add(prefersDark ? "theme-dark" : "theme-light");
    } else {
        body.classList.add("theme-" + value);
    }
}

// Khởi tạo trạng thái active cho các pill toggle từ localStorage
function initPillStates() {
    var textSize = localStorage.getItem("axia-text-size") || "md";
    var contentWidth = localStorage.getItem("axia-content-width") || "standard";
    var theme = localStorage.getItem("axia-theme") || "light";

    setActivePill("text-size", textSize);
    setActivePill("content-width", contentWidth);
    setActivePill("theme", theme);
}

function setActivePill(setting, value) {
    var group = document.querySelector('.setting-group[data-setting="' + setting + '"]');
    if (!group) return;
    group.querySelectorAll(".setting-pill").forEach(function(p) { p.classList.remove("active"); });
    var activePill = group.querySelector('.setting-pill[data-value="' + value + '"]');
    if (activePill) activePill.classList.add("active");
}
