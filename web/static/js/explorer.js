// explorer.js — Cây thư mục, kéo-thả, menu chuột phải, CRUD node qua API
let selectedNodes = [];
let activeNode = null;

function toggleNodeSelection(node, itemEl) {
    const index = selectedNodes.findIndex(n => n.id === node.id);
    if (index > -1) {
        // Bỏ chọn
        selectedNodes.splice(index, 1);
        itemEl.classList.remove("bg-blue-100/70", "border-l-2", "border-blue-500");
    } else {
        // Chọn
        selectedNodes.push(node);
        itemEl.classList.add("bg-blue-100/70", "border-l-2", "border-blue-500");
    }
}

function clearMultiSelection() {
    selectedNodes = [];
    const items = document.querySelectorAll("#explorer-tree div.group");
    items.forEach(item => {
        item.classList.remove("bg-blue-100/70", "border-l-2", "border-blue-500");
    });
}

// Hàm Tải cây thư mục từ API
function loadExplorerTree() {
    fetch('/api/v1/explorer/tree')
        .then(res => res.json())
        .then(res => {
            const container = document.getElementById("explorer-tree");
            if (res.status === 'success') {
                container.innerHTML = "";
                if (res.data && res.data.length > 0) {
                    renderTreeNodes(res.data, container, 0, null);
                    if (typeof htmx !== 'undefined') {
                        htmx.process(container); // Compile các thuộc tính HTMX vừa tạo động
                    }
                    highlightActivePath(getCurrentActiveTitle()); // Cập nhật trạng thái active ngay sau khi render
                } else {
                    container.innerHTML = "<div class='text-xs text-gray-400 italic p-3 text-center'>No documents yet. Use the '+' buttons above to create one.</div>";
                }
            } else {
                container.innerHTML = "<div class='text-xs text-red-500 italic p-2 text-center'>Failed to load folder tree</div>";
            }
        })
        .catch(() => {
            const treeEl = document.getElementById("explorer-tree");
            if (treeEl) {
                treeEl.innerHTML = "<div class='text-xs text-red-500 italic p-2 text-center'>API connection error</div>";
            }
        });
}
window.loadExplorerTree = loadExplorerTree;

// Tự động tô đen trang đang xem và các thư mục cha khi người dùng đổi trang (HTMX swap)
function highlightActivePath(activeTitle) {
    if (!activeTitle) return;
    const items = document.querySelectorAll("#explorer-tree div.group");
    
    // Reset tất cả các icon và text về trạng thái thường
    items.forEach(item => {
        const icon = item.querySelector("span.rounded-full");
        const link = item.querySelector("a");
        const isFolder = item.querySelector("span.text-gray-400") !== null; // Kiểm tra có caret không
        
        if (icon) {
            icon.className = `w-2 h-2 rounded-full shrink-0 mx-0.5 ${isFolder ? 'bg-red-500' : 'bg-green-500'}`;
        }
        if (link) {
            link.classList.remove("text-black", "font-semibold");
        }
    });

    // Tìm node cần tô active
    let activeItem = null;
    items.forEach(item => {
        const link = item.querySelector("a");
        if (link) {
            const href = link.getAttribute("href");
            const title = decodeURIComponent(href.substring(href.indexOf("/wiki/") === 0 ? 6 : 8));
            if (title === activeTitle) {
                activeItem = item;
            }
        }
    });

    if (activeItem) {
        // Tô đen trang hiện tại
        const link = activeItem.querySelector("a");
        if (link) {
            link.classList.add("text-black", "font-semibold");
        }
        const icon = activeItem.querySelector("span.rounded-full");
        if (icon) {
            icon.className = "w-2 h-2 rounded-full shrink-0 mx-0.5 bg-black";
        }

        // Lần ngược lên trên để tô đen các thư mục cha và tự động mở rộng
        let parentElement = activeItem.parentElement; // nodeEl (chứa itemEl và childrenContainer)
        while (parentElement && parentElement.id !== "explorer-tree") {
            if (parentElement.classList.contains("children-container")) {
                const folderNodeEl = parentElement.parentElement;
                if (folderNodeEl) {
                    const folderItemEl = folderNodeEl.querySelector("div.group");
                    if (folderItemEl) {
                        const folderIcon = folderItemEl.querySelector("span.rounded-full");
                        if (folderIcon) {
                            folderIcon.className = "w-2 h-2 rounded-full shrink-0 mx-0.5 bg-black";
                        }
                        // Tự động hiển thị nếu đang đóng
                        parentElement.style.display = "block";
                        const caret = folderItemEl.querySelector("span.text-gray-400");
                        if (caret) caret.innerHTML = "▼";
                    }
                }
            }
            parentElement = parentElement.parentElement;
        }
    }
}

// Tô đen một node cụ thể theo ID (dùng cho folder khi click)
function highlightActivePathByID(activeId) {
    if (!activeId) return;
    const items = document.querySelectorAll("#explorer-tree div.group");
    
    items.forEach(item => {
        const icon = item.querySelector("span.rounded-full");
        const link = item.querySelector("a");
        const isFolder = item.querySelector("span.text-gray-400") !== null;
        
        if (icon) {
            icon.className = `w-2 h-2 rounded-full shrink-0 mx-0.5 ${isFolder ? 'bg-red-500' : 'bg-green-500'}`;
        }
        if (link) {
            link.classList.remove("text-black", "font-semibold");
        }
    });

    const activeItem = document.querySelector(`#explorer-tree div.group[data-id="${activeId}"]`);
    if (activeItem) {
        const link = activeItem.querySelector("a");
        if (link) {
            link.classList.add("text-black", "font-semibold");
        }
        const icon = activeItem.querySelector("span.rounded-full");
        if (icon) {
            icon.className = "w-2 h-2 rounded-full shrink-0 mx-0.5 bg-black";
        }

        let parentElement = activeItem.parentElement;
        while (parentElement && parentElement.id !== "explorer-tree") {
            if (parentElement.classList.contains("children-container")) {
                const folderNodeEl = parentElement.parentElement;
                if (folderNodeEl) {
                    const folderItemEl = folderNodeEl.querySelector("div.group");
                    if (folderItemEl) {
                        const folderIcon = folderItemEl.querySelector("span.rounded-full");
                        if (folderIcon) {
                            folderIcon.className = "w-2 h-2 rounded-full shrink-0 mx-0.5 bg-black";
                        }
                        parentElement.style.display = "block";
                        const caret = folderItemEl.querySelector("span.text-gray-400");
                        if (caret) caret.innerHTML = "▼";
                    }
                }
            }
            parentElement = parentElement.parentElement;
        }
    }
}

// Xóa sạch trạng thái tô đậm con trỏ active (trả về trạng thái mặc định)
function clearActiveSelection() {
    const items = document.querySelectorAll("#explorer-tree div.group");
    items.forEach(item => {
        const icon = item.querySelector("span.rounded-full");
        const link = item.querySelector("a");
        const isFolder = item.querySelector("span.text-gray-400") !== null;
        
        if (icon) {
            icon.className = `w-2 h-2 rounded-full shrink-0 mx-0.5 ${isFolder ? 'bg-red-500' : 'bg-green-500'}`;
        }
        if (link) {
            link.classList.remove("text-black", "font-semibold");
        }
    });
}

// Lắng nghe sự kiện đổi trang của HTMX để di chuyển "mục sách" (active cursor) theo
document.addEventListener("htmx:afterSwap", function(e) {
    if (e.detail.target.id === "wiki-content") {
        highlightActivePath(getCurrentActiveTitle());
    }
});

// Lấy tiêu đề tài liệu hiện tại từ URL
function getCurrentActiveTitle() {
    const path = window.location.pathname;
    if (path.startsWith("/wiki/")) {
        return decodeURIComponent(path.substring(6));
    } else if (path.startsWith("/editor/")) {
        return decodeURIComponent(path.substring(8));
    }
    return "";
}

// Đệ quy vẽ nút, trả về true nếu nhánh này chứa tài liệu đang xem (active)
function renderTreeNodes(nodes, parentEl, depth, parentNodeId) {
    const currentTitle = getCurrentActiveTitle();
    let hasActiveNodeInBranch = false;

    for (const node of nodes) {
        const nodeEl = document.createElement("div");
        nodeEl.className = "flex flex-col select-none";
        
        const itemEl = document.createElement("div");
        itemEl.className = "group flex items-center justify-between py-1 px-1.5 hover:bg-gray-100 rounded cursor-pointer transition text-gray-700";
        itemEl.style.paddingLeft = (depth * 10 + 6) + "px";
        itemEl.setAttribute("data-id", node.id); // Lưu trữ ID tài liệu để dễ truy vấn định vị
        itemEl.setAttribute("data-is-folder", node.is_folder ? "true" : "false");
        itemEl.setAttribute("data-parent-id", parentNodeId || "");

        // Re-apply selected visual state if the node was selected
        const selIndex = selectedNodes.findIndex(n => n.id === node.id);
        if (selIndex > -1) {
            selectedNodes[selIndex] = node; // update metadata reference
            itemEl.classList.add("bg-blue-100/70", "border-l-2", "border-blue-500");
        }

        const leftEl = document.createElement("div");
        leftEl.className = "flex items-center gap-1.5 min-w-0 text-[13px]";

        const rightEl = document.createElement("div");
        rightEl.className = "flex items-center gap-1 opacity-0 group-hover:opacity-100 transition shrink-0 ml-2";

        // Bắt sự kiện chuột phải trên mỗi nút trong cây thư mục
        itemEl.addEventListener("contextmenu", function(e) {
            e.preventDefault();
            e.stopPropagation();
            showContextMenu(e, node);
        });

        // Hỗ trợ Kéo & Thả (Drag & Drop) — Chặn kéo node bị khoá (trừ admin)
        const isUserAdmin = (document.body.getAttribute("data-user-role") || "") === "admin";
        const canDrag = !node.is_locked || isUserAdmin;
        itemEl.setAttribute("draggable", canDrag ? "true" : "false");
        if (!canDrag) {
            itemEl.style.cursor = "default";
        }

        itemEl.addEventListener("dragstart", function(e) {
            if (!canDrag) { e.preventDefault(); return; }
            e.stopPropagation();

            const isSelected = selectedNodes.some(n => n.id === node.id);
            let draggedIds = [node.id];
            if (isSelected) {
                draggedIds = selectedNodes.map(n => n.id);
            }

            e.dataTransfer.setData("text/plain", JSON.stringify(draggedIds));
            e.dataTransfer.effectAllowed = "move";
            
            // Visual indicator
            draggedIds.forEach(id => {
                const el = document.querySelector(`#explorer-tree div.group[data-id="${id}"]`);
                if (el) el.classList.add("opacity-40", "border-dashed", "border", "border-blue-500");
            });
        });

        itemEl.addEventListener("dragend", function(e) {
            const els = document.querySelectorAll("#explorer-tree div.group.opacity-40");
            els.forEach(el => {
                el.classList.remove("opacity-40", "border-dashed", "border", "border-blue-500");
            });
        });

        itemEl.addEventListener("dragover", function(e) {
            e.preventDefault();
            e.stopPropagation();
            // Chặn thả vào thư mục bị khoá (trừ admin, bỏ qua unsorted_bin_folder)
            if (node.is_folder && node.is_locked && !isUserAdmin && node.id !== "unsorted_bin_folder") {
                e.dataTransfer.dropEffect = "none";
                itemEl.classList.add("bg-red-50/50");
                return;
            }

            // Đo vị trí tương đối
            const rect = itemEl.getBoundingClientRect();
            const relY = e.clientY - rect.top;
            const height = rect.height;

            // Xóa các chỉ báo viền cũ
            itemEl.classList.remove("drag-hover-before", "drag-hover-after", "bg-blue-50", "border-blue-300", "border-l-2");

            if (relY < height * 0.25) {
                itemEl.classList.add("drag-hover-before");
            } else if (relY > height * 0.75) {
                itemEl.classList.add("drag-hover-after");
            } else {
                if (node.is_folder) {
                    itemEl.classList.add("bg-blue-50", "border-blue-300", "border-l-2");
                } else {
                    itemEl.classList.add("drag-hover-after");
                }
            }
        });

        itemEl.addEventListener("dragleave", function(e) {
            itemEl.classList.remove("bg-blue-50", "border-blue-300", "border-l-2", "bg-red-50/50", "drag-hover-before", "drag-hover-after");
        });

        itemEl.addEventListener("drop", function(e) {
            e.preventDefault();
            e.stopPropagation();
            itemEl.classList.remove("bg-blue-50", "border-blue-300", "border-l-2", "bg-red-50/50", "drag-hover-before", "drag-hover-after");

            const rect = itemEl.getBoundingClientRect();
            const relY = e.clientY - rect.top;
            const height = rect.height;

            let dropPosition = "inside";
            if (relY < height * 0.25) {
                dropPosition = "before";
            } else if (relY > height * 0.75) {
                dropPosition = "after";
            } else if (!node.is_folder) {
                dropPosition = "after";
            }

            // Chặn thả vào thư mục bị khoá (trừ admin, bỏ qua unsorted_bin_folder)
            const targetParentId = (node.is_folder && dropPosition === "inside") ? node.id : parentNodeId;
            if (node.is_folder && node.is_locked && !isUserAdmin && node.id !== "unsorted_bin_folder" && dropPosition === "inside") {
                window.showToast("The destination folder is locked. Cannot move here.", "error");
                return;
            }

            const rawData = e.dataTransfer.getData("text/plain");
            if (!rawData) return;

            let draggedIds = [];
            try {
                draggedIds = JSON.parse(rawData);
                if (!Array.isArray(draggedIds)) {
                    draggedIds = [draggedIds];
                }
            } catch (err) {
                draggedIds = [rawData];
            }

            // Lọc ra các ID để tránh tự xếp trước/sau chính nó
            draggedIds = draggedIds.filter(id => id !== targetParentId && id !== node.id);
            if (draggedIds.length === 0) return;

            moveNodes(draggedIds, targetParentId, node.id, dropPosition);
        });

        let isThisNodeActive = false;

        if (node.is_folder) {
            // Check if children contain the active node
            const childrenContainer = document.createElement("div");
            childrenContainer.className = "children-container";

            let hasActiveChild = false;
            if (node.children && node.children.length > 0) {
                hasActiveChild = renderTreeNodes(node.children, childrenContainer, depth + 1, node.id);
            }

            const isExpandedFromStorage = localStorage.getItem("folder-expanded-" + node.id) !== "false";
            // Tự động mở rộng nếu chứa trang active, ngược lại dựa theo localStorage
            const isExpanded = hasActiveChild || isExpandedFromStorage;
            childrenContainer.style.display = isExpanded ? "block" : "none";

            // Mũi tên Caret
            const caret = document.createElement("span");
            caret.className = "text-gray-400 text-[9px] w-3 text-center transition-transform";
            caret.innerHTML = isExpanded ? "▼" : "▶";
            leftEl.appendChild(caret);

            if (hasActiveChild) {
                isThisNodeActive = true;
                hasActiveNodeInBranch = true;
            }

            // Icon thư mục (Tròn đỏ, hoặc Tròn đen nếu có mục con đang active)
            const icon = document.createElement("span");
            icon.className = `w-2 h-2 rounded-full shrink-0 mx-0.5 ${isThisNodeActive ? 'bg-black' : 'bg-red-500'}`;
            leftEl.appendChild(icon);

            // Icon ổ khoá SVG (nếu bị khoá)
            if (node.is_locked) {
                const lockIcon = document.createElement("span");
                lockIcon.innerHTML = '<svg class="w-3 h-3 text-amber-500" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>';
                lockIcon.className = "shrink-0 flex items-center";
                leftEl.appendChild(lockIcon);
            }

            // Tên thư mục
            const title = document.createElement("span");
            title.className = "font-medium truncate";
            title.textContent = node.title;
            leftEl.appendChild(title);

            // Click thư mục để đóng mở & Cập nhật tiêu điểm con trỏ định vị sang thư mục này
            itemEl.addEventListener("click", function(e) {
                if (e.ctrlKey || e.metaKey) {
                    e.preventDefault();
                    e.stopPropagation();
                    toggleNodeSelection(node, itemEl);
                    return;
                }
                clearMultiSelection();
                const nowExpanded = childrenContainer.style.display === "none";
                childrenContainer.style.display = nowExpanded ? "block" : "none";
                caret.innerHTML = nowExpanded ? "▼" : "▶";
                localStorage.setItem("folder-expanded-" + node.id, nowExpanded ? "true" : "false");
                
                highlightActivePathByID(node.id);
            });

            // Nút Tạo File trong thư mục này
            const addFileBtn = document.createElement("button");
            addFileBtn.className = "hover:bg-gray-300 text-[10px] rounded px-1 transition text-gray-600 font-bold";
            addFileBtn.title = "Create new page";
            addFileBtn.textContent = "+📄";
            addFileBtn.addEventListener("click", function(e) {
                e.stopPropagation();
                createExplorerNode(node.id, false);
            });

            // Nút Tạo Thư mục con trong thư mục này
            const addFolderBtn = document.createElement("button");
            addFolderBtn.className = "hover:bg-gray-300 text-[10px] rounded px-1 transition text-gray-600 font-bold";
            addFolderBtn.title = "Create subfolder";
            addFolderBtn.textContent = "+📁";
            addFolderBtn.addEventListener("click", function(e) {
                e.stopPropagation();
                createExplorerNode(node.id, true);
            });

            rightEl.appendChild(addFileBtn);
            rightEl.appendChild(addFolderBtn);

            itemEl.appendChild(leftEl);
            itemEl.appendChild(rightEl);
            nodeEl.appendChild(itemEl);

            if (!node.children || node.children.length === 0) {
                const emptyEl = document.createElement("div");
                emptyEl.className = "text-[11px] text-gray-400 italic py-0.5";
                emptyEl.style.paddingLeft = ((depth + 1) * 10 + 20) + "px";
                emptyEl.textContent = "(Empty folder)";
                childrenContainer.appendChild(emptyEl);
            }
            nodeEl.appendChild(childrenContainer);
        } else {
            if (node.title === currentTitle) {
                isThisNodeActive = true;
                hasActiveNodeInBranch = true;
            }

            // Khoảng trống căn hàng cho file
            const spacer = document.createElement("span");
            spacer.className = "w-3";
            leftEl.appendChild(spacer);

            // Icon file (Tròn xanh, hoặc Tròn đen nếu là trang active)
            const icon = document.createElement("span");
            icon.className = `w-2 h-2 rounded-full shrink-0 mx-0.5 ${isThisNodeActive ? 'bg-black' : 'bg-green-500'}`;
            leftEl.appendChild(icon);

            // Icon ổ khoá SVG (nếu bị khoá)
            if (node.is_locked) {
                const lockIcon = document.createElement("span");
                lockIcon.innerHTML = '<svg class="w-3 h-3 text-amber-500" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>';
                lockIcon.className = "shrink-0 flex items-center";
                leftEl.appendChild(lockIcon);
            }

            // Liên kết bài viết
            const titleLink = document.createElement("a");
            titleLink.href = "/wiki/" + encodeURIComponent(node.title);
            titleLink.className = "truncate hover:text-blue-600 hover:underline";
            titleLink.textContent = node.title;
            titleLink.setAttribute("hx-get", "/wiki/" + encodeURIComponent(node.title));
            titleLink.setAttribute("hx-target", "#wiki-content");
            titleLink.setAttribute("hx-swap", "innerHTML");
            titleLink.setAttribute("hx-push-url", "true");
            titleLink.addEventListener("click", e => e.stopPropagation());

            if (isThisNodeActive) {
                titleLink.classList.add("text-black", "font-semibold");
            }

            leftEl.appendChild(titleLink);

            itemEl.appendChild(leftEl);
            nodeEl.appendChild(itemEl);

            // Click cả dòng của Trang tài liệu để kích hoạt link tải HTMX
            itemEl.addEventListener("click", function(e) {
                if (e.ctrlKey || e.metaKey) {
                    e.preventDefault();
                    e.stopPropagation();
                    toggleNodeSelection(node, itemEl);
                    return;
                }
                clearMultiSelection();
                titleLink.click();
            });
        }
        
        parentEl.appendChild(nodeEl);
    }

    return hasActiveNodeInBranch;
}

// Lấy Parent ID dựa trên vị trí hiện tại của con trỏ active (mục sách)
function getActiveParentID() {
    const items = Array.from(document.querySelectorAll("#explorer-tree div.group")).reverse();
    for (const item of items) {
        const icon = item.querySelector("span.rounded-full");
        if (icon && icon.classList.contains("bg-black")) {
            const isFolder = item.getAttribute("data-is-folder") === "true";
            const id = item.getAttribute("data-id");
            const parentId = item.getAttribute("data-parent-id");
            
            if (isFolder) {
                return id; // Nếu con trỏ đang ở thư mục, tạo bên trong thư mục đó
            } else {
                return parentId || null; // Nếu con trỏ đang ở trang, tạo cùng cấp thư mục với trang đó
            }
        }
    }
    return null; // Mặc định tạo ở gốc (root)
}

// Tạo Node mới sử dụng ngữ cảnh vị trí của con trỏ active
function createNodeAtActiveParent(isFolder) {
    const parentId = getActiveParentID();
    createExplorerNode(parentId, isFolder);
}
window.createNodeAtActiveParent = createNodeAtActiveParent;

// Tạo Node mới qua API và dùng các Custom Dialogs
async function createExplorerNode(parentId, isFolder) {
    const typeStr = isFolder ? "folder" : "page";
    const title = await window.showCustomPrompt("➕ Create new", "Enter a name for the new " + typeStr + ":");
    if (!title) return;

    fetch('/api/v1/explorer/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            title: title,
            parent_id: parentId,
            is_folder: isFolder
        })
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === 'success') {
            loadExplorerTree();
            if (!isFolder) {
                // Nếu là file, chuyển hướng tới trang chỉnh sửa ngay lập tức!
                window.location.href = "/editor/" + encodeURIComponent(data.data.Title);
            }
        } else {
            window.showToast(data.message, "error");
        }
    })
    .catch(err => window.showToast("Could not connect to the server.", "error"));
}

// Gọi API di chuyển Node (hỗ trợ hàng loạt & sắp xếp thứ tự)
function moveNodes(ids, parentId, targetId = null, position = "inside") {
    fetch('/api/v1/explorer/move', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            ids: ids,
            parent_id: parentId,
            target_id: targetId,
            position: position
        })
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === 'success') {
            clearMultiSelection();
            loadExplorerTree();
        } else {
            window.showToast("Move failed: " + data.message, "error");
        }
    })
    .catch(() => window.showToast("Could not connect to the server.", "error"));
}

// --- Logic Custom Context Menu ---
function showContextMenu(e, node) {
    // Check if the right-clicked node is already in the selection.
    // If not, clear selection and select ONLY this node
    const isSelected = selectedNodes.some(n => n.id === node.id);
    if (!isSelected) {
        clearMultiSelection();
        const itemEl = document.querySelector(`#explorer-tree div.group[data-id="${node.id}"]`);
        if (itemEl) {
            toggleNodeSelection(node, itemEl);
        }
    }

    activeNode = node;
    const menu = document.getElementById("explorer-context-menu");
    if (!menu) return;

    const isUserAdmin = (document.body.getAttribute("data-user-role") || "") === "admin";
    const lockBtn = document.getElementById("ctx-lock");
    const unlockBtn = document.getElementById("ctx-unlock");
    const renameBtn = document.getElementById("ctx-rename");
    const deleteBtn = document.getElementById("ctx-delete");
    const reportBtn = document.getElementById("ctx-report");

    const count = selectedNodes.length;
    const hasLockedNode = selectedNodes.some(n => n.is_locked);

    if (count > 1) {
        // Bulk action mode
        if (renameBtn) renameBtn.classList.add("hidden"); // Disable rename in bulk

        if (deleteBtn) {
            deleteBtn.querySelector("span:last-child").textContent = `Delete ${count} selected`;
            if (hasLockedNode && !isUserAdmin) {
                deleteBtn.classList.add("hidden");
            } else {
                deleteBtn.classList.remove("hidden");
            }
        }
        if (lockBtn) {
            lockBtn.querySelector("span:last-child").textContent = `Lock ${count} selected`;
        }
        if (unlockBtn) {
            unlockBtn.querySelector("span:last-child").textContent = `Unlock ${count} selected`;
        }
        if (reportBtn) {
            reportBtn.querySelector("span:last-child").textContent = `Report ${count} selected`;
        }

        if (isUserAdmin) {
            lockBtn.classList.remove("hidden");
            unlockBtn.classList.remove("hidden");
        } else {
            lockBtn.classList.add("hidden");
            unlockBtn.classList.add("hidden");
        }
    } else {
        // Single action mode
        const isNodeLocked = node.is_locked;

        if (renameBtn) {
            renameBtn.querySelector("span:last-child").textContent = "Rename";
            if (isNodeLocked && !isUserAdmin) {
                renameBtn.classList.add("hidden");
            } else {
                renameBtn.classList.remove("hidden");
            }
        }
        if (deleteBtn) {
            deleteBtn.querySelector("span:last-child").textContent = "Delete folder/page";
            if (isNodeLocked && !isUserAdmin) {
                deleteBtn.classList.add("hidden");
            } else {
                deleteBtn.classList.remove("hidden");
            }
        }
        if (lockBtn) {
            lockBtn.querySelector("span:last-child").textContent = "Lock";
        }
        if (unlockBtn) {
            unlockBtn.querySelector("span:last-child").textContent = "Unlock";
        }
        if (reportBtn) {
            reportBtn.querySelector("span:last-child").textContent = "Report";
        }

        if (isUserAdmin) {
            if (isNodeLocked) {
                lockBtn.classList.add("hidden");
                unlockBtn.classList.remove("hidden");
            } else {
                lockBtn.classList.remove("hidden");
                unlockBtn.classList.add("hidden");
            }
        } else {
            lockBtn.classList.add("hidden");
            unlockBtn.classList.add("hidden");
        }
    }

    // Hiển thị tạm thời để đo kích thước
    menu.classList.remove("hidden");
    const menuWidth = menu.offsetWidth || 200;
    const menuHeight = menu.offsetHeight || 240;

    let posX = e.clientX;
    let posY = e.clientY;

    // Tránh tràn viền phải
    if (posX + menuWidth > window.innerWidth) {
        posX = window.innerWidth - menuWidth - 10;
    }
    // Tránh tràn viền dưới
    if (posY + menuHeight > window.innerHeight) {
        posY = window.innerHeight - menuHeight - 10;
    }

    if (posX < 0) posX = 0;
    if (posY < 0) posY = 0;

    menu.style.left = (posX + window.scrollX) + "px";
    menu.style.top = (posY + window.scrollY) + "px";

    // Animation hiển thị mượt mà
    setTimeout(() => {
        menu.classList.remove("scale-95", "opacity-0");
        menu.classList.add("scale-100", "opacity-100");
    }, 10);
}

function hideContextMenu() {
    const menu = document.getElementById("explorer-context-menu");
    if (menu && !menu.classList.contains("hidden")) {
        menu.classList.remove("scale-100", "opacity-100");
        menu.classList.add("scale-95", "opacity-0");
        setTimeout(() => {
            menu.classList.add("hidden");
        }, 150);
    }
}

document.addEventListener("click", hideContextMenu);
window.addEventListener("scroll", hideContextMenu);
window.addEventListener("resize", hideContextMenu);
document.addEventListener("keydown", function(e) {
    if (e.key === "Escape") hideContextMenu();
});

// Đăng ký sự kiện thả rơi cấp thư mục gốc trên container
document.addEventListener("DOMContentLoaded", function() {
    const treeContainer = document.getElementById("explorer-tree");
    if (treeContainer) {
        treeContainer.addEventListener("dragover", function(e) {
            e.preventDefault();
            treeContainer.classList.add("bg-blue-50/30");
        });

        treeContainer.addEventListener("dragleave", function(e) {
            treeContainer.classList.remove("bg-blue-50/30");
        });

        treeContainer.addEventListener("drop", function(e) {
            e.preventDefault();
            treeContainer.classList.remove("bg-blue-50/30");
            
            if (e.target === treeContainer || e.target.id === "explorer-tree") {
                const rawData = e.dataTransfer.getData("text/plain");
                if (rawData) {
                    let draggedIds = [];
                    try {
                        draggedIds = JSON.parse(rawData);
                        if (!Array.isArray(draggedIds)) {
                            draggedIds = [draggedIds];
                        }
                    } catch (err) {
                        draggedIds = [rawData];
                    }
                    moveNodes(draggedIds, null, null, "inside");
                }
            }
        });

        treeContainer.addEventListener("click", function(e) {
            if (e.target === treeContainer) {
                clearActiveSelection();
            }
        });
    }

    // Đăng ký sự kiện các nút Context Menu
    const renameBtn = document.getElementById("ctx-rename");
    const deleteBtn = document.getElementById("ctx-delete");
    const lockBtn = document.getElementById("ctx-lock");
    const unlockBtn = document.getElementById("ctx-unlock");
    const reportBtn = document.getElementById("ctx-report");

    if (renameBtn) {
        renameBtn.addEventListener("click", async function() {
            if (selectedNodes.length !== 1) return;
            const node = selectedNodes[0];
            const typeStr = node.is_folder ? "folder" : "page";
            const newTitle = await window.showCustomPrompt("✏️ Rename", "Enter a new name for this " + typeStr + ":", node.title);
            if (!newTitle || newTitle === node.title) return;

            fetch('/api/v1/explorer/rename', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: node.id, title: newTitle })
            })
            .then(res => res.json())
            .then(data => {
                if (data.status === 'success') {
                    const currentPath = window.location.pathname;
                    const oldUrlPath = "/wiki/" + encodeURIComponent(node.title);
                    const oldEditorPath = "/editor/" + encodeURIComponent(node.title);
                    
                    clearMultiSelection();
                    loadExplorerTree();
                    
                    if (currentPath === oldUrlPath) {
                        window.location.href = "/wiki/" + encodeURIComponent(newTitle);
                    } else if (currentPath === oldEditorPath) {
                        window.location.href = "/editor/" + encodeURIComponent(newTitle);
                    }
                } else {
                    window.showToast(data.message, "error");
                }
            })
            .catch(() => window.showToast("Could not connect to the server.", "error"));
        });
    }

    if (deleteBtn) {
        deleteBtn.addEventListener("click", async function() {
            if (selectedNodes.length === 0) return;
            const ids = selectedNodes.map(n => n.id);
            const isBulk = ids.length > 1;
            const confirmMsg = isBulk
                ? `Delete ${ids.length} selected items?\nThis cannot be undone.`
                : `Delete ${selectedNodes[0].is_folder ? "folder" : "page"} '${selectedNodes[0].title}'?\nThis cannot be undone.`;

            if (!(await window.showCustomConfirm("🗑️ Delete selected", confirmMsg))) return;

            fetch('/api/v1/explorer/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ids: ids })
            })
            .then(res => res.json())
            .then(data => {
                if (data.status === 'success') {
                    const currentPath = window.location.pathname;
                    const shouldRedirect = selectedNodes.some(n => {
                        const oldUrlPath = "/wiki/" + encodeURIComponent(n.title);
                        const oldEditorPath = "/editor/" + encodeURIComponent(n.title);
                        return currentPath === oldUrlPath || currentPath === oldEditorPath;
                    });

                    clearMultiSelection();
                    loadExplorerTree();

                    if (shouldRedirect) {
                        window.location.href = "/wiki/Home";
                    }
                } else {
                    window.showToast(data.message, "error");
                }
            })
            .catch(() => window.showToast("Could not connect to the server.", "error"));
        });
    }

    function setLockStatus(locked) {
        if (selectedNodes.length === 0) return;
        const ids = selectedNodes.map(n => n.id);

        fetch('/api/v1/explorer/lock', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ids: ids, locked: locked })
        })
        .then(res => res.json())
        .then(data => {
            if (data.status === 'success') {
                clearMultiSelection();
                loadExplorerTree();
                window.location.reload();
            } else {
                window.showToast(data.message, "error");
            }
        })
        .catch(() => window.showToast("Could not connect to the server.", "error"));
    }

    if (lockBtn) lockBtn.addEventListener("click", () => setLockStatus(true));
    if (unlockBtn) unlockBtn.addEventListener("click", () => setLockStatus(false));

    if (reportBtn) {
        reportBtn.addEventListener("click", async function() {
            if (selectedNodes.length === 0) return;
            const ids = selectedNodes.map(n => n.id);
            const isBulk = ids.length > 1;
            const promptMsg = isBulk
                ? `Enter a reason for reporting these ${ids.length} documents:`
                : `Enter a reason for reporting this document:`;
            const reason = await window.showCustomPrompt("🚨 Report document", promptMsg);
            if (!reason) return;

            fetch('/api/v1/explorer/report', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ids: ids, reason: reason })
            })
            .then(res => res.json())
            .then(data => {
                window.showToast(data.message, "success");
                clearMultiSelection();
                loadExplorerTree();
            })
            .catch(() => window.showToast("Could not connect to the server.", "error"));
        });
    }
});
