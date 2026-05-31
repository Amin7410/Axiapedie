window.formatMarkdown = function(prefix, suffix) {
    const textarea = document.getElementById('content');
    if (!textarea) return;

    textarea.focus();
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    // Use execCommand to preserve undo/redo history (Ctrl+Z / Ctrl+Y)
    const replacement = prefix + selectedText + suffix;
    document.execCommand('insertText', false, replacement);

    // Place cursor after inserted text (before suffix)
    const newCursorPos = start + prefix.length + selectedText.length;
    textarea.setSelectionRange(newCursorPos, newCursorPos);

    textarea.style.height = 'auto';
    textarea.style.height = textarea.scrollHeight + 'px';
};

(function() {
    // Hotkeys handler
    const textarea = document.getElementById('content');
    if (textarea) {
        textarea.addEventListener('keydown', function(e) {
            if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') {
                e.preventDefault();
                window.formatMarkdown('**', '**');
            } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'i') {
                e.preventDefault();
                window.formatMarkdown('*', '*');
            }
        });
    }

    // AJAX Upload ảnh
    const imageUpload = document.getElementById('imageUpload');
    if (imageUpload) {
        imageUpload.addEventListener('change', function(e) {
            if(e.target.files.length === 0) return;
            const file = e.target.files[0];
            
            const formData = new FormData();
            formData.append('file', file);
            
            window.showToast("Uploading image...", "warning");

            fetch('/api/v1/media/upload', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if(data.status === 'success') {
                    window.showToast("Image uploaded successfully.");
                    
                    // Chèn markdown snippet vào textarea (dùng execCommand để giữ undo/redo)
                    const textarea = document.getElementById('content');
                    if (textarea) {
                        textarea.focus();
                        const snippet = "\n" + data.data.markdown_snippet + "\n";
                        document.execCommand('insertText', false, snippet);
                        
                        // Cập nhật lại chiều cao của khung soạn thảo
                        if (typeof autoExpand === 'function') {
                            autoExpand();
                        } else {
                            textarea.style.height = 'auto';
                            textarea.style.height = textarea.scrollHeight + 'px';
                        }
                    }
                    
                    // Reset file input
                    e.target.value = '';
                } else {
                    window.showToast("Error: " + data.message, "error");
                }
            })
            .catch(error => {
                window.showToast("Image upload failed.", "error");
            });
        });
    }
})();

(function() {
    // --- 1. Inline Title Rename Logic ---
    const titleInput = document.getElementById('document-title-input');
    if (titleInput) {
        const handleRename = () => {
            const docId = titleInput.getAttribute('data-doc-id');
            const origTitle = titleInput.getAttribute('data-orig-title');
            const newTitle = titleInput.value.trim();
            if (!newTitle || newTitle === origTitle) {
                titleInput.value = origTitle;
                return;
            }
            
            if (!docId) {
                titleInput.setAttribute('data-orig-title', newTitle);
                const hiddenTitle = document.querySelector('input[name="title"]');
                if (hiddenTitle) hiddenTitle.value = newTitle;
                return;
            }

            fetch('/api/v1/explorer/rename', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: docId, title: newTitle })
            })
            .then(res => res.json())
            .then(data => {
                if (data.status === 'success') {
                    titleInput.setAttribute('data-orig-title', newTitle);
                    const hiddenTitle = document.querySelector('input[name="title"]');
                    if (hiddenTitle) hiddenTitle.value = newTitle;
                    
                    // Update URL
                    const newUrl = '/editor/' + encodeURIComponent(newTitle);
                    window.history.replaceState(null, '', newUrl);
                    
                    // Update Read Button URLs dynamically
                    const readBtn = document.getElementById('read-document-btn');
                    if (readBtn) {
                        const newReadUrl = '/wiki/' + encodeURIComponent(newTitle);
                        readBtn.setAttribute('href', newReadUrl);
                        readBtn.setAttribute('hx-get', newReadUrl);
                        htmx.process(readBtn);
                    }
                    
                    // Reload tree
                    if (typeof loadExplorerTree === 'function') {
                        loadExplorerTree();
                    }
                } else {
                    window.showCustomAlert("Error", "Rename failed: " + data.message);
                    titleInput.value = origTitle;
                }
            })
            .catch(() => {
                window.showCustomAlert("Error", "Could not connect to the server while renaming.");
                titleInput.value = origTitle;
            });
        };

        titleInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                titleInput.blur();
            }
        });
        titleInput.addEventListener('blur', handleRename);
    }

    // Sync subtitle input with hidden form field + auto-expand height
    const subtitleInput = document.getElementById('document-subtitle-input');
    if (subtitleInput) {
        const autoExpandSubtitle = () => {
            const scrollPos = window.scrollY;
            subtitleInput.style.height = 'auto';
            subtitleInput.style.height = subtitleInput.scrollHeight + 'px';
            window.scrollTo(window.scrollX, scrollPos);
        };
        subtitleInput.addEventListener('input', () => {
            const hiddenSubtitle = document.getElementById('subtitle-hidden-input');
            if (hiddenSubtitle) {
                hiddenSubtitle.value = subtitleInput.value;
            }
            autoExpandSubtitle();
        });
        // Auto-expand on page load
        autoExpandSubtitle();
    }

    // --- 2. Advanced Tag Pool System ---
    const hiddenTagsInput = document.getElementById("tags-hidden-input");
    const rawTags = hiddenTagsInput ? hiddenTagsInput.value : "";
    let activeTags = rawTags.split(',').map(t => t.trim()).filter(t => t !== '');
    let tagPool = [];

    const selectedTagsContainer = document.getElementById('selected-tags-container');
    const dropdownList = document.getElementById('tag-pool-dropdown');
    const addTagBtn = document.getElementById('add-tag-btn');

    if (addTagBtn && dropdownList) {
        addTagBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            dropdownList.classList.toggle('hidden');
        });
        document.addEventListener('click', () => {
            dropdownList.classList.add('hidden');
        });
    }

    function renderTags() {
        selectedTagsContainer.innerHTML = '';
        if (activeTags.length === 0) {
            selectedTagsContainer.innerHTML = '<span class="text-xs text-gray-400 italic py-1">No tags selected</span>';
        } else {
            activeTags.forEach(tag => {
                const pill = document.createElement('span');
                pill.className = 'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-50 text-blue-600 border border-blue-100 select-none';
                pill.innerHTML = `
                    #${tag}
                    <button type="button" class="text-blue-400 hover:text-blue-600 font-bold focus:outline-none text-[10px]" onclick="removeTag('${tag}')">×</button>
                `;
                selectedTagsContainer.appendChild(pill);
            });
        }

        if (hiddenTagsInput) {
            hiddenTagsInput.value = activeTags.join(', ');
        }

        if (dropdownList) {
            dropdownList.innerHTML = '';
            const unselected = tagPool.filter(t => !activeTags.includes(t.name));
            if (unselected.length === 0) {
                dropdownList.innerHTML = '<div class="px-3 py-1.5 text-xs text-gray-400 italic">Tag pool is empty</div>';
            } else {
                unselected.forEach(t => {
                    const opt = document.createElement('button');
                    opt.type = 'button';
                    opt.className = 'w-full text-left px-3 py-1.5 text-xs text-gray-700 hover:bg-blue-50 hover:text-blue-600 transition';
                    opt.textContent = '#' + t.name;
                    opt.addEventListener('click', () => {
                        addTag(t.name);
                        dropdownList.classList.add('hidden');
                    });
                    dropdownList.appendChild(opt);
                });
            }
        }
    }

    window.removeTag = function(tag) {
        activeTags = activeTags.filter(t => t !== tag);
        renderTags();
    };

    function addTag(tag) {
        if (!activeTags.includes(tag)) {
            activeTags.push(tag);
            renderTags();
        }
    }

    function fetchTagPool() {
        fetch('/api/v1/tags')
            .then(res => res.json())
            .then(res => {
                if (res.status === 'success' && res.data) {
                    tagPool = res.data;
                    renderTags();
                }
            });
    }

    fetchTagPool();

    // Admin only creation handler
    const createTagBtn = document.getElementById('create-tag-btn');
    const newTagInput = document.getElementById('new-tag-name');
    if (createTagBtn && newTagInput) {
        createTagBtn.addEventListener('click', () => {
            const tagName = newTagInput.value.trim().toLowerCase().replace(/#/g, '');
            if (!tagName) return;

            fetch('/api/v1/tags', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: tagName })
            })
            .then(res => res.json())
            .then(data => {
                if (data.status === 'success') {
                    newTagInput.value = '';
                    // Refresh tag pool and add it automatically
                    fetch('/api/v1/tags')
                        .then(res => res.json())
                        .then(res => {
                            if (res.status === 'success' && res.data) {
                                tagPool = res.data;
                                addTag(tagName);
                            }
                        });
                } else {
                    window.showCustomAlert("Error", "Failed to create tag: " + data.message);
                }
            })
            .catch(() => window.showCustomAlert("Error", "Could not connect to the server."));
        });
        
        newTagInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                createTagBtn.click();
            }
        });
    }

    // --- 3. Save Confirmation Modal Logic ---
    const editorForm = document.getElementById('editor-form');
    const saveBtn = document.getElementById('save-document-btn');
    const modal = document.getElementById('save-confirm-modal');
    const modalBox = document.getElementById('modal-box');
    const modalCommentInput = document.getElementById('modal-comment');
    const formCommentInput = document.getElementById('comment');
    const cancelBtn = document.getElementById('modal-cancel-btn');
    const confirmBtn = document.getElementById('modal-confirm-btn');
    const titleInput = document.getElementById('document-title-input');
    const textarea = document.getElementById('content');

    if (editorForm && titleInput && textarea) {
        const docId = titleInput.getAttribute('data-doc-id');
        const pageTitle = titleInput.value.trim();
        const storageKey = 'axia-draft-' + (docId || pageTitle);

        // Kiểm tra và khôi phục bản nháp chưa lưu
        const savedDraft = localStorage.getItem(storageKey);
        if (savedDraft && savedDraft !== textarea.value) {
            setTimeout(async () => {
                const restore = await window.showCustomConfirm(
                    "📝 Unsaved Draft Found", 
                    "We found an unsaved draft from a previous session. Would you like to restore it?"
                );
                if (restore) {
                    textarea.value = savedDraft;
                    textarea.dispatchEvent(new Event('input')); // kích hoạt tự động giãn cao
                    window.showToast("Draft restored successfully.");
                } else {
                    localStorage.removeItem(storageKey);
                }
            }, 500);
        }

        // Tự động lưu bản nháp sau mỗi 15 giây
        setInterval(() => {
            if (textarea.value) {
                localStorage.setItem(storageKey, textarea.value);
            }
        }, 15000);

        // Xóa bản nháp khi lưu thành công qua HTMX
        editorForm.addEventListener('htmx:afterRequest', (evt) => {
            if (evt.detail.successful) {
                localStorage.removeItem(storageKey);
            }
        });
    }

    if (saveBtn && editorForm && modal) {
        saveBtn.addEventListener('click', () => {
            // Validate the form content before opening the confirmation modal
            if (!editorForm.reportValidity()) {
                return;
            }

            modal.classList.remove('hidden');
            setTimeout(() => {
                modalBox.classList.remove('scale-95', 'opacity-0');
                modalBox.classList.add('scale-100', 'opacity-100');
                modalCommentInput.focus();
            }, 10);
        });

        const closeModal = () => {
            modalBox.classList.remove('scale-100', 'opacity-100');
            modalBox.classList.add('scale-95', 'opacity-0');
            setTimeout(() => {
                modal.classList.add('hidden');
                modalCommentInput.value = '';
            }, 150);
        };

        cancelBtn.addEventListener('click', closeModal);
        modal.addEventListener('click', (e) => {
            if (e.target === modal) closeModal();
        });

        confirmBtn.addEventListener('click', () => {
            const commentVal = modalCommentInput.value.trim();
            if (!commentVal) {
                modalCommentInput.classList.add('border-red-500', 'focus:ring-red-500/20');
                return;
            }
            
            modalCommentInput.classList.remove('border-red-500', 'focus:ring-red-500/20');
            formCommentInput.value = commentVal;
            
            closeModal();
            
            // Programmatically dispatch a submit event to trigger HTMX AJAX POST request
            editorForm.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
        });

        modalCommentInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                confirmBtn.click();
            }
        });
    }

    // --- 4. Auto-Expanding Textarea Logic ---
    const textarea = document.getElementById('content');
    if (textarea) {
        const autoExpand = () => {
            const scrollPos = window.scrollY;
            textarea.style.height = 'auto';
            textarea.style.height = textarea.scrollHeight + 'px';
            window.scrollTo(window.scrollX, scrollPos);
        };
        textarea.addEventListener('input', autoExpand);
        // Trigger on load
        autoExpand();

        // Re-trigger on image insertions or DOM modifications
        const observer = new MutationObserver(autoExpand);
        observer.observe(textarea, { childList: true, characterData: true, subtree: true });
    }



    // --- 6. Full Screen Focus Mode Logic ---
    const focusBtn = document.getElementById('toggle-focus-mode');
    const sidebar = document.getElementById('wiki-sidebar');
    const resizer = document.getElementById('sidebar-resizer');
    const header = document.querySelector('header');
    const mainWrapper = document.querySelector('main');
    const contentBox = document.getElementById('wiki-content');
    const editorHeader = document.querySelector('.select-none.border-b.border-gray-100');
    const titleInputWrapper = document.querySelector('.w-full.md\\:w-2\\/3');
    const tagsArea = document.getElementById('editor-tags-area');
    const stickyToolbar = document.getElementById('editor-sticky-toolbar');
    const boxLabel = document.getElementById('editor-box-label');
    const insertImgBtn = document.getElementById('focus-insert-image-btn');

    if (focusBtn) {
        focusBtn.addEventListener('click', () => {
            const isFocus = document.body.classList.toggle('focus-active');
            if (isFocus) {
                focusBtn.innerHTML = '✨ Exit focus';
                focusBtn.className = "border border-red-200 bg-red-50/50 hover:bg-red-50 active:scale-95 text-red-700 px-3 py-1.5 rounded-lg font-semibold text-xs transition-all duration-150 flex items-center justify-center gap-1 shadow-sm active:translate-y-0.5 active:shadow-inner";
                
                // Hide sidebar and layout parts
                if (sidebar) sidebar.style.display = 'none';
                if (resizer) resizer.style.display = 'none';
                if (header) header.style.display = 'none';

                // Hide Title input and Tags Area in focus mode
                if (titleInputWrapper) titleInputWrapper.style.display = 'none';
                if (tagsArea) tagsArea.style.display = 'none';
                
                // Fullwidth container styling
                if (mainWrapper) mainWrapper.className = 'flex-grow p-4 max-w-5xl w-full mx-auto transition-all duration-300';
                if (contentBox) contentBox.className = 'bg-white border border-gray-200 shadow-xl rounded-xl p-8 md:p-12 min-h-screen';
                
                // Floating top sticky toolbar for document actions
                if (editorHeader) {
                    editorHeader.className = 'flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6 select-none border-b border-gray-100 pb-5 sticky top-0 bg-white/95 backdrop-blur-md z-40 py-4 px-2 shadow-sm rounded-b-xl -mx-4 md:-mx-8';
                }

                // Make editor toolbar sticky and show image insert button
                if (stickyToolbar) {
                    stickyToolbar.className = 'flex justify-between items-center mb-2 bg-white/95 backdrop-blur-md border border-slate-200 rounded-lg p-2 font-sans transition-all duration-150 sticky top-[72px] z-30 shadow-sm';
                }
                if (insertImgBtn) insertImgBtn.classList.remove('hidden');
                if (boxLabel) boxLabel.style.display = 'none';
                
                window.showToast("Fullscreen focus mode enabled.");
            } else {
                focusBtn.innerHTML = '✨ Edit content';
                focusBtn.className = "border border-indigo-200 bg-indigo-50/50 hover:bg-indigo-50 active:scale-95 text-indigo-700 px-3 py-1.5 rounded-lg font-semibold text-xs transition-all duration-150 flex items-center justify-center gap-1 shadow-sm active:translate-y-0.5 active:shadow-inner";
                
                // Restore sidebar state
                const isCollapsed = localStorage.getItem("sidebar-collapsed") === "true";
                const savedWidth = parseInt(localStorage.getItem("sidebar-width")) || 256;
                if (!isCollapsed && sidebar) {
                    sidebar.style.display = 'flex';
                    sidebar.style.width = savedWidth + 'px';
                    if (resizer) resizer.style.display = 'block';
                }
                if (header) header.style.display = 'flex';

                // Restore Title input and Tags Area
                if (titleInputWrapper) titleInputWrapper.style.display = 'block';
                if (tagsArea) tagsArea.style.display = 'flex';
                
                // Restore standard margins
                if (mainWrapper) {
                    const isWide = document.body.classList.contains('content-width-wide');
                    mainWrapper.className = 'flex-grow p-8 w-full mx-auto transition-all duration-300';
                    mainWrapper.style.maxWidth = isWide ? '1600px' : '1350px';
                }
                if (contentBox) contentBox.className = 'bg-white border border-[#a2a9b1] shadow-sm rounded-lg p-10 min-h-[600px]';
                
                if (editorHeader) {
                    editorHeader.className = 'flex flex-col md:flex-row md:items-start justify-between gap-4 mb-6 select-none border-b border-gray-100 pb-5';
                }

                // Restore editor toolbar and hide image insert button
                if (stickyToolbar) {
                    stickyToolbar.className = 'flex justify-between items-center mb-2 bg-slate-50 border border-slate-200 rounded-lg p-2 font-sans transition-all duration-150';
                }
                if (insertImgBtn) insertImgBtn.classList.add('hidden');
                if (boxLabel) boxLabel.style.display = 'block';
                
                window.showToast("Focus mode disabled.");
            }
            
            // Adjust textarea size to new screen boundaries
            if (typeof autoExpand === 'function') {
                autoExpand();
            } else if (textarea) {
                textarea.style.height = 'auto';
                textarea.style.height = textarea.scrollHeight + 'px';
            }
        });

        // Initialize label on load
        focusBtn.innerHTML = '✨ Edit content';
    }
})();
