// Suppliers API
async function loadSuppliersData() {
    try {
        const response = await fetch('/api/admin/suppliers', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });
        if (!response.ok) throw new Error('Ошибка при загрузке поставщиков');
        const suppliers = await response.json();
        // Фильтрация
        const search = document.getElementById('supplierSearch').value.trim().toLowerCase();
        const category = document.getElementById('supplierCategoryFilter').value;
        let filteredSuppliers = suppliers;
        if (search) filteredSuppliers = suppliers.filter(s => s.name.toLowerCase().includes(search));
        if (category) filteredSuppliers = filteredSuppliers.filter(s => s.categories.includes(category));
        // Обновление таблицы
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
        updateRequestSupplierOptions(suppliers);
    } catch (error) {
        console.error('Ошибка загрузки поставщиков:', error);
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
        const response = await fetch('/api/admin/suppliers', {
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

function getSupplierStatusText(status) {
    const statusMap = {
        'active': 'Активный',
        'paused': 'На паузе',
        'archived': 'Архивный'
    };
    return statusMap[status] || status;
}

function updateRequestSupplierOptions(suppliers) {
    const select = document.getElementById('requestSupplierSelect');
    if (!select) return;
    select.innerHTML = '<option value="">Поставщик*</option>' + suppliers.map(s => `<option value="${s.name}">${s.name}</option>`).join('');
}

// Requests API
async function loadRequestsData() {
    try {
        const response = await fetch('/api/admin/requests', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });
        if (!response.ok) throw new Error('Ошибка при загрузке заявок');
        const requests = await response.json();
        // Фильтрация
        const search = document.getElementById('requestSearch').value.trim().toLowerCase();
        const status = document.getElementById('requestStatusFilter').value;
        const branch = document.getElementById('requestBranchFilter').value;
        let filteredRequests = requests;
        if (search) filteredRequests = filteredRequests.filter(r => r.items.join(', ').toLowerCase().includes(search));
        if (status) filteredRequests = filteredRequests.filter(r => r.status === status);
        if (branch) filteredRequests = filteredRequests.filter(r => r.branch === branch);
        // Обновление таблицы
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
        console.error('Ошибка загрузки заявок:', error);
    }
}

async function handleAddRequest(event) {
    event.preventDefault();
    const form = event.target;
    const selectedOptions = Array.from(form.requestItems.selectedOptions).map(opt => opt.value);
    const newRequest = {
        branch: form.requestBranch.value,
        supplier: form.requestSupplier.value,
        items: selectedOptions,
        priority: form.requestPriority.value,
        comment: form.requestComment.value,
        status: 'pending',
        createdAt: new Date().toISOString()
    };
    try {
        const response = await fetch('/api/admin/requests', {
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

function getRequestStatusText(status) {
    const statusMap = {
        'pending': 'Ожидает одобрения',
        'active': 'В обработке',
        'completed': 'Выполнено',
        'rejected': 'Отклонено'
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

document.addEventListener('DOMContentLoaded', function() {
    // Suppliers
    const addSupplierForm = document.getElementById('addSupplierForm');
    if (addSupplierForm) addSupplierForm.addEventListener('submit', handleAddSupplier);
    document.getElementById('supplierSearch').addEventListener('input', loadSuppliersData);
    document.getElementById('supplierCategoryFilter').addEventListener('change', loadSuppliersData);
    loadSuppliersData();
    // Requests
    const addRequestForm = document.getElementById('addRequestForm');
    if (addRequestForm) addRequestForm.addEventListener('submit', handleAddRequest);
    document.getElementById('requestSearch').addEventListener('input', loadRequestsData);
    document.getElementById('requestStatusFilter').addEventListener('change', loadRequestsData);
    document.getElementById('requestBranchFilter').addEventListener('change', loadRequestsData);
    loadRequestsData();
}); 