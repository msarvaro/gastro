// kitchen.js

document.addEventListener('DOMContentLoaded', () => {
    // Add event listeners for tab switching
    document.querySelectorAll('.tab-button').forEach(button => {
        button.addEventListener('click', () => {
            const tab = button.dataset.tab;
            openTab(tab);
        });
    });

    // Load user info on page load
    loadUserInfo();

    // Load the default tab (queue)
    openTab('queue');
});

function openTab(tabName) {
    // Hide all tab panes
    document.querySelectorAll('.tab-pane').forEach(pane => {
        pane.classList.remove('active');
    });

    // Deactivate all tab buttons
    document.querySelectorAll('.tab-button').forEach(button => {
        button.classList.remove('active');
    });

    // Show the selected tab pane and activate the button
    document.getElementById(tabName).classList.add('active');
    document.querySelector(`.tab-button[data-tab="${tabName}"]`).classList.add('active');

    // Load data based on the tab
    if (tabName === 'queue') {
        loadKitchenOrders();
    } else if (tabName === 'history') {
        loadKitchenHistory();
    } else if (tabName === 'inventory') {
        loadInventory();
    }
}

async function loadUserInfo() {
    try {
        // Attempt to load user info from token if available, otherwise fetch from a placeholder endpoint
        const token = localStorage.getItem('token');
        if (token) {
            try {
                // Simple client-side decode to get role (assuming non-sensitive claims)
                const base64Url = token.split('.')[1];
                const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
                const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
                    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
                }).join(''));

                const claims = JSON.parse(jsonPayload);
                const username = claims.username || 'User'; // Assuming username claim or default
                const role = claims.role || 'Unknown Role';

                const userInfoEl = document.querySelector('.user-info');
                if(userInfoEl) {
                     userInfoEl.innerHTML = `Logged in as: ${username} (${role})`; // Display username and role from token
                }
                console.log('User info loaded from token.');
                return; // Exit if info loaded from token

            } catch (e) {
                console.error('Failed to decode token:', e);
                // Fallback to fetching from endpoint if token decode fails
            }
        }

        // Fallback: Fetch from the backend endpoint (assuming /api/kitchen/profile exists or will exist)
        // Note: A general /api/profile endpoint is recommended for all roles
        console.log('Attempting to fetch user info from backend...');
        const resp = await fetch('/api/kitchen/profile', { headers: { 'Authorization': `Bearer ${token}` } }); // Using a kitchen profile endpoint placeholder
        if (!resp.ok) throw new Error('Failed to load user info from backend');
        const user = await resp.json();

        const userInfoEl = document.querySelector('.user-info');
        if(userInfoEl) {
             userInfoEl.innerHTML = `Logged in as: ${user.username} (${user.role})`; // Assuming username and role are available
        }
        console.log('User info loaded from backend.', user);

    } catch (error) {
         console.error('Error loading user info:', error);
         const userInfoEl = document.querySelector('.user-info');
         if(userInfoEl) {
              userInfoEl.innerHTML = 'Failed to load user info'; // Display error message
         }
    }
}

async function loadKitchenOrders() {
    const ordersListEl = document.getElementById('kitchenOrdersList');
    ordersListEl.innerHTML = '<div class="loading">Загрузка заказов...</div>';
    try {
        const resp = await fetch('/api/kitchen/orders', { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } });
        if (!resp.ok) throw new Error('Failed to load kitchen orders');
        const data = await resp.json();
        console.log('Kitchen Orders:', data);

        if (!data.orders || data.orders.length === 0) {
            ordersListEl.innerHTML = '<div class="no-orders">Нет заказов в очереди</div>';
            return;
        }

        ordersListEl.innerHTML = data.orders.map(order => `
            <div class="order-card order-card--${order.status}">
                <div class="order-card__header">
                    <div class="order-card__id">#${order.id}</div>
                    <div class="order-card__info">
                        <div class="order-card__table">Стол ${order.table_id}</div>
                        <div class="order-card__time">${formatOrderTime(order.created_at)}</div>
                    </div>
                </div>
                <div class="order-card__items">
                    ${order.items.map(item => `
                        <div>${item.quantity} x ${item.name} (${item.category})</div>
                    `).join('')}
                </div>
                 <div class="order-card__footer">
                     <div class="order-card__waiter">Официант: ${order.waiter_id}</div>
                     <button class="status-button status-button--ready" onclick="updateOrderStatusByCook(${order.id}, 'ready')">Готово</button>
                 </div>
            </div>
        `).join('');

    } catch (error) {
        console.error('Error loading kitchen orders:', error);
        ordersListEl.innerHTML = '<div class="error">Ошибка загрузки заказов.</div>';
    }
}

async function loadKitchenHistory() {
     const historyListEl = document.getElementById('kitchenHistoryList');
     historyListEl.innerHTML = '<div class="loading">Загрузка истории...</div>';
    try {
        const resp = await fetch('/api/kitchen/history', { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } });
        if (!resp.ok) throw new Error('Failed to load kitchen history');
        const data = await resp.json();
        console.log('Kitchen History Data:', data);

        if (!data.orders || data.orders.length === 0) {
            historyListEl.innerHTML = '<div class="no-orders">История заказов пуста</div>';
            return;
        }

        historyListEl.innerHTML = data.orders.map(order => `
            <div class="order-card order-card--${order.status}">
                <div class="order-card__header">
                    <div class="order-card__id">#${order.id}</div>
                    <div class="order-card__info">
                        <div class="order-card__table">Стол ${order.table_id}</div>
                        <div class="order-card__time">${formatOrderTime(order.created_at)}</div>
                    </div>
                </div>
                <div class="order-card__items">
                    ${order.items.map(item => `
                        <div>${item.quantity} x ${item.name} (${item.category})</div>
                    `).join('')}
                </div>
                 <div class="order-card__footer">
                     <div class="order-card__waiter">Официант: ${order.waiter_id}</div>
                     <div class="status-badge status-badge--${order.status}">${getStatusText(order.status)}</div>
                 </div>
            </div>
        `).join('');

    } catch (error) {
        console.error('Error loading kitchen history:', error);
        historyListEl.innerHTML = '<div class="error">Ошибка загрузки истории.</div>';
    }
}

async function loadInventory() {
     const inventoryListEl = document.getElementById('inventoryList');
     inventoryListEl.innerHTML = '<div class="loading">Загрузка запасов...</div>';
    try {
        const resp = await fetch('/api/kitchen/inventory', { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } });
        if (!resp.ok) throw new Error('Failed to load inventory');
        const data = await resp.json();
        console.log('Inventory:', data);

        if (!data.items || data.items.length === 0) {
            inventoryListEl.innerHTML = '<div class="no-inventory">Запасы не найдены</div>';
            return;
        }

        inventoryListEl.innerHTML = data.items.map(item => `
            <div class="inventory-item">
                <div class="item-name">${item.name}</div>
                <div class="item-quantity">Количество: <span id="inventory-quantity-${item.id}">${item.quantity}</span> ${item.unit}</div>
                <div class="item-actions">
                    <button onclick="promptUpdateInventory(${item.id}, '${item.name}', ${item.quantity})">Редактировать</button>
                </div>
            </div>
        `).join('');

    } catch (error) {
        console.error('Error loading inventory:', error);
        inventoryListEl.innerHTML = '<div class="error">Ошибка загрузки запасов.</div>';
    }
}

async function updateOrderStatusByCook(orderId, status) {
    try {
        const resp = await fetch(`/api/kitchen/orders/${orderId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify({ status: status })
        });

        if (!resp.ok) {
            const errorText = await resp.text();
            throw new Error(`Failed to update order status: ${resp.status} - ${errorText}`);
        }

        // Reload the queue after successful update
        loadKitchenOrders();

    } catch (error) {
        console.error('Error updating order status:', error);
        alert('Не удалось обновить статус заказа: ' + error.message);
    }
}

async function promptUpdateInventory(itemId, itemName, currentQuantity) {
    const newQuantityStr = prompt(`Введите новое количество для "${itemName}" (Текущее: ${currentQuantity}):`);
    if (newQuantityStr === null) { // User cancelled prompt
        return;
    }

    const newQuantity = parseFloat(newQuantityStr);

    if (isNaN(newQuantity) || newQuantity < 0) {
        alert('Пожалуйста, введите корректное положительное число.');
        return;
    }

    // Optional: Add a confirmation step here if desired

    try {
        const resp = await fetch(`/api/kitchen/inventory/${itemId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify({ quantity: newQuantity })
        });

        if (!resp.ok) {
             const errorText = await resp.text();
             throw new Error(`Failed to update inventory: ${resp.status} - ${errorText}`);
        }

        // Update the displayed quantity on the page
        document.getElementById(`inventory-quantity-${itemId}`).textContent = newQuantity;
        alert('Запас успешно обновлен.');

    } catch (error) {
        console.error('Error updating inventory:', error);
        alert('Не удалось обновить запас: ' + error.message);
    }
}

// Helper functions (can be reused or adapted from waiter.js)
function formatOrderTime(dateString) {
     if (!dateString) {
        return "Не указано"; 
    }
    const date = new Date(dateString);
    if (isNaN(date.getTime())) { 
        return "Некорректная дата";
    }
    return date.toLocaleString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function getStatusText(status) {
    switch (status) {
        case 'new': return 'Новый';
        case 'accepted': return 'Принят';
        case 'preparing': return 'Готовится';
        case 'ready': return 'Готов';
        case 'served': return 'Подан';
        case 'completed': return 'Завершен';
        case 'cancelled': return 'Отменен';
        default: return status;
    }
}

// Add more helper functions if needed, e.g., for fetching user info for the header 