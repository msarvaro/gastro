document.addEventListener('DOMContentLoaded', function() {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    loadUsers();
    loadStats();

    // Находим элемент поиска
    const searchInput = document.querySelector('.search-input');
    console.log('Search Input found:', searchInput); // Проверка наличия элемента поиска
    
    if (searchInput) {
        // Добавляем обработчик события input
        searchInput.addEventListener('input', function() {
            console.log('Search event triggered'); // Проверка срабатывания события
            
            // Получаем значение поиска и приводим к нижнему регистру
            const searchText = this.value.toLowerCase();
            console.log('Search text:', searchText); // Проверка поискового запроса
            
            // Находим таблицу и строки
            const tbody = document.querySelector('.user-table tbody');
            const rows = tbody.querySelectorAll('tr:not(.no-results)');
            console.log('Found rows:', rows.length); // Проверка количества найденных строк
            
            let visibleRows = false;
            
            // Проходим по каждой строке
            rows.forEach(row => {
                // Собираем текст для поиска
                const cells = Array.from(row.cells).slice(0, -1);
                const rowText = cells.map(cell => {
                    const statusBadge = cell.querySelector('.status-badge');
                    return statusBadge ? statusBadge.textContent : cell.textContent;
                }).join(' ').toLowerCase();
                
                console.log('Row text:', rowText); // Проверка текста строки
                
                // Проверяем совпадение
                const isVisible = rowText.includes(searchText);
                console.log('Is visible:', isVisible); // Проверка видимости строки
                
                // Показываем или скрываем строку
                row.style.display = isVisible ? '' : 'none';
                if (isVisible) visibleRows = true;
            });

            // Управляем сообщением "не найдено"
            const noResultsRow = tbody.querySelector('.no-results');
            
            if (!visibleRows) {
                if (!noResultsRow) {
                    const tr = document.createElement('tr');
                    tr.className = 'no-results';
                    tr.innerHTML = `
                        <td colspan="6" class="no-results">
                            Ничего не найдено по вашему запросу
                        </td>
                    `;
                    tbody.appendChild(tr);
                }
            } else if (noResultsRow) {
                noResultsRow.remove();
            }

            // Обновляем счетчик после фильтрации
            updateUserCount();
        });

        // Очистка по Escape
        searchInput.addEventListener('keydown', function(e) {
            if (e.key === 'Escape') {
                this.value = '';
                this.dispatchEvent(new Event('input'));
            }
        });
    } else {
        console.error('Search input not found!'); // Ошибка если элемент не найден
    }

    // Фильтры
    const filterBtn = document.querySelector('.filter-btn');
    const filterDropdown = document.querySelector('.filter-dropdown');
    const filterSelects = document.querySelectorAll('.filter-select');

    // Функция фильтрации строк таблицы
    function filterRows(type, value) {
        const rows = document.querySelectorAll('.user-table tbody tr:not(.no-results)');
        let hasVisibleRows = false;
        
        rows.forEach(row => {
            if (value === 'all') {
                row.style.display = '';
                hasVisibleRows = true;
                return;
            }

            let cellValue;
            if (type === 'role') {
                // Получаем значение роли из data-атрибута или из текста
                const roleCell = row.cells[1];
                cellValue = roleCell.getAttribute('data-role') || roleCell.textContent.toLowerCase();
                
                // Сопоставляем значения
                const roleMatches = {
                    'manager': ['manager', 'менеджер'],
                    'waiter': ['waiter', 'официант'],
                    'cook': ['cook', 'повар']
                };
                
                if (roleMatches[value] && roleMatches[value].includes(cellValue.toLowerCase())) {
                    row.style.display = '';
                    hasVisibleRows = true;
                } else {
                    row.style.display = 'none';
                }
                return;
            } else if (type === 'status') {
                cellValue = row.querySelector('.status-badge').classList.contains('active') ? 'active' : 'inactive';
                
                if (cellValue === value) {
                    row.style.display = '';
                    hasVisibleRows = true;
                } else {
                    row.style.display = 'none';
                }
            }
        });

        // Показываем сообщение, только если нет видимых строк
        const tbody = document.querySelector('.user-table tbody');
        const noResults = document.querySelector('.no-results');

        if (!hasVisibleRows) {
            if (!noResults) {
                const tr = document.createElement('tr');
                tr.className = 'no-results';
                tr.innerHTML = `
                    <td colspan="6" class="no-results">
                        Ничего не найдено по выбранным фильтрам
                    </td>
                `;
                tbody.appendChild(tr);
            }
        } else if (noResults) {
            noResults.remove();
        }
    }

    // Открытие/закрытие основного дропдауна
    if (filterBtn && filterDropdown) {
        filterBtn.addEventListener('click', function(e) {
            e.stopPropagation();
            filterDropdown.classList.toggle('show');
            filterSelects.forEach(select => select.classList.remove('active'));
        });
    }

    // Обработка клика на селекты фильтров
    filterSelects.forEach(select => {
        select.addEventListener('click', function(e) {
            e.stopPropagation();
            filterSelects.forEach(s => {
                if (s !== select) s.classList.remove('active');
            });
            select.classList.toggle('active');
        });
    });

    // Обработка выбора опции
    document.querySelectorAll('.option').forEach(option => {
        option.addEventListener('click', function(e) {
            e.stopPropagation();
            const type = this.closest('.filter-item').querySelector('.filter-select').dataset.type;
            const value = this.dataset.value;
            
            // Обновляем визуальное состояние
            this.closest('.filter-options').querySelectorAll('.option').forEach(opt => {
                opt.classList.remove('selected');
            });
            this.classList.add('selected');
            
            // Закрываем дропдауны
            this.closest('.filter-item').querySelector('.filter-select').classList.remove('active');
            
            // Применяем фильтр
            filterRows(type, value);
        });
    });

    // Закрытие при клике вне фильтра
    document.addEventListener('click', function(e) {
        if (!filterDropdown.contains(e.target) && !filterBtn.contains(e.target)) {
            filterDropdown.classList.remove('show');
            filterSelects.forEach(select => select.classList.remove('active'));
        }
    });

    // Обработчик для кнопки добавления пользователя
    const addUserBtn = document.querySelector('.add-user-btn');
    if (addUserBtn) {
        addUserBtn.addEventListener('click', showAddUserModal);
    }

    // Закрытие модальных окон
    document.querySelectorAll('.modal .close-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            btn.closest('.modal').style.display = 'none';
        });
    });
});

async function loadUsers() {
    const token = localStorage.getItem('token');
    try {
        const response = await fetch('/api/admin/users', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        let users = Array.isArray(data) ? data : (data.users || []);
        
        const tbody = document.querySelector('.user-table tbody');
        if (!tbody) return;

        tbody.innerHTML = '';
        
        if (users.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="no-results">Нет пользователей</td>
                </tr>`;
        } else {
            users.forEach(user => {
                const tr = document.createElement('tr');
                tr.setAttribute('data-user-id', user.id);
                tr.innerHTML = `
                    <td>${user.username || ''}</td>
                    <td data-role="${user.role || ''}">${translateRole(user.role || '')}</td>
                    <td><span class="status-badge ${user.status || ''}">${translateStatus(user.status || '')}</span></td>
                    <td>${formatDate(user.last_active) || ''}</td>
                    <td>${formatDate(user.created_at) || ''}</td>
                    <td class="actions">
                        <button onclick="editUser(${user.id})" class="edit-btn" title="Редактировать">
                            <img src="/static/images/edit.svg" alt="Редактировать" class="icon">
                        </button>
                        <button onclick="deleteUser(${user.id})" class="delete-btn" title="Удалить">
                            <img src="/static/images/delete.svg" alt="Удалить" class="icon">
                        </button>
                    </td>
                `;
                tbody.appendChild(tr);
            });
        }

        // Обновляем счетчик после загрузки пользователей
        updateUserCount();

    } catch (error) {
        console.error('Error loading users:', error);
    }
}

function showAddUserModal() {
    const modal = document.getElementById('addUserModal');
    if (!modal) return;

    const form = modal.querySelector('form');
    if (!form) return;

    form.reset();
    modal.style.display = 'block';

    form.onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const userData = {
            username: formData.get('username'),
            email: formData.get('email'),
            password: formData.get('password'),
            role: formData.get('role'),
            status: 'active'
        };

        alert('submit!');

        console.log('Создание пользователя:', userData);

        try {
            const response = await fetch('/api/admin/users', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(userData)
            });

            const responseText = await response.text();
            console.log('Server response:', responseText);

            if (response.ok) {
                modal.style.display = 'none';
                loadUsers();
                showNotification('Пользователь успешно создан', 'success');
            } else {
                if (responseText.includes('idx_users_email')) {
                    showNotification('Пользователь с таким email уже существует', 'error');
                } else {
                    showNotification('Ошибка при создании пользователя', 'error');
                }
            }
        } catch (error) {
            console.error('Error adding user:', error);
            showNotification('Ошибка при создании пользователя', 'error');
        }
    };
}

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
    form.elements['role'].value = userRow.cells[1].textContent;
    form.elements['status'].value = userRow.querySelector('.status-badge').classList.contains('active') ? 'active' : 'inactive';

    modal.style.display = 'block';

    form.onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const userData = {
            username: formData.get('username'),
            role: formData.get('role'),
            status: formData.get('status')
        };

        try {
            const response = await fetch(`/api/admin/users/${id}`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(userData)
            });

            if (response.ok) {
                modal.style.display = 'none';
                loadUsers();
                showNotification('Пользователь успешно обновлен', 'success');
            } else {
                showNotification('Ошибка при обновлении пользователя', 'error');
            }
        } catch (error) {
            console.error('Error updating user:', error);
            showNotification('Ошибка при обновлении пользователя', 'error');
        }
    };
}

function deleteUser(id) {
    if (confirm('Вы уверены, что хотите удалить этого пользователя?')) {
        fetch(`/api/admin/users/${id}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => {
            if (response.ok) {
                loadUsers();
                showNotification('Пользователь успешно удален', 'success');
            } else {
                showNotification('Ошибка при удалении пользователя', 'error');
            }
        })
        .catch(error => {
            console.error('Error deleting user:', error);
            showNotification('Ошибка при удалении пользователя', 'error');
        });
    }
}

function showNotification(message, type) {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    document.body.appendChild(notification);

    setTimeout(() => {
        notification.classList.add('fade-out');
        setTimeout(() => notification.remove(), 500);
    }, 3000);
}

function getRoleBadgeClass(role) {
    const classes = {
        'waiter': 'role-waiter',
        'cook': 'role-cook',
        'manager': 'role-manager',
        'client': 'role-client'
    };
    return classes[role] || 'role-default';
}

function getStatusBadgeClass(status) {
    const classes = {
        'active': 'status-active',
        'inactive': 'status-inactive',
        'blocked': 'status-blocked'
    };
    return classes[status] || 'status-default';
}

function translateRole(role) {
    const translations = {
        'waiter': 'Официант',
        'cook': 'Повар',
        'manager': 'Менеджер',
        'client': 'Клиент'
    };
    return translations[role] || role;
}

function translateStatus(status) {
    const translations = {
        'active': 'Активен',
        'inactive': 'Неактивен',
        'blocked': 'Заблокирован'
    };
    return translations[status] || status;
}

function formatDate(dateString) {
    if (!dateString) return '';
    return new Date(dateString).toLocaleString('ru-RU', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
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

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    window.location.href = '/';
}

// Инициализация статистики
async function loadStats() {
    const token = localStorage.getItem('token');
    try {
        const response = await fetch('/api/admin/stats', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            const stats = await response.json();
            updateStats(stats);
        }
    } catch (error) {
        console.error('Error loading stats:', error);
    }
}

function updateStats(stats) {
    const statValues = document.querySelectorAll('.stat-value');
    statValues.forEach(element => {
        const statType = element.parentElement.querySelector('.stat-title').textContent;
        switch(statType) {
            case 'Активные пользователи':
                element.textContent = stats.active || 0;
                break;
            case 'Официанты':
                element.textContent = stats.waiters || 0;
                break;
            case 'Повара':
                element.textContent = stats.cooks || 0;
                break;
            case 'Клиенты':
                element.textContent = stats.clients || 0;
                break;
        }
    });
}

// Глобальные переменные для фильтров и поиска
let activeFilters = {
    role: 'all',
    status: 'all',
    search: ''
};

// Функция поиска и фильтрации пользователей
function filterUsers(users) {
    return users.filter(user => {
        // Проверка на соответствие фильтрам
        const roleMatch = activeFilters.role === 'all' || user.role === activeFilters.role;
        const statusMatch = activeFilters.status === 'all' || user.status === activeFilters.status;
        
        // Поиск по имени пользователя (без учета регистра)
        const searchMatch = !activeFilters.search || 
            user.username.toLowerCase().includes(activeFilters.search.toLowerCase());

        return roleMatch && statusMatch && searchMatch;
    });
}

// Функция обновления счетчика пользователей
function updateUserCount() {
    const totalUsers = document.querySelectorAll('.user-table tbody tr:not(.no-results)').length;
    const userCountElement = document.getElementById('total-users');
    if (userCountElement) {
        userCountElement.textContent = totalUsers;
    }
}