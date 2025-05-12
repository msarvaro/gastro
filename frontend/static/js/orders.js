document.addEventListener('DOMContentLoaded', function() {
    // Проверка авторизации
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    
    if (!token || role !== 'waiter') {
        window.location.href = '/';
        return;
    }

    loadOrders();
    updateOrdersStatus();
});

async function loadOrders() {
    try {
        const orders = await waiterApi.getOrders();
        
        // Обновляем статистику
        updateOrdersStatus();

        const ordersList = document.getElementById('ordersList');
        if (!orders || !orders.length) {
            ordersList.innerHTML = '<div class="no-orders">Нет активных заказов</div>';
            return;
        }

        // Сортируем заказы по времени (новые сверху)
        orders.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

        ordersList.innerHTML = orders.map(order => `
            <div class="order-card ${getOrderStatusClass(order.status)}">
                <div class="order-card__header">
                    <div class="order-card__id">#${order.id}</div>
                    <div class="order-card__info">
                        <div class="order-card__table">Стол ${order.tableId}</div>
                        <div class="order-card__time">${waiterApi.formatOrderTime(order.createdAt)}</div>
                    </div>
                </div>
                <div class="order-card__items">
                    ${waiterApi.formatOrderItems(order.items)}
                </div>
                ${order.comment ? `
                    <div class="order-card__comment">
                        <div class="comment-icon">💬</div>
                        <div class="comment-text">${order.comment}</div>
                    </div>
                ` : ''}
                <div class="order-card__footer">
                    <div class="order-card__total">${waiterApi.formatMoney(order.total)} KZT</div>
                    <div class="order-actions">
                        ${getActionButtons(order)}
                    </div>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading orders:', error);
        const ordersList = document.getElementById('ordersList');
        ordersList.innerHTML = '<div class="error-message">Ошибка загрузки заказов. Пожалуйста, попробуйте позже.</div>';
    }
}

function getOrderStatusClass(status) {
    const statusClasses = {
        'new': 'order-card--blue',
        'accepted': 'order-card--purple',
        'preparing': 'order-card--orange',
        'ready': 'order-card--green',
        'served': 'order-card--gray'
    };
    return statusClasses[status] || '';
}

function getStatusBadgeClass(status) {
    const badgeClasses = {
        'new': 'status-badge--new',
        'accepted': 'status-badge--accepted',
        'preparing': 'status-badge--preparing',
        'ready': 'status-badge--ready',
        'served': 'status-badge--served'
    };
    return badgeClasses[status] || '';
}

function getStatusText(status) {
    const statusTexts = {
        'new': 'Новый',
        'accepted': 'Принят',
        'preparing': 'Готовится',
        'ready': 'Готов',
        'served': 'Подан'
    };
    return statusTexts[status] || status;
}

function getActionButtons(order) {
    const statusButtons = {
        'new': {
            next: 'accepted',
            text: 'Принять',
            class: 'status-badge--new'
        },
        'accepted': {
            next: 'preparing',
            text: 'Готовится',
            class: 'status-badge--accepted'
        },
        'preparing': {
            next: 'ready',
            text: 'Готов',
            class: 'status-badge--preparing'
        },
        'ready': {
            next: 'served',
            text: 'Подан',
            class: 'status-badge--ready'
        },
        'served': {
            next: 'completed',
            text: 'Оплачен',
            class: 'status-badge--served',
            action: 'completeOrder'
        }
    };

    const currentStatus = statusButtons[order.status];
    if (!currentStatus) return '';

    const onClickAction = currentStatus.action === 'completeOrder' 
        ? `completeOrder(${order.id})`
        : `updateOrderStatus(${order.id}, '${currentStatus.next}')`;

    return `
        <button onclick="${onClickAction}" 
                class="status-badge ${currentStatus.class}">
            ${currentStatus.text}
        </button>
    `;
}

async function updateOrderStatus(orderId, newStatus) {
    try {
        await waiterApi.updateOrder(orderId, { status: newStatus });
        loadOrders(); // Reload orders after update
    } catch (error) {
        console.error('Error updating order status:', error);
        alert('Ошибка при обновлении статуса заказа. Пожалуйста, попробуйте позже.');
    }
}

async function cancelOrder(orderId) {
    try {
        await waiterApi.updateOrder(orderId, { status: 'cancelled' });
        loadOrders(); // Reload orders after cancellation
    } catch (error) {
        console.error('Error cancelling order:', error);
        alert('Ошибка при отмене заказа. Пожалуйста, попробуйте позже.');
    }
}

async function completeOrder(orderId) {
    try {
        await waiterApi.updateOrder(orderId, { status: 'completed' });
        // Перенаправляем на страницу истории
        window.location.href = '/waiter/history';
    } catch (error) {
        console.error('Error completing order:', error);
        alert('Ошибка при завершении заказа. Пожалуйста, попробуйте позже.');
    }
}

async function updateOrdersStatus() {
    try {
        const status = await waiterApi.getOrderStatus();
        
        // Обновляем заголовок
        document.querySelector('.orders-status__title').textContent = 
            `${status.total} активных заказов`;
        document.querySelector('.orders-status__subtitle').textContent = 
            `Новых: ${status.new} | В работе: ${status.accepted + status.preparing} | Готовых: ${status.ready}`;
    } catch (error) {
        console.error('Error updating orders status:', error);
        document.querySelector('.orders-status__title').textContent = 'Ошибка загрузки статуса';
        document.querySelector('.orders-status__subtitle').textContent = '';
    }
}

function formatOrderTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatOrderItems(items) {
    return items.map(item => item.name).join(', ');
}

function viewOrderDetails(orderId) {
    // Здесь можно добавить логику просмотра деталей заказа
    console.log(`Viewing order details for order #${orderId}`);
}

// Функция для создания нового заказа
function createNewOrder() {
    window.location.href = '/waiter/create-order';
}

// Добавляем форматирование денег как в history.js
function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

// Обновляем обработчик создания заказа
async function createOrder(orderData) {
    try {
        await waiterApi.createOrder(orderData);
        // Обновляем статистику сразу после создания заказа
        updateOrdersStatus();
        loadOrders();
    } catch (error) {
        console.error('Error creating order:', error);
        alert('Ошибка при создании заказа: ' + error.message);
    }
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    window.location.href = '/';
}
