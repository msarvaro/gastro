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
    // Check if business is selected before proceeding
    if (!window.api.checkBusinessSelected()) {
        return; // Will redirect to business selection page
    }
    
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
            '/waiter/profile': 'profile'
        };
        
        const activeSection = sections[currentPath] || 'tables';
        showSection(activeSection);
    
    // Инициализируем счетчики фильтров при загрузке страницы
    updateTableFilterBadge();
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

// Глобальная переменная для хранения текущих фильтров столов
window.tableFilters = {
    statuses: [] // Массив активных фильтров: ['free', 'reserved', 'occupied'] или пустой массив (все)
};

// Пример функций для загрузки данных (реализуйте по аналогии с вашими API)
async function loadTables() {
    try {
        const data = await window.api.call('/api/waiter/tables');
        if (!data) return; // Request failed or redirect happened
        
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

        // Сохраняем данные столов в глобальной переменной для использования в фильтрации
        window.allTables = data.tables || [];
        
        renderTablesWithFilter();
        
        // Добавляем обработчик для кнопки фильтра
        const filterButton = document.querySelector('.filter-button');
        if (filterButton) {
            filterButton.addEventListener('click', showTableFiltersModal);
        }
        
        // Обновляем счетчик фильтров при загрузке таблиц
        updateTableFilterBadge();
    } catch (e) {
        console.error('Failed to load tables:', e);
        const tablesStatusInfo = document.getElementById('tablesStatusInfo');
        if (tablesStatusInfo) tablesStatusInfo.textContent = 'Ошибка загрузки столов';
    }
}

// Функция для отображения столов с учетом текущего фильтра
function renderTablesWithFilter() {
    if (!window.allTables) return;

        const grid = document.getElementById('tablesGrid');
    if (!grid) return;
    
    // Применяем фильтры, только если есть активные фильтры
    const filterStatuses = window.tableFilters.statuses;
    const filteredTables = filterStatuses.length > 0 ? 
        window.allTables.filter(table => filterStatuses.includes(table.status)) : 
        window.allTables;
    
    if (filteredTables.length === 0) {
        grid.innerHTML = '<p class="no-tables-message">Столы не найдены</p>';
        return;
    }
    
    grid.innerHTML = filteredTables.map(table => `
                <div class="table-card table-card--${table.status.toLowerCase()}" data-table-id="${table.id}" data-table-status="${table.status.toLowerCase()}"> 
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
                        `).join('') : ''}
                    </div>
                </div>
            `).join('');
            
    // Добавляем обработчик клика для карточек столов
            grid.querySelectorAll('.table-card').forEach(tableCard => {
                tableCard.addEventListener('click', function() {
                    const tableId = this.dataset.tableId;
                    const currentStatus = this.dataset.tableStatus;
                    showTableStatusModal(tableId, currentStatus);
                });
            });
        }

// Функция для обновления бейджа фильтров
function updateTableFilterBadge() {
    const filterBadge = document.querySelector('.filter-button .filter-button__badge');
    if (!filterBadge) return;
    
    // Количество активных фильтров
    const activeFilterCount = window.tableFilters.statuses.length;
    
    // Обновляем текст и видимость бейджа
    filterBadge.textContent = activeFilterCount;
    filterBadge.style.display = activeFilterCount > 0 ? 'flex' : 'none';
}

async function loadOrders() {
    try {
        const data = await window.api.call('/api/waiter/orders');
        if (!data) return; // Request failed or redirect happened
        
        document.getElementById('ordersStatusInfo').textContent = `${data.stats.total_active_orders || 0} активных заказов`;
        const list = document.getElementById('ordersList');
        if (!data.orders || data.orders.length === 0) {
            list.innerHTML = '<div class="no-orders">Нет активных заказов</div>';
            return;
        }
        
        // Сохраняем все заказы для фильтрации
        window.allOrders = data.orders || [];
        
        // Рендерим заказы с учетом фильтров
        renderOrdersWithFilter();
        
        // Добавляем обработчики для фильтров
        setupOrderFilters();
    } catch (e) {
        console.error('Failed to load orders:', e);
        document.getElementById('ordersList').innerHTML = '<div class="error-message">Ошибка загрузки заказов</div>';
    }
}

// Функция для настройки фильтров заказов
function setupOrderFilters() {
    const timeFilterBtn = document.querySelector('#section-orders .filter-button--time');
    const statusFilterBtn = document.querySelector('#section-orders .filter-button--filter');
    
    if (timeFilterBtn) {
        timeFilterBtn.addEventListener('click', showOrderTimeFilterModal);
    }
    
    if (statusFilterBtn) {
        statusFilterBtn.addEventListener('click', showOrderStatusFilterModal);
    }
}

// Глобальные переменные для хранения текущих фильтров заказов
window.orderFilters = {
    sortBy: 'newest', // newest, oldest
    statuses: [] // Активные статусы фильтра ['new', 'accepted', ...]
};

// Функция для рендеринга заказов с учетом фильтров
function renderOrdersWithFilter() {
    if (!window.allOrders) return;
    
    const list = document.getElementById('ordersList');
    if (!list) return;
    
    // Применяем фильтры
    let filteredOrders = [...window.allOrders];
    
    // Фильтруем по статусам, если есть активные фильтры
    if (window.orderFilters.statuses.length > 0) {
        filteredOrders = filteredOrders.filter(order => 
            window.orderFilters.statuses.includes(order.status)
        );
    }
    
    // Сортируем по времени
    filteredOrders.sort((a, b) => {
        const dateA = new Date(a.created_at);
        const dateB = new Date(b.created_at);
        
        if (window.orderFilters.sortBy === 'newest') {
            return dateB - dateA; // От новых к старым
        } else {
            return dateA - dateB; // От старых к новым
        }
    });
    
    if (filteredOrders.length === 0) {
        list.innerHTML = '<div class="no-orders">Нет заказов, соответствующих фильтрам</div>';
        return;
    }
    
    list.innerHTML = filteredOrders.map(order => `
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
    
    // Обновление счетчиков активных фильтров для заказов
    updateOrderFilterBadges();
}

// Обновление счетчиков активных фильтров для заказов
function updateOrderFilterBadges() {
    const statusFilterBadge = document.querySelector('#section-orders .filter-button--filter .filter-button__badge');
    if (statusFilterBadge) {
        const activeFilters = window.orderFilters.statuses.length;
        statusFilterBadge.textContent = activeFilters;
        statusFilterBadge.style.display = activeFilters > 0 ? 'flex' : 'none';
    }
    
    // Отдельно обновляем бейдж для времени (сортировки)
    const timeFilterBadge = document.querySelector('#section-orders .filter-button--time .filter-button__badge');
    if (timeFilterBadge) {
        // Сортировка "Сначала новые" считается дефолтной и не влияет на счетчик
        const hasTimeFilter = window.orderFilters.sortBy !== 'newest';
        timeFilterBadge.textContent = hasTimeFilter ? '1' : '0';
        timeFilterBadge.style.display = hasTimeFilter ? 'flex' : 'none';
    }
}

// Модальное окно для фильтра по времени
function showOrderTimeFilterModal() {
    let modal = document.getElementById('orderTimeFilterModal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'orderTimeFilterModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal__content">
                <div class="modal__header">
                    <h2>Сортировка по времени</h2>
                    <button class="close-modal-btn">&times;</button>
                </div>
                <div class="modal__body">
                    <div class="filter-options">
                        <button class="filter-option filter-option--newest active">Сначала новые</button>
                        <button class="filter-option filter-option--oldest">Сначала старые</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Закрытие модального окна
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            modal.style.display = 'none';
        });
        
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
    }
    
    // Установка активного фильтра
    const filterOptions = modal.querySelectorAll('.filter-option');
    filterOptions.forEach(option => {
        // Обновляем активный фильтр
        if ((option.classList.contains('filter-option--newest') && window.orderFilters.sortBy === 'newest') ||
            (option.classList.contains('filter-option--oldest') && window.orderFilters.sortBy === 'oldest')) {
            option.classList.add('active');
        } else {
            option.classList.remove('active');
        }
        
        // Удаляем старые обработчики
        const newOption = option.cloneNode(true);
        option.parentNode.replaceChild(newOption, option);
        
        // Добавляем новые обработчики
        newOption.addEventListener('click', function() {
            filterOptions.forEach(opt => opt.classList.remove('active'));
            this.classList.add('active');
            
            if (this.classList.contains('filter-option--newest')) {
                window.orderFilters.sortBy = 'newest';
            } else {
                window.orderFilters.sortBy = 'oldest';
            }
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                this.blur();
                document.activeElement.blur();
            }, 100);
            
            renderOrdersWithFilter();
            modal.style.display = 'none';
        });
    });
    
    modal.style.display = 'block';
}

// Модальное окно для фильтра по статусу
function showOrderStatusFilterModal() {
    let modal = document.getElementById('orderStatusFilterModal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'orderStatusFilterModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal__content">
                <div class="modal__header">
                    <h2>Фильтр по статусу</h2>
                    <button class="close-modal-btn">&times;</button>
                </div>
                <div class="modal__body">
                    <div class="filter-options filter-options--checkboxes">
                        <label class="filter-checkbox">
                            <input type="checkbox" value="new">
                            <span class="status-badge status-badge--new">Новый</span>
                        </label>
                        <label class="filter-checkbox">
                            <input type="checkbox" value="accepted">
                            <span class="status-badge status-badge--accepted">Принят</span>
                        </label>
                        <label class="filter-checkbox">
                            <input type="checkbox" value="preparing">
                            <span class="status-badge status-badge--preparing">Готовится</span>
                        </label>
                        <label class="filter-checkbox">
                            <input type="checkbox" value="ready">
                            <span class="status-badge status-badge--ready">Готов</span>
                        </label>
                        <label class="filter-checkbox">
                            <input type="checkbox" value="served">
                            <span class="status-badge status-badge--served">Подан</span>
                        </label>
                    </div>
                    <div class="filter-actions">
                        <button class="clear-filters-btn">Сбросить</button>
                        <button class="apply-filters-btn">Применить</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Закрытие модального окна
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            modal.style.display = 'none';
        });
        
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
        
        // Кнопки действий
        modal.querySelector('.clear-filters-btn').addEventListener('click', () => {
            modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
                checkbox.checked = false;
            });
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                document.activeElement.blur();
            }, 100);
        });
        
        modal.querySelector('.apply-filters-btn').addEventListener('click', () => {
            const selectedStatuses = [];
            modal.querySelectorAll('input[type="checkbox"]:checked').forEach(checkbox => {
                selectedStatuses.push(checkbox.value);
            });
            
            window.orderFilters.statuses = selectedStatuses;
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                document.activeElement.blur();
            }, 100);
            
            renderOrdersWithFilter();
            modal.style.display = 'none';
        });
    }
    
    // Установка текущих фильтров
    modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
        checkbox.checked = window.orderFilters.statuses.includes(checkbox.value);
    });
    
    modal.style.display = 'block';
}

async function loadHistory() {
    try {
        const data = await window.api.call('/api/waiter/history');
        if (!data) return; // Request failed or redirect happened

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
        
        // Сохраняем заказы для фильтрации
        window.historyOrders = data.orders || [];
        
        // Рендерим историю с учетом фильтров
        renderHistoryWithFilter();
        
        // Настраиваем фильтры для истории
        setupHistoryFilters();
    } catch (e) {
        console.error('Failed to load history:', e);
        document.getElementById('historyList').innerHTML = '<div class="error-message">Ошибка загрузки истории</div>';
        const historyMainStatEl = document.getElementById('historyMainStat');
        const historySubStatEl = document.getElementById('historySubStat');
        if (historyMainStatEl) historyMainStatEl.textContent = 'История заказов';
        if (historySubStatEl) historySubStatEl.textContent = 'Ошибка загрузки статистики';
    }
}

// Функция для настройки фильтров истории заказов
function setupHistoryFilters() {
    const timeFilterBtn = document.querySelector('#section-history .filter-button--time');
    const statusFilterBtn = document.querySelector('#section-history .filter-button--filter');
    
    if (timeFilterBtn) {
        timeFilterBtn.addEventListener('click', showHistoryTimeFilterModal);
    }
    
    if (statusFilterBtn) {
        statusFilterBtn.addEventListener('click', showHistoryStatusFilterModal);
    }
}

// Глобальные переменные для хранения фильтров истории
window.historyFilters = {
    sortBy: 'newest', // newest, oldest
    statuses: [], // Статусы для фильтрации ['completed', 'cancelled']
    dateRange: null // Объект с date_from и date_to или null
};

// Функция для рендеринга истории с учетом фильтров
function renderHistoryWithFilter() {
    if (!window.historyOrders) return;
    
    const historyList = document.getElementById('historyList');
    if (!historyList) return;
    
    // Применяем фильтры
    let filteredHistory = [...window.historyOrders];
    
    // Фильтруем по статусам
    if (window.historyFilters.statuses.length > 0) {
        filteredHistory = filteredHistory.filter(order => 
            window.historyFilters.statuses.includes(order.status)
        );
    }
    
    // Фильтруем по датам, если задан диапазон
    if (window.historyFilters.dateRange) {
        const dateFrom = new Date(window.historyFilters.dateRange.date_from);
        const dateTo = new Date(window.historyFilters.dateRange.date_to);
        
        filteredHistory = filteredHistory.filter(order => {
            const orderDate = new Date(order.completed_at || order.cancelled_at);
            return orderDate >= dateFrom && orderDate <= dateTo;
        });
    }
    
    // Сортируем по времени
    filteredHistory.sort((a, b) => {
        const dateA = new Date(a.completed_at || a.cancelled_at);
        const dateB = new Date(b.completed_at || b.cancelled_at);
        
        if (window.historyFilters.sortBy === 'newest') {
            return dateB - dateA; // От новых к старым
        } else {
            return dateA - dateB; // От старых к новым
        }
    });
    
    if (filteredHistory.length === 0) {
        historyList.innerHTML = '<div class="no-orders">Нет заказов, соответствующих фильтрам</div>';
        return;
    }
    
    historyList.innerHTML = filteredHistory.map(order => `
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
    
    // Обновление счетчиков активных фильтров для истории
    updateHistoryFilterBadges();
}

// Обновление счетчиков активных фильтров для истории
function updateHistoryFilterBadges() {
    const statusFilterBadge = document.querySelector('#section-history .filter-button--filter .filter-button__badge');
    if (statusFilterBadge) {
        // Считаем активные фильтры: статусы, если выбраны + диапазон дат, если указан
        let activeFilters = window.historyFilters.statuses.length;
        
        // Если есть фильтр по дате, добавляем его как еще один активный фильтр
        if (window.historyFilters.dateRange) {
            activeFilters += 1;
        }
        
        statusFilterBadge.textContent = activeFilters;
        statusFilterBadge.style.display = activeFilters > 0 ? 'flex' : 'none';
    }
    
    // Отдельно обновляем бейдж для времени (сортировки)
    const timeFilterBadge = document.querySelector('#section-history .filter-button--time .filter-button__badge');
    if (timeFilterBadge) {
        // Сортировка "Сначала новые" считается дефолтной и не влияет на счетчик
        const hasTimeFilter = window.historyFilters.sortBy !== 'newest';
        timeFilterBadge.textContent = hasTimeFilter ? '1' : '0';
        timeFilterBadge.style.display = hasTimeFilter ? 'flex' : 'none';
    }
}

// Модальное окно для сортировки истории по времени
function showHistoryTimeFilterModal() {
    let modal = document.getElementById('historyTimeFilterModal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'historyTimeFilterModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal__content">
                <div class="modal__header">
                    <h2>Сортировка по времени</h2>
                    <button class="close-modal-btn">&times;</button>
                </div>
                <div class="modal__body">
                    <div class="filter-options">
                        <button class="filter-option filter-option--newest active">Сначала новые</button>
                        <button class="filter-option filter-option--oldest">Сначала старые</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Закрытие модального окна
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            modal.style.display = 'none';
        });
        
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
    }
    
    // Установка активного фильтра
    const filterOptions = modal.querySelectorAll('.filter-option');
    filterOptions.forEach(option => {
        // Обновляем активный фильтр
        if ((option.classList.contains('filter-option--newest') && window.historyFilters.sortBy === 'newest') ||
            (option.classList.contains('filter-option--oldest') && window.historyFilters.sortBy === 'oldest')) {
            option.classList.add('active');
        } else {
            option.classList.remove('active');
        }
        
        // Удаляем старые обработчики
        const newOption = option.cloneNode(true);
        option.parentNode.replaceChild(newOption, option);
        
        // Добавляем новые обработчики
        newOption.addEventListener('click', function() {
            filterOptions.forEach(opt => opt.classList.remove('active'));
            this.classList.add('active');
            
            if (this.classList.contains('filter-option--newest')) {
                window.historyFilters.sortBy = 'newest';
            } else {
                window.historyFilters.sortBy = 'oldest';
            }
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                this.blur();
                document.activeElement.blur();
            }, 100);
            
            renderHistoryWithFilter();
            modal.style.display = 'none';
        });
    });
    
    modal.style.display = 'block';
}

// Модальное окно для фильтра истории по статусу и дате
function showHistoryStatusFilterModal() {
    let modal = document.getElementById('historyStatusFilterModal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'historyStatusFilterModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal__content">
                <div class="modal__header">
                    <h2>Фильтр истории</h2>
                    <button class="close-modal-btn">&times;</button>
                </div>
                <div class="modal__body">
                    <h3>Статус заказа</h3>
                    <div class="filter-options filter-options--checkboxes">
                        <label class="filter-checkbox">
                            <input type="checkbox" value="completed">
                            <span class="status-badge status-badge--paid">Оплачен</span>
                        </label>
                        <label class="filter-checkbox">
                            <input type="checkbox" value="cancelled">
                            <span class="status-badge status-badge--cancelled">Отменён</span>
                        </label>
                    </div>
                    
                    <h3 style="margin-top: 16px;">Период</h3>
                    <div class="date-range-picker">
                        <div class="date-input">
                            <label>С</label>
                            <input type="date" id="date-from">
                        </div>
                        <div class="date-input">
                            <label>По</label>
                            <input type="date" id="date-to">
                        </div>
                    </div>
                    
                    <div class="filter-actions">
                        <button class="clear-filters-btn">Сбросить</button>
                        <button class="apply-filters-btn">Применить</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Закрытие модального окна
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            modal.style.display = 'none';
        });
        
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
        
        // Кнопка сброса фильтров
        modal.querySelector('.clear-filters-btn').addEventListener('click', () => {
            // Сбрасываем чекбоксы
            modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
                checkbox.checked = false;
            });
            
            // Сбрасываем даты
            modal.querySelector('#date-from').value = '';
            modal.querySelector('#date-to').value = '';
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                document.activeElement.blur();
            }, 100);
        });
        
        // Кнопка применения фильтров
        modal.querySelector('.apply-filters-btn').addEventListener('click', () => {
            // Собираем статусы
            const selectedStatuses = [];
            modal.querySelectorAll('input[type="checkbox"]:checked').forEach(checkbox => {
                selectedStatuses.push(checkbox.value);
            });
            
            // Собираем диапазон дат
            const dateFrom = modal.querySelector('#date-from').value;
            const dateTo = modal.querySelector('#date-to').value;
            
            // Обновляем фильтры
            window.historyFilters.statuses = selectedStatuses;
            
            if (dateFrom && dateTo) {
                window.historyFilters.dateRange = {
                    date_from: dateFrom,
                    date_to: dateTo
                };
            } else {
                window.historyFilters.dateRange = null;
            }
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                document.activeElement.blur();
            }, 100);
            
            renderHistoryWithFilter();
            modal.style.display = 'none';
        });
    }
    
    // Устанавливаем текущие значения фильтров
    modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
        checkbox.checked = window.historyFilters.statuses.includes(checkbox.value);
    });
    
    // Устанавливаем значения дат
    if (window.historyFilters.dateRange) {
        modal.querySelector('#date-from').value = window.historyFilters.dateRange.date_from;
        modal.querySelector('#date-to').value = window.historyFilters.dateRange.date_to;
    } else {
        modal.querySelector('#date-from').value = '';
        modal.querySelector('#date-to').value = '';
    }
    
    modal.style.display = 'block';
}

const formatTime = (dateOrTimeString) => {
    console.log(dateOrTimeString);
    if (!dateOrTimeString) return '';
    
    // Если это строка даты-времени, извлекаем только время
    if (typeof dateOrTimeString === 'string' && dateOrTimeString.includes('T')) {
        const timePart = dateOrTimeString.split('T')[1] || '00:00';
        return timePart.substring(0, 5); // HH:MM
    }
    
    // Если это строка времени в формате HH:MM:SS, возвращаем только HH:MM
    if (typeof dateOrTimeString === 'string' && dateOrTimeString.includes(':')) {
        return dateOrTimeString.substring(0, 5); // HH:MM
    }
    
    // Если это объект Date, извлекаем время напрямую без создания нового объекта Date
    if (dateOrTimeString instanceof Date) {
        const hours = dateOrTimeString.getHours().toString().padStart(2, '0');
        const minutes = dateOrTimeString.getMinutes().toString().padStart(2, '0');
        return `${hours}:${minutes}`;
    }
    
    return dateOrTimeString; // Если ничего не подходит, вернем как есть
};

const formatDate = (date) => {
    return date.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
};

async function loadProfile() {
    try {
        const profileData = await window.api.call('/api/waiter/profile');
        if (!profileData) return; // Request failed or redirect happened
        
        console.log("Профиль загружен:", profileData);
        console.log("Имя:", profileData.name, "Тип:", typeof profileData.name);
        console.log("Логин:", profileData.username, "Тип:", typeof profileData.username);
        
        // Обновляем шапку профиля
        const profileHeaderEl = document.getElementById('profileHeaderName');
        if (profileHeaderEl) {
            profileHeaderEl.textContent = profileData.name || profileData.username;
        }                // Обновляем шапку профиля с информацией о смене, если она есть        
        const profileStatusEl = document.querySelector('.profile-status');        
        if (profileStatusEl) {            
            if (profileData.current_shift) {                
                const shift = profileData.current_shift;                
                const shiftId = shift.id; // Используем ID смены или значение по умолчанию                
                const shiftDate = formatShiftDate(shift.date) || new Date().toLocaleDateString('ru-RU');             
                profileStatusEl.innerHTML = `                    
                <div>Смена #${shiftId}</div>                    
                <div>${shiftDate}</div>
                `;            
            } else {                
                // Если нет активной смены, устанавливаем дефолтные значения в шапке                
                profileStatusEl.innerHTML = `                   
                <div>Нет активной смены</div>                    
                <div>Проверьте расписание</div>                
                `;            
            }        
        }
        
        // Получаем контейнер для профиля
        const profileContainer = document.getElementById('section-profile');
        if (!profileContainer) return;
        
        // Обновляем основную информацию профиля (имя и роль)
        const userDetailsHtml = `
            <div class="profile-user-details">
                <div class="profile-user-details__avatar"></div>
                <div class="profile-user-details__info">
                    <div>${profileData.name || profileData.username}</div>
                    <div>Официант • ${profileData.email || ''}</div>
                </div>
            </div>
        `;
        
        // Информация о текущей смене
        let shiftInfoHtml = `
            <div class="profile-info-card">
                <div class="profile-info-card__header">
                    <span style="font-size: 18px; margin-right: 8px;">🕒</span>
                    Информация о смене
                </div>
        `;
        
        if (profileData.current_shift) {
            const shift = profileData.current_shift;
            console.log('Данные текущей смены:', shift);
            console.log('Дата смены:', shift.date);
            console.log('Время начала:', shift.start_time);
            console.log('Время окончания:', shift.end_time);
            
            // Получаем текущую дату
            const now = new Date();
            const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
            
            // Извлекаем только время из строк времени
            let startHour = 0, startMinute = 0, endHour = 0, endMinute = 0;
            
            // Парсим время из строк
            if (typeof shift.start_time === 'string') {
                const timeMatch = shift.start_time.match(/(\d{1,2}):(\d{1,2})/);
                if (timeMatch) {
                    startHour = parseInt(timeMatch[1], 10);
                    startMinute = parseInt(timeMatch[2], 10);
                } else if (shift.start_time.includes('T')) {
                    const timePart = shift.start_time.split('T')[1] || '';
                    const timeComponents = timePart.split(':');
                    if (timeComponents.length >= 2) {
                        startHour = parseInt(timeComponents[0], 10);
                        startMinute = parseInt(timeComponents[1], 10);
                    }
                }
            }
            
            if (typeof shift.end_time === 'string') {
                const timeMatch = shift.end_time.match(/(\d{1,2}):(\d{1,2})/);
                if (timeMatch) {
                    endHour = parseInt(timeMatch[1], 10);
                    endMinute = parseInt(timeMatch[2], 10);
                } else if (shift.end_time.includes('T')) {
                    const timePart = shift.end_time.split('T')[1] || '';
                    const timeComponents = timePart.split(':');
                    if (timeComponents.length >= 2) {
                        endHour = parseInt(timeComponents[0], 10);
                        endMinute = parseInt(timeComponents[1], 10);
                    }
                }
            }
            
            // Создаем объекты Date с текущим днем и временем из смены
            const startDate = new Date(today);
            startDate.setHours(startHour, startMinute, 0, 0);
            
            const endDate = new Date(today);
            endDate.setHours(endHour, endMinute, 0, 0);
            
            // Если конец смены раньше начала (смена заканчивается на следующий день)
            if (endDate < startDate) {
                endDate.setDate(endDate.getDate() + 1);
            }
            
            console.log('Исправленное время начала:', startDate);
            console.log('Исправленное время окончания:', endDate);
            
            let timeLeftText = '';
            
            // Получаем удобные для отображения времена
            const startTime = formatTime(shift.start_time);
            const endTime = formatTime(shift.end_time);

            // Форматируем дату смены с использованием специальной функции
            const formattedShiftDate = formatShiftDate(shift.date) || new Date().toLocaleDateString('ru-RU');

            // Вычисляем оставшееся время
            if (now < startDate) {
                // Смена еще не началась
                const diffMs = startDate - now;
                const diffHrs = Math.floor(diffMs / 3600000); // часы
                const diffMins = Math.round((diffMs % 3600000) / 60000); // оставшиеся минуты
                
                if (diffHrs > 0) {
                    timeLeftText = `${diffHrs} ч ${diffMins} мин до начала смены`;
                } else {
                    timeLeftText = `${diffMins} мин до начала смены`;
                }
            } else if (now < endDate) {
                // Смена идет в данный момент
                const diffMs = endDate - now;
                const diffHrs = Math.floor(diffMs / 3600000); // часы
                const diffMins = Math.round((diffMs % 3600000) / 60000); // оставшиеся минуты
                
                if (diffHrs > 0) {
                    timeLeftText = `${diffHrs} ч ${diffMins} мин до конца смены`;
                } else {
                    timeLeftText = `${diffMins} мин до конца смены`;
                }
            } else {
                timeLeftText = 'Смена завершена';
            }
            
            shiftInfoHtml += `
                <div class="profile-info-card__content">
                    <p><b>Текущая смена:</b> ${formattedShiftDate}</p>
                    <p><b>Время:</b> ${startTime} - ${endTime}</p>
                    <p><b>Статус:</b> <span class="status-text status-text--${now < startDate ? 'new' : (now < endDate ? 'ready' : 'completed')}">${now < startDate ? 'Запланирована' : (now < endDate ? 'Активна' : 'Завершена')}</span></p>
                    <p><b>${now < startDate ? 'До начала:' : 'До конца:'}</b> ${timeLeftText}</p>
                    ${profileData.current_shift_manager ? `<p><b>Менеджер:</b> ${profileData.current_shift_manager}</p>` : ''}
                </div>
            `;
        } else {
            shiftInfoHtml += `
                <div class="profile-info-card__content">
                    <p>В данный момент нет активной смены.</p>
                </div>
            `;
        }
        
        // Добавляем будущие смены если есть
        if (profileData.upcoming_shifts && profileData.upcoming_shifts.length > 0) {
            shiftInfoHtml += `<div class="profile-info-card__header" style="margin-top: 16px;">Предстоящие смены</div>`;
            shiftInfoHtml += `<div class="profile-info-card__content profile-info-card__content--flex">`;
            
            profileData.upcoming_shifts.forEach(shift => {
                // Форматируем дату и время с использованием специальных функций
                const formattedDate = formatShiftDate(shift.date);
                const startTime = formatTime(shift.start_time);
                const endTime = formatTime(shift.end_time);
                
                shiftInfoHtml += `
                    <div class="profile-info-card__item">
                        <span>${formattedDate}</span>
                        ${startTime} - ${endTime}
                    </div>
                `;
            });
            
            shiftInfoHtml += `</div>`;
        }
        
        shiftInfoHtml += `</div>`;
        
        // Назначенные столы
        let assignedTablesHtml = `
            <div class="profile-info-card">
                <div class="profile-info-card__header">
                    <span style="font-size: 18px; margin-right: 8px;">🍽️</span>
                    Назначенные столы
                </div>
                <div class="profile-info-card__content profile-info-card__content--flex">
        `;
        
        if (profileData.assigned_tables && profileData.assigned_tables.length > 0) {
            profileData.assigned_tables.forEach(table => {
                assignedTablesHtml += `
                    <div class="profile-table">
                        Стол №${table.number}
                    <span>${table.seats} мест • ${                            
                        table.status === 'free' ? 'Свободен' :                             
                        (table.status === 'reserved' ? 'Забронирован' : 'Занят')}</span>                    
                    </div>
                `;
            });
        } else {
            assignedTablesHtml += `<p>Нет назначенных столов</p>`;
        }
        
        assignedTablesHtml += `</div></div>`;
        
        // Статистика по заказам
        let orderStatsHtml = `
            <div class="profile-info-card">
                <div class="profile-info-card__header">
                    <span style="font-size: 18px; margin-right: 8px;">📋</span>
                    Активные заказы
                </div>
                <div class="profile-info-card__content profile-info-card__content--grid">
                    <div class="profile-info-card__item">
                        <span>${profileData.order_stats.new}</span>
                        Новые
                    </div>
                    <div class="profile-info-card__item">
                        <span>${profileData.order_stats.accepted}</span>
                        Принятые
                    </div>
                    <div class="profile-info-card__item">
                        <span>${profileData.order_stats.preparing}</span>
                        Готовятся
                    </div>
                    <div class="profile-info-card__item">
                        <span>${profileData.order_stats.ready}</span>
                        Готовы
                    </div>
                    <div class="profile-info-card__item">
                        <span>${profileData.order_stats.served}</span>
                        Поданы
                    </div>
                    <div class="profile-info-card__item">
                        <span>${profileData.order_stats.total}</span>
                        Всего
                    </div>
                </div>
            </div>
        `;
        
        // Показатели эффективности
        let performanceHtml = `
            <div class="profile-info-card">
                <div class="profile-info-card__header">
                    <span style="font-size: 18px; margin-right: 8px;">📊</span>
                    Эффективность
                </div>
                <div class="profile-info-card__content profile-info-card__content--grid">
                    <div class="profile-info-card__item">
                        <span>${profileData.performance_data.tables_served}</span>
                        Столов обслужено
                    </div>
                    <div class="profile-info-card__item">
                        <span>${profileData.performance_data.orders_completed}</span>
                        Заказов выполнено
                    </div>
                </div>
            </div>
        `;
        
        // Обновляем контент страницы
        const contentWrapper = profileContainer.querySelector('.content-wrapper') || profileContainer;
        contentWrapper.innerHTML = userDetailsHtml + shiftInfoHtml + assignedTablesHtml + orderStatsHtml + performanceHtml;
        
    } catch (error) {
        console.error('Ошибка при загрузке профиля:', error);
        
        // Отображаем сообщение об ошибке
        const profileContainer = document.getElementById('section-profile');
        if (profileContainer) {
            const contentWrapper = profileContainer.querySelector('.content-wrapper') || profileContainer;
            contentWrapper.innerHTML = `
                <div class="error-message" style="padding: 20px; text-align: center;">
                    <p>Не удалось загрузить данные профиля</p>
                    <button onclick="loadProfile()" style="margin-top: 10px; padding: 8px 16px;">Попробовать снова</button>
                </div>
            `;
        }
    }
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
    try {
        const data = await window.api.call('/api/waiter/tables');
        if (!data) return; // Request failed or redirect happened
        
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
    } catch (e) {
        console.error('Failed to render tables:', e);
        alert('Ошибка загрузки столов');
    }
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
        const createdOrder = await window.api.call('/api/waiter/orders', 'POST', payload);
        if (!createdOrder) return; // Request failed or redirect happened
        
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
        await window.api.call(`/api/waiter/orders/${orderId}/status`, 'PUT', { status: newStatus });
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

// Форматирует дату в российском формате дд.мм.гггг
function formatShiftDate(dateString) {
    if (!dateString) return '';
    
    try {
        // Проверяем формат YYYY-MM-DD
        if (typeof dateString === 'string' && dateString.match(/^\d{4}-\d{2}-\d{2}$/)) {
            const [year, month, day] = dateString.split('-');
            return `${day}.${month}.${year}`;
        }
        
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            return dateString;
        }
        
        return date.toLocaleDateString('ru-RU');
    } catch (error) {
        console.error('Ошибка при форматировании даты смены:', error);
        return dateString;
    }
} 

// Add this function to handle table status updates
async function updateTableStatus(tableId, newStatus) {
    try {
        const response = await window.api.call(`/api/waiter/tables/${tableId}/status`, 'PUT', { status: newStatus });
        return true; // Success
    } catch (error) {
        console.error('Error updating table status:', error);
        return false; // Failed
    }
}

// Удаляем весь встроенный CSS-код 
// Add table status modal
function showTableStatusModal(tableId, currentStatus) {
    // Create modal if it doesn't exist
    let modal = document.getElementById('tableStatusModal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'tableStatusModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal__content">
                <div class="modal__header">
                    <h2>Изменить статус стола</h2>
                    <button class="close-modal-btn">&times;</button>
                </div>
                <div class="modal__body">
                    <div class="status-options">
                        <button class="status-option status-option--free">Свободен</button>
                        <button class="status-option status-option--occupied">Занят</button>
                        <button class="status-option status-option--reserved">Забронирован</button>
                    </div>
                </div>
                <div class="modal__footer">
                    <p class="modal__message" style="display: none;"></p>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Add event listener to close button
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            modal.style.display = 'none';
        });
        
        // Close modal when clicking outside
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
    }
    
    // Store the tableId in the modal for reference
    modal.setAttribute('data-table-id', tableId);
    
    // Get all status buttons
    const statusButtons = modal.querySelectorAll('.status-option');
    
    // Remove active class from all buttons first
    statusButtons.forEach(button => {
        button.classList.remove('active');
        
        // Remove all event listeners by cloning and replacing
        const newButton = button.cloneNode(true);
        button.parentNode.replaceChild(newButton, button);
    });
    
    // Re-query buttons after replacing them
    const newStatusButtons = modal.querySelectorAll('.status-option');
    
    // Add active class to current status button
    newStatusButtons.forEach(button => {
        if ((button.classList.contains('status-option--free') && currentStatus === 'free') ||
            (button.classList.contains('status-option--occupied') && currentStatus === 'occupied') ||
            (button.classList.contains('status-option--reserved') && currentStatus === 'reserved')) {
            button.classList.add('active');
        }
        
        // Add click event handler
        button.addEventListener('click', async () => {
            let newStatus;
            if (button.classList.contains('status-option--free')) {
                newStatus = 'free';
            } else if (button.classList.contains('status-option--occupied')) {
                newStatus = 'occupied';
            } else if (button.classList.contains('status-option--reserved')) {
                newStatus = 'reserved';
            }
            
            // Immediately update UI
            newStatusButtons.forEach(btn => btn.classList.remove('active'));
            button.classList.add('active');
            
            if (newStatus) {
                const messageElement = modal.querySelector('.modal__message');
                const success = await updateTableStatus(tableId, newStatus);
                
                if (success) {
                    // Update was successful
                    messageElement.textContent = 'Статус стола успешно обновлен';
                    messageElement.style.color = 'green';
                    messageElement.style.display = 'block';
                    
                    // Refresh tables data
                    await loadTables();
                    
                    // Close modal after a delay
                    setTimeout(() => {
                        modal.style.display = 'none';
                        messageElement.style.display = 'none';
                    }, 1500);
                } else {
                    // Handle error
                    messageElement.textContent = 'Ошибка при обновлении статуса стола';
                    messageElement.style.color = 'red';
                    messageElement.style.display = 'block';
                }
            }
        });
    });
    
    // Show the modal
    modal.style.display = 'block';
}

// Function to generate table elements in the grid
function generateTableElements(tables) {
    const tablesGrid = document.getElementById('tablesGrid');
    if (!tablesGrid) return;
    
    tablesGrid.innerHTML = '';
    
    tables.forEach(table => {
        const tableElement = document.createElement('div');
        tableElement.className = `table-item table-item--${table.status}`;
        tableElement.setAttribute('data-table-id', table.id);
        
        tableElement.innerHTML = `
            <div class="table-item__header">
                <div class="table-item__number">№${table.number}</div>
                <div class="table-item__seats">${table.seats} мест</div>
            </div>
            <div class="table-item__status">${translateTableStatus(table.status)}</div>
        `;
        
        // Add orders if any
        if (table.orders && table.orders.length > 0) {
            const ordersElement = document.createElement('div');
            ordersElement.className = 'table-item__orders';
            
            table.orders.forEach(order => {
                const orderElement = document.createElement('div');
                orderElement.className = 'table-item__order';
                orderElement.innerHTML = `
                    <span class="order-id">Заказ #${order.id}</span>
                    <span class="order-time">${formatTableTime(new Date(order.time))}</span>
                `;
                ordersElement.appendChild(orderElement);
            });
            
            tableElement.appendChild(ordersElement);
        }
        
        // Add click event to update status
        tableElement.addEventListener('click', () => {
            showTableStatusModal(table.id, table.status);
        });
        
        tablesGrid.appendChild(tableElement);
    });
}

// Helper function to translate table status to Russian
function translateTableStatus(status) {
    const translations = {
        'free': 'Свободен',
        'occupied': 'Занят',
        'reserved': 'Забронирован'
    };
    return translations[status] || status;
}

// Helper function to format time
function formatTableTime(date) {
    return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

// Function to update tables statistics
function updateTablesStats(stats) {
    const tablesStatusInfo = document.getElementById('tablesStatusInfo');
    if (tablesStatusInfo) {
        tablesStatusInfo.textContent = `${stats.free} столов свободно из ${stats.total} (${stats.occupied} занято, ${stats.reserved} забронировано)`;
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    loadTables();
    
    // Add tab switching logic
    const tabItems = document.querySelectorAll('.tab-item');
    const sections = document.querySelectorAll('section');
    
    tabItems.forEach(tab => {
        tab.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetSection = this.getAttribute('data-section');
            
            // Update active tab
            tabItems.forEach(t => t.classList.remove('tab-item--active'));
            this.classList.add('tab-item--active');
            
            // Show target section, hide others
            sections.forEach(section => {
                if (section.id === `section-${targetSection}`) {
                    section.style.display = 'block';
                } else {
                    section.style.display = 'none';
                }
            });
            
            // Load data based on active tab
            if (targetSection === 'tables') {
                loadTables();
            } else if (targetSection === 'orders') {
                // loadOrders(); // Implement this function if needed
            } else if (targetSection === 'history') {
                // loadOrderHistory(); // Implement this function if needed
            } else if (targetSection === 'profile') {
                // loadProfile(); // Implement this function if needed
            }
        });
    });
}); 

// Функция для отображения модального окна с фильтрами столов
function showTableFiltersModal() {
    // Создаем модальное окно, если оно еще не существует
    let modal = document.getElementById('tableFiltersModal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'tableFiltersModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal__content">
                <div class="modal__header">
                    <h2>Фильтр столов</h2>
                    <button class="close-modal-btn">&times;</button>
                </div>
                <div class="modal__body">
                    <div class="filter-options filter-options--checkboxes">
                        <label class="filter-checkbox">
                            <input type="checkbox" value="free">
                            <span class="status-badge status-badge--free">Свободные</span>
                        </label>
                        <label class="filter-checkbox">
                            <input type="checkbox" value="reserved">
                            <span class="status-badge status-badge--reserved">Забронированные</span>
                        </label>
                        <label class="filter-checkbox">
                            <input type="checkbox" value="occupied">
                            <span class="status-badge status-badge--occupied">Занятые</span>
                        </label>
                    </div>
                    <div class="filter-actions">
                        <button class="clear-filters-btn">Сбросить</button>
                        <button class="apply-filters-btn">Применить</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Добавляем обработчик для закрытия модального окна
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            modal.style.display = 'none';
        });
        
        // Закрытие модального окна при клике вне его содержимого
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
        
        // Кнопка сброса фильтров
        modal.querySelector('.clear-filters-btn').addEventListener('click', () => {
            modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
                checkbox.checked = false;
            });
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                document.activeElement.blur();
            }, 100);
        });
        
        // Кнопка применения фильтров
        modal.querySelector('.apply-filters-btn').addEventListener('click', () => {
            // Собираем выбранные статусы
            const selectedStatuses = [];
            modal.querySelectorAll('input[type="checkbox"]:checked').forEach(checkbox => {
                selectedStatuses.push(checkbox.value);
            });
            
            // Обновляем фильтры
            window.tableFilters.statuses = selectedStatuses;
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                document.activeElement.blur();
            }, 100);
            
            // Обновляем отображение таблиц и счетчик фильтров
            renderTablesWithFilter();
            updateTableFilterBadge();
            
            modal.style.display = 'none';
        });
    }
    
    // Устанавливаем текущие значения фильтров в чекбоксы
    modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
        checkbox.checked = window.tableFilters.statuses.includes(checkbox.value);
    });
    
    // Отображаем модальное окно
    modal.style.display = 'block';
}