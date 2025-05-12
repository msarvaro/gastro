// Waiter API functions
const API_BASE = '/api/waiter';

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

// Table functions
async function getTables() {
    return apiCall('/tables');
}

async function getTableStatus() {
    return apiCall('/tables/status');
}

async function updateTableStatus(tableId, status) {
    return apiCall(`/tables/${tableId}/status`, 'PUT', { status });
}

// Order functions
async function getOrders() {
    return apiCall('/orders');
}

async function getOrderById(orderId) {
    return apiCall(`/orders/${orderId}`);
}

async function createOrder(orderData) {
    return apiCall('/orders', 'POST', orderData);
}

async function updateOrder(orderId, orderData) {
    return apiCall(`/orders/${orderId}`, 'PUT', orderData);
}

async function getOrderStatus() {
    return apiCall('/orders/status');
}

async function getOrderHistory() {
    return apiCall('/orders/history');
}

// Helper functions for formatting
function formatOrderTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatOrderItems(items) {
    if (!items || !items.length) return '';
    return items.map(item => `${item.name} x${item.quantity}`).join(', ');
}

function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

// Export functions
window.waiterApi = {
    // Table functions
    getTables,
    getTableStatus,
    updateTableStatus,
    
    // Order functions
    getOrders,
    getOrderById,
    createOrder,
    updateOrder,
    getOrderStatus,
    getOrderHistory,
    
    // Helper functions
    formatOrderTime,
    formatOrderItems,
    formatMoney
}; 