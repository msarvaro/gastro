// waiter.js: объединённая логика для панели официанта

// Menu API functions
const API_BASE = '/api/menu';
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
    if (data) options.body = JSON.stringify(data);
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, options);
        if (response.status === 401 || response.status === 403) {
            window.location.href = '/';
            return null;
        }
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
}
window.menuApi = {
    getMenuItems: () => apiCall('/items'),
    getMenuItemsByCategory: (category) => apiCall(`/items/category/${encodeURIComponent(category)}`),
    addMenuItem: (itemData) => apiCall('/items', 'POST', itemData),
    updateMenuItem: (itemId, itemData) => apiCall(`/items/${itemId}`, 'PUT', itemData),
    deleteMenuItem: (itemId) => apiCall(`/items/${itemId}`, 'DELETE'),
    getCategories: () => apiCall('/categories'),
    addCategory: (categoryData) => apiCall('/categories', 'POST', categoryData),
    updateCategory: (categoryId, categoryData) => apiCall(`/categories/${categoryId}`, 'PUT', categoryData),
    deleteCategory: (categoryId) => apiCall(`/categories/${categoryId}`, 'DELETE')
};

// Основная инициализация
let currentOrderData = { tableId: null, items: [], comment: '', total: 0 };
document.addEventListener('DOMContentLoaded', function() {
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    if (!token || role !== 'waiter') {
        window.location.href = '/';
        return;
    }
    // Навигация по секциям
    document.querySelectorAll('.tab-item').forEach(tab => {
        tab.addEventListener('click', function(e) {
            e.preventDefault();
            const section = this.getAttribute('data-section');
            showSection(section);
            document.querySelectorAll('.tab-item').forEach(t => t.classList.remove('tab-item--active'));
            this.classList.add('tab-item--active');
        });
    });
    // Кнопка "Добавить заказ"
    const showCreateOrderBtn = document.getElementById('showCreateOrderBtn');
    if (showCreateOrderBtn) {
        showCreateOrderBtn.addEventListener('click', function() {
            document.getElementById('create-order-section').style.display = 'block';
            document.querySelector('.orders-section').style.display = 'none';
            resetOrderForm(); // Reset form and set initial UI state (menu/order disabled)
            renderMenu(); // Load menu data
            // Explicitly ensure table selection is the first step
            document.getElementById('selectTableBtn').focus(); 
        });
    }
    // Кнопка "Назад" в создании заказа
    const backToOrdersBtn = document.getElementById('backToOrdersBtn');
    if (backToOrdersBtn) {
        backToOrdersBtn.addEventListener('click', function() {
            document.getElementById('create-order-section').style.display = 'none';
            document.querySelector('.orders-section').style.display = 'block';
        });
    }
    document.getElementById('selectTableBtn').addEventListener('click', showTableModal);
    document.querySelector('.close-modal-btn').addEventListener('click', closeTableModal);
    document.querySelector('.create-order-btn').addEventListener('click', showConfirmOrderModal);
    document.getElementById('confirmOrderBtn').addEventListener('click', createOrder);
    document.getElementById('cancelOrderBtn').addEventListener('click', hideConfirmOrderModal);
    document.getElementById('clearOrderBtn').addEventListener('click', clearOrder);
 
    const currentPath = window.location.pathname;
        const sections = {
            '/waiter': 'tables',
            '/waiter/orders': 'orders',
            '/waiter/history': 'history',
            '/waiter/profile': 'profile',
        };
        
        const activeSection = sections[currentPath] || 'tables';
        showSection(activeSection);
});

function showSection(section) {
    ['tables', 'orders', 'history', 'profile'].forEach(s => {
        const el = document.getElementById('section-' + s);
        if (el) el.style.display = (s === section) ? 'block' : 'none';
    });
    if (section === 'orders') {
        document.getElementById('create-order-section').style.display = 'none';
        document.querySelector('.orders-section').style.display = 'block';
        loadOrders();
    }
    if (section === 'tables') loadTables();
    if (section === 'history') loadHistory();
    if (section === 'profile') loadProfile();
}

// Пример функций для загрузки данных (реализуйте по аналогии с вашими API)
async function loadTables() {
    try {
        const resp = await fetch('/api/waiter/tables', { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } });
        if (!resp.ok) throw new Error('Failed to load tables');
        const data = await resp.json();
        
        const tablesStatusInfo = document.getElementById('tablesStatusInfo');
        if (tablesStatusInfo && data.stats) {
            // Updated to match new CSS structure for header
            let occupancyPercentage = 0;
            if (data.stats.total > 0) {
                occupancyPercentage = Math.round(((data.stats.total - data.stats.free) / data.stats.total) * 100);
            }
            tablesStatusInfo.innerHTML = `
                <span class="occupancy-percentage">${occupancyPercentage}% занято</span><br>
                <span class="occupancy-status__subtitle">Количество свободных столов: ${data.stats.free} из ${data.stats.total}</span>
            `;
        }

        const grid = document.getElementById('tablesGrid');
        if (grid && data.tables) {
            grid.innerHTML = data.tables.map(table => `
                <div class="table-card table-card--${table.status.toLowerCase()}"> 
                    <div class="table-card__header">
                        <span class="status-dot status-dot--${table.status.toLowerCase()}"></span>
                        <span class="table-card__title">№${table.number}</span>
                        <span class="table-card__seats">${table.seats} мест</span>
                    </div>
                    <div class="table-card__content"> 
                        ${table.orders && table.orders.length ? table.orders.map(order => `
                            <div class="table-order">
                                <div class="table-order__id_container"> 
                                    <span class="table-order__id">#${order.id}</span>
                                    ${order.comment ? `<span class="table-order__comment-indicator" title="Есть комментарий">💬</span>` : ''}
                                </div>
                                <div class="table-order__time">${order.time}</div>
                                ${order.comment ? `<div class="table-order__comment-text">${order.comment}</div>` : ''}
                            </div>
                        `).join('') : (table.status.toLowerCase() === 'free' || table.status.toLowerCase() === 'available' ? '<p class="table-empty-message">Свободен</p>' : '')}
                    </div>
                </div>
            `).join('');
        }
    } catch (e) {
        console.error('Failed to load tables:', e);
        const tablesStatusInfo = document.getElementById('tablesStatusInfo');
        if (tablesStatusInfo) tablesStatusInfo.textContent = 'Ошибка загрузки столов';
    }
}
async function loadOrders() {
  
        const resp = await fetch('/api/waiter/orders', { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } });
        if (!resp.ok) throw new Error('Failed to load orders');
        const data = await resp.json();
        document.getElementById('ordersStatusInfo').textContent = `${data.stats.total_active_orders || 0} активных заказов`;
        const list = document.getElementById('ordersList');
        if (!data.orders || data.orders.length === 0) {
            list.innerHTML = '<div class="no-orders">Нет активных заказов</div>';
            return;
        }
        list.innerHTML = data.orders.map(order => `
            <div class="order-card order-card--${order.status}">
                <div class="order-card__header">
                    <div class="order-card__id">#${order.id}</div>
                    <div class="order-card__info">
                        <div class="order-card__table">Стол ${order.table_id}</div>
                        <div class="order-card__time">${formatOrderTime(order.created_at)}</div>
                    </div>
                </div>
                <div class="order-card__items">${order.items.map(item => item.name).join(', ')}</div>
                <div class="order-card__footer">
                    <div class="order-card__total">${formatMoney(order.total_amount)} KZT</div>
                    <div class="order-actions">
                        <button class="status-badge status-badge--${order.status}" onclick="updateOrderStatus(${order.id}, '${getNextStatus(order.status)}')">${getStatusText(order.status)}</button>
                    </div>
                </div>
            </div>
        `).join('');
}
async function loadHistory() {
    try {
        const resp = await fetch('/api/waiter/history', { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } });
        if (!resp.ok) throw new Error('Failed to load history');
        const data = await resp.json();

        const historyMainStatEl = document.getElementById('historyMainStat');
        const historySubStatEl = document.getElementById('historySubStat');

        if (data.stats && historyMainStatEl && historySubStatEl) {
            let mainStatText = '';
            let subStatText = '';

            // Используем data.stats.completed_total для количества и data.stats.total_amount_all для суммы
            // Замените на актуальные поля, если они другие
            if (data.stats.completed_total !== undefined) {
                mainStatText = `Выполнено ${data.stats.completed_total} заказов`;
            }
            if (data.stats.completed_amount_total !== undefined) { 
                subStatText = `Сумма заказов: ${formatMoney(data.stats.completed_amount_total)}`;
            }
            console.log(data.stats.completed_total);
            console.log(formatMoney(data.stats.completed_amount_total));

            historyMainStatEl.textContent = mainStatText;
            historySubStatEl.textContent = subStatText;

        } else if (historyMainStatEl && historySubStatEl) {
            historyMainStatEl.textContent = 'История заказов';
            historySubStatEl.textContent = 'Статистика недоступна';
        }
        
        const historyList = document.getElementById('historyList');
        if (!data.orders || !data.orders.length) {
            historyList.innerHTML = '<div class="no-orders">История заказов пуста</div>';
            if (historyMainStatEl) historyMainStatEl.textContent = 'История заказов пуста';
            if (historySubStatEl) historySubStatEl.textContent = ''; 
            return;
        }
        historyList.innerHTML = data.orders.map(order => `
            <div class="order-card ${order.status === 'completed' ? 'order-card--green' : 'order-card--red'}">
                <div class="order-card__header">
                    <div class="order-card__id">#${order.id}</div>
                    <div class="order-card__info">
                        <div class="order-card__table">Стол ${order.table_id}</div>
                        <div class="order-card__time">${formatOrderTime(order.completed_at || order.cancelled_at)}</div>
                    </div>
                </div>
                <div class="order-card__items">${order.items.map(item => item.name).join(', ')}</div>
                <div class="order-card__footer">
                    <div class="order-card__total">${formatMoney(order.total_amount)} KZT</div>
                    <div class="status-badge ${order.status === 'completed' ? 'status-badge--paid' : 'status-badge--cancelled'}">${order.status === 'completed' ? 'Оплачен' : 'Отменён'}</div>
                </div>
            </div>
        `).join('');
    } catch (e) {
        document.getElementById('historyList').innerHTML = '<div class="error-message">Ошибка загрузки истории</div>';
        const historyMainStatEl = document.getElementById('historyMainStat');
        const historySubStatEl = document.getElementById('historySubStat');
        if (historyMainStatEl) historyMainStatEl.textContent = 'История заказов';
        if (historySubStatEl) historySubStatEl.textContent = 'Ошибка загрузки статистики';
    }
}
async function loadProfile() {
    // Здесь можно реализовать загрузку профиля через API, если потребуется
}

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
function formatMoney(amount) {
    if (amount === undefined || amount === null) {
        return Number(0).toLocaleString('ru-RU'); 
    }
    const numberAmount = parseFloat(amount);
    if (isNaN(numberAmount)) {
        console.warn('formatMoney received a non-numeric value:', amount);
        return "0"; 
    }
    return numberAmount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}
function getStatusText(status) {
    const statusTexts = {
        'new': 'Новый',
        'accepted': 'Принят',
        'preparing': 'Готовится',
        'ready': 'Готов',
        'served': 'Подан',
        'completed': 'Оплачен',
        'cancelled': 'Отменён'
    };
    return statusTexts[status] || status;
}
function getNextStatus(status) {
    const flow = {
        'new': 'accepted',
        'accepted': 'preparing',
        'preparing': 'ready',
        'ready': 'served',
        'served': 'completed'
    };
    return flow[status] || status;
}
async function renderTables() {
        const resp = await fetch('/api/waiter/tables', { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } });
        if (!resp.ok) throw new Error('Failed to load tables');
        const data = await resp.json();
        const grid = document.querySelector('.table-modal__grid');
        console.log(data);
        grid.innerHTML = data.tables.map(table => `
            <div class="table-option ${table.status === 'occupied' ? 'occupied' : ''}" 
                 data-table-id="${table.id}">
                <div class="table-number">Стол ${table.number}</div>
                <div class="table-seats">${table.seats} мест</div>
            </div>
        `).join('');

        // Добавляем обработчики для каждого стола
        grid.querySelectorAll('.table-option:not(.occupied)').forEach(tableEl => {
            tableEl.addEventListener('click', () => {
                const rawTableId = tableEl.dataset.tableId;
                const parsedTableId = parseInt(rawTableId);
                console.log('[Waiter LOG] Table selected: rawTableId =', rawTableId, '(type:', typeof rawTableId, ')');
                console.log('[Waiter LOG] Table selected: parsedTableId =', parsedTableId, '(type:', typeof parsedTableId, ')');
                
                if (isNaN(parsedTableId) || parsedTableId <= 0) { // Assuming table IDs are positive
                    console.error('[Waiter LOG] Invalid Table ID after parse:', parsedTableId, 'from raw value:', rawTableId);
                    alert("Выбран неверный ID стола. Пожалуйста, попробуйте еще раз.");
                    return;
                }

                currentOrderData.tableId = parsedTableId;
                console.log('[Waiter LOG] currentOrderData.tableId has been set to:', currentOrderData.tableId, '(type:', typeof currentOrderData.tableId, ')');
                
                const table = data.tables.find(t => t.id === parsedTableId);
                if (!table) {
                    console.error('[Waiter LOG] Could not find table object for ID:', parsedTableId);
                    alert("Не удалось найти информацию о выбранном столе.");
                    return;
                }
                
                const selectedTableTextEl = document.getElementById('selectedTableText');
                if (selectedTableTextEl) selectedTableTextEl.textContent = `Стол №${table.number}`;
                
                const selectTableBtn = document.getElementById('selectTableBtn');
                if(selectTableBtn) selectTableBtn.textContent = `Стол №${table.number}`;

                tableModal.classList.remove('active');
                
                // Активируем форму создания заказа (меню и детали заказа)
                setCreateOrderInteractive(true);
            });
        });
    }

function showTableModal() {
    document.getElementById('tableModal').classList.add('active');
    renderTables();
}
function closeTableModal() {
    document.getElementById('tableModal').classList.remove('active');
}
function showConfirmOrderModal() {
    document.getElementById('confirmOrderModal').classList.add('active');
}
function hideConfirmOrderModal() {
    document.getElementById('confirmOrderModal').classList.remove('active');
}
function resetOrderForm() {
    currentOrderData = { tableId: null, items: [], comment: '', total: 0 };
    
    const selectedTableText = document.getElementById('selectedTableText');
    if(selectedTableText) selectedTableText.textContent = 'Выберите стол';
    
    const orderCommentInput = document.getElementById('order-comment-input');
    if(orderCommentInput) orderCommentInput.value = '';

    renderCurrentOrder(); // Clear the displayed order items and total

    const menuCategoriesContainer = document.getElementById('menu-categories-container');
    if (menuCategoriesContainer) {
        menuCategoriesContainer.querySelectorAll('.category-button').forEach(btn => btn.classList.remove('active'));
        const allButton = menuCategoriesContainer.querySelector('button[data-category-id="all"]');
        if (allButton) allButton.classList.add('active');
    }
    const menuDishesContainer = document.getElementById('menu-dishes-container');
    if(menuDishesContainer) menuDishesContainer.innerHTML = '<p>Выберите стол для активации меню.</p>'; 
    
    // Делаем секцию создания заказа (кроме выбора стола) неактивной
    setCreateOrderInteractive(false);
}
function clearOrder() {
    if (confirm('Вы уверены, что хотите очистить заказ?')) {
        resetOrderForm();
    }
}
async function createOrder() {
    console.log('[Waiter LOG] In createOrder(): currentOrderData.tableId =', currentOrderData.tableId, '(type:', typeof currentOrderData.tableId, ')');
    console.log('[Waiter LOG] In createOrder(): Client-side check ' + (!currentOrderData.tableId ? 'true' : 'false') + ' evaluates to:', !currentOrderData.tableId);

    if (!currentOrderData.tableId && currentOrderData.tableId !== 0) { // Allow 0 if it were valid, but error is "required" so 0 is bad
        alert("Пожалуйста, выберите стол.");
        return;
    }
    if (currentOrderData.items.length === 0) {
        alert("Пожалуйста, добавьте блюда в заказ.");
        return;
    }

    const orderCommentInput = document.getElementById('order-comment-input');
    currentOrderData.comment = orderCommentInput ? orderCommentInput.value : '';

    const payload = {
        tableId: currentOrderData.tableId,
        comment: currentOrderData.comment,
        items: currentOrderData.items.map(item => ({
            dishId: item.id,
            quantity: item.quantity,
            notes: item.notes || "" // Assuming notes might be added later
        }))
    };
    console.log(payload);

    try {
        const response = await fetch('/api/waiter/orders', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify(payload)
        });
        if (!response.ok) {
            let errorMsg = `Ошибка создания заказа: ${response.status}`;
            try {
                const errorData = await response.json(); // Attempt to parse JSON error response
                if (errorData && errorData.error) {
                    errorMsg = errorData.error; // Use specific error from backend
                }
            } catch (e) {
                // If response is not JSON or errorData.error is not present, stick to generic
                console.warn("Could not parse error response as JSON from backend", e);
            }
            throw new Error(errorMsg);
        }

        const createdOrder = await response.json();
        console.log('Order created:', createdOrder);
        alert('Заказ успешно создан!');
        
        hideConfirmOrderModal(); // Assuming this is still relevant for a final confirmation step
                               // If not, it can be removed or repurposed.
                               // The summary mentioned createOrderBtn -> showConfirmOrderModal -> createOrder
                               // So this flow implies a confirmation modal.

        // Switch back to orders view and refresh
        document.getElementById('create-order-section').style.display = 'none';
        document.querySelector('.orders-section').style.display = 'block';
        showSection('orders'); // This will also call loadOrders()
        resetOrderForm(); 

    } catch (error) {
        console.error('Failed to create order:', error);
        alert(`Не удалось создать заказ: ${error.message}`);
    }
}

let allMenuItems = [];
let allCategories = [];

async function renderMenu() {
    const menuCategoriesContainer = document.getElementById('menu-categories-container');
    const menuDishesContainer = document.getElementById('menu-dishes-container');

    if (!menuCategoriesContainer || !menuDishesContainer) {
        console.error('Menu containers not found in HTML.');
        return;
    }

    try {
        allCategories = await window.menuApi.getCategories();
        allMenuItems = await window.menuApi.getMenuItems();

        // Render categories
        menuCategoriesContainer.innerHTML = '<button class="category-button active" data-category-id="all">Все</button>'; // "All" button
        allCategories.forEach(category => {
            const button = document.createElement('button');
            button.className = 'category-button';
            button.textContent = category.name;
            button.dataset.categoryId = category.id;
            menuCategoriesContainer.appendChild(button);
        });

        // Add event listeners to category buttons
        menuCategoriesContainer.querySelectorAll('.category-button').forEach(button => {
            button.addEventListener('click', function() {
                menuCategoriesContainer.querySelectorAll('.category-button').forEach(btn => btn.classList.remove('active'));
                this.classList.add('active');
                filterDishesByCategory(this.dataset.categoryId);
            });
        });

        // Initial render of all dishes
        filterDishesByCategory('all'); 
    } catch (error) {
        console.error('Failed to load menu:', error);
        menuDishesContainer.innerHTML = '<p class="error-message">Не удалось загрузить меню.</p>';
    }
}

function filterDishesByCategory(categoryId) {
    const menuDishesContainer = document.getElementById('menu-dishes-container');
    menuDishesContainer.innerHTML = ''; // Clear previous dishes

    const itemsToDisplay = categoryId === 'all' 
        ? allMenuItems 
        : allMenuItems.filter(item => item.category_id === parseInt(categoryId));

    if (itemsToDisplay.length === 0) {
        menuDishesContainer.innerHTML = '<p>Нет блюд в этой категории.</p>';
        return;
    }
    
    itemsToDisplay.forEach(dish => {
        if (!dish.is_available) return; // Skip unavailable dishes

        const dishCard = document.createElement('div');
        dishCard.className = 'dish-card';
        // Note: dish properties are based on common patterns (id, name, price, description, image_url, category_id, is_available)
        // Adjust if your menu item structure is different (e.g. from menuApi.getMenuItems())
        dishCard.innerHTML = `
            <div class="dish-card__image-container">
                ${dish.image_url ? `<img src="${dish.image_url}" alt="${dish.name}" class="dish-card__image">` : '<div class="dish-card__image_placeholder">Нет фото</div>'}
            </div>
            <div class="dish-card__details">
                <h4 class="dish-card__name">${dish.name}</h4>
                <p class="dish-card__price">${formatMoney(dish.price)} KZT</p>
                ${dish.description ? `<p class="dish-card__description">${dish.description}</p>` : ''}
            </div>
            <button class="dish-card__add-btn" data-dish-id="${dish.id}">+</button>
        `;
        // Add event listener to the add button
        dishCard.querySelector('.dish-card__add-btn').addEventListener('click', () => {
            addDishToOrder(dish); // Pass the full dish object
        });
        menuDishesContainer.appendChild(dishCard);
    });
}

function addDishToOrder(dish) {
    const existingItem = currentOrderData.items.find(item => item.id === dish.id);
    if (existingItem) {
        existingItem.quantity++;
    } else {
        currentOrderData.items.push({ 
            id: dish.id, 
            name: dish.name, 
            price: dish.price, 
            quantity: 1,
            // notes: "" // Optional: initialize notes if you have a way to add them per item
        });
    }
    renderCurrentOrder();
}

function removeDishFromOrder(dishId) {
    const itemIndex = currentOrderData.items.findIndex(item => item.id === dishId);
    if (itemIndex > -1) {
        currentOrderData.items[itemIndex].quantity--;
        if (currentOrderData.items[itemIndex].quantity <= 0) {
            currentOrderData.items.splice(itemIndex, 1);
        }
    }
    renderCurrentOrder();
}

function renderCurrentOrder() {
    const currentOrderItemsContainer = document.getElementById('current-order-items');
    const currentOrderTotalEl = document.getElementById('current-order-total');

    if (!currentOrderItemsContainer || !currentOrderTotalEl) {
        console.error('Current order display elements not found.');
        return;
    }

    currentOrderItemsContainer.innerHTML = '';
    let totalAmount = 0;

    if (currentOrderData.items.length === 0) {
        currentOrderItemsContainer.innerHTML = '<p class="empty-order-message">Заказ пуст. Добавьте блюда из меню.</p>';
    } else {
        currentOrderData.items.forEach(item => {
            const itemElement = document.createElement('div');
            itemElement.className = 'current-order-item';
            itemElement.innerHTML = `
                <span class="item-name">${item.name}</span>
                <div class="item-controls">
                    <button class="item-quantity-btn" onclick="decrementOrderItem(${item.id})">-</button>
                    <span class="item-quantity">${item.quantity}</span>
                    <button class="item-quantity-btn" onclick="incrementOrderItem(${item.id})">+</button>
                </div>
                <span class="item-price">${formatMoney(item.price * item.quantity)} KZT</span>
                <button class="item-remove-btn" onclick="removeDishFromOrder(${item.id})">&times;</button>
            `;
            currentOrderItemsContainer.appendChild(itemElement);
            totalAmount += item.price * item.quantity;
        });
    }
    
    currentOrderData.total = totalAmount;
    currentOrderTotalEl.textContent = `Итого: ${formatMoney(totalAmount)} KZT`;
    
    // Update create order button state
    const createOrderBtn = document.querySelector('.create-order-btn'); // The one that shows confirm modal
    if (createOrderBtn) {
        if (currentOrderData.items.length > 0 && currentOrderData.tableId) {
            createOrderBtn.disabled = false;
        } else {
            createOrderBtn.disabled = true;
        }
    }
}

// Helper functions for increment/decrement buttons in current order
function incrementOrderItem(dishId) {
    const dish = allMenuItems.find(d => d.id === dishId); // Assuming allMenuItems is populated
    if (dish) {
        addDishToOrder(dish); // addDishToOrder handles incrementing quantity
    }
}

function decrementOrderItem(dishId) {
    // This is essentially the same as removeDishFromOrder's logic for decrementing.
    // For simplicity, we can directly call removeDishFromOrder which handles decrementing and removal if quantity reaches 0.
    // If a different behavior is needed (e.g., never fully removing via '-' button, only via 'x'), this function would be different.
    removeDishFromOrder(dishId);
}

async function updateOrderStatus(orderId, newStatus) {
    try {
        await fetch(`/api/waiter/orders/${orderId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify({ status: newStatus })
        });
        loadOrders();
    } catch (e) {
        alert('Ошибка при обновлении статуса заказа');
    }
}

// Helper function to enable/disable create order UI parts
function setCreateOrderInteractive(isInteractive) {
    const UIElementsToToggle = [
        document.getElementById('menu-categories-container'),
        document.getElementById('menu-dishes-container'),
        document.getElementById('current-order-items'),
        // document.getElementById('current-order-total'), // Total usually just display
        document.getElementById('order-comment-input'),
        document.querySelector('.create-order-btn') // The main button to finalize order
    ];

    UIElementsToToggle.forEach(element => {
        if (element) {
            if (isInteractive) {
                element.style.opacity = '1';
                element.style.pointerEvents = 'auto';
                if (element.tagName === 'BUTTON' || element.tagName === 'INPUT' || element.tagName === 'TEXTAREA') {
                    element.disabled = false;
                }
            } else {
                element.style.opacity = '0.5';
                element.style.pointerEvents = 'none';
                if (element.tagName === 'BUTTON' || element.tagName === 'INPUT' || element.tagName === 'TEXTAREA') {
                    element.disabled = true;
                }
            }
        }
    });

    // Specifically for dishes container message when not interactive
    const menuDishesContainer = document.getElementById('menu-dishes-container');
    if (menuDishesContainer && !isInteractive && !currentOrderData.tableId) {
         menuDishesContainer.innerHTML = '<p>Пожалуйста, сначала выберите стол, чтобы активировать меню.</p>';
    } else if (menuDishesContainer && isInteractive && menuDishesContainer.innerHTML.includes('Пожалуйста, сначала выберите стол')) {
        // If it became interactive and still has the message, trigger menu render if needed
        // renderMenu(); // Or just clear it if renderMenu was already called
        // filterDishesByCategory('all'); // assuming renderMenu already populated allMenuItems
         menuDishesContainer.innerHTML = '<p>Загрузка блюд...</p>'; // Placeholder, filterDishesByCategory will fill it
         if(allMenuItems.length > 0) { // if menu items were already fetched by renderMenu()
            filterDishesByCategory('all');
         }
    }
}

// Call at the end of DOMContentLoaded
document.addEventListener('DOMContentLoaded', function() {
    // ... existing listeners ...
    
    // Initial state for create order section if it's visible by default (e.g. direct navigation)
    // For SPA, this is better handled by the click on showCreateOrderBtn
    if (document.getElementById('create-order-section').style.display === 'block') {
       if (!currentOrderData.tableId) {
            setCreateOrderInteractive(false);
            const menuDishesContainer = document.getElementById('menu-dishes-container');
            if (menuDishesContainer) {
                 menuDishesContainer.innerHTML = '<p>Пожалуйста, сначала выберите стол, чтобы активировать меню.</p>';
            }
       }
    }
}); 