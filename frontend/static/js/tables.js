document.addEventListener('DOMContentLoaded', function() {
    // Проверка авторизации
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser || currentUser.role !== 'waiter') {
        window.location.href = 'login.html';
        return;
    }

    loadTables();
    updateTablesStatus();
});

// Функция загрузки и отображения столов
function loadTables() {
    try {
        // Получаем активные заказы
        const ordersData = JSON.parse(localStorage.getItem('orders') || '{"orders":[]}');
        const activeOrders = ordersData.orders || [];
        
        // Получаем резервации столов
        const reservationsData = JSON.parse(localStorage.getItem('reservations') || '{"reservations":[]}');
        const activeReservations = reservationsData.reservations || [];
        
        // Получаем контейнер для столов
        const tablesGrid = document.getElementById('tablesGrid');
        let tablesHTML = '';
        
        // Создаем 6 столов (как в create-order.html)
        for (let i = 1; i <= 6; i++) {
            // Проверяем, есть ли активные заказы для данного стола
            const tableOrders = activeOrders.filter(order => order.tableId === i);
            const isOccupied = tableOrders.length > 0;
            
            // Проверяем, забронирован ли стол
            const isReserved = activeReservations.some(reservation => reservation.tableId === i);
            
            // Определяем класс для стола на основе его статуса
            let tableClass = 'table-card--free';
            let statusText = 'Свободен';
            
            if (isOccupied) {
                tableClass = 'table-card--occupied';
                statusText = 'Занят';
            } else if (isReserved) {
                tableClass = 'table-card--reserved';
                statusText = 'Забронирован';
            }
            
            tablesHTML += `
                <div class="table-card ${tableClass}" 
                     onclick="handleTableClick(${i})">
                    <div class="table-card__header">
                        <div class="table-card__number">Стол ${i}</div>
                        <div class="table-card__status ${tableClass.replace('table-card--', 'status--')}">
                            ${statusText}
                        </div>
                    </div>
                    ${isOccupied ? `
                        <div class="table-card__orders">
                            ${tableOrders.map(order => `
                                <div class="table-order">
                                    <div class="table-order__header">
                                        <div class="table-order__id">#${order.id}</div>
                                        <div class="table-order__time">${formatOrderTime(order.createdAt)}</div>
                                    </div>
                                    <div class="table-order__items">${formatOrderItems(order.items)}</div>
                                    ${order.comment ? `
                                        <div class="table-order__comment">
                                            <div class="comment-icon">💬</div>
                                            <div class="comment-text">${order.comment}</div>
                                        </div>
                                    ` : ''}
                                    <div class="table-order__footer">
                                        <div class="table-order__total">${formatMoney(order.total)} KZT</div>
                                        <div class="table-order__status-badge ${getStatusBadgeClass(order.status)}">
                                            ${getStatusText(order.status)}
                                        </div>
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    ` : ''}
                </div>
            `;
        }
        
        tablesGrid.innerHTML = tablesHTML;
        
        // Обновляем количество активных столов
        const activeTables = new Set(activeOrders.map(order => order.tableId)).size;
        document.getElementById('activeTables').textContent = activeTables;
        
    } catch (error) {
        console.error('Error loading tables:', error);
    }
}

// Обработчик клика по столу
function handleTableClick(tableId) {
    window.location.href = `create-order.html?table=${tableId}`;
}

// Вспомогательные функции
function getStatusBadgeClass(status) {
    const statusClasses = {
        'new': 'status-badge--new',
        'accepted': 'status-badge--accepted',
        'preparing': 'status-badge--preparing',
        'ready': 'status-badge--ready',
        'served': 'status-badge--served'
    };
    return statusClasses[status] || '';
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

function formatOrderItems(items) {
    if (!items || !items.length) return '';
    return items.map(item => `${item.name} x${item.quantity}`).join(', ');
}

function formatOrderTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

// Обновляем столы каждые 30 секунд
setInterval(loadTables, 30000);

// Загружаем столы при загрузке страницы
document.addEventListener('DOMContentLoaded', loadTables);

function updateTablesStatus() {
    const statusInfo = document.getElementById('tablesStatusInfo');
    const freeTables = document.querySelectorAll('.table-card--free').length;
    const reservedTables = document.querySelectorAll('.table-card--reserved').length;
    const occupiedTables = document.querySelectorAll('.table-card--occupied').length;
    const totalTables = document.querySelectorAll('.table-card').length;
    statusInfo.textContent = `Свободно ${freeTables} из ${totalTables} столов (${occupiedTables} занято, ${reservedTables} забронировано)`;
}

function logout() {
    localStorage.removeItem('currentUser');
    window.location.href = 'login.html';
}