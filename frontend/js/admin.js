const CATEGORIES_URL = `${API_URL}/categories`;

const tableBody = document.getElementById('productsTableBody');
const productModal = document.getElementById('productModal');
const productForm = document.getElementById('productForm');
const modalTitle = document.getElementById('modalTitle');
const categorySelect = document.getElementById('productCategory');


document.addEventListener('DOMContentLoaded', () => {
    loadProducts();
    loadCategories();
});

async function loadCategories() {
    try {
        const response = await fetch(CATEGORIES_URL);
        const categories = await response.json();
        categorySelect.innerHTML = categories.map(c => 
            `<option value="${c.id}">${c.name}</option>`
        ).join('');
    } catch (error) {
        console.error('Error cargando categorías:', error);
    }
}

async function loadProducts() {
    try {
        const response = await fetch(`${API_URL}/products`);
        const products = await response.json();
        renderTable(products);
    } catch (error) {
        console.error('Error cargando productos:', error);
        alert('Error al cargar los productos');
    }
}

function renderTable(products) {
    tableBody.innerHTML = products.map(p => `
        <tr>
            <td>${p.id}</td>
            <td><img src="${p.image_url}" alt="${p.name}" style="width: 50px; height: 50px; object-fit: cover;"></td>
            <td>${p.name}</td>
            <td>${p.category}</td>
            <td>$${p.price}</td>
            <td>${p.stock ? '✅' : '❌'}</td>
            <td>
                <button class="action-btn edit-btn" onclick="editProduct(${p.id})">✏️</button>
                <button class="action-btn delete-btn" onclick="deleteProduct(${p.id})">🗑️</button>
            </td>
        </tr>
    `).join('');
}

function openProductModal() {
    productForm.reset();
    document.getElementById('productId').value = '';
    modalTitle.textContent = 'Nuevo Producto';
    productModal.classList.add('active');
}

function closeProductModal() {
    productModal.classList.remove('active');
}

productForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const id = document.getElementById('productId').value;
    const productData = {
        name: document.getElementById('productName').value,
        description: document.getElementById('productDescription').value,
        price: parseFloat(document.getElementById('productPrice').value),
        category_id: parseInt(document.getElementById('productCategory').value),
        image_url: document.getElementById('productImage').value,
        watering: document.getElementById('productWatering').value,
        light: document.getElementById('productLight').value,
        stock: document.getElementById('productStock').checked
    };

    try {
        const method = id ? 'PUT' : 'POST';
        const url = id ? `${API_URL}/products?id=${id}` : `${API_URL}/products`;

        if (id) productData.id = parseInt(id);

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(productData)
        });

        if (!response.ok) throw new Error('Error al guardar el producto');

        closeProductModal();
        loadProducts();
        alert(id ? 'Producto actualizado' : 'Producto creado');
    } catch (error) {
        console.error('Error:', error);
        alert('Error al guardar el producto');
    }
});


window.editProduct = async (id) => {
    try {
        const response = await fetch(`${API_URL}/products`);
        const products = await response.json();
        const product = products.find(p => p.id === id);

        if (!product) return;

        document.getElementById('productId').value = product.id;
        document.getElementById('productName').value = product.name;
        document.getElementById('productDescription').value = product.description;
        document.getElementById('productPrice').value = product.price;
        document.getElementById('productCategory').value = product.category_id;
        document.getElementById('productImage').value = product.image_url;
        document.getElementById('productWatering').value = product.watering;
        document.getElementById('productLight').value = product.light;
        document.getElementById('productStock').checked = product.stock;

        modalTitle.textContent = 'Editar Producto';
        productModal.classList.add('active');
    } catch (error) {
        console.error('Error obteniendo detalles del producto:', error);
    }
};


window.deleteProduct = async (id) => {
    if (!confirm('¿Estás seguro de eliminar este producto?')) return;

    try {
        const response = await fetch(`${API_URL}/products?id=${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) throw new Error('Error al eliminar el producto');

        loadProducts();
    } catch (error) {
        console.error('Error:', error);
        alert('Error al eliminar el producto');
    }
};


productModal.addEventListener('click', (e) => {
    if (e.target === productModal) closeProductModal();
});
