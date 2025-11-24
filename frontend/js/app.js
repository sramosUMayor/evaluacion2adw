let cart = JSON.parse(localStorage.getItem('cart')) || [];
let products = [];
let currentPage = 1;
const ITEMS_PER_PAGE = 9;

const productsGrid = document.querySelector('.productos-grid');
const cartBadge = document.querySelector('.badge');
const cartItemsContainer = document.getElementById('cart-items-container');
const cartTotalElement = document.getElementById('cart-total');
const checkoutForm = document.getElementById('checkout-form');

document.addEventListener('DOMContentLoaded', () => {
    updateCartBadge();

    const path = window.location.pathname;
    if (path.includes('catalogo.html')) {
        renderBreadcrumbs([{label: 'Inicio', url: 'index.html'}, {label: 'Catálogo'}]);
        const urlParams = new URLSearchParams(window.location.search);
        const category = urlParams.get('category');

        if (category) {
            const categorySelect = document.getElementById('categorySelect');
            if (categorySelect) categorySelect.value = category;
            fetchProducts(0, category);
        } else {
            fetchProducts();
        }
        setupSearchAndFilter();
    } else if (path.includes('index.html') || path.endsWith('/') || path.endsWith('/frontend/')) {
        fetchProducts(4);
    } else if (path.includes('carrito.html')) {
        renderBreadcrumbs([{label: 'Inicio', url: 'index.html'}, {label: 'Mi Carrito'}]);
        renderCart();
        fetchProducts(3);
    } else if (path.includes('checkout.html')) {
        renderBreadcrumbs([{label: 'Inicio', url: 'index.html'}, {label: 'Mi Carrito', url: 'carrito.html'}, {label: 'Checkout'}]);
        renderCheckoutSummary();
        setupCheckoutForm();
        setupDeliveryToggle();
    }

    createBackToTopButton();
});

function showModal(message, title = 'Flora Verde') {
    const modal = document.getElementById('customModal');
    if (!modal) {
        console.warn('Componente modal aún no cargado');
        alert(`${title}: ${message}`);
        return;
    }
    document.getElementById('modalTitle').textContent = title;
    document.getElementById('modalMessage').textContent = message;
    modal.classList.add('active');
}

function closeModal() {
    const modal = document.getElementById('customModal');
    if (modal) modal.classList.remove('active');
}


function setupDeliveryToggle() {
    const deliveryRadios = document.querySelectorAll('input[name="deliveryMethod"]');
    const addressContainer = document.getElementById('address-container');
    const addressInput = document.getElementById('address');

    if (!addressContainer || !addressInput) return;

    deliveryRadios.forEach(radio => {
        radio.addEventListener('change', (e) => {
            if (e.target.value === 'pickup') {
                addressContainer.style.display = 'none';
                addressInput.required = false;
            } else {
                addressContainer.style.display = 'block';
                addressInput.required = true;
            }
        });
    });
}


function createBackToTopButton() {
    const button = document.createElement('button');
    button.id = 'backToTopBtn';
    button.innerHTML = '↑';
    button.ariaLabel = 'Volver arriba';
    button.title = 'Volver arriba';
    document.body.appendChild(button);

    window.addEventListener('scroll', () => {
        if (window.scrollY > 300) {
            button.classList.add('visible');
        } else {
            button.classList.remove('visible');
        }
    });

    button.addEventListener('click', () => {
        window.scrollTo({
            top: 0,
            behavior: 'smooth'
        });
    });
}


async function fetchProducts(limit = 0, category = '') {
    try {
        let url = `${API_URL}/products`;
        if (category) {
            url += `?category=${encodeURIComponent(category)}`;
        }

        const response = await fetch(url);
        if (!response.ok) throw new Error('Error al obtener productos');
        products = await response.json();

        const productsToRender = limit > 0 ? products.slice(0, limit) : products;
        renderProducts(productsToRender);
    } catch (error) {
        console.error('Error:', error);
        if (productsGrid) {
            productsGrid.innerHTML = '<p class="error">Error al cargar los productos. Por favor intente más tarde.</p>';
        }
    }
}


async function renderProducts(productsToRender) {
    if (!productsGrid) return;

    let template = '';
    if (typeof getTemplate === 'function') {
        template = await getTemplate('product-card');
    }

    if (!template) {
        console.error('Plantilla de producto no encontrada');
        return;
    }

    productsGrid.innerHTML = productsToRender.map(product => {
        let html = template;
        html = html.replace(/{{image_url}}/g, product.image_url);
        html = html.replace(/{{name}}/g, product.name);
        html = html.replace(/{{description}}/g, product.description);
        html = html.replace(/{{watering}}/g, product.watering);
        html = html.replace(/{{light}}/g, product.light);
        html = html.replace(/{{price}}/g, product.price.toLocaleString('es-CL'));
        html = html.replace(/{{stock_label}}/g, product.stock ? '✓ En stock' : '✗ Agotado');
        html = html.replace(/{{id}}/g, product.id);
        html = html.replace(/{{disabled}}/g, !product.stock ? 'disabled' : '');
        return html;
    }).join('');
}


function addToCart(productId) {
    const product = products.find(p => p.id === productId);
    if (!product) return;

    const existingItem = cart.find(item => item.id === productId);
    if (existingItem) {
        existingItem.quantity += 1;
    } else {
        cart.push({ ...product, quantity: 1 });
    }

    saveCart();
    updateCartBadge();
    showModal('Producto agregado al carrito', '¡Éxito!');
}

function removeFromCart(productId) {
    cart = cart.filter(item => item.id !== productId);
    saveCart();
    renderCart();
    updateCartBadge();
}

function updateQuantity(productId, change) {
    const item = cart.find(item => item.id === productId);
    if (item) {
        item.quantity += change;
        if (item.quantity <= 0) {
            removeFromCart(productId);
        } else {
            saveCart();
            renderCart();
            updateCartBadge();
        }
    }
}

function saveCart() {
    localStorage.setItem('cart', JSON.stringify(cart));
}

function updateCartBadge() {
    if (cartBadge) {
        const count = cart.reduce((sum, item) => sum + item.quantity, 0);
        cartBadge.textContent = count;
    }
}


function renderCart() {
    const container = document.querySelector('.col-lg-8');
    if (!container) return;

    if (cart.length === 0) {
        container.innerHTML = '<p>Tu carrito está vacío. <a href="catalogo.html">Ir al catálogo</a></p>';
        updateTotals();
        return;
    }

    const itemsHtml = cart.map(item => `
        <div class="cart-item" style="display: flex; gap: 1rem; margin-bottom: 1.5rem; padding-bottom: 1.5rem; border-bottom: 1px solid #eee;">
            <img src="${item.image_url}" alt="${item.name}" style="width: 100px; height: 100px; object-fit: cover; border-radius: 8px;">
            <div style="flex: 1;">
                <div style="display: flex; justify-content: space-between; margin-bottom: 0.5rem;">
                    <h3 style="font-size: 1.1rem; margin: 0;">${item.name}</h3>
                    <button onclick="removeFromCart(${item.id})" style="background: none; border: none; color: #dc3545; cursor: pointer;">✕</button>
                </div>
                <p style="color: #6c757d; font-size: 0.9rem;">${item.category}</p>
                <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 1rem;">
                    <div style="display: flex; align-items: center; gap: 0.5rem;">
                        <button onclick="updateQuantity(${item.id}, -1)" class="btn btn-secondary" style="padding: 0.25rem 0.5rem;">-</button>
                        <span>${item.quantity}</span>
                        <button onclick="updateQuantity(${item.id}, 1)" class="btn btn-secondary" style="padding: 0.25rem 0.5rem;">+</button>
                    </div>
                    <p style="font-weight: bold;">$${(item.price * item.quantity).toLocaleString('es-CL')}</p>
                </div>
            </div>
        </div>
    `).join('');




    container.innerHTML = itemsHtml + `
        <div style="display: flex; justify-content: space-between; margin-top: 2rem;">
            <a href="catalogo.html" class="btn btn-outline-primary">← Seguir Comprando</a>
            <button onclick="cart = []; saveCart(); renderCart(); updateCartBadge();" class="btn btn-outline-secondary">Vaciar Carrito</button>
        </div>
    `;

    updateTotals();
}

function updateTotals() {
    const subtotal = cart.reduce((sum, item) => sum + (item.price * item.quantity), 0);
    const totalElement = document.querySelector('.resumen-total');
    const subtotalElement = document.querySelector('.resumen-subtotal');

    if (subtotalElement) subtotalElement.textContent = `$${subtotal.toLocaleString('es-CL')}`;
    if (totalElement) totalElement.textContent = `$${subtotal.toLocaleString('es-CL')}`;
}


function proceedToCheckout() {
    if (cart.length === 0) {
        showModal('El carrito está vacío', 'Carrito Vacío');
        return;
    }
    window.location.href = 'checkout.html';
}

function renderCheckoutSummary() {
    const itemsContainer = document.getElementById('checkout-items');
    if (!itemsContainer) return;

    if (cart.length === 0) {
        window.location.href = 'carrito.html';
        return;
    }

    itemsContainer.innerHTML = cart.map(item => `
        <div style="display: flex; justify-content: space-between; margin-bottom: 1rem; font-size: 0.9rem;">
            <span>${item.quantity}x ${item.name}</span>
            <span>$${(item.price * item.quantity).toLocaleString('es-CL')}</span>
        </div>
    `).join('');

    updateTotals();
}

function setupCheckoutForm() {
    if (!checkoutForm) return;

    checkoutForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const submitButton = checkoutForm.querySelector('button[type="submit"]');
        const originalButtonText = submitButton.textContent;


        submitButton.disabled = true;
        submitButton.textContent = 'Procesando...';

        const formData = new FormData(checkoutForm);
        const order = {
            customer_name: formData.get('name'),
            customer_email: formData.get('email'),
            address: formData.get('address'),
            items: cart.map(item => ({
                product_id: item.id,
                quantity: item.quantity
            })),
            total: cart.reduce((sum, item) => sum + (item.price * item.quantity), 0)
        };

        try {
            const response = await fetch(`${API_URL}/orders`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(order)
            });

            if (response.ok) {
                const result = await response.json();

                cart = [];
                saveCart();

                window.location.href = `thank-you.html?orderId=${result.id}`;
            } else {
                throw new Error('Error al procesar el pedido');
            }
        } catch (error) {
            console.error('Error:', error);
            showModal('Hubo un error al procesar su pedido. Por favor intente nuevamente.', 'Error');


            submitButton.disabled = false;
            submitButton.textContent = originalButtonText;
        }
    });
}


function setupSearchAndFilter() {
    const searchInput = document.getElementById('searchInput');
    const categorySelect = document.getElementById('categorySelect');
    const sortSelect = document.getElementById('sortSelect');

    if (searchInput) {
        searchInput.addEventListener('input', filterProducts);
    }
    if (categorySelect) {
        categorySelect.addEventListener('change', () => {
            const category = categorySelect.value;

            const newUrl = new URL(window.location);
            if (category) {
                newUrl.searchParams.set('category', category);
            } else {
                newUrl.searchParams.delete('category');
            }
            window.history.pushState({}, '', newUrl);

            fetchProducts(0, category);
        });
    }
    if (sortSelect) {
        sortSelect.addEventListener('change', filterProducts);
    }
}



function filterProducts() {
    const searchInput = document.getElementById('searchInput');
    const sortSelect = document.getElementById('sortSelect');

    let filtered = [...products];

    
    if (searchInput && searchInput.value) {
        const term = searchInput.value.toLowerCase();
        filtered = filtered.filter(p => 
            p.name.toLowerCase().includes(term) || 
            p.description.toLowerCase().includes(term)
        );
    }

    
    if (sortSelect && sortSelect.value) {
        switch (sortSelect.value) {
            case 'price-asc':
                filtered.sort((a, b) => a.price - b.price);
                break;
            case 'price-desc':
                filtered.sort((a, b) => b.price - a.price);
                break;
            case 'name-asc':
                filtered.sort((a, b) => a.name.localeCompare(b.name));
                break;
        }
    }

    const paginationContainer = document.getElementById('pagination-container');
    if (paginationContainer) {
        const totalPages = Math.ceil(filtered.length / ITEMS_PER_PAGE);
        
        if (currentPage > totalPages) currentPage = 1;
        if (currentPage < 1) currentPage = 1;
        if (totalPages === 0) currentPage = 1;

        const start = (currentPage - 1) * ITEMS_PER_PAGE;
        const end = start + ITEMS_PER_PAGE;
        const paginatedProducts = filtered.slice(start, end);
        
        renderProducts(paginatedProducts);
        renderPagination(totalPages);
    } else {
        renderProducts(filtered);
    }
}

function renderPagination(totalPages) {
    const container = document.getElementById('pagination-container');
    if (!container) return;

    if (totalPages <= 1) {
        container.innerHTML = '';
        return;
    }

    let html = `
        <div style="display: flex; justify-content: center; align-items: center; gap: 1rem; margin-top: 3rem;">
            <button onclick="changePage(${currentPage - 1})" class="btn btn-secondary" ${currentPage === 1 ? 'disabled style="opacity: 0.5;"' : ''}>◄ Anterior</button>
            <div style="display: flex; gap: 0.5rem;">
    `;

    for (let i = 1; i <= totalPages; i++) {
        html += `<button onclick="changePage(${i})" class="btn ${currentPage === i ? 'btn-primary' : 'btn-secondary'}" style="padding: 0.5rem 1rem;">${i}</button>`;
    }

    html += `
            </div>
            <button onclick="changePage(${currentPage + 1})" class="btn btn-secondary" ${currentPage === totalPages ? 'disabled style="opacity: 0.5;"' : ''}>Siguiente ►</button>
        </div>
    `;

    container.innerHTML = html;
}

function changePage(newPage) {
    currentPage = newPage;
    filterProducts();
    window.scrollTo({ top: 0, behavior: 'smooth' });
}

window.addToCart = addToCart;
window.removeFromCart = removeFromCart;
window.updateQuantity = updateQuantity;
window.proceedToCheckout = proceedToCheckout;
window.closeModal = closeModal;
window.showModal = showModal;
window.updateTotals = updateTotals;
window.changePage = changePage;
