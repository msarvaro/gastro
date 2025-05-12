// Menu API functions
const API_BASE = '/api/menu';

// Helper function for API calls
async function apiCall(endpoint, method = 'GET', data = null) {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/';
        return null;
    }

    const options = {
        method,
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        }
    };

    if (data) {
        options.body = JSON.stringify(data);
    }

    try {
        const response = await fetch(`${API_BASE}${endpoint}`, options);
        if (response.status === 401 || response.status === 403) {
            window.location.href = '/';
            return null;
        }
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
}

// Get all menu items
async function getMenuItems() {
    return apiCall('/items');
}

// Get menu items by category
async function getMenuItemsByCategory(category) {
    return apiCall(`/items/category/${encodeURIComponent(category)}`);
}

// Add new menu item (admin only)
async function addMenuItem(itemData) {
    return apiCall('/items', 'POST', itemData);
}

// Update menu item (admin only)
async function updateMenuItem(itemId, itemData) {
    return apiCall(`/items/${itemId}`, 'PUT', itemData);
}

// Delete menu item (admin only)
async function deleteMenuItem(itemId) {
    return apiCall(`/items/${itemId}`, 'DELETE');
}

// Get all categories
async function getCategories() {
    return apiCall('/categories');
}

// Add new category (admin only)
async function addCategory(categoryData) {
    return apiCall('/categories', 'POST', categoryData);
}

// Update category (admin only)
async function updateCategory(categoryId, categoryData) {
    return apiCall(`/categories/${categoryId}`, 'PUT', categoryData);
}

// Delete category (admin only)
async function deleteCategory(categoryId) {
    return apiCall(`/categories/${categoryId}`, 'DELETE');
}

// Export functions
window.menuApi = {
    getMenuItems,
    getMenuItemsByCategory,
    addMenuItem,
    updateMenuItem,
    deleteMenuItem,
    getCategories,
    addCategory,
    updateCategory,
    deleteCategory
}; 