// search.js — Tìm kiếm nhanh (Ctrl+K, /) và autocomplete gợi ý
document.addEventListener("DOMContentLoaded", function() {
    const searchInput = document.getElementById("search-input");
    const suggestionsBox = document.getElementById("search-suggestions");
    const searchForm = document.getElementById("search-form");
    let activeSuggestionIndex = -1;

    if (!searchInput || !suggestionsBox) return;

    // Phím tắt Ctrl+K và / để focus ô tìm kiếm
    document.addEventListener("keydown", function(e) {
        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
            e.preventDefault();
            searchInput.focus();
        } else if (e.key === '/' && document.activeElement !== searchInput && document.activeElement.tagName !== 'INPUT' && document.activeElement.tagName !== 'TEXTAREA') {
            e.preventDefault();
            searchInput.focus();
        }
    });

    // Đóng gợi ý khi click ra ngoài
    document.addEventListener("click", function(e) {
        if (searchForm && !searchForm.contains(e.target)) {
            suggestionsBox.classList.add("hidden");
        }
    });

    // Lấy trang từ cây explorer (tìm client-side)
    function getPagesFromTree() {
        const links = document.querySelectorAll("#explorer-tree a");
        return Array.from(links).map(link => {
            const item = link.closest("div.group");
            const isFolder = item ? item.getAttribute("data-is-folder") === "true" : false;
            return {
                title: link.textContent.trim(),
                href: link.getAttribute("href"),
                isFolder: isFolder
            };
        });
    }

    searchInput.addEventListener("focus", function() {
        showSuggestions(searchInput.value);
    });

    searchInput.addEventListener("input", function() {
        showSuggestions(searchInput.value);
    });

    // Điều hướng gợi ý bằng phím
    searchInput.addEventListener("keydown", function(e) {
        const items = suggestionsBox.querySelectorAll(".suggestion-item");
        if (suggestionsBox.classList.contains("hidden") || items.length === 0) return;

        if (e.key === "ArrowDown") {
            e.preventDefault();
            activeSuggestionIndex = (activeSuggestionIndex + 1) % items.length;
            highlightSuggestion(items);
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            activeSuggestionIndex = (activeSuggestionIndex - 1 + items.length) % items.length;
            highlightSuggestion(items);
        } else if (e.key === "Enter") {
            if (activeSuggestionIndex > -1 && activeSuggestionIndex < items.length) {
                e.preventDefault();
                items[activeSuggestionIndex].click();
            }
        } else if (e.key === "Escape") {
            suggestionsBox.classList.add("hidden");
            searchInput.blur();
        }
    });

    function highlightSuggestion(items) {
        items.forEach((item, index) => {
            if (index === activeSuggestionIndex) {
                item.classList.add("bg-blue-50", "text-blue-600");
                item.scrollIntoView({ block: "nearest" });
            } else {
                item.classList.remove("bg-blue-50", "text-blue-600");
            }
        });
    }

    let cachedTags = null;
    function fetchTagsIfNeeded() {
        if (cachedTags !== null) return Promise.resolve(cachedTags);
        return fetch('/api/v1/tags')
            .then(res => res.json())
            .then(res => {
                if (res.status === 'success' && res.data) {
                    cachedTags = res.data;
                    return cachedTags;
                }
                return [];
            })
            .catch(() => []);
    }

    function showSuggestions(val) {
        const query = val.trim().toLowerCase();
        suggestionsBox.innerHTML = "";
        activeSuggestionIndex = -1;

        if (val.length === 0) {
            suggestionsBox.classList.add("hidden");
            return;
        }

        // Gợi ý tag khi gõ #
        const words = val.split(/\s+/);
        const currentWord = words[words.length - 1];

        if (currentWord.startsWith("#")) {
            const tagSearch = currentWord.substring(1).toLowerCase();
            fetchTagsIfNeeded().then(tags => {
                const matches = tags.filter(t => t.name.toLowerCase().includes(tagSearch)).slice(0, 8);
                
                if (matches.length === 0) {
                    suggestionsBox.classList.add("hidden");
                    return;
                }

                matches.forEach(t => {
                    const itemEl = document.createElement("button");
                    itemEl.type = "button";
                    itemEl.className = "suggestion-item w-full text-left block px-4 py-2.5 hover:bg-blue-50 text-gray-700 transition flex items-center justify-between border-b border-gray-100 last:border-b-0";
                    itemEl.innerHTML = `
                        <div class="flex items-center gap-2 truncate">
                            <span class="text-xs bg-blue-50 text-blue-600 px-1.5 py-0.5 rounded font-mono">#</span>
                            <span class="truncate font-medium text-blue-600">${t.name}</span>
                        </div>
                        <span class="text-[10px] text-gray-400 font-mono">Complete tag</span>
                    `;

                    itemEl.addEventListener("click", function(e) {
                        e.preventDefault();
                        e.stopPropagation();
                        
                        words[words.length - 1] = "#" + t.name;
                        searchInput.value = words.join(" ") + " ";
                        searchInput.focus();
                        suggestionsBox.classList.add("hidden");
                    });

                    suggestionsBox.appendChild(itemEl);
                });
                suggestionsBox.classList.remove("hidden");
            });
            return;
        }

        // Gợi ý trang khớp
        const pages = getPagesFromTree();
        const matches = pages.filter(p => p.title.toLowerCase().includes(query)).slice(0, 8);

        if (matches.length === 0) {
            const noResultEl = document.createElement("div");
            noResultEl.className = "px-4 py-2.5 text-gray-400 italic text-center select-none";
            noResultEl.textContent = "No matching results";
            suggestionsBox.appendChild(noResultEl);
        } else {
            matches.forEach(p => {
                const itemEl = document.createElement("a");
                itemEl.href = p.href;
                itemEl.className = "suggestion-item block px-4 py-2.5 hover:bg-blue-50 text-gray-700 transition flex items-center justify-between border-b border-gray-100 last:border-b-0";
                
                const icon = p.isFolder ? "📁" : "📄";
                itemEl.innerHTML = `
                    <div class="flex items-center gap-2 truncate">
                        <span class="text-xs opacity-60">${icon}</span>
                        <span class="truncate">${p.title}</span>
                    </div>
                    <span class="text-[10px] text-gray-400 font-mono">Go to</span>
                `;
                
                itemEl.setAttribute("hx-get", p.href);
                itemEl.setAttribute("hx-target", "#wiki-content");
                itemEl.setAttribute("hx-swap", "innerHTML");
                itemEl.setAttribute("hx-push-url", "true");
                
                itemEl.addEventListener("click", function() {
                    suggestionsBox.classList.add("hidden");
                    searchInput.value = p.title;
                });
                
                suggestionsBox.appendChild(itemEl);
            });
            if (typeof htmx !== 'undefined') {
                htmx.process(suggestionsBox);
            }
        }

        suggestionsBox.classList.remove("hidden");
    }
});
