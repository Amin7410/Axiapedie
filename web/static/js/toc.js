/**
 * Table of Contents (TOC) Generator for Axia Wiki
 * Generates dynamic TOC from heading tags (h2, h3) in articles,
 * handles smooth scrolling, scroll spying, and collapsible mobile views.
 */

function generateTOC() {
    const wikiContent = document.querySelector('.wiki-content');
    const desktopTocList = document.getElementById('article-toc');
    const mobileTocList = document.getElementById('mobile-toc-list');
    const desktopContainer = document.getElementById('article-toc-container');
    const mobileContainer = document.getElementById('mobile-toc-container');

    if (!wikiContent) return;

    // Find all h2 and h3 headings in the article content
    const headings = wikiContent.querySelectorAll('h2, h3');

    // If less than 2 headings, hide the TOC components
    if (headings.length < 2) {
        if (desktopContainer) desktopContainer.classList.add('hidden');
        if (mobileContainer) mobileContainer.classList.add('hidden');
        return;
    }

    // Show containers
    if (desktopContainer) desktopContainer.classList.remove('hidden');
    if (mobileContainer) mobileContainer.classList.remove('hidden');

    // Clear previous TOC items
    if (desktopTocList) desktopTocList.innerHTML = '';
    if (mobileTocList) mobileTocList.innerHTML = '';

    headings.forEach((heading, index) => {
        let id = heading.id;
        if (!id) {
            // Generate clean ID from heading text if missing
            const text = heading.textContent || '';
            id = 'heading-' + index + '-' + text.toLowerCase().trim()
                .replace(/[^a-z0-9\u00C0-\u024F\u1E00-\u1EFF]+/g, '-') // Support unicode chars (Vietnamese)
                .replace(/(^-|-$)/g, '');
            heading.id = id;
        }

        const title = heading.textContent || '';
        const level = heading.tagName.toLowerCase(); // 'h2' or 'h3'

        // 1. Build desktop TOC list item
        if (desktopTocList) {
            const li = document.createElement('li');
            const a = document.createElement('a');
            a.href = '#' + id;
            a.textContent = title;
            a.className = 'block transition-all duration-150 py-0.5 hover:text-blue-600 relative pl-3 border-l-2 border-transparent';
            if (level === 'h3') {
                a.classList.add('ml-3', 'text-[11.5px]');
            } else {
                a.classList.add('text-[13px]');
            }
            a.setAttribute('data-heading-id', id);

            // Smooth scroll click handler
            a.addEventListener('click', (e) => {
                e.preventDefault();
                const target = document.getElementById(id);
                if (target) {
                    const headerOffset = 80; // Offset for sticky headers
                    const elementPosition = target.getBoundingClientRect().top;
                    const offsetPosition = elementPosition + window.pageYOffset - headerOffset;

                    window.scrollTo({
                        top: offsetPosition,
                        behavior: 'smooth'
                    });
                    history.pushState(null, '', '#' + id);
                }
            });

            li.appendChild(a);
            desktopTocList.appendChild(li);
        }

        // 2. Build mobile TOC list item
        if (mobileTocList) {
            const li = document.createElement('li');
            const a = document.createElement('a');
            a.href = '#' + id;
            a.textContent = title;
            a.className = 'block py-1 hover:text-blue-600 transition';
            if (level === 'h3') {
                a.classList.add('pl-4', 'text-xs');
            } else {
                a.classList.add('text-sm', 'font-medium');
            }

            a.addEventListener('click', (e) => {
                e.preventDefault();
                const target = document.getElementById(id);
                if (target) {
                    const headerOffset = 80;
                    const elementPosition = target.getBoundingClientRect().top;
                    const offsetPosition = elementPosition + window.pageYOffset - headerOffset;

                    window.scrollTo({
                        top: offsetPosition,
                        behavior: 'smooth'
                    });
                    history.pushState(null, '', '#' + id);

                    // Collapse mobile menu after selection
                    const mobileTocContent = document.getElementById('mobile-toc-content');
                    if (mobileTocContent) mobileTocContent.classList.add('hidden');
                    const toggleIcon = document.getElementById('mobile-toc-toggle-icon');
                    if (toggleIcon) toggleIcon.style.transform = 'rotate(0deg)';
                }
            });

            li.appendChild(a);
            mobileTocList.appendChild(li);
        }
    });

    // 3. Scroll Spying (Highlight active headings on desktop scroll)
    if (desktopTocList && typeof IntersectionObserver !== 'undefined') {
        const observerOptions = {
            root: null,
            rootMargin: '-80px 0px -75% 0px', // Trigger highlight near top of viewport
            threshold: 0
        };

        const observer = new IntersectionObserver((entries) => {
            // Find headings currently intersecting
            entries.forEach((entry) => {
                const id = entry.target.id;
                const link = desktopTocList.querySelector(`a[data-heading-id="${id}"]`);
                if (!link) return;

                if (entry.isIntersecting) {
                    // Remove active style from all TOC links
                    desktopTocList.querySelectorAll('a').forEach((l) => {
                        l.classList.remove('text-blue-600', 'border-blue-600', 'font-semibold');
                        l.classList.add('text-slate-500', 'border-transparent');
                    });

                    // Add active style to current link
                    link.classList.add('text-blue-600', 'border-blue-600', 'font-semibold');
                    link.classList.remove('text-slate-500', 'border-transparent');
                }
            });
        }, observerOptions);

        headings.forEach((heading) => observer.observe(heading));
    }
}

// Mobile TOC Collapsible toggle
function toggleMobileToc() {
    const content = document.getElementById('mobile-toc-content');
    const toggleIcon = document.getElementById('mobile-toc-toggle-icon');
    if (!content) return;

    const isHidden = content.classList.contains('hidden');
    if (isHidden) {
        content.classList.remove('hidden');
        if (toggleIcon) toggleIcon.style.transform = 'rotate(180deg)';
    } else {
        content.classList.add('hidden');
        if (toggleIcon) toggleIcon.style.transform = 'rotate(0deg)';
    }
}

// Initialise TOC on load
document.addEventListener('DOMContentLoaded', generateTOC);

// Re-initialise TOC after HTMX swap
document.addEventListener('htmx:afterSwap', function (evt) {
    if (evt.detail.target.id === 'wiki-content') {
        generateTOC();
    }
});
