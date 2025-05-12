document.addEventListener('DOMContentLoaded', function() {
    // Проверка авторизации
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    
    if (!token || role !== 'waiter') {
        window.location.href = '/';
        return;
    }

    loadTables();
    updateTablesStatus();
});

// Функция загрузки и отображения столов
async function loadTables() {
    try {
        const tables = await waiterApi.getTables();
        const tableStatus = await waiterApi.getTableStatus();
        
        // Получаем контейнер для столов
        const tablesGrid = document.getElementById('tablesGrid');
        let tablesHTML = '';
        
        // Создаем HTML для каждого стола
        for (const table of tables) {
            const status = tableStatus[table.id] || 'free';
            const tableClass = `table-card--${status}`;
            const statusText = getStatusText(status);
            
            tablesHTML += `
                <div class="table-card ${tableClass}" 
                     onclick="handleTableClick(${table.id})">
                    <div class="table-card__header">
                        <div class="table-card__number">Стол ${table.number}</div>
                        <div class="table-card__status ${tableClass.replace('table-card--', 'status--')}">
                            ${statusText}
                        </div>
                    </div>
                    ${status === 'occupied' ? `
                        <div class="table-card__orders">
                            ${table.orders ? table.orders.map(order => `
                                <div class="table-order">
                                    <div class="table-order__header">
                                        <div class="table-order__id">#${order.id}</div>
                                        <div class="table-order__time">${waiterApi.formatOrderTime(order.createdAt)}</div>
                                    </div>
                                    <div class="table-order__items">${waiterApi.formatOrderItems(order.items)}</div>
                                    ${order.comment ? `
                                        <div class="table-order__comment">
                                            <div class="comment-icon">💬</div>
                                            <div class="comment-text">${order.comment}</div>
                                        </div>
                                    ` : ''}
                                    <div class="table-order__footer">
                                        <div class="table-order__total">${waiterApi.formatMoney(order.total)} KZT</div>
                                        <div class="table-order__status-badge ${getStatusBadgeClass(order.status)}">
                                            ${getStatusText(order.status)}
                                        </div>
                                    </div>
                                </div>
                            `).join('') : ''}
                        </div>
                    ` : ''}
                </div>
            `;
        }
        
        tablesGrid.innerHTML = tablesHTML;
        updateTablesStatus();
        
    } catch (error) {
        console.error('Error loading tables:', error);
        const tablesGrid = document.getElementById('tablesGrid');
        tablesGrid.innerHTML = '<div class="error-message">Ошибка загрузки столов. Пожалуйста, попробуйте позже.</div>';
    }
}

// Обработчик клика по столу
function handleTableClick(tableId) {
    window.location.href = `/waiter/create-order?table=${tableId}`;
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

async function updateTablesStatus() {
    try {
        const tables = await waiterApi.getTables();
        const tableStatus = await waiterApi.getTableStatus();
        
        const freeTables = tables.filter(table => tableStatus[table.id] === 'free').length;
        const reservedTables = tables.filter(table => tableStatus[table.id] === 'reserved').length;
        const occupiedTables = tables.filter(table => tableStatus[table.id] === 'occupied').length;
        const totalTables = tables.length;
        
        const statusInfo = document.getElementById('tablesStatusInfo');
        statusInfo.textContent = `Свободно ${freeTables} из ${totalTables} столов (${occupiedTables} занято, ${reservedTables} забронировано)`;
    } catch (error) {
        console.error('Error updating tables status:', error);
        const statusInfo = document.getElementById('tablesStatusInfo');
        statusInfo.textContent = 'Ошибка загрузки статуса столов';
    }
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    window.location.href = '/';
}