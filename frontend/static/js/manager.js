document.addEventListener('DOMContentLoaded', async function() {
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    
    if (!token || role !== 'manager') {
        window.location.href = '/';
        return;
    }

    // Проверяем токен через API
    try {
        const resp = await fetch('/api/manager/dashboard', {
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

        // Если авторизация успешна, загружаем данные
        await loadDashboardData();
        setupEventListeners();
        
        // Показываем основную секцию
        const currentPath = window.location.pathname;
        if (currentPath === '/manager') {
            showSection('main');
        } else if (currentPath === '/manager/inventory') {
            showSection('inventory');
            showInventoryTab('stock');
        } else if (currentPath === '/manager/menu') {
            showSection('menu');
        } else if (currentPath === '/manager/finances') {
            showSection('finances');
        } else if (currentPath === '/manager/staff') {
            showSection('staff');
        } else if (currentPath === '/manager/settings') {
            showSection('settings');
        } else if (currentPath === '/manager/analytics') {
            showSection('analytics');
        }
    } catch (e) {
        console.error('Auth check failed:', e);
        localStorage.removeItem('token');
        localStorage.removeItem('role');
        document.cookie = 'auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        window.location.href = '/';
    }
});

async function loadDashboardData() {
    try {
        const token = localStorage.getItem('token');
        // Fetch orders history from API
        const response = await fetch('/api/manager/orders/history', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load dashboard data');
        }

        const orders = await response.json();
        
        // Calculate statistics
        const completedOrders = orders.filter(order => order.status === 'completed');
        const totalRevenue = completedOrders.reduce((sum, order) => sum + order.total, 0);
        const visitorCount = completedOrders.length;
        const averageCheck = visitorCount > 0 ? totalRevenue / visitorCount : 0;

        // Update dashboard cards
        document.querySelector('.card:nth-child(1) .value').textContent = 
            `${formatMoney(totalRevenue)}₸`;
        document.querySelector('.card:nth-child(2) .value').textContent = 
            `${visitorCount}`;
        document.querySelector('.card:nth-child(3) .value').textContent = 
            `${formatMoney(Math.round(averageCheck))}₸`;

        // Calculate daily comparison
        const today = new Date().toDateString();
        const yesterday = new Date(Date.now() - 86400000).toDateString();
        
        const todayOrders = completedOrders.filter(order => 
            new Date(order.completed_at).toDateString() === today);
        const yesterdayOrders = completedOrders.filter(order => 
            new Date(order.completed_at).toDateString() === yesterday);

        const todayRevenue = todayOrders.reduce((sum, order) => sum + order.total, 0);
        const yesterdayRevenue = yesterdayOrders.reduce((sum, order) => sum + order.total, 0);
        const revenueChange = yesterdayRevenue ? ((todayRevenue - yesterdayRevenue) / yesterdayRevenue * 100).toFixed(0) : 0;
        
        const visitorChange = yesterdayOrders.length ? 
            ((todayOrders.length - yesterdayOrders.length) / yesterdayOrders.length * 100).toFixed(0) : 0;
        
        const todayAvgCheck = todayOrders.length ? todayRevenue / todayOrders.length : 0;
        const yesterdayAvgCheck = yesterdayOrders.length ? yesterdayRevenue / yesterdayOrders.length : 0;
        const avgCheckChange = yesterdayAvgCheck ? 
            ((todayAvgCheck - yesterdayAvgCheck) / yesterdayAvgCheck * 100).toFixed(0) : 0;

        // Update comparison indicators
        updateComparisonIndicator(1, revenueChange);
        updateComparisonIndicator(2, visitorChange);
        updateComparisonIndicator(3, avgCheckChange);

    } catch (error) {
        console.error('Error loading dashboard data:', error);
        // Show error message to user
        const cards = document.querySelectorAll('.card .value');
        cards.forEach(card => {
            card.textContent = '—';
        });
    }
}

function updateComparisonIndicator(cardIndex, change) {
    const indicator = document.querySelector(`.card:nth-child(${cardIndex}) .desc span`);
    const color = change > 0 ? '#006FFD' : '#5D7285';
    const sign = change > 0 ? '+' : '-';
    indicator.textContent = `${sign}${change}% от вчера`;
    indicator.style.color = color;
}

function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

function setupEventListeners() {
    // Sidebar toggle
    document.querySelector('.logo').addEventListener('click', toggleSidebar);
    
    // Navigation menu
    document.querySelectorAll('.sidebar nav ul li').forEach(item => {
        item.addEventListener('click', function() {
            const section = this.getAttribute('data-section');
            if (section) {
                navigateTo(section);
            }
        });
    });

    // Inventory tabs
    document.querySelectorAll('.tabs .tab-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const tab = this.getAttribute('onclick').match(/'([^']+)'/)[1];
            showInventoryTab(tab);
        });
    });

    // Modal controls
    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.addEventListener('click', function() {
            const modalId = this.closest('.modal').id;
            closeModal(modalId);
        });
    });

    // Add Product button
    document.getElementById('addProductBtn').addEventListener('click', function() {
        openModal('addProductModal');
    });

    // Add Request button
    document.getElementById('addRequestBtn').addEventListener('click', function() {
        openModal('addRequestModal');
    });

    // Add Supplier button
    document.getElementById('addSupplierBtn').addEventListener('click', function() {
        openModal('addSupplierModal');
    });

    // Form submissions
    document.getElementById('addProductForm').addEventListener('submit', handleAddProduct);
    document.getElementById('addRequestForm').addEventListener('submit', handleAddRequest);
    document.getElementById('addSupplierForm').addEventListener('submit', handleAddSupplier);
    document.getElementById('editSupplierModal form').addEventListener('submit', handleEditSupplier);

    // Поиск и фильтры для склада
    document.getElementById('productSearch').addEventListener('input', loadInventoryData);
    document.getElementById('productCategoryFilter').addEventListener('change', loadInventoryData);
    document.getElementById('productBranchFilter').addEventListener('change', loadInventoryData);

    // Поиск и фильтры для заявок
    document.getElementById('requestSearch').addEventListener('input', loadRequestsData);
    document.getElementById('requestStatusFilter').addEventListener('change', loadRequestsData);
    document.getElementById('requestBranchFilter').addEventListener('change', loadRequestsData);

    // Поиск и фильтры для поставщиков
    document.getElementById('supplierSearch').addEventListener('input', loadSuppliersData);
    document.getElementById('supplierCategoryFilter').addEventListener('change', loadSuppliersData);
}

// Sidebar toggle
function toggleSidebar() {
    document.getElementById('sidebar').classList.toggle('closed');
}

// Section switching
function showSection(section) {
    const sections = ['main', 'finances', 'menu', 'inventory', 'staff', 'settings', 'analytics'];
    sections.forEach(s => {
        const el = document.getElementById(s + '-section');
        if (el) el.style.display = (s === section) ? '' : 'none';
    });
    
    // Highlight active menu
    document.querySelectorAll('.sidebar nav ul li').forEach((li, idx) => {
        li.classList.toggle('active', idx === sections.indexOf(section));
    });
}

// Inventory tabs
function showInventoryTab(tab) {
    const tabs = ['stock', 'requests', 'suppliers', 'history'];
    tabs.forEach(t => {
        const el = document.getElementById('inventory-' + t + '-tab');
        if (el) el.style.display = (t === tab) ? '' : 'none';
    });
    
    // Highlight active tab
    document.querySelectorAll('.tabs .tab-btn').forEach((btn, idx) => {
        btn.classList.toggle('active', idx === tabs.indexOf(tab));
    });
}

// Modal controls
function openModal(id) {
    document.getElementById(id).classList.add('active');
}

function closeModal(id) {
    document.getElementById(id).classList.remove('active');
}

// Initialize default view
showSection('main');
showInventoryTab('stock');

// Inventory Management
async function loadInventoryData() {
    try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/manager/inventory', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load inventory data');
        }

        const data = await response.json();
        const items = data.items || [];

        // Фильтрация
        const search = document.getElementById('productSearch').value.trim().toLowerCase();
        const category = document.getElementById('productCategoryFilter').value;
        const branch = document.getElementById('productBranchFilter').value;
        let filteredItems = items;
        if (search) filteredItems = items.filter(i => i.name.toLowerCase().includes(search));
        if (category) filteredItems = filteredItems.filter(i => i.category === category);
        if (branch) filteredItems = filteredItems.filter(i => i.branch === branch);

        // Update inventory cards
        document.querySelector('#inventory-stock-tab .card:nth-child(1) .value').textContent = `${filteredItems.length}`;
        document.querySelector('#inventory-stock-tab .card:nth-child(2) .value').textContent = 
            `${filteredItems.filter(item => getStatusClass(item)==='low').length}`;
        document.querySelector('#inventory-stock-tab .card:nth-child(3) .value').textContent = 
            `${filteredItems.filter(item => item.status === 'pending').length}`;

        // Update inventory table
        const table = document.querySelector('#inventory-stock-tab table');
        if (table) {
            const tbody = table.querySelector('tbody');
            if (tbody) {
                tbody.innerHTML = filteredItems.map(item => `
                    <tr>
                        <td>${item.name}</td>
                        <td>${item.category}</td>
                        <td>${item.quantity} ${item.unit}</td>
                        <td>${item.minQuantity} ${item.unit}</td>
                        <td><span class="status-${getStatusClass(item)}">${getStatusText(item)}</span></td>
                    </tr>
                `).join('');
            }
        }
    } catch (error) {
        console.error('Error loading inventory data:', error);
    }
}

function getStatusClass(item) {
    if (item.quantity < item.minQuantity/2) return 'critical';
    if (item.quantity < item.minQuantity) return 'low';
    return 'ok';
}

function getStatusText(item) {
    if (item.quantity < item.minQuantity/2) return 'Критично';
    if (item.quantity < item.minQuantity) return 'Низкий';
    return 'В норме';
}

// Requests Management
async function loadRequestsData() {
    try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/manager/requests', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load requests data');
        }

        const data = await response.json();
        const requests = data.requests || [];
        
        // Фильтрация
        const search = document.getElementById('requestSearch').value.trim().toLowerCase();
        const status = document.getElementById('requestStatusFilter').value;
        const branch = document.getElementById('requestBranchFilter').value;
        let filteredRequests = requests;
        if (search) filteredRequests = requests.filter(r => r.items.join(', ').toLowerCase().includes(search));
        if (status) filteredRequests = filteredRequests.filter(r => r.status === status);
        if (branch) filteredRequests = filteredRequests.filter(r => r.branch === branch);

        // Update requests cards
        document.querySelector('#inventory-requests-tab .card:nth-child(1) .value').textContent = 
            `${filteredRequests.filter(req => req.status === 'active').length}`;
        document.querySelector('#inventory-requests-tab .card:nth-child(2) .value').textContent = 
            `${filteredRequests.filter(req => req.status === 'pending').length}`;
        document.querySelector('#inventory-requests-tab .card:nth-child(3) .value').textContent = 
            `${filteredRequests.filter(req => req.status === 'completed' && 
                new Date(req.completedAt).getMonth() === new Date().getMonth()).length}`;

        // Update requests table
        const table = document.querySelector('#inventory-requests-tab table');
        if (table) {
            const tbody = table.querySelector('tbody');
            if (tbody) {
                tbody.innerHTML = filteredRequests.map(request => `
                    <tr>
                        <td>${request.items.join(', ')}</td>
                        <td>${request.branch}</td>
                        <td>${formatDate(request.createdAt)}</td>
                        <td><span class="status-${request.status}">${getRequestStatusText(request.status)}</span></td>
                        <td>
                            ${request.status === 'pending' ? 
                                `<button onclick="approveRequest('${request.id}')">✔</button> 
                                 <button onclick="rejectRequest('${request.id}')">✖</button>` :
                                `<button onclick="showRequestDetails('${request.id}')">Подробнее</button>`
                            }
                        </td>
                    </tr>
                `).join('');
            }
        }
    } catch (error) {
        console.error('Error loading requests data:', error);
    }
}

function getRequestStatusText(status) {
    const statusMap = {
        'pending': 'Ожидает одобрения',
        'active': 'В обработке',
        'completed': 'Выполнено',
        'rejected': 'Отклонено'
    };
    return statusMap[status] || status;
}

// Suppliers Management
async function loadSuppliersData() {
    try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/manager/suppliers', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load suppliers data');
        }

        const data = await response.json();
        const suppliers = data.suppliers || [];
        
        // Фильтрация
        const search = document.getElementById('supplierSearch').value.trim().toLowerCase();
        const category = document.getElementById('supplierCategoryFilter').value;
        let filteredSuppliers = suppliers;
        if (search) filteredSuppliers = suppliers.filter(s => s.name.toLowerCase().includes(search));
        if (category) filteredSuppliers = filteredSuppliers.filter(s => s.categories.includes(category));
        
        // Update suppliers table
        const table = document.getElementById('suppliersTable');
        if (table) {
            const tbody = table.querySelector('tbody');
            if (tbody) {
                tbody.innerHTML = filteredSuppliers.map(supplier => `
                    <tr>
                        <td>${supplier.name}</td>
                        <td>${supplier.categories.join(', ')}</td>
                        <td>${supplier.phone}<br>${supplier.email}<br>${supplier.address}</td>
                        <td><span class="status-${supplier.status}">${getSupplierStatusText(supplier.status)}</span></td>
                        <td><button onclick="editSupplier('${supplier.id}')">Редактировать</button></td>
                    </tr>
                `).join('');
            }
        }
    } catch (error) {
        console.error('Error loading suppliers data:', error);
    }
}

function getSupplierStatusText(status) {
    const statusMap = {
        'active': 'Активный',
        'paused': 'На паузе',
        'archived': 'Архивный'
    };
    return statusMap[status] || status;
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric'
    });
}

// Обновляем функцию навигации
function navigateTo(section) {
    const routes = {
        'main': '/manager',
        'inventory': '/manager/inventory',
        'menu': '/manager/menu',
        'finances': '/manager/finances',
        'staff': '/manager/staff',
        'settings': '/manager/settings',
        'analytics': '/manager/analytics'
    };

    if (routes[section]) {
        window.location.href = routes[section];
    }
} 