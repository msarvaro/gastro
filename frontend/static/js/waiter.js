// waiter.js: объединённая логика для панели официанта

// Add a default food image constant at the top of the file
const DEFAULT_FOOD_IMAGE = "https://cdn.pixabay.com/photo/2018/06/01/20/30/food-3447416_1280.jpg";

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
    getCategories: () => {
        const businessId = localStorage.getItem('business_id');
        return apiCall(`/categories?business_id=${businessId}`);
    },
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
            document.getElementById('tab-bar').style.display = 'none';
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
            document.getElementById('tab-bar').style.display = 'flex';
        });
    }
    document.getElementById('selectTableBtn').addEventListener('click', showTableModal);
    document.querySelector('.close-modal-btn').addEventListener('click', closeTableModal);
    document.querySelector('.create-order-btn').addEventListener('click', showConfirmOrderModal);
    document.getElementById('confirmOrderBtn').addEventListener('click', createOrder)
    document.getElementById('cancelOrderBtn').addEventListener('click', hideConfirmOrderModal);
    document.querySelector('.close-order-modal-btn').addEventListener('click', hideConfirmOrderModal);
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

    // Настраиваем фильтры заказов и истории, даже если соответствующие секции не активны сейчас
    console.log('Инициализирую фильтры заказов из DOMContentLoaded');
    setupOrderFilters();
    setupHistoryFilters();
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
        
        const tablesStatusInfoTitle = document.getElementById('tablesStatusInfoTitle');
        const tableStatusInfoSubtitle = document.getElementById('tableStatusInfoSubtitle');
        if (tablesStatusInfoTitle && tableStatusInfoSubtitle && data.stats) {
            // Updated to match new CSS structure for header
            let occupancyPercentage = 0;
            if (data.stats.total > 0) {
                occupancyPercentage = Math.round(((data.stats.total - data.stats.free) / data.stats.total) * 100);
            }
            tablesStatusInfoTitle.innerHTML = `${occupancyPercentage}% занято`;
            tableStatusInfoSubtitle.innerHTML = `Количество свободных столов: ${data.stats.free} из ${data.stats.total}`;
        }
        
        // Сохраняем данные столов в глобальной переменной для использования в фильтрации
        window.allTables = data.tables || [];
        
        // Отрисовываем столы с учетом фильтров
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
    
    grid.innerHTML = filteredTables.map(table => {
        const status = table.status || 'free'; // Default to 'free' if status is null/undefined
        const statusLower = status.toLowerCase();
        
        return `
                <div class="table-card table-card--${statusLower}" data-table-id="${table.id}" data-table-status="${statusLower}"> 
                    <div class="table-card__header">
                        <div class = table-card__number>
                            <span class="status-dot status-dot--${statusLower}"></span>
                            <span class="table-card__title">№${table.number}</span>
                        </div>
                        <span class="table-card__seats">${table.seats} мест</span>
                    </div>
                    <div class="table-card__content"> 
                        ${table.orders && table.orders.length ? `
                            <div class="table-card__orders">
                                ${table.orders.map(order => `
                                    <div class="table-order">
                                        <span class="table-order__id table-order__id--${order.status ? order.status.toLowerCase() : 'unknown'}">#${order.id}</span>
                                        <span class="table-order__time">${formatTableTime(new Date(order.time))}</span>
                                        
                                    </div>
                                `).join('')}
                            </div>
                        ` : ''}
                    </div>
                </div>
            `;
    }).join('');
            
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
        
        document.getElementById('ordersStatusInfoTitle').textContent = `${data.stats.total_active_orders} активных заказов`;
        document.getElementById('ordersStatusInfoSubtitle').textContent = `Новых: ${data.stats.new} | В работе: ${data.stats.accepted + data.stats.preparing} | Готовых: ${data.stats.ready + data.stats.served}`
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
    
    console.log('setupOrderFilters вызвана');
    console.log('timeFilterBtn:', timeFilterBtn);
    console.log('statusFilterBtn:', statusFilterBtn);
    
    if (timeFilterBtn) {
        console.log('Добавляем обработчик click для timeFilterBtn');
        
        // Проверим, есть ли уже обработчик клика
        const oldClone = timeFilterBtn.cloneNode(true);
        timeFilterBtn.parentNode.replaceChild(oldClone, timeFilterBtn);
        
        oldClone.addEventListener('click', function(e) {
            e.preventDefault();
            console.log('Клик на кнопке фильтра времени');
            showOrderTimeFilterModal();
        });
        
        // Добавляем визуальную индикацию того, что обработчик привязан
        oldClone.style.cursor = 'pointer';
    }
    
    if (statusFilterBtn) {
        console.log('Добавляем обработчик click для statusFilterBtn');
        
        // Проверим, есть ли уже обработчик клика
        const oldClone = statusFilterBtn.cloneNode(true);
        statusFilterBtn.parentNode.replaceChild(oldClone, statusFilterBtn);
        
        oldClone.addEventListener('click', function(e) {
            e.preventDefault();
            console.log('Клик на кнопке фильтра статусов');
            showOrderStatusFilterModal();
        });
        
        // Добавляем визуальную индикацию того, что обработчик привязан
        oldClone.style.cursor = 'pointer';
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
    console.log('showOrderTimeFilterModal вызвана');
    let modal = document.getElementById('orderTimeFilterModal');
    if (!modal) {
        console.log('Создаю модальное окно для фильтра по времени');
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
        console.log('Модальное окно добавлено в DOM');
        
        // Закрытие модального окна
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            console.log('Нажата кнопка закрытия модального окна по времени');
            modal.classList.remove('active');
        });
        
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                console.log('Клик вне модального окна по времени');
                modal.classList.remove('active');
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
            console.log('Выбран фильтр времени:', this.textContent);
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
            modal.classList.remove('active');
        });
    });
    
    console.log('Отображаю модальное окно по времени (добавляю класс active)');
    modal.classList.add('active');
}

// Модальное окно для фильтра по статусу
function showOrderStatusFilterModal() {
    console.log('showOrderStatusFilterModal вызвана');
    let modal = document.getElementById('orderStatusFilterModal');
    if (!modal) {
        console.log('Создаю модальное окно для фильтра по статусу');
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
        console.log('Модальное окно для статусов добавлено в DOM');
        
        // Закрытие модального окна
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            console.log('Нажата кнопка закрытия модального окна статусов');
            modal.classList.remove('active');
        });
        
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                console.log('Клик вне модального окна статусов');
                modal.classList.remove('active');
            }
        });
        
        // Кнопки действий
        modal.querySelector('.clear-filters-btn').addEventListener('click', () => {
            console.log('Нажата кнопка сброса фильтров');
            modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
                checkbox.checked = false;
            });
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                document.activeElement.blur();
            }, 100);
        });
        
        modal.querySelector('.apply-filters-btn').addEventListener('click', () => {
            console.log('Нажата кнопка применения фильтров');
            const selectedStatuses = [];
            modal.querySelectorAll('input[type="checkbox"]:checked').forEach(checkbox => {
                selectedStatuses.push(checkbox.value);
            });
            
            window.orderFilters.statuses = selectedStatuses;
            console.log('Выбранные статусы:', selectedStatuses);
            
            // Убираем фокус с кнопки
            setTimeout(() => {
                document.activeElement.blur();
            }, 100);
            
            renderOrdersWithFilter();
            modal.classList.remove('active');
        });
    }
    
    // Установка текущих фильтров
    modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
        checkbox.checked = window.orderFilters.statuses.includes(checkbox.value);
    });
    
    console.log('Отображаю модальное окно статусов (добавляю класс active)');
    modal.classList.add('active');
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
    console.log('setupHistoryFilters вызвана');
    const timeFilterBtn = document.querySelector('#section-history .filter-button--time');
    const statusFilterBtn = document.querySelector('#section-history .filter-button--filter');
    
    console.log('timeFilterBtn история:', timeFilterBtn);
    console.log('statusFilterBtn история:', statusFilterBtn);
    
    if (timeFilterBtn) {
        console.log('Добавляем обработчик click для timeFilterBtn истории');
        
        // Проверим, есть ли уже обработчик клика
        const oldClone = timeFilterBtn.cloneNode(true);
        timeFilterBtn.parentNode.replaceChild(oldClone, timeFilterBtn);
        
        oldClone.addEventListener('click', function(e) {
            e.preventDefault();
            console.log('Клик на кнопке фильтра времени истории');
            showHistoryTimeFilterModal();
        });
        
        // Добавляем визуальную индикацию того, что обработчик привязан
        oldClone.style.cursor = 'pointer';
    }
    
    if (statusFilterBtn) {
        console.log('Добавляем обработчик click для statusFilterBtn истории');
        
        // Проверим, есть ли уже обработчик клика
        const oldClone = statusFilterBtn.cloneNode(true);
        statusFilterBtn.parentNode.replaceChild(oldClone, statusFilterBtn);
        
        oldClone.addEventListener('click', function(e) {
            e.preventDefault();
            console.log('Клик на кнопке фильтра статусов истории');
            showHistoryStatusFilterModal();
        });
        
        // Добавляем визуальную индикацию того, что обработчик привязан
        oldClone.style.cursor = 'pointer';
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
            modal.classList.remove('active');
        });
        
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.classList.remove('active');
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
            modal.classList.remove('active');
        });
    });
    
    modal.classList.add('active');
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
            modal.classList.remove('active');
        });
        
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.classList.remove('active');
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
            modal.classList.remove('active');
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
    
    modal.classList.add('active');
}

const formatTime = (dateOrTimeString) => {
    if (!dateOrTimeString) return '';
    
    try {
        // Простое извлечение времени из строки ISO с T с помощью регулярного выражения
        if (typeof dateOrTimeString === 'string' && dateOrTimeString.includes('T')) {
            const match = dateOrTimeString.match(/T(\d{2}):(\d{2})/);
            if (match && match.length >= 3) {
                return `${match[1]}:${match[2]}`; // Просто часы и минуты
            }
        }
        
        // Если это строка времени в формате HH:MM:SS, возвращаем только HH:MM
        if (typeof dateOrTimeString === 'string' && dateOrTimeString.includes(':')) {
            return dateOrTimeString.substring(0, 5); // HH:MM
        }
        
        // Если это объект Date
        if (dateOrTimeString instanceof Date) {
            const hours = dateOrTimeString.getUTCHours().toString().padStart(2, '0');
            const minutes = dateOrTimeString.getUTCMinutes().toString().padStart(2, '0');
            return `${hours}:${minutes}`;
        }
        
        return dateOrTimeString; // Если ничего не подходит, вернем как есть
    } catch (e) {
        console.error('Ошибка форматирования времени:', e);
        return dateOrTimeString; // В случае ошибки возвращаем исходное значение
    }
};

const formatDate = (date) => {
    return date.toLocaleDateString('ru-RU', { 
        day: '2-digit', 
        month: '2-digit', 
        year: 'numeric',
        timeZone: 'UTC'
    });
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
                const formattedDate = formatShiftDate(shift.date) || new Date().toLocaleDateString('ru-RU');
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
    
    try {
        // Convert to string if it's not already a string
        const dateStr = typeof dateString === 'string' ? dateString : String(dateString);
        
        // Если строка уже содержит Z в конце, используем ее как есть
        // в противном случае добавляем Z для указания UTC
        const normalizedDateString = dateStr.endsWith('Z') 
            ? dateStr 
            : dateStr + 'Z';

        // Если формат ISO с T, можно просто извлечь дату и время через регулярное выражение
        if (dateStr.includes('T')) {
            const match = dateStr.match(/(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/) || [];
            if (match.length >= 6) {
                const [_, year, month, day, hour, minute] = match;
                return `${day}.${month}.${year.substring(2)} ${hour}:${minute}`;
            }
        }
        
        // Если не удалось извлечь через регулярку, пробуем через Date
        const date = new Date(normalizedDateString);
        
        if (isNaN(date.getTime())) {
            console.warn('Некорректная дата:', dateString);
            return dateStr; // Возвращаем исходную строку вместо "Некорректная дата"
        }
        
        return date.toLocaleString('ru-RU', {
            day: '2-digit',
            month: '2-digit',
            year: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            timeZone: 'UTC' // Используем UTC как исходный часовой пояс
        });
    } catch (e) {
        console.error('Ошибка форматирования времени заказа:', e, dateString);
        return typeof dateString === 'string' ? dateString : String(dateString); // Возвращаем исходную строку вместо "Ошибка даты"
    }
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
            <div class="table-option ${table.status === 'occupied' ? 'occupied' : table.status === 'reserved' ? 'reserved' : 'free'}" 
                 data-table-id="${table.id}">
                <div class="table-number">Стол ${table.number}</div>
                <div class="table-seats">${table.seats} мест</div>
            </div>
        `).join('');

        // Добавляем обработчики для всех столов, независимо от статуса
        grid.querySelectorAll('.table-option').forEach(tableEl => {
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
    // Render order summary in the confirm modal
    const modal = document.getElementById('confirmOrderModal');
    const summaryTable = modal.querySelector('.order-summary__table');
    const summaryComment = modal.querySelector('.order-summary__comment');
    const summaryTotal = modal.querySelector('.order-summary__total');

    // Table info
    let tableText = '';
    if (currentOrderData.tableId) {
        tableText = `<b>Стол №${currentOrderData.tableId}</b>`;
    } else {
        tableText = '<b>Стол не выбран</b>';
    }

    // Items list
    let itemsHtml = '';
    if (currentOrderData.items && currentOrderData.items.length > 0) {
        itemsHtml = '<ul style="padding-left: 18px; margin: 8px 0;">' +
            currentOrderData.items.map(item =>
                `<li>${item.name} &times; ${item.quantity} — <span style='white-space:nowrap;'>${formatMoney(item.price * item.quantity)} KZT</span></li>`
            ).join('') + '</ul>';
    } else {
        itemsHtml = '<i>Нет блюд в заказе</i>';
    }

    summaryTable.innerHTML = `${tableText}${itemsHtml}`;

    // Comment
    if (currentOrderData.comment && currentOrderData.comment.trim() !== '') {
        summaryComment.style.display = '';
        summaryComment.textContent = `Комментарий: ${currentOrderData.comment}`;
    } else {
        summaryComment.style.display = 'none';
        summaryComment.textContent = '';
    }

    // Total
    summaryTotal.innerHTML = `<b>Итого:</b> ${formatMoney(currentOrderData.total)} KZT`;

    modal.classList.add('active');
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
        document.getElementById('tab-bar').style.display = 'flex';
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
        dishCard.innerHTML = `
            <img src="${dish.image_url || DEFAULT_FOOD_IMAGE}" alt="${dish.name}" class="dish-card__image" style="width:48px;height:48px;border-radius:50%;object-fit:cover;flex-shrink:0;">
            <div class="dish-card__details" style="flex:1;min-width:0;">
                <div class="dish-card__name" style="font-weight:600;font-size:16px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${dish.name}</div>
                ${dish.description ? `<div class="dish-card__description" style="color:#888;font-size:13px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${dish.description}</div>` : ''}
                <div class="dish-card__price" style="color:#006FFD;font-weight:600;font-size:15px;">${formatMoney(dish.price)} KZT</div>
            </div>
            <button class="dish-card__add-btn" data-dish-id="${dish.id}">+</button>
        `;
        // Клик по всей карточке добавляет блюдо
        dishCard.addEventListener('click', (e) => {
            // Не срабатывает, если клик по кнопке
            if (e.target.classList.contains('dish-card__add-btn')) return;
            addDishToOrder(dish);
        });
        // Клик по кнопке тоже добавляет
        dishCard.querySelector('.dish-card__add-btn').addEventListener('click', (e) => {
            e.stopPropagation();
            addDishToOrder(dish);
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
    // Удаляем все блюда с указанным id
    currentOrderData.items = currentOrderData.items.filter(item => item.id !== dishId);
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
        document.getElementById('tab-bar').style.display = 'none';
       if (!currentOrderData.tableId) {
            setCreateOrderInteractive(true);
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
        
        // Если это ISO строка даты с часовым поясом
        if (typeof dateString === 'string' && dateString.includes('T')) {
            // Добавляем Z, чтобы указать, что это UTC
            const dateTimeString = dateString.endsWith('Z') ? dateString : dateString + 'Z';
            const date = new Date(dateTimeString);
            
            if (isNaN(date.getTime())) {
                return dateString;
            }
            
            // Получаем компоненты даты в UTC
            const day = date.getUTCDate().toString().padStart(2, '0');
            const month = (date.getUTCMonth() + 1).toString().padStart(2, '0');
            const year = date.getUTCFullYear();
            
            return `${day}.${month}.${year}`;
        }
        
        // Для других форматов
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            return dateString;
        }
        
        return date.toLocaleDateString('ru-RU', { timeZone: 'UTC' });
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
            modal.classList.remove('active');
        });
        
        // Close modal when clicking outside
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.classList.remove('active');
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
                        modal.classList.remove('active');
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
    modal.classList.add('active');
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
    try {
        // Если это строка
        if (typeof date === 'string') {
            if (date.includes('T')) {
                const match = date.match(/T(\d{2}):(\d{2})/);
                if (match && match.length >= 3) {
                    return `${match[1]}:${match[2]}`;
                }
            }
        }
        
        // Если это объект Date
        if (date instanceof Date) {
            return date.toLocaleTimeString('ru-RU', { 
                hour: '2-digit', 
                minute: '2-digit',
                timeZone: 'UTC' 
            });
        }
        
        // Если ничего не сработало, возвращаем как есть или пустую строку
        return date || '';
    } catch (e) {
        console.error('Ошибка форматирования времени стола:', e);
        return date || '';
    }
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
        
        // Добавляем обработчик на кнопку закрытия
        modal.querySelector('.close-modal-btn').addEventListener('click', () => {
            modal.classList.remove('active');
        });
        
        // Закрываем модальное окно при клике вне его
        window.addEventListener('click', (event) => {
            if (event.target === modal) {
                modal.classList.remove('active');
            }
        });
        
        // Обработчик для кнопки "Сбросить"
        modal.querySelector('.clear-filters-btn').addEventListener('click', () => {
            window.tableFilters.statuses = [];
            modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
                checkbox.checked = false;
            });
        });
        
        // Обработчик для кнопки "Применить"
        modal.querySelector('.apply-filters-btn').addEventListener('click', () => {
            const selectedStatuses = [];
            modal.querySelectorAll('input[type="checkbox"]:checked').forEach(checkbox => {
                selectedStatuses.push(checkbox.value);
            });
            window.tableFilters.statuses = selectedStatuses;
            renderTablesWithFilter();
            updateTableFilterBadge();
            modal.classList.remove('active');
        });
    }
    
    // Устанавливаем состояние чекбоксов в соответствии с текущими фильтрами
    modal.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
        checkbox.checked = window.tableFilters.statuses.includes(checkbox.value);
    });
    
    // Показываем модальное окно
    modal.classList.add('active');
}