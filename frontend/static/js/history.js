document.addEventListener('DOMContentLoaded', function() {
    loadHistory();
});

function loadHistory() {
    try {
        const historyData = JSON.parse(localStorage.getItem('orderHistory') || '{"orders":[]}');
        const orders = historyData.orders || [];
        
        // Подсчитываем статистику
        const completedOrders = orders.filter(order => order.status === 'completed');
        const totalAmount = completedOrders.reduce((sum, order) => sum + order.total, 0);
        
        // Обновляем заголовок
        document.querySelector('.orders-status__title').textContent = 
            `Выполнено ${completedOrders.length} заказов`;
        document.querySelector('.orders-status__subtitle').textContent = 
            `Сумма заказов: ${formatMoney(totalAmount)} KZT`;

        const historyList = document.getElementById('historyList');
        if (!orders.length) {
            historyList.innerHTML = '<div class="no-orders">История заказов пуста</div>';
            return;
        }

        // Сортируем заказы по времени (новые сверху)
        orders.sort((a, b) => new Date(b.completedAt || b.cancelledAt) - new Date(a.completedAt || a.cancelledAt));

        historyList.innerHTML = orders.map(order => `
            <div class="order-card ${order.status === 'completed' ? 'order-card--green' : 'order-card--red'}">
                <div class="order-card__header">
                    <div class="order-card__id">#${order.id}</div>
                    <div class="order-card__info">
                        <div class="order-card__table">Стол ${order.tableId}</div>
                        <div class="order-card__time">${formatOrderTime(order.completedAt || order.cancelledAt)}</div>
                    </div>
                </div>
                <div class="order-card__items">
                    ${formatOrderItems(order.items)}
                </div>
                <div class="order-card__footer">
                    <div class="order-card__total">${formatMoney(order.total)} KZT</div>
                    <div class="status-badge ${order.status === 'completed' ? 'status-badge--paid' : 'status-badge--cancelled'}">
                        ${order.status === 'completed' ? 'Оплачен' : 'Отменён'}
                    </div>
                </div>
            </div>
        `).join('');

    } catch (error) {
        console.error('Error loading history:', error);
    }
}

function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

function formatOrderItems(items) {
    return items.map(item => item.name).join(', ');
}

function formatOrderTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit',
        day: '2-digit',
        month: '2-digit',
        year: '2-digit'
    });
} 