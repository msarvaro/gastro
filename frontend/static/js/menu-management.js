// Global variables
let currentCategoryId = null;
let deleteItemId = null;
let deleteItemType = null;

// Initialize the page
document.addEventListener('DOMContentLoaded', async () => {
    // Check authorization
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    if (!token || role !== 'manager') {
        window.location.href = '/';
        return;
    }

    // Load initial data
    await loadCategories();
    await loadMenuItems();

    // Add event listeners
    document.getElementById('addCategoryBtn').addEventListener('click', () => openCategoryModal());
    document.getElementById('addMenuItemBtn').addEventListener('click', () => openMenuItemModal());
    document.getElementById('searchMenuItems').addEventListener('input', handleSearch);
    document.getElementById('categoryForm').addEventListener('submit', handleCategorySubmit);
    document.getElementById('menuItemForm').addEventListener('submit', handleMenuItemSubmit);
    document.getElementById('confirmDeleteBtn').addEventListener('click', handleDelete);
});

// Load categories
async function loadCategories() {
    try {
        const categories = await menuApi.getCategories();
        const categoriesList = document.getElementById('categoriesList');
        const categorySelect = document.getElementById('menuItemCategory');
        
        // Clear existing options
        categoriesList.innerHTML = '';
        categorySelect.innerHTML = '<option value="">Select Category</option>';
        
        // Add categories to both lists
        categories.forEach(category => {
            // Add to categories list
            const categoryElement = createCategoryElement(category);
            categoriesList.appendChild(categoryElement);
            
            // Add to select dropdown
            const option = document.createElement('option');
            option.value = category.id;
            option.textContent = category.name;
            categorySelect.appendChild(option);
        });
    } catch (error) {
        showError('Failed to load categories');
        console.error('Error loading categories:', error);
    }
}

// Load menu items
async function loadMenuItems(categoryId = null) {
    try {
        const items = await menuApi.getMenuItems(categoryId);
        const menuItemsList = document.getElementById('menuItemsList');
        menuItemsList.innerHTML = '';
        
        items.forEach(item => {
            const itemElement = createMenuItemElement(item);
            menuItemsList.appendChild(itemElement);
        });
    } catch (error) {
        showError('Failed to load menu items');
        console.error('Error loading menu items:', error);
    }
}

// Create category element
function createCategoryElement(category) {
    const div = document.createElement('div');
    div.className = 'category-item';
    div.innerHTML = `
        <div class="category-header">
            <h3>${category.name}</h3>
            <div class="category-actions">
                <button onclick="editCategory(${category.id})" class="btn btn-icon">
                    <i class="fas fa-edit"></i>
                </button>
                <button onclick="deleteItem(${category.id}, 'category')" class="btn btn-icon">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        </div>
        <p>${category.description || ''}</p>
    `;
    div.addEventListener('click', () => filterByCategory(category.id));
    return div;
}

// Create menu item element
function createMenuItemElement(item) {
    const div = document.createElement('div');
    div.className = `menu-item ${!item.is_available ? 'unavailable' : ''}`;
    div.innerHTML = `
        <div class="menu-item-header">
            <h3>${item.name}</h3>
            <div class="menu-item-actions">
                <button onclick="editMenuItem(${item.id})" class="btn btn-icon">
                    <i class="fas fa-edit"></i>
                </button>
                <button onclick="deleteItem(${item.id}, 'menuItem')" class="btn btn-icon">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        </div>
        <p class="menu-item-description">${item.description || ''}</p>
        <div class="menu-item-details">
            <span class="price">${formatMoney(item.price)}</span>
            <span class="status ${item.is_available ? 'available' : 'unavailable'}">
                ${item.is_available ? 'Available' : 'Unavailable'}
            </span>
        </div>
        ${item.image_url ? `<img src="${item.image_url}" alt="${item.name}" class="menu-item-image">` : ''}
    `;
    return div;
}

// Modal functions
function openCategoryModal(category = null) {
    const modal = document.getElementById('categoryModal');
    const title = document.getElementById('categoryModalTitle');
    const form = document.getElementById('categoryForm');
    const idInput = document.getElementById('categoryId');
    const nameInput = document.getElementById('categoryName');
    const descInput = document.getElementById('categoryDescription');

    title.textContent = category ? 'Edit Category' : 'Add Category';
    idInput.value = category ? category.id : '';
    nameInput.value = category ? category.name : '';
    descInput.value = category ? category.description : '';
    
    modal.style.display = 'block';
}

function openMenuItemModal(item = null) {
    const modal = document.getElementById('menuItemModal');
    const title = document.getElementById('menuItemModalTitle');
    const form = document.getElementById('menuItemForm');
    const idInput = document.getElementById('menuItemId');
    const nameInput = document.getElementById('menuItemName');
    const categoryInput = document.getElementById('menuItemCategory');
    const descInput = document.getElementById('menuItemDescription');
    const priceInput = document.getElementById('menuItemPrice');
    const imageInput = document.getElementById('menuItemImage');
    const availableInput = document.getElementById('menuItemAvailable');

    title.textContent = item ? 'Edit Menu Item' : 'Add Menu Item';
    idInput.value = item ? item.id : '';
    nameInput.value = item ? item.name : '';
    categoryInput.value = item ? item.category_id : '';
    descInput.value = item ? item.description : '';
    priceInput.value = item ? item.price : '';
    imageInput.value = item ? item.image_url : '';
    availableInput.checked = item ? item.is_available : true;
    
    modal.style.display = 'block';
}

function closeModal(modalId) {
    document.getElementById(modalId).style.display = 'none';
}

// Event handlers
async function handleCategorySubmit(e) {
    e.preventDefault();
    const id = document.getElementById('categoryId').value;
    const category = {
        name: document.getElementById('categoryName').value,
        description: document.getElementById('categoryDescription').value
    };

    try {
        if (id) {
            await menuApi.updateCategory(id, category);
        } else {
            await menuApi.createCategory(category);
        }
        closeModal('categoryModal');
        await loadCategories();
        showSuccess('Category saved successfully');
    } catch (error) {
        showError('Failed to save category');
        console.error('Error saving category:', error);
    }
}

async function handleMenuItemSubmit(e) {
    e.preventDefault();
    const id = document.getElementById('menuItemId').value;
    const menuItem = {
        name: document.getElementById('menuItemName').value,
        category_id: parseInt(document.getElementById('menuItemCategory').value),
        description: document.getElementById('menuItemDescription').value,
        price: parseFloat(document.getElementById('menuItemPrice').value),
        image_url: document.getElementById('menuItemImage').value,
        is_available: document.getElementById('menuItemAvailable').checked
    };

    try {
        if (id) {
            await menuApi.updateMenuItem(id, menuItem);
        } else {
            await menuApi.createMenuItem(menuItem);
        }
        closeModal('menuItemModal');
        await loadMenuItems(currentCategoryId);
        showSuccess('Menu item saved successfully');
    } catch (error) {
        showError('Failed to save menu item');
        console.error('Error saving menu item:', error);
    }
}

function deleteItem(id, type) {
    deleteItemId = id;
    deleteItemType = type;
    const modal = document.getElementById('deleteModal');
    const message = document.getElementById('deleteModalMessage');
    message.textContent = `Are you sure you want to delete this ${type === 'category' ? 'category' : 'menu item'}?`;
    modal.style.display = 'block';
}

async function handleDelete() {
    try {
        if (deleteItemType === 'category') {
            await menuApi.deleteCategory(deleteItemId);
            await loadCategories();
        } else {
            await menuApi.deleteMenuItem(deleteItemId);
            await loadMenuItems(currentCategoryId);
        }
        closeModal('deleteModal');
        showSuccess('Item deleted successfully');
    } catch (error) {
        showError('Failed to delete item');
        console.error('Error deleting item:', error);
    }
}

function filterByCategory(categoryId) {
    currentCategoryId = categoryId;
    loadMenuItems(categoryId);
}

function handleSearch(e) {
    const query = e.target.value.toLowerCase();
    const items = document.querySelectorAll('.menu-item');
    items.forEach(item => {
        const name = item.querySelector('h3').textContent.toLowerCase();
        const description = item.querySelector('.menu-item-description').textContent.toLowerCase();
        const visible = name.includes(query) || description.includes(query);
        item.style.display = visible ? 'block' : 'none';
    });
}

// Helper functions
function formatMoney(amount) {
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'KZT'
    }).format(amount);
}

function showSuccess(message) {
    // Implement your success notification here
    alert(message); // Replace with a better UI notification
}

function showError(message) {
    // Implement your error notification here
    alert(message); // Replace with a better UI notification
}

// Edit functions
function editCategory(id) {
    const category = categories.find(c => c.id === id);
    if (category) {
        openCategoryModal(category);
    }
}

function editMenuItem(id) {
    const item = menuItems.find(i => i.id === id);
    if (item) {
        openMenuItemModal(item);
    }
} 