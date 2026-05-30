// sidebar.js - Quản lý co giãn, đóng mở Sidebar và lưu trạng thái UI
document.addEventListener("DOMContentLoaded", function() {
    const sidebar = document.getElementById("wiki-sidebar");
    const resizer = document.getElementById("sidebar-resizer");
    const toggleBtn = document.getElementById("sidebar-toggle");
    
    let isResizing = false;
    let isCollapsed = localStorage.getItem("sidebar-collapsed") === "true";
    let savedWidth = parseInt(localStorage.getItem("sidebar-width")) || 256;

    // Áp dụng trạng thái đã lưu cho sidebar khi tải trang
    if (sidebar && resizer) {
        if (isCollapsed) {
            sidebar.style.display = "none";
            resizer.style.display = "none";
        } else {
            sidebar.style.width = savedWidth + "px";
            sidebar.style.display = "flex";
            resizer.style.display = "block";
        }
    }

    // Bật tắt Sidebar
    if (toggleBtn && sidebar && resizer) {
        toggleBtn.addEventListener("click", function() {
            if (sidebar.style.display === "none") {
                sidebar.style.display = "flex";
                resizer.style.display = "block";
                sidebar.style.width = savedWidth + "px";
                localStorage.setItem("sidebar-collapsed", "false");
            } else {
                sidebar.style.display = "none";
                resizer.style.display = "none";
                localStorage.setItem("sidebar-collapsed", "true");
            }
        });
    }

    // Logic Kéo giãn rộng/hẹp
    if (resizer && sidebar) {
        resizer.addEventListener("mousedown", function(e) {
            isResizing = true;
            document.body.style.cursor = "col-resize";
            document.body.style.userSelect = "none";
        });

        document.addEventListener("mousemove", function(e) {
            if (!isResizing) return;
            
            let newWidth = e.clientX;
            
            // Giới hạn chiều rộng từ 160px tới 480px
            if (newWidth < 160) newWidth = 160;
            if (newWidth > 480) newWidth = 480;
            
            sidebar.style.width = newWidth + "px";
            savedWidth = newWidth;
            localStorage.setItem("sidebar-width", newWidth);
        });

        document.addEventListener("mouseup", function() {
            if (isResizing) {
                isResizing = false;
                document.body.style.cursor = "default";
                document.body.style.userSelect = "";
            }
        });
    }

    // Toggle dropdown tài khoản (Top Right)
    const accountBtn = document.getElementById("account-dropdown-btn");
    const accountMenu = document.getElementById("account-dropdown-menu");
    
    if (accountBtn && accountMenu) {
        accountBtn.addEventListener("click", function(e) {
            e.stopPropagation();
            // Đóng settings panel nếu đang mở
            var settingsPanel = document.getElementById("settings-panel");
            if (settingsPanel && !settingsPanel.classList.contains("hidden")) {
                settingsPanel.classList.add("hidden", "scale-95", "opacity-0");
            }
            accountMenu.classList.toggle("hidden");
        });

        document.addEventListener("click", function() {
            accountMenu.classList.add("hidden");
        });
    }

    // Tải Cây Thư mục lần đầu
    if (typeof window.loadExplorerTree === 'function') {
        window.loadExplorerTree();
    }
    
    // Xóa sạch chế độ Focus khi chuyển trang qua HTMX
    document.body.addEventListener('htmx:beforeOnLoad', function(evt) {
        if (evt.detail.target && evt.detail.target.id === 'wiki-content') {
            // Bảo toàn các class settings (theme, text-size, content-width)
            var settingsClasses = [];
            document.body.classList.forEach(function(c) {
                if (c.startsWith('text-size-') || c.startsWith('content-width-') || c.startsWith('theme-')) {
                    settingsClasses.push(c);
                }
            });
            document.body.className = "bg-[#f1f5f9] text-[#202122] min-h-screen flex flex-col overflow-x-hidden";
            settingsClasses.forEach(function(c) { document.body.classList.add(c); });
            
            // Khôi phục layout nguyên bản trước khi nạp trang mới
            const sidebarEl = document.getElementById("wiki-sidebar");
            const resizerEl = document.getElementById("sidebar-resizer");
            const header = document.querySelector('header');
            const mainWrapper = document.querySelector('main');
            
            const isCollapsedVal = localStorage.getItem("sidebar-collapsed") === "true";
            const savedWidthVal = parseInt(localStorage.getItem("sidebar-width")) || 256;
            
            if (sidebarEl) {
                if (isCollapsedVal) {
                    sidebarEl.style.display = 'none';
                } else {
                    sidebarEl.style.display = 'flex';
                    sidebarEl.style.width = savedWidthVal + "px";
                }
            }
            if (resizerEl) {
                resizerEl.style.display = isCollapsedVal ? 'none' : 'block';
            }
            if (header) header.style.display = 'flex';
            // Tôn trọng cài đặt chiều rộng nội dung
            var isWide = document.body.classList.contains('content-width-wide');
            if (mainWrapper) {
                mainWrapper.className = 'flex-grow p-8 w-full mx-auto';
                mainWrapper.style.maxWidth = isWide ? '1600px' : '1350px';
            }
        }
    });

    // --- Back to Top Floating Button Logic ---
    const backToTopBtn = document.getElementById('back-to-top-btn');
    if (backToTopBtn) {
        window.addEventListener('scroll', function() {
            if (window.scrollY > 300) {
                backToTopBtn.classList.remove('hidden');
            } else {
                backToTopBtn.classList.add('hidden');
            }
        });
        
        backToTopBtn.addEventListener('click', function() {
            window.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        });
    }
});
