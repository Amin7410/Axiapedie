// dialogs.js — Toast notifications và custom modal (prompt / confirm / alert)
window.showToast = function (message, type = 'success') {
    const container = document.getElementById('toast-container');
    if (!container) return;

    // Tối đa 3 toast cùng lúc
    while (container.children.length >= 3) {
        container.children[0].remove();
    }

    const toast = document.createElement('div');
    toast.className = `p-3 rounded-lg shadow-lg border text-xs font-semibold pointer-events-auto transform translate-y-2 opacity-0 transition-all duration-300 flex items-center gap-2 bg-white/60 backdrop-blur-md cursor-pointer select-none`;

    if (type === 'error') {
        toast.classList.add('border-red-200', 'text-red-900', 'bg-red-50/60');
        toast.innerHTML = `<span class="text-sm">❌</span> <span class="flex-grow">${message}</span>`;
    } else if (type === 'warning') {
        toast.classList.add('border-amber-200', 'text-amber-900', 'bg-amber-50/60');
        toast.innerHTML = `<span class="text-sm">⚠️</span> <span class="flex-grow">${message}</span>`;
    } else {
        toast.classList.add('border-emerald-200', 'text-emerald-900', 'bg-emerald-50/60');
        toast.innerHTML = `<span class="text-sm">✨</span> <span class="flex-grow">${message}</span>`;
    }

    container.appendChild(toast);

    // Hiệu ứng xuất hiện
    setTimeout(() => {
        toast.classList.remove('translate-y-2', 'opacity-0');
    }, 10);

    // Tự đóng sau 3 giây
    const autoDismissTimeout = setTimeout(() => {
        dismissToast();
    }, 3000);

    // Đóng khi click vào thông báo
    toast.addEventListener('click', () => {
        clearTimeout(autoDismissTimeout);
        dismissToast();
    });

    function dismissToast() {
        toast.classList.add('opacity-0', 'translate-y-1');
        setTimeout(() => {
            toast.remove();
        }, 300);
    }
};

window.showCustomPrompt = function (title, message, defaultValue = "", placeholder = "") {
    return new Promise((resolve) => {
        const modal = document.getElementById('custom-dialog-modal');
        const modalBox = document.getElementById('custom-dialog-box');
        const titleEl = document.getElementById('custom-dialog-title');
        const messageEl = document.getElementById('custom-dialog-message');
        const inputContainer = document.getElementById('custom-dialog-input-container');
        const inputEl = document.getElementById('custom-dialog-input');
        const cancelBtn = document.getElementById('custom-dialog-cancel-btn');
        const confirmBtn = document.getElementById('custom-dialog-confirm-btn');

        if (!modal || !modalBox) {
            resolve(prompt(message, defaultValue));
            return;
        }

        if (modal._closeTimeout) {
            clearTimeout(modal._closeTimeout);
            modal._closeTimeout = null;
        }

        titleEl.textContent = title;
        messageEl.textContent = message;
        inputContainer.classList.remove('hidden');
        inputEl.value = defaultValue;
        inputEl.placeholder = placeholder;
        cancelBtn.classList.remove('hidden');

        modal.classList.remove('hidden');
        setTimeout(() => {
            modalBox.classList.remove('scale-95', 'opacity-0');
            modalBox.classList.add('scale-100', 'opacity-100');
            inputEl.focus();
            inputEl.select();
        }, 10);

        const closeDialog = () => {
            modalBox.classList.remove('scale-100', 'opacity-100');
            modalBox.classList.add('scale-95', 'opacity-0');
            if (modal._closeTimeout) clearTimeout(modal._closeTimeout);
            modal._closeTimeout = setTimeout(() => {
                modal.classList.add('hidden');
                modal._closeTimeout = null;
            }, 150);
        };

        const handleConfirm = () => {
            const val = inputEl.value;
            closeDialog();
            resolve(val);
        };

        const handleCancel = () => {
            closeDialog();
            resolve(null);
        };

        confirmBtn.onclick = handleConfirm;
        cancelBtn.onclick = handleCancel;

        inputEl.onkeydown = (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                handleConfirm();
            } else if (e.key === 'Escape') {
                e.preventDefault();
                handleCancel();
            }
        };
    });
};

window.showCustomConfirm = function (title, message) {
    return new Promise((resolve) => {
        const modal = document.getElementById('custom-dialog-modal');
        const modalBox = document.getElementById('custom-dialog-box');
        const titleEl = document.getElementById('custom-dialog-title');
        const messageEl = document.getElementById('custom-dialog-message');
        const inputContainer = document.getElementById('custom-dialog-input-container');
        const cancelBtn = document.getElementById('custom-dialog-cancel-btn');
        const confirmBtn = document.getElementById('custom-dialog-confirm-btn');

        if (!modal || !modalBox) {
            resolve(confirm(message));
            return;
        }

        if (modal._closeTimeout) {
            clearTimeout(modal._closeTimeout);
            modal._closeTimeout = null;
        }

        titleEl.textContent = title;
        messageEl.textContent = message;
        inputContainer.classList.add('hidden');
        cancelBtn.classList.remove('hidden');

        modal.classList.remove('hidden');
        setTimeout(() => {
            modalBox.classList.remove('scale-95', 'opacity-0');
            modalBox.classList.add('scale-100', 'opacity-100');
            confirmBtn.focus();
        }, 10);

        const closeDialog = () => {
            modalBox.classList.remove('scale-100', 'opacity-100');
            modalBox.classList.add('scale-95', 'opacity-0');
            if (modal._closeTimeout) clearTimeout(modal._closeTimeout);
            modal._closeTimeout = setTimeout(() => {
                modal.classList.add('hidden');
                modal._closeTimeout = null;
            }, 150);
        };

        const handleConfirm = () => {
            closeDialog();
            resolve(true);
        };

        const handleCancel = () => {
            closeDialog();
            resolve(false);
        };

        confirmBtn.onclick = handleConfirm;
        cancelBtn.onclick = handleCancel;

        modal.onkeydown = (e) => {
            if (e.key === 'Escape') {
                e.preventDefault();
                handleCancel();
            }
        };
    });
};

window.showCustomAlert = function (title, message) {
    const type = (title && title.toLowerCase() === 'error') ? 'error' : 'success';
    window.showToast(message, type);
    return Promise.resolve();
};
