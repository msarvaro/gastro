// kitchen.js

document.addEventListener('DOMContentLoaded', () => {
    // Инициализируем активную вкладку на основе URL или используем 'queue' по умолчанию
    const defaultTab = window.location.hash.substring(1) || 'queue';
    initNav();
    openTab(defaultTab);
    
    // Обновляем данные каждые 30 секунд
    setInterval(() => {
        const activeTab = document.querySelector('.nav-link.active').dataset.tab;
        if (activeTab) {
            openTab(activeTab, false);
        }
    }, 30000);
});

function initNav() {
    // Добавляем обработчики на навигацию
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const tab = link.dataset.tab;
            
            // Обновляем хэш без автоматической прокрутки
            history.pushState(null, '', `#${tab}`);
            
            openTab(tab);
        });
    });
}

function openTab(tabName, updateUI = true) {
    // Скрываем все вкладки
    document.querySelectorAll('.tab-pane').forEach(pane => {
        pane.classList.remove('active');
    });
    
    // Деактивируем все кнопки навигации
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    
    // Показываем выбранную вкладку и активируем кнопку
    const tabPane = document.getElementById(tabName);
    if (tabPane) {
        tabPane.classList.add('active');
    }
    
    const navLink = document.querySelector(`.nav-link[data-tab="${tabName}"]`);
    if (navLink) {
        navLink.classList.add('active');
    }
    
    // Прокрутка в начало страницы при смене вкладки
    window.scrollTo(0, 0);
    
    // Загружаем данные в зависимости от выбранной вкладки
    if (tabName === 'queue') {
        loadKitchenOrders();
    } else if (tabName === 'inventory') {
        loadInventory();
    } else if (tabName === 'history') {
        loadKitchenHistory();
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
        const response = await fetch('/api/kitchen/orders', {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка загрузки заказов: ${response.status}`);
        }
        
        const data = await response.json();
        
        // Обновляем счетчик активных заказов
        const queueStatus = document.querySelector('#queueStatus span');
        if (queueStatus) {
            queueStatus.textContent = data.orders?.length || 0;
        }
        
        // Если нет заказов
        if (!data.orders || data.orders.length === 0) {
            ordersListEl.innerHTML = '<div class="no-orders">Нет активных заказов</div>';
            return;
        }
        
        // Отрисовываем заказы
        ordersListEl.innerHTML = data.orders.map(order => `
            <div class="order-card order-card--${order.status}">
                <div class="order-card__header">
                    <div class="order-card__id">Заказ #${order.id}</div>
                    <div class="order-card__info">
                        <div class="order-card__table">Стол ${order.table_id}</div>
                        <div class="order-card__time">${formatOrderTime(order.created_at)}</div>
                    </div>
                </div>
                <div class="order-card__items">
                    ${order.items.map(item => `
                        <div>${item.quantity} × ${item.name}</div>
                    `).join('')}
                </div>
                <div class="order-card__footer">
                    <div class="order-card__waiter">официант: ${order.waiter_name || order.waiter_id}</div>
                    <button class="status-button status-button--ready" onclick="updateOrderStatusByCook(${order.id}, 'ready')">Готово</button>
                </div>
            </div>
        `).join('');
        
    } catch (error) {
        console.error('Ошибка загрузки заказов:', error);
        ordersListEl.innerHTML = `<div class="error">Ошибка загрузки заказов: ${error.message}</div>`;
    }
}

async function loadKitchenHistory() {
    const historyListEl = document.getElementById('kitchenHistoryList');
    historyListEl.innerHTML = '<div class="loading">Загрузка истории...</div>';
    
    try {
        const response = await fetch('/api/kitchen/history', {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка загрузки истории: ${response.status}`);
        }
        
        const data = await response.json();
        
        // Обновляем счетчик заказов в истории
        const historyStatus = document.querySelector('#historyStatus span');
        if (historyStatus) {
            historyStatus.textContent = data.total || data.orders?.length || 0;
        }
        
        // Если нет заказов в истории
        if (!data.orders || data.orders.length === 0) {
            historyListEl.innerHTML = '<div class="no-orders">История заказов пуста</div>';
            return;
        }
        
        // Отрисовываем историю заказов
        historyListEl.innerHTML = data.orders.map(order => `
            <div class="order-card ${order.status === 'completed' ? 'order-card--completed' : ''}">
                <div class="order-card__header">
                    <div class="order-card__id">Заказ #${order.id}</div>
                    <div class="order-card__info">
                        <div class="order-card__table">Стол ${order.table_id}</div>
                        <div class="order-card__time">${formatOrderTime(order.completed_at || order.created_at)}</div>
                    </div>
                </div>
                <div class="order-card__items">
                    ${order.items.map(item => `
                        <div>${item.quantity} × ${item.name}</div>
                    `).join('')}
                </div>
                <div class="order-card__footer">
                    <div class="order-card__waiter">официант: ${order.waiter_name || order.waiter_id}</div>
                    <div class="status-badge status-badge--${order.status}">${getStatusText(order.status)}</div>
                </div>
            </div>
        `).join('');
        
    } catch (error) {
        console.error('Ошибка загрузки истории заказов:', error);
        historyListEl.innerHTML = `<div class="error">Ошибка загрузки истории: ${error.message}</div>`;
    }
}

async function loadInventory() {
    const inventoryTableEl = document.getElementById('inventoryList');
    inventoryTableEl.innerHTML = '<tr><td colspan="6" class="loading-cell">Загрузка запасов...</td></tr>';
    
    try {
        const response = await fetch('/api/kitchen/inventory', {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка загрузки запасов: ${response.status}`);
        }
        
        const data = await response.json();
        
        // Обновляем статус последнего обновления
        const inventoryStatus = document.querySelector('#inventoryStatus span');
        if (inventoryStatus && data.last_updated) {
            inventoryStatus.textContent = formatTimeAgo(data.last_updated);
        }
        
        // Если нет запасов
        if (!data.items || data.items.length === 0) {
            inventoryTableEl.innerHTML = '<tr><td colspan="6" class="loading-cell">Нет данных о запасах</td></tr>';
            return;
        }
        
        // Отрисовываем таблицу запасов
        inventoryTableEl.innerHTML = data.items.map(item => `
            <tr>
                <td>
                    ${getStatusIndicator(item.quantity, item.critical_level)}
                    ${item.name}
                </td>
                <td>${item.category}</td>
                <td>${item.quantity} ${item.unit}</td>
                <td>${item.used_today || 0} ${item.unit}</td>
                <td>${item.waste_today || 0} ${item.unit}</td>
                <td>${formatTimeAgo(item.updated_at)}</td>
            </tr>
        `).join('');
        
    } catch (error) {
        console.error('Ошибка загрузки запасов:', error);
        inventoryTableEl.innerHTML = `<tr><td colspan="6" class="loading-cell error">Ошибка загрузки запасов: ${error.message}</td></tr>`;
    }
}

async function updateOrderStatusByCook(orderId, status) {
    try {
        const response = await fetch(`/api/kitchen/orders/${orderId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify({ status: status })
        });

        if (!response.ok) {
            throw new Error(`Ошибка при обновлении статуса: ${response.status}`);
        }

        // Обновляем список заказов
        loadKitchenOrders();
        
    } catch (error) {
        console.error('Ошибка при обновлении статуса заказа:', error);
        alert(`Ошибка при обновлении статуса заказа: ${error.message}`);
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

// Вспомогательные функции
function formatOrderTime(dateString) {
    if (!dateString) return "Не указано";
    
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
    const statusMap = {
        'new': 'Новый',
        'accepted': 'Принят',
        'preparing': 'Готовится',
        'ready': 'Готов',
        'served': 'Подан',
        'completed': 'Завершен',
        'cancelled': 'Отменен'
    };
    
    return statusMap[status] || status;
}

function getStatusIndicator(quantity, criticalLevel) {
    if (quantity <= criticalLevel) {
        return '<span class="status-indicator status-indicator--red"></span>';
    } else if (quantity <= criticalLevel * 2) {
        return '<span class="status-indicator status-indicator--yellow"></span>';
    } else {
        return '<span class="status-indicator status-indicator--green"></span>';
    }
}

function formatTimeAgo(timestamp) {
    if (!timestamp) return "Не указано";
    
    const date = new Date(timestamp);
    if (isNaN(date.getTime())) {
        return "Некорректная дата";
    }
    
    const now = new Date();
    const diff = Math.floor((now - date) / 1000); // разница в секундах
    
    if (diff < 60) return "Только что";
    if (diff < 3600) return `${Math.floor(diff / 60)} минут назад`;
    if (diff < 86400) return `${Math.floor(diff / 3600)} часов назад`;
    
    return date.toLocaleDateString('ru-RU');
}

// Add more helper functions if needed, e.g., for fetching user info for the header 