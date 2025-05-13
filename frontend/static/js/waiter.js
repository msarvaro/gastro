document.addEventListener('DOMContentLoaded', async function() {
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    
    if (!token || role !== 'waiter') {
        window.location.href = '/';
        return;
    }

    // Initialize event listeners
    setupEventListeners();

    // Check token and load initial data
    try {
        const resp = await fetch('/api/waiter/dashboard', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (!resp.ok) {
            console.error('Auth check failed:', resp.status);
            localStorage.removeItem('token');
            localStorage.removeItem('role');
            document.cookie = 'auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
            window.location.href = '/';
            return;
        }

        // Load initial data based on current section
        const currentPath = window.location.pathname;
        const sections = {
            '/waiter': 'tables',
            '/waiter/orders': 'orders',
            '/waiter/history': 'history',
            '/waiter/profile': 'profile'
        };
        
        const activeSection = sections[currentPath] || 'tables';
        showSection(activeSection);
        
        // Highlight active menu item
        document.querySelectorAll('.sidebar nav ul li').forEach(li => {
            const route = li.getAttribute('data-route');
            li.classList.toggle('active', route === currentPath);
        });
    } catch (e) {
        console.error('Error during initialization:', e);
    }
});

function setupEventListeners() {
    // Navigation menu
    document.querySelectorAll('.sidebar nav ul li').forEach(item => {
        item.addEventListener('click', function() {
            const route = this.getAttribute('data-route');
            if (route) {
                window.location.href = route;
            }
        });
    });

    // Filter buttons
    document.querySelectorAll('.filter-btn').forEach(button => {
        button.addEventListener('click', function() {
            const filter = this.getAttribute('data-filter');
            const section = this.closest('.section').id;
            
            // Update active state
            this.closest('.filters').querySelectorAll('.filter-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            this.classList.add('active');
            
            // Apply filter
            switch(section) {
                case 'tables-section':
                    filterTables(filter);
                    break;
                case 'orders-section':
                    filterOrders(filter);
                    break;
                case 'history-section':
                    filterHistory(filter);
                    break;
            }
        });
    });

    // Add order button
    const addOrderBtn = document.getElementById('add-order-btn');
    if (addOrderBtn) {
        addOrderBtn.addEventListener('click', () => {
            showModal('create-order-modal');
            loadFreeTables();
            loadMenuItems();
        });
    }

    // Close modals
    document.querySelectorAll('.close-modal').forEach(button => {
        button.addEventListener('click', function() {
            const modal = this.closest('.modal');
            if (modal) {
                closeModal(modal.id);
            }
        });
    });

    // Create order form
    const createOrderForm = document.getElementById('create-order-form');
    if (createOrderForm) {
        createOrderForm.addEventListener('submit', handleCreateOrder);
    }

    // Menu search
    const menuSearch = document.getElementById('menu-search');
    if (menuSearch) {
        menuSearch.addEventListener('input', debounce(filterMenuItems, 300));
    }

    // Menu categories
    document.querySelectorAll('.category-btn').forEach(button => {
        button.addEventListener('click', function() {
            const category = this.getAttribute('data-category');
            filterMenuByCategory(category);
            
            // Update active state
            this.closest('.categories').querySelectorAll('.category-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            this.classList.add('active');
        });
    });

    // Profile form
    const profileForm = document.getElementById('profile-form');
    if (profileForm) {
        profileForm.addEventListener('submit', handleProfileUpdate);
    }
}

// Section switching
function showSection(section) {
    const sections = ['tables', 'orders', 'history', 'profile'];
    sections.forEach(s => {
        const el = document.getElementById(s + '-section');
        if (el) el.style.display = (s === section) ? 'block' : 'none';
    });
    
    // Load section data
    switch(section) {
        case 'tables':
            loadTables();
            break;
        case 'orders':
            loadOrders();
            break;
        case 'history':
            loadHistory();
            break;
        case 'profile':
            loadProfile();
            break;
    }
}

// Tables Management
async function loadTables() {
    try {
        const response = await fetch('/api/waiter/tables', {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load tables');
        }

        const data = await response.json();
        const tables = data.tables;
        const stats = data.stats;

        // Update statistics
        document.getElementById('free-tables').textContent = stats.free;
        document.getElementById('occupancy').textContent = `${Math.round(stats.occupancy_percentage)}%`;

        // Render tables
        const grid = document.getElementById('tables-grid');
        grid.innerHTML = tables.map(table => `
            <div class="table-card ${table.status}" data-id="${table.id}">
                <div class="table-number">Стол ${table.number}</div>
                <div class="table-seats">${table.seats} мест</div>
                ${table.current_order ? `
                    <div class="table-order">
                        <div class="order-time">Занят: ${formatTime(table.occupied_at)}</div>
                        <div class="order-items">${table.order_items || 'Нет заказов'}</div>
                    </div>
                ` : ''}
                ${table.reserved_at ? `
                    <div class="table-reservation">
                        <div class="reservation-time">Бронь: ${formatTime(table.reserved_at)}</div>
                    </div>
                ` : ''}
            </div>
        `).join('');

        // Add click handlers
        document.querySelectorAll('.table-card').forEach(card => {
            card.addEventListener('click', () => {
                const tableId = card.getAttribute('data-id');
                showTableDetails(tableId);
            });
        });

    } catch (error) {
        console.error('Error loading tables:', error);
    }
}

// Orders Management
async function loadOrders() {
    try {
        const response = await fetch('/api/waiter/orders', {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load orders');
        }

        const data = await response.json();
        const orders = data.orders;
        const stats = data.stats;

        // Update statistics
        document.getElementById('active-orders').textContent = stats.active;
        document.getElementById('new-orders').textContent = stats.new;
        document.getElementById('in-progress-orders').textContent = stats.in_progress;
        document.getElementById('ready-orders').textContent = stats.ready;

        // Render orders
        const list = document.getElementById('orders-list');
        list.innerHTML = orders.map(order => `
            <div class="order-card ${order.status}" data-id="${order.id}">
                <div class="order-header">
                    <div class="order-info">
                        <span class="order-number">Заказ #${order.id}</span>
                        <span class="table-number">Стол ${order.table_number}</span>
                    </div>
                    <div class="order-time">${formatTime(order.created_at)}</div>
                </div>
                <div class="order-items">
                    ${order.items.map(item => `
                        <div class="order-item">
                            <span class="item-name">${item.name}</span>
                            <span class="item-quantity">x${item.quantity}</span>
                            <span class="item-price">${formatMoney(item.price)} ₸</span>
                        </div>
                    `).join('')}
                </div>
                <div class="order-footer">
                    <div class="order-total">Итого: ${formatMoney(order.total_amount)} ₸</div>
                    <div class="order-actions">
                        <button class="status-btn ${order.status}" onclick="updateOrderStatus(${order.id}, '${getNextStatus(order.status)}')">
                            ${getStatusText(order.status)}
                        </button>
                    </div>
                </div>
            </div>
        `).join('');

    } catch (error) {
        console.error('Error loading orders:', error);
    }
}

// History Management
async function loadHistory() {
    try {
        const response = await fetch('/api/waiter/history', {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load history');
        }

        const data = await response.json();
        const orders = data.orders;
        const stats = data.stats;

        // Update statistics
        document.getElementById('completed-orders').textContent = stats.completed;
        document.getElementById('cancelled-orders').textContent = stats.cancelled;
        document.getElementById('total-amount').textContent = `${formatMoney(stats.total_amount)} ₸`;

        // Render history
        const list = document.getElementById('history-list');
        list.innerHTML = orders.map(order => `
            <div class="history-card ${order.status}" data-id="${order.id}">
                <div class="history-header">
                    <div class="history-info">
                        <span class="order-number">Заказ #${order.id}</span>
                        <span class="table-number">Стол ${order.table_number}</span>
                    </div>
                    <div class="history-time">${formatTime(order.completed_at || order.created_at)}</div>
                </div>
                <div class="history-items">
                    ${order.items.map(item => `
                        <div class="history-item">
                            <span class="item-name">${item.name}</span>
                            <span class="item-quantity">x${item.quantity}</span>
                            <span class="item-price">${formatMoney(item.price)} ₸</span>
                        </div>
                    `).join('')}
                </div>
                <div class="history-footer">
                    <div class="history-total">Итого: ${formatMoney(order.total_amount)} ₸</div>
                    <div class="history-status">${getStatusText(order.status)}</div>
                </div>
            </div>
        `).join('');

    } catch (error) {
        console.error('Error loading history:', error);
    }
}

// Profile Management
async function loadProfile() {
    try {
        const response = await fetch('/api/waiter/profile', {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load profile');
        }

        const profile = await response.json();

        // Update profile info
        document.getElementById('profile-name').textContent = `${profile.firstname} ${profile.lastname}`;
        document.getElementById('profile-firstname').value = profile.firstname;
        document.getElementById('profile-lastname').value = profile.lastname;
        document.getElementById('profile-email').value = profile.email;
        document.getElementById('profile-phone').value = profile.phone;

        // Update statistics
        document.getElementById('orders-today').textContent = profile.stats.orders_today;
        document.getElementById('average-check').textContent = `${formatMoney(profile.stats.average_check)} ₸`;
        document.getElementById('rating').textContent = profile.stats.rating.toFixed(1);

    } catch (error) {
        console.error('Error loading profile:', error);
    }
}

// Utility functions
function formatTime(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

function getStatusText(status) {
    const statusMap = {
        'new': 'Новый',
        'in_progress': 'В работе',
        'ready': 'Готов',
        'completed': 'Выполнен',
        'cancelled': 'Отменен'
    };
    return statusMap[status] || status;
}

function getNextStatus(currentStatus) {
    const statusFlow = {
        'new': 'in_progress',
        'in_progress': 'ready',
        'ready': 'completed'
    };
    return statusFlow[currentStatus] || currentStatus;
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Modal controls
function showModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.add('active');
    }
}

function closeModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.remove('active');
    }
}

// Event handlers
async function handleCreateOrder(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    
    const order = {
        table_id: parseInt(formData.get('table_id')),
        items: Array.from(document.querySelectorAll('.selected-item')).map(item => ({
            menu_item_id: parseInt(item.getAttribute('data-id')),
            quantity: parseInt(item.querySelector('.quantity').textContent),
            price: parseFloat(item.getAttribute('data-price'))
        })),
        comment: formData.get('comment')
    };

    try {
        const response = await fetch('/api/waiter/orders', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify(order)
        });

        if (!response.ok) {
            throw new Error('Failed to create order');
        }

        closeModal('create-order-modal');
        loadOrders();
        loadTables();
    } catch (error) {
        console.error('Error creating order:', error);
        alert('Ошибка при создании заказа: ' + error.message);
    }
}

async function handleProfileUpdate(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    
    const profile = {
        firstname: formData.get('firstname'),
        lastname: formData.get('lastname'),
        email: formData.get('email'),
        phone: formData.get('phone')
    };

    try {
        const response = await fetch('/api/waiter/profile', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify(profile)
        });

        if (!response.ok) {
            throw new Error('Failed to update profile');
        }

        alert('Профиль успешно обновлен');
        loadProfile();
    } catch (error) {
        console.error('Error updating profile:', error);
        alert('Ошибка при обновлении профиля: ' + error.message);
    }
}

async function updateOrderStatus(orderId, newStatus) {
    try {
        const response = await fetch(`/api/waiter/orders/${orderId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify({ status: newStatus })
        });

        if (!response.ok) {
            throw new Error('Failed to update order status');
        }

        loadOrders();
        loadTables();
    } catch (error) {
        console.error('Error updating order status:', error);
        alert('Ошибка при обновлении статуса заказа: ' + error.message);
    }
} 