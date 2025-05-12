document.addEventListener('DOMContentLoaded', async function() {
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    
    if (!token || role !== 'manager') {
        window.location.href = '/';
        return;
    }

    // Инициализация обработчиков событий
    setupEventListeners();

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
        
        // Показываем основную секцию
        const currentPath = window.location.pathname;
        const sections = {
            '/manager': 'main',
            '/manager/inventory': 'inventory',
            '/manager/menu': 'menu',
            '/manager/finances': 'finances',
            '/manager/staff': 'staff',
            '/manager/settings': 'settings',
            '/manager/analytics': 'analytics'
        };
        
        const activeSection = sections[currentPath] || 'main';
        showSection(activeSection);
        
        // Подсвечиваем активный пункт меню
        document.querySelectorAll('.sidebar nav ul li').forEach(li => {
            const route = li.getAttribute('data-route');
            li.classList.toggle('active', route === currentPath);
        });
        
        if (activeSection === 'inventory') {
            showInventoryTab('stock');
        }
    } catch (e) {
        console.error('Auth check failed:', e);
        localStorage.removeItem('token');
        localStorage.removeItem('role');
        document.cookie = 'auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        window.location.href = '/';
    }
});

async function handleAddProduct(event) {
    event.preventDefault();
    const form = event.target;

    try {
        const quantity = parseFloat(form.productQuantity.value.replace(',', '.'));
        const minQuantity = parseFloat(form.productMinQuantity.value.replace(',', '.'));

        if (isNaN(quantity) || isNaN(minQuantity)) {
            throw new Error('Некорректные числовые значения');
        }

        const newProduct = {
            name: form.productName.value,
            category: form.productCategory.value,
            quantity: quantity,
            unit: form.productUnit.value,
            min_quantity: minQuantity,
            min_unit: form.productMinUnit.value
        };

        const response = await fetch('/api/manager/inventory', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify(newProduct)
        });

        if (!response.ok) {
            throw new Error('Ошибка при добавлении продукта');
        }

        form.reset();
        closeModal('addProductModal');
        loadInventoryData();
    } catch (error) {
        alert('Ошибка при добавлении продукта: ' + error.message);
    }
}

async function handleAddSupplier(event) {
    event.preventDefault();
    const form = event.target;
    const newSupplier = {
        name: form.supplierName.value,
        categories: form.supplierCategory.value.split(',').map(cat => cat.trim()),
        phone: form.supplierPhone.value,
        email: form.supplierEmail.value,
        address: form.supplierAddress.value,
        status: 'active'
    };
    try {
        const response = await fetch('/api/manager/suppliers', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify(newSupplier)
        });
        if (!response.ok) throw new Error('Ошибка при добавлении поставщика');
        closeModal('addSupplierModal');
        loadSuppliersData();
    } catch (error) {
        alert('Ошибка при добавлении поставщика: ' + error.message);
    }
}

async function handleAddRequest(event) {
    event.preventDefault();
    const form = event.target;

    // Получаем список товаров из скрытого поля
    let items = [];
    try {
        items = JSON.parse(form.requestItemsData.value);
    } catch (e) {
        alert('Ошибка: не удалось прочитать список товаров');
        return;
    }

    const newRequest = {
        branch: form.requestBranch.value,
        supplier: form.requestSupplier.value,
        items: items, // теперь это массив объектов {id, name, qty, unit}
        priority: form.requestPriority.value,
        comment: form.requestComment.value,
        status: 'pending',
        createdAt: new Date().toISOString()
    };
    try {
        const response = await fetch('/api/manager/requests', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify(newRequest)
        });
        if (!response.ok) throw new Error('Ошибка при создании заявки');
        closeModal('addRequestModal');
        loadRequestsData();
    } catch (error) {
        alert('Ошибка при создании заявки: ' + error.message);
    }
}

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
        const cards = document.querySelectorAll('.card .value');
        if (cards.length >= 3) {
            cards[0].textContent = `${formatMoney(totalRevenue)}₸`;
            cards[1].textContent = `${visitorCount}`;
            cards[2].textContent = `${formatMoney(Math.round(averageCheck))}₸`;
        }

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
        const indicators = document.querySelectorAll('.card .desc span');
        if (indicators.length >= 3) {
            updateComparisonIndicator(1, revenueChange);
            updateComparisonIndicator(2, visitorChange);
            updateComparisonIndicator(3, avgCheckChange);
        }

    } catch (error) {
        console.error('Error loading dashboard data:', error);
        // Show error message to user
        const cards = document.querySelectorAll('.card .value');
        cards.forEach(card => {
            card.textContent = '—';
        });
        const indicators = document.querySelectorAll('.card .desc span');
        indicators.forEach(indicator => {
            indicator.textContent = '';
        });
    }
}

function updateComparisonIndicator(cardIndex, change) {
    const indicator = document.querySelector(`.card:nth-child(${cardIndex}) .desc span`);
    if (!indicator) return;
    
    const color = change > 0 ? '#006FFD' : '#5D7285';
    const sign = change > 0 ? '+' : '';
    indicator.textContent = `${sign}${change}% от вчера`;
    indicator.style.color = color;
}

function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

function setupEventListeners() {
    // Sidebar toggle
    const logo = document.querySelector('.logo');
    if (logo) {
        logo.addEventListener('click', function(e) {
            e.preventDefault();
            e.stopPropagation();
            toggleSidebar();
        });
    }
    
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
    document.querySelectorAll('.filter-button').forEach(button => {
        button.addEventListener('click', function() {
            const type = this.classList.contains('filter-button--time') ? 'time' : 'filter';
            toggleFilterDropdown(type);
        });
    });

    // Close modals when clicking outside
    window.addEventListener('click', function(event) {
        if (event.target.classList.contains('modal')) {
            closeModal(event.target.id);
        }
    });

    // Close modals when clicking close button
    document.querySelectorAll('.close-modal').forEach(button => {
        button.addEventListener('click', function() {
            const modal = this.closest('.modal');
            if (modal) {
                closeModal(modal.id);
            }
        });
    });

    // Add product button
    const addProductBtn = document.getElementById('addProductBtn');
    if (addProductBtn) {
        addProductBtn.addEventListener('click', () => {
            showModal('addProductModal');
        });
    }

    // Add request button
    const addRequestBtn = document.getElementById('addRequestBtn');
    if (addRequestBtn) {
        addRequestBtn.addEventListener('click', () => {
            showModal('addRequestModal');
            populateSupplierSelect();
            populateRequestItemsSelect();
        });
    }

    // Add supplier button
    const addSupplierBtn = document.getElementById('addSupplierBtn');
    if (addSupplierBtn) {
        addSupplierBtn.addEventListener('click', () => {
            showModal('addSupplierModal');
        });
    }

    // Form submissions
    const addProductForm = document.getElementById('addProductForm');
    if (addProductForm) {
        addProductForm.addEventListener('submit', handleAddProduct);
    }

    const addRequestForm = document.getElementById('addRequestForm');
    if (addRequestForm) {
        addRequestForm.addEventListener('submit', handleAddRequest);
    }

    const addSupplierForm = document.getElementById('addSupplierForm');
    if (addSupplierForm) {
        addSupplierForm.addEventListener('submit', handleAddSupplier);
    }

    // Search inputs
    const productSearch = document.getElementById('productSearch');
    if (productSearch) {
        productSearch.addEventListener('input', debounce(loadInventoryData, 300));
    }

    const requestSearch = document.getElementById('requestSearch');
    if (requestSearch) {
        requestSearch.addEventListener('input', debounce(loadRequestsData, 300));
    }

    const supplierSearch = document.getElementById('supplierSearch');
    if (supplierSearch) {
        supplierSearch.addEventListener('input', debounce(loadSuppliersData, 300));
    }

    // Filter selects
    const productCategoryFilter = document.getElementById('productCategoryFilter');
    if (productCategoryFilter) {
        productCategoryFilter.addEventListener('change', loadInventoryData);
    }

    const productBranchFilter = document.getElementById('productBranchFilter');
    if (productBranchFilter) {
        productBranchFilter.addEventListener('change', loadInventoryData);
    }

    const requestStatusFilter = document.getElementById('requestStatusFilter');
    if (requestStatusFilter) {
        requestStatusFilter.addEventListener('change', loadRequestsData);
    }

    const requestBranchFilter = document.getElementById('requestBranchFilter');
    if (requestBranchFilter) {
        requestBranchFilter.addEventListener('change', loadRequestsData);
    }

    const supplierCategoryFilter = document.getElementById('supplierCategoryFilter');
    if (supplierCategoryFilter) {
        supplierCategoryFilter.addEventListener('change', loadSuppliersData);
    }

    // Add event listeners for menu management buttons
    const addCategoryBtn = document.getElementById('addCategoryBtn');
    if (addCategoryBtn) {
        addCategoryBtn.addEventListener('click', function() {
            showModal('addCategoryModal');
        });
    }
    const addMenuItemBtn = document.getElementById('addMenuItemBtn');
    if (addMenuItemBtn) {
        addMenuItemBtn.addEventListener('click', function() {
            showModal('addMenuItemModal');
        });
    }
}

// Sidebar toggle
function toggleSidebar() {
    const sidebar = document.getElementById('sidebar');
    if (sidebar) {
        sidebar.classList.toggle('closed');
        console.log('Sidebar toggled:', sidebar.classList.contains('closed')); // Debug log
        
        // Обновляем отступ основного контента
        const mainContent = document.querySelector('.main-content');
        if (mainContent) {
            if (sidebar.classList.contains('closed')) {
                mainContent.style.marginLeft = '88px';
            } else {
                mainContent.style.marginLeft = '260px';
            }
        }
    }
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
        const searchInput = document.getElementById('productSearch');
        const categorySelect = document.getElementById('productCategoryFilter');
        const branchSelect = document.getElementById('productBranchFilter');
        
        const search = searchInput ? searchInput.value.trim().toLowerCase() : '';
        const category = categorySelect ? categorySelect.value : '';
        const branch = branchSelect ? branchSelect.value : '';
        
        let filteredItems = items;
        if (search) filteredItems = items.filter(i => i.name.toLowerCase().includes(search));
        if (category) filteredItems = filteredItems.filter(i => i.category === category);
        if (branch) filteredItems = filteredItems.filter(i => i.branch === branch);

        console.log('Загружено позиций:', filteredItems.length);
        console.log('Карточки:', document.querySelectorAll('#inventory-stock-tab .card .value'));

        // Update inventory cards
        const cards = document.querySelectorAll('#inventory-stock-tab .card .value');
        if (cards.length >= 3) {
            cards[0].textContent = `${filteredItems.length}`;
            cards[1].textContent = `${filteredItems.filter(item => getStatusClass(item)==='low').length}`;
            cards[2].textContent = `${filteredItems.filter(item => item.status === 'pending').length}`;
        }

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
        // Show error message to user
        const cards = document.querySelectorAll('#inventory-stock-tab .card .value');
        cards.forEach(card => {
            card.textContent = '—';
        });
        const table = document.querySelector('#inventory-stock-tab table tbody');
        if (table) {
            table.innerHTML = '<tr><td colspan="5">Ошибка загрузки данных</td></tr>';
        }
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

// Utility functions
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

// Menu Management Functions
async function loadMenuData() {
    console.log('loadMenuData called');
    try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/menu', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load menu data');
        }

        const data = await response.json();
        updateMenuUI(data.categories, data.items);
    } catch (error) {
        console.error('Error loading menu data:', error);
        showError('Ошибка загрузки меню');
    }
}

function updateMenuUI(categories, items) {
    // Update categories
    const categoriesContainer = document.querySelector('#menu-section .categories');
    if (categoriesContainer) {
        categoriesContainer.innerHTML = categories.map(category => `
            <div class="category-card" data-category-id="${category.id}">
                <h3>${category.name}</h3>
                <div class="category-actions">
                    <button onclick="editCategory(${category.id})">✏️</button>
                    <button onclick="deleteCategory(${category.id})">🗑️</button>
                </div>
            </div>
        `).join('');
    }

    // Update menu items
    const itemsContainer = document.querySelector('#menu-section .menu-items');
    if (itemsContainer) {
        itemsContainer.innerHTML = items.map(item => `
            <div class="menu-item-card" data-item-id="${item.id}">
                <img src="${item.image_url || '../static/images/placeholder.jpg'}" alt="${item.name}">
                <div class="item-details">
                    <h4>${item.name}</h4>
                    <p>${item.description || ''}</p>
                    <div class="item-meta">
                        <span class="price">${formatMoney(item.price)}₸</span>
                        <span class="prep-time">${item.preparation_time} мин</span>
                    </div>
                    <div class="item-actions">
                        <button onclick="editMenuItem(${item.id})">✏️</button>
                        <button onclick="deleteMenuItem(${item.id})">🗑️</button>
                    </div>
                </div>
            </div>
        `).join('');
    }

    // Update category select in add/edit item forms
    const categorySelects = document.querySelectorAll('select[name="itemCategory"]');
    categorySelects.forEach(select => {
        select.innerHTML = `
            <option value="">Выберите категорию*</option>
            ${categories.map(category => `
                <option value="${category.id}">${category.name}</option>
            `).join('')}
        `;
    });
}

async function addCategory(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    
    try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/menu/categories', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                name: formData.get('categoryName'),
                description: formData.get('categoryDescription')
            })
        });

        if (!response.ok) {
            throw new Error('Failed to add category');
        }

        closeModal('addCategoryModal');
        form.reset();
        await loadMenuData();
        showSuccess('Категория успешно добавлена');
    } catch (error) {
        console.error('Error adding category:', error);
        showError('Ошибка при добавлении категории');
    }
}

async function editCategory(categoryId) {
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/menu/categories/${categoryId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load category data');
        }

        const category = await response.json();
        
        // Fill the edit form
        const form = document.getElementById('editCategoryForm');
        form.querySelector('[name="categoryId"]').value = category.id;
        form.querySelector('[name="categoryName"]').value = category.name;
        form.querySelector('[name="categoryDescription"]').value = category.description || '';
        
        showModal('editCategoryModal');
    } catch (error) {
        console.error('Error loading category:', error);
        showError('Ошибка при загрузке категории');
    }
}

async function updateCategory(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    const categoryId = formData.get('categoryId');
    
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/menu/categories/${categoryId}`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                name: formData.get('categoryName'),
                description: formData.get('categoryDescription')
            })
        });

        if (!response.ok) {
            throw new Error('Failed to update category');
        }

        closeModal('editCategoryModal');
        await loadMenuData();
        showSuccess('Категория успешно обновлена');
    } catch (error) {
        console.error('Error updating category:', error);
        showError('Ошибка при обновлении категории');
    }
}

async function deleteCategory(categoryId) {
    if (!confirm('Вы уверены, что хотите удалить эту категорию?')) {
        return;
    }
    
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/menu/categories/${categoryId}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to delete category');
        }

        await loadMenuData();
        showSuccess('Категория успешно удалена');
    } catch (error) {
        console.error('Error deleting category:', error);
        showError('Ошибка при удалении категории');
    }
}

async function addMenuItem(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    
    try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/menu/items', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                category_id: formData.get('itemCategory'),
                name: formData.get('itemName'),
                description: formData.get('itemDescription'),
                price: parseFloat(formData.get('itemPrice')),
                preparation_time: parseInt(formData.get('itemPrepTime')),
                calories: parseInt(formData.get('itemCalories')) || null,
                allergens: formData.get('itemAllergens') ? 
                    formData.get('itemAllergens').split(',').map(a => a.trim()) : []
            })
        });

        if (!response.ok) {
            throw new Error('Failed to add menu item');
        }

        // Handle image upload if present
        const imageFile = formData.get('itemImage');
        if (imageFile && imageFile.size > 0) {
            const itemId = (await response.json()).id;
            await uploadMenuItemImage(itemId, imageFile);
        }

        closeModal('addMenuItemModal');
        form.reset();
        await loadMenuData();
        showSuccess('Блюдо успешно добавлено');
    } catch (error) {
        console.error('Error adding menu item:', error);
        showError('Ошибка при добавлении блюда');
    }
}

async function editMenuItem(itemId) {
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/menu/items/${itemId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to load menu item data');
        }

        const item = await response.json();
        
        // Fill the edit form
        const form = document.getElementById('editMenuItemForm');
        form.querySelector('[name="itemId"]').value = item.id;
        form.querySelector('[name="itemCategory"]').value = item.category_id;
        form.querySelector('[name="itemName"]').value = item.name;
        form.querySelector('[name="itemDescription"]').value = item.description || '';
        form.querySelector('[name="itemPrice"]').value = item.price;
        form.querySelector('[name="itemPrepTime"]').value = item.preparation_time;
        form.querySelector('[name="itemCalories"]').value = item.calories || '';
        form.querySelector('[name="itemAllergens"]').value = item.allergens ? item.allergens.join(', ') : '';
        
        showModal('editMenuItemModal');
    } catch (error) {
        console.error('Error loading menu item:', error);
        showError('Ошибка при загрузке блюда');
    }
}

async function updateMenuItem(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    const itemId = formData.get('itemId');
    
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/menu/items/${itemId}`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                category_id: formData.get('itemCategory'),
                name: formData.get('itemName'),
                description: formData.get('itemDescription'),
                price: parseFloat(formData.get('itemPrice')),
                preparation_time: parseInt(formData.get('itemPrepTime')),
                calories: parseInt(formData.get('itemCalories')) || null,
                allergens: formData.get('itemAllergens') ? 
                    formData.get('itemAllergens').split(',').map(a => a.trim()) : []
            })
        });

        if (!response.ok) {
            throw new Error('Failed to update menu item');
        }

        // Handle image upload if present
        const imageFile = formData.get('itemImage');
        if (imageFile && imageFile.size > 0) {
            await uploadMenuItemImage(itemId, imageFile);
        }

        closeModal('editMenuItemModal');
        await loadMenuData();
        showSuccess('Блюдо успешно обновлено');
    } catch (error) {
        console.error('Error updating menu item:', error);
        showError('Ошибка при обновлении блюда');
    }
}

async function deleteMenuItem(itemId) {
    if (!confirm('Вы уверены, что хотите удалить это блюдо?')) {
        return;
    }
    
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/menu/items/${itemId}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            throw new Error('Failed to delete menu item');
        }

        await loadMenuData();
        showSuccess('Блюдо успешно удалено');
    } catch (error) {
        console.error('Error deleting menu item:', error);
        showError('Ошибка при удалении блюда');
    }
}

async function uploadMenuItemImage(itemId, file) {
    const formData = new FormData();
    formData.append('image', file);
    
    const token = localStorage.getItem('token');
    const response = await fetch(`/api/menu/items/${itemId}/image`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData
    });

    if (!response.ok) {
        throw new Error('Failed to upload image');
    }
}

// Utility functions for notifications
function showSuccess(message) {
    // You can implement a proper notification system here
    alert(message);
}

function showError(message) {
    // You can implement a proper notification system here
    alert(message);
}

// Initialize menu when the page loads
document.addEventListener('DOMContentLoaded', function() {
    // ... existing code ...
    
    // Load menu data if we're on the menu section
    if (window.location.pathname === '/manager/menu') {
        loadMenuData();
    }

    if (window.location.pathname === '/manager/inventory') {
        loadInventoryData();
        loadSuppliersData();
        loadRequestsData();
    }
});

async function populateSupplierSelect() {
    const token = localStorage.getItem('token');
    const response = await fetch('/api/manager/suppliers', {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    if (!response.ok) return;
    const data = await response.json();
    const suppliers = data.suppliers || [];
    const select = document.getElementById('requestSupplierSelect');
    if (select) {
        select.innerHTML = '<option value="">Поставщик*</option>' +
            suppliers.map(s => `<option value="${s.id}">${s.name}</option>`).join('');
    }
}

async function populateRequestItemsSelect() {
    const token = localStorage.getItem('token');
    const response = await fetch('/api/manager/inventory', {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    if (!response.ok) return;
    const data = await response.json();
    const items = data.items || [];
    const select = document.getElementById('requestItemsSelect');
    if (select) {
        select.innerHTML = items.map(item => 
            `<option value="${item.id}">${item.name}</option>`
        ).join('');
    }
}

// --- Request Items Dynamic Logic ---
let requestItems = [];

document.getElementById('addRequestItemBtn').addEventListener('click', async function() {
    // Показываем форму выбора товара
    const form = document.getElementById('addRequestItemForm');
    form.style.display = '';
    // Заполняем select товарами
    const token = localStorage.getItem('token');
    const resp = await fetch('/api/manager/inventory', {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    const data = await resp.json();
    const items = data.items || [];
    const select = document.getElementById('requestItemSelect');
    select.innerHTML = items.map(item => `<option value="${item.id}" data-unit="${item.unit}">${item.name}</option>`).join('');
    // Заполняем select единиц
    const unitSelect = document.getElementById('requestItemUnit');
    if (items.length > 0) {
        unitSelect.innerHTML = `<option value="${items[0].unit}">${items[0].unit}</option>`;
    }
    // При смене товара — меняем единицу
    select.onchange = function() {
        const selected = select.options[select.selectedIndex];
        unitSelect.innerHTML = `<option value="${selected.dataset.unit}">${selected.dataset.unit}</option>`;
    };
});

document.getElementById('confirmAddRequestItemBtn').addEventListener('click', function() {
    const select = document.getElementById('requestItemSelect');
    const qty = document.getElementById('requestItemQty').value;
    const unit = document.getElementById('requestItemUnit').value;
    if (!select.value || !qty) return;
    // Проверка на дубли
    if (requestItems.some(i => i.id === select.value)) {
        alert('Этот товар уже добавлен');
        return;
    }
    requestItems.push({
        id: select.value,
        name: select.options[select.selectedIndex].text,
        qty: qty,
        unit: unit
    });
    renderRequestItemsList();
    document.getElementById('addRequestItemForm').style.display = 'none';
    document.getElementById('requestItemQty').value = '';
});

document.getElementById('cancelAddRequestItemBtn').addEventListener('click', function() {
    document.getElementById('addRequestItemForm').style.display = 'none';
});

function renderRequestItemsList() {
    const list = document.getElementById('requestItemsList');
    list.innerHTML = requestItems.map((item, idx) =>
        `<div style="margin-bottom:4px;">
            ${item.name} <b>${item.qty} ${item.unit}</b>
            <button type="button" onclick="removeRequestItem(${idx})" style="margin-left:8px;">×</button>
        </div>`
    ).join('');
}
window.removeRequestItem = function(idx) {
    requestItems.splice(idx, 1);
    renderRequestItemsList();
};

// --- При отправке формы заявки ---
document.getElementById('addRequestForm').addEventListener('submit', function(e) {
    if (requestItems.length === 0) {
        alert('Добавьте хотя бы один товар');
        e.preventDefault();
        return false;
    }
    // Добавляем товары в скрытое поле для отправки
    this.requestItemsData.value = JSON.stringify(requestItems);
}); 