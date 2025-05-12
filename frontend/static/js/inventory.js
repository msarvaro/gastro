// Inventory Management Functions
async function loadInventoryData() {
    try {
        const response = await fetch('/api/admin/inventory', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });

        if (!response.ok) {
            throw new Error('Ошибка при загрузке данных инвентаря');
        }

        const items = await response.json();

        // Фильтрация
        const search = document.getElementById('productSearch').value.trim().toLowerCase();
        const category = document.getElementById('productCategoryFilter').value;
        const branch = document.getElementById('productBranchFilter').value;
        let filteredItems = items;
        if (search) filteredItems = items.filter(i => i.name.toLowerCase().includes(search));
        if (category) filteredItems = items.filter(i => i.category === category);
        if (branch) filteredItems = items.filter(i => i.branch === branch);

        // Update inventory cards
        document.querySelector('#inventory-stock-tab .card:nth-child(1) .value').textContent = `${filteredItems.length}`;
        document.querySelector('#inventory-stock-tab .card:nth-child(2) .value').textContent = `${filteredItems.filter(item => getStatusClass(item)==='low').length}`;
        document.querySelector('#inventory-stock-tab .card:nth-child(3) .value').textContent = `${filteredItems.filter(item => item.status === 'pending').length}`;

        // Update inventory table
        const table = document.querySelector('#inventory-stock-tab table');
        if (table) {
            const tbody = table.querySelector('tbody');
            if (tbody) {
                tbody.innerHTML = filteredItems.map(item => `
                    <tr>
                        <td>${item.name}</td>
                        <td>${item.category}</td>
                        <td>${item.quantity} ${item.unit}</td>
                        <td>${item.min_quantity} ${item.min_unit}</td>
                        <td><span class="status-${getStatusClass(item)}">${getStatusText(item)}</span></td>
                    </tr>
                `).join('');
            }
        }
    } catch (error) {
        console.error('Error loading inventory data:', error);
    }
}

function getStatusClass(item) {
    if (item.quantity < item.min_quantity/2) return 'critical';
    if (item.quantity < item.min_quantity) return 'low';
    return 'ok';
}

function getStatusText(item) {
    if (item.quantity < item.min_quantity/2) return 'Критично';
    if (item.quantity < item.min_quantity) return 'Низкий';
    return 'В норме';
}

async function handleAddProduct(event) {
    event.preventDefault();
    const form = event.target;

    try {
        const quantity = parseFloat(form.productQuantity.value.replace(',', '.'));
        const minQuantity = parseFloat(form.productMinQuantity.value.replace(',', '.'));

        if (isNaN(quantity) || isNaN(minQuantity)) {
            throw new Error('Некорректные числовые значения');
        }

        const newProduct = {
            name: form.productName.value,
            category: form.productCategory.value,
            quantity: quantity,
            unit: form.productUnit.value,
            min_quantity: minQuantity,
            min_unit: form.productMinUnit.value
        };

        const response = await fetch('/api/admin/inventory', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify(newProduct)
        });

        if (!response.ok) {
            throw new Error('Ошибка при добавлении продукта');
        }

        form.reset();
        closeModal('addProductModal');
        loadInventoryData();
    } catch (error) {
        alert('Ошибка при добавлении продукта: ' + error.message);
    }
}

// Initialize event listeners
document.addEventListener('DOMContentLoaded', function() {
    // Add Product form submission
    const addProductForm = document.getElementById('addProductForm');
    if (addProductForm) {
        addProductForm.addEventListener('submit', handleAddProduct);
    }

    // Search and filter inputs
    const productSearch = document.getElementById('productSearch');
    const productCategoryFilter = document.getElementById('productCategoryFilter');
    const productBranchFilter = document.getElementById('productBranchFilter');

    if (productSearch) {
        productSearch.addEventListener('input', loadInventoryData);
    }
    if (productCategoryFilter) {
        productCategoryFilter.addEventListener('change', loadInventoryData);
    }
    if (productBranchFilter) {
        productBranchFilter.addEventListener('change', loadInventoryData);
    }

    // Initial load
    loadInventoryData();
}); 