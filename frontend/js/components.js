const API_URL = window.location.protocol === 'file:' ? 'http://localhost:8080/api' : '/api';

document.addEventListener('DOMContentLoaded', () => {
    loadComponent('header-placeholder', 'components/header.html', initHeader);
    loadComponent('features-placeholder', 'components/features.html');
    loadComponent('order-summary-placeholder', 'components/order-summary.html', () => {
        if (typeof window.updateTotals === 'function') window.updateTotals();
    });
    loadComponent('footer-placeholder', 'components/footer.html');
    loadModal();
});

async function loadModal() {
    if (document.getElementById('customModal')) return;
    
    try {
        const response = await fetch('components/modal.html');
        if (!response.ok) throw new Error('Error al cargar el componente modal');
        const html = await response.text();
        document.body.insertAdjacentHTML('beforeend', html);
        
        const modal = document.getElementById('customModal');
        if (modal) {
            modal.addEventListener('click', (e) => {
                if (e.target.id === 'customModal') closeModal();
            });
        }
    } catch (error) {
        console.error(error);
    }
}

let templates = {};

async function getTemplate(name) {
    if (templates[name]) return templates[name];
    
    try {
        const response = await fetch(`components/${name}.html`);
        if (!response.ok) throw new Error(`Error al cargar la plantilla ${name}`);
        const html = await response.text();
        templates[name] = html;
        return html;
    } catch (error) {
        console.error(error);
        return '';
    }
}

async function loadComponent(elementId, filePath, callback) {
    const element = document.getElementById(elementId);
    if (!element) return;

    try {
        const response = await fetch(filePath);
        if (!response.ok) throw new Error(`Error al cargar ${filePath}`);
        const html = await response.text();
        element.innerHTML = html;
        if (callback) callback();
    } catch (error) {
        console.error(error);
    }
}

function initHeader() {
    const menuToggle = document.getElementById('menuToggle');
    const navMenu = document.getElementById('navMenu');
    
    if (menuToggle && navMenu) {
        menuToggle.addEventListener('click', function() {
            navMenu.classList.toggle('active');
        });

        const navLinks = document.querySelectorAll('.nav-menu a');
        navLinks.forEach(link => {
            link.addEventListener('click', () => {
                navMenu.classList.remove('active');
            });
        });
    }

    const currentPage = window.location.pathname.split('/').pop() || 'index.html';
    const links = document.querySelectorAll('.nav-menu a[data-page]');
    
    links.forEach(link => {
        if (link.getAttribute('data-page') === currentPage) {
            link.classList.add('active');
        }
    });

    if (typeof updateCartCount === 'function') {
        updateCartCount();
    }
}

function renderBreadcrumbs(items) {
    const container = document.querySelector('.breadcrumbs');
    if (!container) return;

    const html = items.map((item, index) => {
        if (index === items.length - 1) {
            return `<span>${item.label}</span>`;
        }
        return `<a href="${item.url}">${item.label}</a>`;
    }).join(' &gt; ');

    container.innerHTML = html;
}