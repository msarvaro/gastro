document.addEventListener('DOMContentLoaded', function() {
    // Проверка авторизации
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser || currentUser.role !== 'waiter') {
        window.location.href = 'login.html';
        return;
    }

    loadOrders();
    updateOrdersStatus();
});

function loadOrders() {
    try {
        const ordersData = JSON.parse(localStorage.getItem('orders') || '{"orders":[]}');
        const orders = ordersData.orders || [];
        
        // Обновляем статистику
        updateOrdersStatus();

        const ordersList = document.getElementById('ordersList');
        if (!orders.length) {
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
                        <div class="order-card__time">${formatOrderTime(order.createdAt)}</div>
                    </div>
                </div>
                <div class="order-card__items">
                    ${formatOrderItems(order.items)}
                </div>
                ${order.comment ? `
                    <div class="order-card__comment">
                        <div class="comment-icon">💬</div>
                        <div class="comment-text">${order.comment}</div>
                    </div>
                ` : ''}
                <div class="order-card__footer">
                    <div class="order-card__total">${formatMoney(order.total)} KZT</div>
                    <div class="order-actions">
                        ${getActionButtons(order)}
                        ${order.status !== 'served' ? `
                            <button onclick="cancelOrder(${order.id})" class="status-badge status-badge--cancel">
                                Отменить
                            </button>
                        ` : ''}
                    </div>
                </div>
            </div>
        `).join('');

    } catch (error) {
        console.error('Error loading orders:', error);
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

function updateOrderStatus(orderId, newStatus) {
    try {
        const ordersData = JSON.parse(localStorage.getItem('orders') || '{"orders":[]}');
        const orderIndex = ordersData.orders.findIndex(order => order.id === orderId);
        
        if (orderIndex !== -1) {
            ordersData.orders[orderIndex].status = newStatus;
            localStorage.setItem('orders', JSON.stringify(ordersData));
            
            // Обновляем статистику при изменении статуса
            updateOrdersStatus();
            loadOrders();
        }
    } catch (error) {
        console.error('Error updating order status:', error);
    }
}

function cancelOrder(orderId) {
    try {
        const ordersData = JSON.parse(localStorage.getItem('orders') || '{"orders":[]}');
        const orderIndex = ordersData.orders.findIndex(order => order.id === orderId);
        
        if (orderIndex !== -1) {
            const cancelledOrder = ordersData.orders[orderIndex];
            cancelledOrder.status = 'cancelled';
            cancelledOrder.cancelledAt = new Date().toISOString();
            
            // Добавляем в историю
            const historyData = JSON.parse(localStorage.getItem('orderHistory') || '{"orders":[]}');
            historyData.orders.push(cancelledOrder);
            localStorage.setItem('orderHistory', JSON.stringify(historyData));
            
            // Удаляем из активных заказов
            ordersData.orders.splice(orderIndex, 1);
            localStorage.setItem('orders', JSON.stringify(ordersData));
            
            // Обновляем статистику после отмены
            updateOrdersStatus();
            loadOrders();
        }
    } catch (error) {
        console.error('Error cancelling order:', error);
    }
}

function completeOrder(orderId) {
    try {
        const ordersData = JSON.parse(localStorage.getItem('orders') || '{"orders":[]}');
        const orderIndex = ordersData.orders.findIndex(order => order.id === orderId);
        
        if (orderIndex !== -1) {
            const completedOrder = ordersData.orders[orderIndex];
            completedOrder.status = 'completed';
            completedOrder.completedAt = new Date().toISOString();
            
            // Добавляем в историю
            const historyData = JSON.parse(localStorage.getItem('orderHistory') || '{"orders":[]}');
            historyData.orders.push(completedOrder);
            localStorage.setItem('orderHistory', JSON.stringify(historyData));
            
            // Удаляем из активных заказов
            ordersData.orders.splice(orderIndex, 1);
            localStorage.setItem('orders', JSON.stringify(ordersData));
            
            // Перенаправляем на страницу истории
            window.location.href = 'history.html';
        }
    } catch (error) {
        console.error('Error completing order:', error);
    }
}

function updateOrdersStatus() {
    try {
        const ordersData = JSON.parse(localStorage.getItem('orders') || '{"orders":[]}');
        const orders = ordersData.orders || [];
        
        // Подсчитываем статистику по статусам
        const stats = {
            total: orders.length,
            new: orders.filter(order => order.status === 'new').length,
            preparing: orders.filter(order => ['accepted', 'preparing'].includes(order.status)).length,
            ready: orders.filter(order => order.status === 'ready').length
        };
        
        // Обновляем заголовок
        document.querySelector('.orders-status__title').textContent = 
            `${stats.total} активных заказов`;
        document.querySelector('.orders-status__subtitle').textContent = 
            `Новых: ${stats.new} | В работе: ${stats.preparing} | Готовых: ${stats.ready}`;
    } catch (error) {
        console.error('Error updating orders status:', error);
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
    window.location.href = 'create-order.html';
}

// Добавляем форматирование денег как в history.js
function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

// Обновляем обработчик создания заказа
function createOrder(orderData) {
    try {
        const ordersData = JSON.parse(localStorage.getItem('orders') || '{"orders":[]}');
        
        // Добавляем новый заказ
        const newOrder = {
            ...orderData,
            id: Date.now(),
            status: 'new',
            createdAt: new Date().toISOString()
        };
        
        ordersData.orders.push(newOrder);
        localStorage.setItem('orders', JSON.stringify(ordersData));
        
        // Обновляем статистику сразу после создания заказа
        updateOrdersStatus();
        loadOrders();
    } catch (error) {
        console.error('Error creating order:', error);
    }
}