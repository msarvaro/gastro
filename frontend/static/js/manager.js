document.addEventListener('DOMContentLoaded', async function() {

    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');

    console.log(`Auth check - Token: ${token ? 'exists' : 'missing'}, Role: ${role}`);

    if (!token || (role !== 'manager' && role !== 'admin')) {
        console.error("Authentication failed - missing token or invalid role");
        window.location.href = '/';
        return;
    }

    // Initialize sidebar state
    const sidebar = document.getElementById('sidebar');
    const mainContent = document.querySelector('.main-content');
    const state = localStorage.getItem('sidebarState') || 'open';

    if (sidebar && mainContent) {
        // Set initial state without transitions 
        if (state === 'closed') {
            sidebar.classList.add('closed');
            mainContent.style.marginLeft = '88px';
        } else {
            sidebar.classList.remove('closed');
            mainContent.style.marginLeft = '279px';
        }

    }

    // Инициализация обработчиков событий
    setupEventListeners();

    // Проверяем токен через API
    try {
        const token = localStorage.getItem('token'); // Ensure token is fetched here if not already available
        const resp = await fetch('/api/manager/dashboard', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (!resp.ok) {
            console.error('Auth check failed (dashboard API call):', resp.status, await resp.text().catch(() => 'Could not get error text')); 
            localStorage.removeItem('token');
            localStorage.removeItem('role');
            document.cookie = 'auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
            alert("DEBUG: Redirecting because /api/manager/dashboard call failed. Status: " + resp.status); 
            window.location.href = '/';
            return;
        }

        console.log("manager.js: /api/manager/dashboard call OK. Calling loadDashboardData...");
        await loadDashboardData(); 
        console.log("manager.js: loadDashboardData completed without throwing to outer catch.");
        
        const currentPath = window.location.pathname;
        const sections = {
            '/manager': 'main',
            '/manager/inventory': 'inventory',
            '/manager/menu': 'menu',
            '/manager/staff': 'staff',
        };
        
        const activeSection = sections[currentPath] || 'main';
        showSection(activeSection);
        
        document.querySelectorAll('.sidebar nav ul li').forEach(li => {
            const route = li.getAttribute('data-route');
            li.classList.toggle('active', route === currentPath);
        });
        
        // Специальная обработка для разделов при прямом переходе/обновлении страницы
        if (activeSection === 'inventory') {
            showInventoryTab('stock');
        }
    } catch (e) { 
        console.error('manager.js: OUTER CATCH BLOCK triggered. Error:', e.message, e.stack);
        alert(`DEBUG: OUTER CATCH in manager.js. Error: ${e.message}\n\nStack: ${e.stack}`);
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

    // Преобразуем массив объектов в массив строк
    const itemsAsStrings = items.map(item => `${item.name} ${item.qty} ${item.unit}`);

    const newRequest = {
        branch: form.requestBranch.value,
        supplier_id: parseInt(form.requestSupplier.value, 10),
        items: itemsAsStrings, // теперь массив строк
        priority: form.requestPriority.value,
        comment: form.requestComment.value,
        status: 'pending',
        created_at: new Date().toISOString()
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
        
        const response = await fetch('/api/manager/orders/history', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (!response.ok) {
            const errorText = await response.text();
            console.error("loadDashboardData: /api/manager/orders/history failed. Status:", response.status, "Response text:", errorText);
            throw new Error(`Failed to load order history. Status: ${response.status}. Details: ${errorText}`);
        }

        const data = await response.json();

        // Проверяем структуру ответа и получаем массив заказов
        const orders = Array.isArray(data) ? data : (data.orders || []);

        const completedOrders = orders.filter(order => order.status === 'completed');
        
        const totalRevenue = completedOrders.reduce((sum, order) => sum + (order.total || 0), 0);
        const visitorCount = completedOrders.length;
        const averageCheck = visitorCount > 0 ? totalRevenue / visitorCount : 0;

        const cards = document.querySelectorAll('#main-section .card .value');
        if (cards.length >= 3) {
            cards[0].textContent = `${formatMoney(totalRevenue)}₸`;
            cards[1].textContent = `${visitorCount}`;
            cards[2].textContent = `${formatMoney(Math.round(averageCheck))}₸`;
        } else {
            console.warn("loadDashboardData: Could not find all dashboard cards to update values.");
        }

        const today = new Date().toDateString();
        const yesterday = new Date(Date.now() - 86400000).toDateString();
        
        const todayOrders = completedOrders.filter(order => 
            order.completed_at && new Date(order.completed_at).toDateString() === today);
        const yesterdayOrders = completedOrders.filter(order => 
            order.completed_at && new Date(order.completed_at).toDateString() === yesterday);

        const todayRevenue = todayOrders.reduce((sum, order) => sum + (order.total || 0), 0);
        const yesterdayRevenue = yesterdayOrders.reduce((sum, order) => sum + (order.total || 0), 0);
        
        let revenueChange = 0;
        if (yesterdayRevenue !== 0) {
            revenueChange = ((todayRevenue - yesterdayRevenue) / yesterdayRevenue * 100);
        }

        let visitorChange = 0;
        if (yesterdayOrders.length !== 0) {
            visitorChange = ((todayOrders.length - yesterdayOrders.length) / yesterdayOrders.length * 100);
        }
        
        const todayAvgCheck = todayOrders.length ? todayRevenue / todayOrders.length : 0;
        const yesterdayAvgCheck = yesterdayOrders.length ? yesterdayRevenue / yesterdayOrders.length : 0;
        let avgCheckChange = 0;
        if (yesterdayAvgCheck !== 0) {
            avgCheckChange = ((todayAvgCheck - yesterdayAvgCheck) / yesterdayAvgCheck * 100);
        }

        const indicators = document.querySelectorAll('#main-section .card .desc span');
        if (indicators.length >= 3) {
            updateComparisonIndicator(indicators[0], revenueChange);
            updateComparisonIndicator(indicators[1], visitorChange);
            updateComparisonIndicator(indicators[2], avgCheckChange);
        } else {
            console.warn("loadDashboardData: Could not find all dashboard indicator spans to update.");
        }
    } catch (error) {
        console.error('loadDashboardData: CRITICAL ERROR caught inside loadDashboardData:', error.message, error.stack);
        throw error; 
    }
}

function updateComparisonIndicator(indicatorElement, change) {
    if (!indicatorElement) return;
    
    const roundedChange = Math.round(change);
    let color = '#5D7285';
    let sign = '';

    if (roundedChange > 0) {
        color = '#006FFD';
        sign = '+';
    } else if (roundedChange < 0) {
        // Negative change, color already set to default, sign is handled by number itself
    } else {
        // Default color and no sign is fine
    }
    
    indicatorElement.textContent = `${sign}${roundedChange}% от вчера`;
    indicatorElement.style.color = color;
}

function formatMoney(amount) {
    return amount.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

function setupEventListeners() {
    // Sidebar click handler
    const sidebar = document.getElementById('sidebar');
    const mainContent = document.querySelector('.main-content');

    if (sidebar && mainContent) {
        // Open sidebar when clicking on it (if closed)
        sidebar.addEventListener('click', function(e) {
            if (sidebar.classList.contains('closed')) {
                sidebar.classList.remove('closed');
                mainContent.style.marginLeft = '279px';
                localStorage.setItem('sidebarState', 'open');
            e.stopPropagation();
            }
        });

        // Close sidebar when clicking on main content (if open)
        mainContent.addEventListener('click', function() {
            if (!sidebar.classList.contains('closed')) {
                sidebar.classList.add('closed');
                mainContent.style.marginLeft = '88px';
                localStorage.setItem('sidebarState', 'closed');
            }
        });
    }
    
    // Navigation menu
    document.querySelectorAll('.sidebar nav ul li').forEach(item => {
        item.addEventListener('click', function() {
            const section = this.getAttribute('data-section');
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

    // Inventory tab switching (JS only, no inline onclick)
    document.querySelectorAll('#inventory-section .tab-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const tab = this.getAttribute('data-tab');
            if (tab) {
                showInventoryTab(tab);
            }
        });
    });
    
    // Обработчики для персонала
    setupStaffEventListeners();
}

// Настройка обработчиков событий для секции персонала
function setupStaffEventListeners() {
    // Вкладки в секции персонала
    document.querySelectorAll('#staff-section .tab-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const tab = this.getAttribute('data-tab');
            showStaffTab(tab);
        });
    });
    
    // Кнопка добавления пользователя
    const addUserBtn = document.getElementById('addUserBtn');
    if (addUserBtn) {
        addUserBtn.addEventListener('click', showAddUserModal);
    }
    
    // Форма добавления пользователя
    const addUserForm = document.getElementById('addUserForm');
    if (addUserForm) {
        addUserForm.addEventListener('submit', (e) => {
            e.preventDefault();
            const formData = new FormData(addUserForm);
            
            const userData = {
                username: formData.get('username'),
                name: formData.get('name'),
                email: formData.get('email'),
                password: formData.get('password'),
                role: formData.get('role'),
                status: 'active'
            };

            fetch(getUsersApiEndpoint(), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify(userData)
            })
            .then(response => {
                if (response.ok) {
                    closeModal('addUserModal');
                    loadUsers();
                    showSuccess('Пользователь успешно создан');
                } else {
                    return response.text().then(text => {
                        if (text.includes('idx_users_email')) {
                            showError('Пользователь с таким email уже существует');
                        } else {
                            showError('Ошибка при создании пользователя');
                        }
                    });
                }
            })
            .catch(error => {
                console.error('Error adding user:', error);
                showError('Ошибка при создании пользователя');
            });
        });
    }
    
    // Форма редактирования пользователя
    const editUserForm = document.getElementById('editUserForm');
    if (editUserForm) {
        editUserForm.addEventListener('submit', (e) => {
            e.preventDefault();
            const userId = editUserForm.getAttribute('data-user-id');
            if (!userId) return;
            
            const formData = new FormData(editUserForm);
            const userData = {
                username: formData.get('username'),
                name: formData.get('name'),
                role: formData.get('role'),
                status: formData.get('status')
            };

            fetch(`${getUsersApiEndpoint()}/${userId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify(userData)
            })
            .then(response => {
                if (response.ok) {
                    closeModal('editUserModal');
                    loadUsers();
                    showSuccess('Пользователь успешно обновлен');
                } else {
                    showError('Ошибка при обновлении пользователя');
                }
            })
            .catch(error => {
                console.error('Error updating user:', error);
                showError('Ошибка при обновлении пользователя');
            });
        });
    }
    
    // Кнопка добавления смены
    const addShiftBtn = document.getElementById('addShiftBtn');
    if (addShiftBtn) {
        addShiftBtn.addEventListener('click', showAddShiftModal);
    }
    
    // Форма смены
    const shiftForm = document.getElementById('shiftForm');
    if (shiftForm) {
        shiftForm.addEventListener('submit', saveShift);
    }
    
    // Кнопка отмены в форме смены
    const cancelShiftBtn = document.getElementById('cancelShiftBtn');
    if (cancelShiftBtn) {
        cancelShiftBtn.addEventListener('click', () => closeModal('shiftModal'));
    }
    
    // Фильтры пользователей
    const userSearch = document.getElementById('userSearch');
    if (userSearch) {
        userSearch.addEventListener('input', debounce(() => loadUsers(), 300));
    }
    
    const userRoleFilter = document.getElementById('userRoleFilter');
    if (userRoleFilter) {
        userRoleFilter.addEventListener('change', () => loadUsers());
    }
    
    const userStatusFilter = document.getElementById('userStatusFilter');
    if (userStatusFilter) {
        userStatusFilter.addEventListener('change', () => loadUsers());
    }
    
    // Фильтры смен
    const shiftSearch = document.getElementById('shiftSearch');
    if (shiftSearch) {
        shiftSearch.addEventListener('input', debounce(() => loadShifts(), 300));
    }
    
    const shiftStatusFilter = document.getElementById('shiftStatusFilter');
    if (shiftStatusFilter) {
        shiftStatusFilter.addEventListener('change', () => loadShifts());
    }
    
    // Пагинация для пользователей
    const userPagination = document.querySelector('#staff-users-tab .pagination');
    if (userPagination) {
        const prevBtn = userPagination.querySelector('.prev');
        const nextBtn = userPagination.querySelector('.next');
        
        if (prevBtn) {
            prevBtn.addEventListener('click', () => {
                const currentPage = parseInt(localStorage.getItem('userPage') || '1');
                if (currentPage > 1) {
                    localStorage.setItem('userPage', (currentPage - 1).toString());
                    loadUsers();
                }
            });
        }
        
        if (nextBtn) {
            nextBtn.addEventListener('click', () => {
                const currentPage = parseInt(localStorage.getItem('userPage') || '1');
                localStorage.setItem('userPage', (currentPage + 1).toString());
                loadUsers();
            });
        }
    }
    
    // Пагинация для смен
    const shiftPagination = document.querySelector('#staff-shifts-tab .pagination');
    if (shiftPagination) {
        const prevBtn = shiftPagination.querySelector('.prev');
        const nextBtn = shiftPagination.querySelector('.next');
        
        if (prevBtn) {
            prevBtn.addEventListener('click', () => {
                const currentPage = parseInt(localStorage.getItem('shiftPage') || '1');
                if (currentPage > 1) {
                    localStorage.setItem('shiftPage', (currentPage - 1).toString());
                    loadShifts();
                }
            });
        }
        
        if (nextBtn) {
            nextBtn.addEventListener('click', () => {
                const currentPage = parseInt(localStorage.getItem('shiftPage') || '1');
                localStorage.setItem('shiftPage', (currentPage + 1).toString());
                loadShifts();
            });
        }
    }
}

// Section switching
function showSection(sectionName) { 
    const knownSectionNames = ['main', 'menu', 'inventory', 'staff']; 

    knownSectionNames.forEach(s_name => {
        const el = document.getElementById(s_name + '-section');
        if (el) {
            el.style.display = (s_name === sectionName) ? 'block' : 'none';
        } else {
            console.warn(`showSection: Element with ID '${s_name}-section' not found.`);
        }
    });

    // Загружаем данные для соответствующего раздела
    if (sectionName === 'staff') {
        // По умолчанию загружаем данные пользователей (первая вкладка)
        console.log('Автоматически загружаем данные пользователей при показе раздела персонала');
        
        // Определяем активную вкладку или используем "users" по умолчанию
        const activeTab = document.querySelector('#staff-section .tab-btn.active');
        const tabName = activeTab ? activeTab.getAttribute('data-tab') : 'users';
        
        // Загружаем данные в зависимости от активной вкладки
        showStaffTab(tabName);
    }

    // The active menu item highlighting is handled in the DOMContentLoaded scope
    // based on currentPath, so it's removed from here to avoid the ReferenceError.
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
                        <td>${item.min_quantity} ${item.unit}</td>
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
    if (item.quantity < item.min_quantity/2) return 'critical';
    if (item.quantity < item.min_quantity) return 'low';
    return 'ok';
}

function getStatusText(item) {
    if (item.quantity < item.min_quantity/2) return 'Критично';
    if (item.quantity < item.min_quantity) return 'Низкий';
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
                        <td>${formatDate(request.created_at)}</td>
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
    if (window.location.pathname === '/manager/orders') {
        loadOrdersData();
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

// Показать вкладку в секции персонала
function showStaffTab(tab) {
    console.log(`showStaffTab: Показываем вкладку ${tab}`);
    
    // Если не указана вкладка или указана неверно, используем "users" по умолчанию
    if (!tab || !['users', 'shifts'].includes(tab)) {
        console.log(`showStaffTab: Указана неверная вкладка ${tab}, используем users по умолчанию`);
        tab = 'users';
    }
    
    // Скрываем все вкладки
    document.querySelectorAll('#staff-section .tab-content').forEach(el => {
        el.style.display = 'none';
    });

    // Убираем активное состояние у всех кнопок
    document.querySelectorAll('#staff-section .tab-btn').forEach(el => {
        el.classList.remove('active');
    });

    // Показываем выбранную вкладку
    const tabContent = document.getElementById(`staff-${tab}-tab`);
    if (tabContent) {
        tabContent.style.display = 'block';
    } else {
        console.error(`showStaffTab: Элемент с ID staff-${tab}-tab не найден`);
    }

    // Добавляем активное состояние кнопке
    const tabBtn = document.querySelector(`#staff-section .tab-btn[data-tab="${tab}"]`);
    if (tabBtn) {
        tabBtn.classList.add('active');
    } else {
        console.error(`showStaffTab: Кнопка с атрибутом data-tab="${tab}" не найдена`);
    }

    // Загружаем данные в зависимости от выбранной вкладки
    if (tab === 'users') {
        loadUsers();
    } else if (tab === 'shifts') {
        loadShifts();
    }
}

// Определяет API-эндпоинт для пользователей
function getUsersApiEndpoint() {
    // Всегда используем эндпоинт менеджера независимо от роли
    return '/api/manager/users';
}

function getShiftsApiEndpoint() {
    // Всегда используем эндпоинт менеджера независимо от роли
    return '/api/manager/shifts';
}

// Создаем демо-данные для пользователей, если API недоступен
function populateDummyUsers(tbody) {
    const dummyUsers = [
        {
            id: 'dummy1',
            username: 'ivanov',
            name: 'Иванов Иван',
            role: 'manager',
            status: 'active',
            last_active: '2023-06-15T10:30:00',
            created_at: '2023-01-10T08:00:00'
        },
        {
            id: 'dummy2',
            username: 'petrov',
            name: 'Петров Петр',
            role: 'waiter',
            status: 'active',
            last_active: '2023-06-14T18:45:00',
            created_at: '2023-02-15T09:30:00'
        },
        {
            id: 'dummy3',
            username: 'sidorov',
            name: 'Сидоров Сидор',
            role: 'cook',
            status: 'inactive',
            last_active: '2023-05-20T14:20:00',
            created_at: '2023-03-05T10:15:00'
        }
    ];

    tbody.innerHTML = '';
    
    dummyUsers.forEach(user => {
        // Преобразуем даты в правильный формат
        const formattedLastActive = formatUserDate(user.last_active);
        const formattedCreatedAt = formatUserDate(user.created_at);
        
        const tr = document.createElement('tr');
        tr.setAttribute('data-user-id', user.id);
        tr.innerHTML = `
            <td>${user.username || ''}</td>
            <td>${user.name || ''}</td>
            <td data-role="${user.role || ''}">${translateRole(user.role || '')}</td>
            <td><span class="status-badge ${user.status || ''}">${translateStatus(user.status || '')}</span></td>
            <td>${formattedLastActive}</td>
            <td>${formattedCreatedAt}</td>
            <td class="actions">
                <button onclick="editUser('${user.id}')" class="edit-btn" title="Редактировать">
                    <img src="../static/images/edit.svg" alt="Редактировать" class="icon">
                </button>
                <button onclick="deleteUser('${user.id}')" class="delete-btn" title="Удалить">
                    <img src="../static/images/delete.svg" alt="Удалить" class="icon">
                </button>
            </td>
        `;
        tbody.appendChild(tr);
    });
    
    // Обновляем счетчики на основе демо-данных
    updateUserCount(dummyUsers.length);
    
    const activeUsers = dummyUsers.filter(user => user.status === 'active').length;
    const newUsers = dummyUsers.filter(user => {
        if (!user.created_at) return false;
        const oneWeekAgo = new Date();
        oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);
        const createdDate = new Date(user.created_at);
        return createdDate >= oneWeekAgo;
    }).length;
    
    const cardValues = document.querySelectorAll('#staff-users-tab .card .value');
    if (cardValues.length >= 3) {
        cardValues[1].textContent = activeUsers;
        cardValues[2].textContent = newUsers;
    }
}

// Загрузка пользователей
async function loadUsers() {
    try {
        // Очищаем таблицу и показываем индикатор загрузки
        const tbody = document.querySelector('#users-table tbody');
        if (!tbody) return;
        
        tbody.innerHTML = '<tr><td colspan="7" class="loading">Загрузка данных...</td></tr>';
        
        // Получаем фильтры, если они есть
        const userSearch = document.getElementById('userSearch');
        const userRoleFilter = document.getElementById('userRoleFilter');
        const userStatusFilter = document.getElementById('userStatusFilter');
        const page = parseInt(localStorage.getItem('userPage') || '1');
        
        let queryParams = '?';
        
        if (userSearch && userSearch.value) {
            queryParams += `search=${encodeURIComponent(userSearch.value)}&`;
        }
        
        if (userRoleFilter && userRoleFilter.value) {
            queryParams += `role=${encodeURIComponent(userRoleFilter.value)}&`;
        }
        
        if (userStatusFilter && userStatusFilter.value) {
            queryParams += `status=${encodeURIComponent(userStatusFilter.value)}&`;
        }
        
        // Добавляем параметр страницы
        queryParams += `page=${page}&limit=10&`;
        
        // Если адрес заканчивается на & или ?, удаляем последний символ
        if (queryParams.endsWith('&') || queryParams.endsWith('?')) {
            queryParams = queryParams.slice(0, -1);
        }
        
        // Если queryParams только ?, удаляем его
        if (queryParams === '?') {
            queryParams = '';
        }
        
        const endpoint = getUsersApiEndpoint() + queryParams;
        const token = localStorage.getItem('token');
        console.log(`loadUsers: Using token: ${token ? 'Token exists' : 'No token!'}`);
        
        const response = await fetch(endpoint, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            credentials: 'include' // Включаем куки для кросс-доменных запросов, если они есть
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        let users = Array.isArray(data) ? data : (data.users || []);

        tbody.innerHTML = '';

        if (users.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="7" class="no-results">Нет пользователей</td>
                </tr>`;
        } else {
            users.forEach(user => {
                // Преобразуем даты в правильный формат
                const formattedLastActive = formatUserDate(user.last_active);
                const formattedCreatedAt = formatUserDate(user.created_at);
                
                const tr = document.createElement('tr');
                tr.setAttribute('data-user-id', user.id);
                tr.innerHTML = `
                    <td>${user.username || ''}</td>
                    <td>${user.name || ''}</td>
                    <td data-role="${user.role || ''}">${translateRole(user.role || '')}</td>
                    <td><span class="status-badge ${user.status || ''}">${translateStatus(user.status || '')}</span></td>
                    <td>${formattedLastActive}</td>
                    <td>${formattedCreatedAt}</td>
                    <td class="actions">
                        <button onclick="editUser(${user.id})" class="edit-btn" title="Редактировать">
                            <img src="../static/images/edit.svg" alt="Редактировать" class="icon">
                        </button>
                        <button onclick="deleteUser(${user.id})" class="delete-btn" title="Удалить">
                            <img src="../static/images/delete.svg" alt="Удалить" class="icon">
                        </button>
                    </td>
                `;
                tbody.appendChild(tr);
            });
        }

        // Обновляем счетчики после загрузки пользователей
        updateUserCount(users.length);
        
        // Обновляем другие счетчики в карточках
        const activeUsers = users.filter(user => user.status === 'active').length;
        const newUsers = users.filter(user => {
            if (!user.created_at) return false;
            const oneWeekAgo = new Date();
            oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);
            const createdDate = new Date(user.created_at);
            return createdDate >= oneWeekAgo;
        }).length;
        
        const cardValues = document.querySelectorAll('#staff-users-tab .card .value');
        if (cardValues.length >= 3) {
            cardValues[1].textContent = activeUsers;
            cardValues[2].textContent = newUsers;
        }
        
        // Обновляем пагинацию, если она есть
        const pagination = document.querySelector('#staff-users-tab .pagination span');
        if (pagination && data.total !== undefined) {
            const page = data.page || 1;
            const pages = Math.ceil(data.total / (data.per_page || 10)) || 1;
            pagination.textContent = `${page} из ${pages} страниц`;
        }

    } catch (error) {
        const errorMessage = `Ошибка при загрузке пользователей: ${error.message}. Возможно, API-эндпоинт не существует.`;
        console.error('Error loading users:', error);
        const tbody = document.querySelector('#users-table tbody');
        if (tbody) {
            tbody.innerHTML = `<tr><td colspan="7" class="error">${errorMessage}</td></tr>`;
        }
        
        // Вместо всплывающего окна с ошибкой просто показываем сообщение в таблице
        console.log("Ошибка загрузки пользователей подавлена для улучшения UX");
        
        // Создаем демо-данные для отображения, чтобы интерфейс не был пустым
        populateDummyUsers(tbody);
    }
}

// Форматирование даты пользователя
function formatUserDate(dateString) {
    if (!dateString) return '—';
    
    try {
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            console.error('Невалидная дата для пользователя:', dateString);
            return '—';
        }
        
        return date.toLocaleString('ru-RU', {
            year: 'numeric', 
            month: 'long', 
            day: 'numeric',
            hour: '2-digit', 
            minute: '2-digit'
        });
    } catch (error) {
        console.error('Ошибка при форматировании даты:', error);
        return '—';
    }
}

// Обновление счетчика пользователей
function updateUserCount(count) {
    const userCountElement = document.getElementById('total-users');
    if (userCountElement) {
        userCountElement.textContent = count || 0;
    }
}

// Функции для перевода ролей и статусов
function translateRole(role) {
    const roles = {
        'admin': 'Администратор',
        'manager': 'Менеджер',
        'waiter': 'Официант',
        'cook': 'Повар',
        'client': 'Клиент'
    };
    return roles[role] || role;
}

function translateStatus(status) {
    const statuses = {
        'active': 'Активен',
        'inactive': 'Неактивен',
        'blocked': 'Заблокирован'
    };
    return statuses[status] || status;
}

// Добавление пользователя
function showAddUserModal() {
    const modal = document.getElementById('addUserModal');
    if (!modal) return;

    const form = modal.querySelector('form');
    if (!form) return;

    form.reset();
    showModal('addUserModal');

    form.onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        
        const userData = {
            username: formData.get('username'),
            name: formData.get('name'),
            email: formData.get('email'),
            password: formData.get('password'),
            role: formData.get('role'),
            status: 'active'
        };

        try {
            const endpoint = getUsersApiEndpoint();
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify(userData)
            });

            if (response.ok) {
                closeModal('addUserModal');
                loadUsers();
                showSuccess('Пользователь успешно создан');
            } else {
                const responseText = await response.text();
                if (responseText.includes('idx_users_email')) {
                    showError('Пользователь с таким email уже существует');
                } else {
                    showError('Ошибка при создании пользователя');
                }
            }
        } catch (error) {
            console.error('Error adding user:', error);
            showError('Ошибка при создании пользователя');
        }
    };
}

// Редактирование пользователя
function editUser(id) {
    const modal = document.getElementById('editUserModal');
    if (!modal) return;

    const form = modal.querySelector('form');
    if (!form) return;

    // Найдем пользователя в таблице
    const userRow = document.querySelector(`tr[data-user-id="${id}"]`);
    if (!userRow) return;

    // Заполним форму текущими данными
    form.elements['username'].value = userRow.cells[0].textContent;
    form.elements['name'].value = userRow.cells[1].textContent;
    form.elements['role'].value = userRow.cells[2].getAttribute('data-role');
    form.elements['status'].value = userRow.querySelector('.status-badge').classList.contains('active') ? 'active' : 'inactive';

    // Сохраним ID пользователя в атрибуте формы
    form.setAttribute('data-user-id', id);

    showModal('editUserModal');

    form.onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const userData = {
            username: formData.get('username'),
            name: formData.get('name'),
            role: formData.get('role'),
            status: formData.get('status')
        };

        try {
            const endpoint = `${getUsersApiEndpoint()}/${id}`;
            const response = await fetch(endpoint, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify(userData)
            });

            if (response.ok) {
                closeModal('editUserModal');
                loadUsers();
                showSuccess('Пользователь успешно обновлен');
            } else {
                showError('Ошибка при обновлении пользователя');
            }
        } catch (error) {
            console.error('Error updating user:', error);
            showError('Ошибка при обновлении пользователя');
        }
    };
}

// Удаление пользователя
function deleteUser(id) {
    if (confirm('Вы уверены, что хотите удалить этого пользователя?')) {
        const endpoint = `${getUsersApiEndpoint()}/${id}`;
        fetch(endpoint, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => {
            if (response.ok) {
                loadUsers();
                showSuccess('Пользователь успешно удален');
            } else {
                showError('Ошибка при удалении пользователя');
            }
        })
        .catch(error => {
            console.error('Error deleting user:', error);
            showError('Ошибка при удалении пользователя');
        });
    }
}

// Загрузка смен
async function loadShifts() {
    try {
        console.log("loadShifts: Attempting to load shifts...");
        // Очищаем таблицу и показываем индикатор загрузки
        const tbody = document.getElementById('shifts-tbody');
        if (!tbody) {
            console.error("loadShifts: shifts-tbody element not found");
            return;
        }
        
        tbody.innerHTML = '<tr><td colspan="5" class="loading">Загрузка данных...</td></tr>';
        
        // Получаем фильтры, если они есть
        const shiftSearch = document.getElementById('shiftSearch');
        const shiftStatusFilter = document.getElementById('shiftStatusFilter');
        const page = parseInt(localStorage.getItem('shiftPage') || '1');
        
        let queryParams = '?';
        
        if (shiftSearch && shiftSearch.value) {
            queryParams += `search=${encodeURIComponent(shiftSearch.value)}&`;
        }
        
        if (shiftStatusFilter && shiftStatusFilter.value) {
            queryParams += `status=${encodeURIComponent(shiftStatusFilter.value)}&`;
        }
        
        // Добавляем параметр страницы
        queryParams += `page=${page}&limit=10&`;
        
        // Если адрес заканчивается на & или ?, удаляем последний символ
        if (queryParams.endsWith('&') || queryParams.endsWith('?')) {
            queryParams = queryParams.slice(0, -1);
        }
        
        // Если queryParams только ?, удаляем его
        if (queryParams === '?') {
            queryParams = '';
        }
        
        const endpoint = getShiftsApiEndpoint() + queryParams;
        const token = localStorage.getItem('token');
        console.log(`loadShifts: Using token: ${token ? 'Token exists' : 'No token!'}`);
        
        const response = await fetch(endpoint, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            credentials: 'include' // Включаем куки для кросс-доменных запросов, если они есть
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        // Обрабатываем как массив или объект с массивом shifts
        const shifts = Array.isArray(data) ? data : (data.shifts || []);
        
        tbody.innerHTML = '';

        if (shifts.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" class="no-results">Нет данных о сменах</td></tr>';
            return;
        }

        shifts.forEach(shift => {
            const date = formatShiftDate(shift.date);
            const startTime = formatShiftTime(shift.start_time);
            const endTime = formatShiftTime(shift.end_time);
            const timeRange = `${startTime} - ${endTime}`;
            const status = translateShiftStatus(shift.status || 'active');
            const notes = shift.notes || '';
            
            const row = document.createElement('tr');
            row.setAttribute('data-shift-id', shift.id);
            row.innerHTML = `
                <td>${date}</td>
                <td>${timeRange}</td>
                <td>${shift.manager_name || 'Не назначен'}</td>
                <td><span class="status-${shift.status || 'active'}">${status}</span></td>
                <td class="actions">
                    <button onclick="editShift(${shift.id})" class="edit-btn" title="Редактировать смену">
                        <img src="../static/images/edit.svg" alt="Редактировать" class="icon">
                    </button>
                    <button onclick="confirmDeleteShift(${shift.id})" class="delete-btn" title="Удалить смену">
                        <img src="../static/images/delete.svg" alt="Удалить" class="icon">
                    </button>
                </td>
            `;

            if (notes) {
                row.setAttribute('title', notes);
            }

            tbody.appendChild(row);
        });
        
        // Обновляем пагинацию, если она есть
        const pagination = document.querySelector('#staff-shifts-tab .pagination span');
        if (pagination && data.total !== undefined) {
            const page = data.page || 1;
            const pages = Math.ceil(data.total / (data.per_page || 10)) || 1;
            pagination.textContent = `${page} из ${pages} страниц`;
        }

    } catch (error) {
        const errorMessage = `Ошибка при загрузке смен: ${error.message}. Возможно, API-эндпоинт не существует.`;
        console.error('Error loading shifts:', error);
        const tbody = document.getElementById('shifts-tbody');
        if (tbody) {
            tbody.innerHTML = `<tr><td colspan="5" class="error">${errorMessage}</td></tr>`;
        }
        
        // Вместо всплывающего окна с ошибкой просто показываем сообщение в таблице
        console.log("Ошибка загрузки данных о сменах подавлена для улучшения UX");
    }
}

// Функция для форматирования даты смены
function formatShiftDate(dateString) {
    if (!dateString) return '';
    
    try {
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            return dateString;
        }
        
        return date.toLocaleDateString('ru-RU', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    } catch (error) {
        console.error('Ошибка при форматировании даты смены:', error);
        return dateString;
    }
}

// Функция для форматирования времени смены
function formatShiftTime(timeString) {
    if (!timeString) return '';
    
    // Если это уже формат HH:MM
    if (typeof timeString === 'string' && timeString.match(/^\d{1,2}:\d{2}$/)) {
        // Форматируем для корректного отображения
        const [hours, minutes] = timeString.split(':');
        return `${hours.padStart(2, '0')}:${minutes.padStart(2, '0')}`;
    }
    
    // Если время в формате даты-времени
    if (typeof timeString === 'string' && timeString.includes('T')) {
        const timePart = timeString.split('T')[1] || '00:00';
        return timePart.substring(0, 5); // HH:MM
    }
    
    return timeString;
}

// Функция для перевода статусов смен
function translateShiftStatus(status) {
    const translations = {
        'active': 'Активна',
        'pending': 'Ожидает',
        'completed': 'Завершена',
        'canceled': 'Отменена'
    };
    return translations[status] || 'Активна'; // По умолчанию "Активна"
}

// Заполнение выпадающего списка менеджеров
async function populateManagerDropdown(selectedManagerId = null) {
    const managers = await loadManagers();
    const select = document.getElementById('shift-manager');
    if (!select) return;
    
    select.innerHTML = '<option value="">Выберите менеджера</option>';
    
    managers.forEach(manager => {
        const displayName = manager.name || manager.username;
        select.innerHTML += `<option value="${manager.id}" ${manager.id === selectedManagerId ? 'selected' : ''}>${displayName}</option>`;
    });
}

// Заполнение чекбоксов сотрудников
async function populateEmployeeCheckboxes(selectedEmployeeIds = []) {
    const employees = await loadEmployees();
    const container = document.getElementById('shift-employees-container');
    if (!container) return;
    
    container.innerHTML = '';
    
    // Группируем сотрудников по ролям для лучшей организации
    const groups = {
        'waiter': {title: 'Официанты', employees: []},
        'cook': {title: 'Повара', employees: []}
    };
    
    // Set для отслеживания добавленных ID, чтобы избежать дублирования
    const addedEmployeeIds = new Set();
    
    employees.forEach(employee => {
        // Проверяем, не добавлен ли сотрудник уже и есть ли его роль в группах
        if (!addedEmployeeIds.has(employee.id) && groups[employee.role]) {
            addedEmployeeIds.add(employee.id);
            groups[employee.role].employees.push(employee);
        }
    });
    
    // Создаем группы с чекбоксами
    for (const role in groups) {
        if (groups[role].employees.length > 0) {
            const groupDiv = document.createElement('div');
            groupDiv.className = 'employee-group';
            
            const titleDiv = document.createElement('div');
            titleDiv.className = 'employee-group-title';
            titleDiv.textContent = groups[role].title;
            groupDiv.appendChild(titleDiv);
            
            // Сортируем сотрудников по имени для удобства
            groups[role].employees.sort((a, b) => {
                const nameA = (a.name || a.username || '').toLowerCase();
                const nameB = (b.name || b.username || '').toLowerCase();
                return nameA.localeCompare(nameB);
            });
            
            groups[role].employees.forEach(employee => {
                const itemDiv = document.createElement('div');
                itemDiv.className = 'employee-item';
                
                const checkbox = document.createElement('input');
                checkbox.type = 'checkbox';
                checkbox.id = `employee-${employee.id}`;
                checkbox.name = 'employee';
                checkbox.value = employee.id;
                checkbox.checked = selectedEmployeeIds.includes(employee.id);
                
                const label = document.createElement('label');
                label.htmlFor = `employee-${employee.id}`;
                label.textContent = employee.name || employee.username;
                
                itemDiv.appendChild(checkbox);
                itemDiv.appendChild(label);
                groupDiv.appendChild(itemDiv);
            });
            
            container.appendChild(groupDiv);
        }
    }
}

// Показ модального окна добавления смены
async function showAddShiftModal() {
    const modal = document.getElementById('shiftModal');
    if (!modal) return;
    
    const form = document.getElementById('shiftForm');
    if (!form) return;
    
    // Сбрасываем форму и устанавливаем сегодняшнюю дату
    form.reset();
    document.getElementById('shift-id').value = '';
    document.getElementById('shiftModalTitle').textContent = 'Добавить смену';
    
    // Устанавливаем текущую дату
    const today = new Date();
    const formattedDate = today.toISOString().split('T')[0];
    document.getElementById('shift-date').value = formattedDate;
    
    // Загружаем менеджеров и сотрудников
    await populateManagerDropdown();
    await populateEmployeeCheckboxes();
    
    showModal('shiftModal');
}

// Показ модального окна редактирования смены
async function editShift(shiftId) {
    try {
        const endpoint = `${getShiftsApiEndpoint()}/${shiftId}`;
        const response = await fetch(endpoint, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });
        
        if (!response.ok) {
            throw new Error('Не удалось загрузить данные смены');
        }
        
        const shift = await response.json();
        
        // Настраиваем форму редактирования
        document.getElementById('shiftModalTitle').textContent = 'Редактировать смену';
        document.getElementById('shift-id').value = shift.id;
        
        // Устанавливаем дату в формате YYYY-MM-DD
        const dateInput = document.getElementById('shift-date');
        dateInput.value = shift.date;
        
        // Очищаем формат времени к HH:MM
        const cleanTimeFormat = (timeStr) => {
            if (!timeStr) return '';
            
            // Если это уже формат HH:MM, возвращаем его
            if (typeof timeStr === 'string' && timeStr.match(/^\d{1,2}:\d{2}$/)) {
                return timeStr;
            }
            
            // Если это полный формат времени HH:MM:SS
            if (typeof timeStr === 'string' && timeStr.match(/^\d{1,2}:\d{2}:\d{2}$/)) {
                return timeStr.substring(0, 5);
            }
            
            // Для формата с датой (ISO или другой)
            if (typeof timeStr === 'string' && timeStr.includes('T')) {
                const timePart = timeStr.split('T')[1] || '00:00';
                return timePart.substring(0, 5);
            }
            
            return timeStr;
        };
        
        document.getElementById('shift-start-time').value = cleanTimeFormat(shift.start_time);
        document.getElementById('shift-end-time').value = cleanTimeFormat(shift.end_time);
        
        // Устанавливаем статус и примечания
        document.getElementById('shift-status').value = shift.status || 'active';
        document.getElementById('shift-notes').value = shift.notes || '';
        
        // Загружаем и устанавливаем менеджера
        await populateManagerDropdown(shift.manager_id);
        
        // Загружаем и отмечаем сотрудников
        const employeeIds = shift.employees ? shift.employees.map(emp => emp.id) : [];
        await populateEmployeeCheckboxes(employeeIds);
        
        showModal('shiftModal');
    } catch (error) {
        console.error('Error loading shift details:', error);
        showError('Ошибка при загрузке данных смены');
    }
}

// Сохранение смены (добавление или редактирование)
async function saveShift(e) {
    e.preventDefault();
    
    try {
        // Получаем ID смены (если редактирование) или пустую строку (если создание)
        const shiftId = document.getElementById('shift-id').value;
        
        // Собираем данные смены
        const date = document.getElementById('shift-date').value;
        const startTime = document.getElementById('shift-start-time').value;
        const endTime = document.getElementById('shift-end-time').value;
        const managerId = document.getElementById('shift-manager').value;
        const status = document.getElementById('shift-status').value;
        const notes = document.getElementById('shift-notes').value;
        
        // Собираем выбранных сотрудников
        const selectedEmployees = [];
        document.querySelectorAll('#shift-employees-container input[type="checkbox"]:checked').forEach(checkbox => {
            selectedEmployees.push(parseInt(checkbox.value, 10));
        });
        
        const shiftData = {
            date: date,
            start_time: startTime,
            end_time: endTime,
            manager_id: managerId || null,
            status: status,
            notes: notes,
            employees: selectedEmployees
        };
        
        // Определяем метод и URL в зависимости от того, создание или редактирование
        const method = shiftId ? 'PUT' : 'POST';
        const url = shiftId ? `${getShiftsApiEndpoint()}/${shiftId}` : getShiftsApiEndpoint();
        
        // Отправляем запрос
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify(shiftData)
        });
        
        // Обрабатываем результат
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
        }
        
        closeModal('shiftModal');
        loadShifts();
        showSuccess(`Смена успешно ${shiftId ? 'обновлена' : 'создана'}`);
    } catch (error) {
        console.error('Error saving shift:', error);
        showError(`Ошибка при сохранении смены: ${error.message}`);
    }
}

// Подтверждение удаления смены
async function confirmDeleteShift(shiftId) {
    if (confirm('Вы уверены, что хотите удалить эту смену?')) {
        try {
            const endpoint = `${getShiftsApiEndpoint()}/${shiftId}`;
            const response = await fetch(endpoint, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to delete shift');
            }
            
            loadShifts();
            showSuccess('Смена успешно удалена');
        } catch (error) {
            console.error('Error deleting shift:', error);
            showError('Ошибка при удалении смены');
        }
    }
}

// Загрузка менеджеров для выпадающего списка
async function loadManagers() {
    try {
        const endpoint = getUsersApiEndpoint() + '?role=manager';
        const response = await fetch(endpoint, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });

        if (response.ok) {
            const data = await response.json();
            const managers = Array.isArray(data) ? data : (data.users || []);
            // Фильтруем, чтобы получить только менеджеров и только активных
            return managers.filter(user => 
                user.role === 'manager' && (!user.status || user.status === 'active')
            );
        } else {
            console.error('Failed to load managers');
            return [];
        }
    } catch (error) {
        console.error('Error loading managers:', error);
        showError('Ошибка при загрузке списка менеджеров');
        return [];
    }
}

// Загрузка сотрудников для чекбоксов
async function loadEmployees() {
    try {
        // Загружаем только официантов и поваров
        const roleTypes = ['waiter', 'cook'];
        let allEmployeesList = [];
        
        // Временная карта для отслеживания уже добавленных сотрудников по ID
        const employeeMap = new Map();
        
        for (const role of roleTypes) {
            try {
                const endpoint = getUsersApiEndpoint() + `?role=${role}&status=active`;
                const response = await fetch(endpoint, {
                    method: 'GET',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    }
                });
    
                if (response.ok) {
                    const data = await response.json();
                    const employees = Array.isArray(data) ? data : (data.users || []);
                    
                    // Добавляем только уникальных сотрудников
                    employees.forEach(employee => {
                        if (!employeeMap.has(employee.id)) {
                            employeeMap.set(employee.id, employee);
                            allEmployeesList.push(employee);
                        }
                    });
                }
            } catch (e) {
                console.error(`Failed to load ${role}s:`, e);
            }
        }
        
        return allEmployeesList;
    } catch (error) {
        console.error('Failed to load employees:', error);
        showError('Ошибка при загрузке списка сотрудников');
        return [];
    }
}
