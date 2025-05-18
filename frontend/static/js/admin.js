document.addEventListener('DOMContentLoaded', function() {
    // Перенаправляем с админ-панели на панель менеджера
    window.location.href = '/manager';
});

document.addEventListener('DOMContentLoaded', function() {
    // Check for authentication either from cookie or localStorage (backward compatibility)
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    const hasAuthCookie = document.cookie.split(';').some(c => c.trim().startsWith('auth_token='));
    
    if (!token && !hasAuthCookie) {
        window.location.href = '/';
        return;
    }

    // Добавляем кнопку перехода в панель менеджера для администраторов
    if (role === 'admin') {
        const header = document.querySelector('.main-nav');
        if (header) {
            const managerLink = document.createElement('div');
            managerLink.className = 'nav-item';
            managerLink.innerHTML = `
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    <polyline points="9 22 9 12 15 12 15 22" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <span>Панель менеджера</span>
            `;
            managerLink.addEventListener('click', () => {
                window.location.href = '/manager';
            });
            header.appendChild(managerLink);
        }
    }

    loadUsers();
    loadStats();

    // Находим элементы поиска
    const searchInputs = document.querySelectorAll('.search-input');
    
    searchInputs.forEach(searchInput => {
        // Добавляем обработчик события input
        searchInput.addEventListener('input', function() {
            // Определяем текущую активную секцию
            const activeSection = document.querySelector('.section.active');
            const isUsersSection = activeSection && activeSection.id === 'section-users';
            const isShiftsSection = activeSection && activeSection.id === 'section-shifts';
            
            // Получаем значение поиска и приводим к нижнему регистру
            const searchText = this.value.toLowerCase();
            
            // Находим таблицу и строки в активной секции
            const tbody = activeSection.querySelector('.user-table tbody, .shift-table tbody');
            if (!tbody) return;
            
            const rows = tbody.querySelectorAll('tr:not(.no-results)');
            let visibleRows = false;

            // Проходим по каждой строке
            rows.forEach(row => {
                // Собираем текст для поиска
                const cells = Array.from(row.cells).slice(0, -1);
                const rowText = cells.map(cell => {
                    const statusBadge = cell.querySelector('.status-badge');
                    return statusBadge ? statusBadge.textContent : cell.textContent;
                }).join(' ').toLowerCase();

                // Проверяем совпадение
                const isVisible = rowText.includes(searchText);

                // Показываем или скрываем строку
                row.style.display = isVisible ? '' : 'none';
                if (isVisible) visibleRows = true;
            });

            // Управляем сообщением "не найдено"
            const noResultsRow = tbody.querySelector('.no-results');
            const colSpan = isUsersSection ? 7 : (isShiftsSection ? 5 : 5);

            if (!visibleRows) {
                if (!noResultsRow) {
                    const tr = document.createElement('tr');
                    tr.className = 'no-results';
                    tr.innerHTML = `
                        <td colspan="${colSpan}" class="no-results">
                            Ничего не найдено по вашему запросу
                        </td>
                    `;
                    tbody.appendChild(tr);
                }
            } else if (noResultsRow) {
                noResultsRow.remove();
            }

            // Обновляем счетчик после фильтрации
            if (isUsersSection) {
                updateUserCount();
            }
        });

        // Очистка по Escape
        searchInput.addEventListener('keydown', function(e) {
            if (e.key === 'Escape') {
                this.value = '';
                this.dispatchEvent(new Event('input'));
            }
        });
    });

    // Фильтры
    const filterBtns = document.querySelectorAll('.filter-btn');
    const filterDropdowns = document.querySelectorAll('.filter-dropdown');
    const filterSelects = document.querySelectorAll('.filter-select');

    // Открытие/закрытие дропдаунов фильтров
    filterBtns.forEach((btn, index) => {
        btn.addEventListener('click', function(e) {
            e.stopPropagation();
            
            // Закрываем все дропдауны
            filterDropdowns.forEach(dropdown => dropdown.classList.remove('show'));
            
            // Открываем текущий дропдаун
            const dropdown = btn.nextElementSibling;
            if (dropdown && dropdown.classList.contains('filter-dropdown')) {
                dropdown.classList.toggle('show');
            }
            
            // Сбрасываем активные селекты
            filterSelects.forEach(select => select.classList.remove('active'));
        });
    });

    // Обработчики для селектов фильтров
    filterSelects.forEach(select => {
        select.addEventListener('click', function() {
            const isActive = this.classList.contains('active');
            // Сначала отключаем все активные селекты
            filterSelects.forEach(s => s.classList.remove('active'));
            
            if (!isActive) {
                this.classList.add('active');
            }
        });
    });

    // Обработчики опций фильтра
    document.querySelectorAll('.filter-options .option').forEach(option => {
        option.addEventListener('click', function() {
            const select = this.closest('.filter-item').querySelector('.filter-select');
            const filterType = select.getAttribute('data-type');
            const value = this.getAttribute('data-value');
            const activeSection = document.querySelector('.section.active');
            
            // Обновляем внешний вид опций
            this.closest('.filter-options').querySelectorAll('.option').forEach(opt => {
                opt.classList.remove('selected');
            });
            this.classList.add('selected');
            
            // Закрываем дропдауны
            select.classList.remove('active');
            this.closest('.filter-dropdown').classList.remove('show');
            
            // Применяем фильтр в зависимости от активной секции
            if (activeSection.id === 'section-users') {
                filterRows(filterType, value);
            } else if (activeSection.id === 'section-shifts') {
                filterShifts(filterType, value);
            }
        });
    });
    
    // Закрытие дропдаунов по клику вне
    document.addEventListener('click', function() {
        filterDropdowns.forEach(dropdown => dropdown.classList.remove('show'));
        filterSelects.forEach(select => select.classList.remove('active'));
    });

    // Обработчик для переключения между вкладками меню
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        item.addEventListener('click', function() {
            const section = this.getAttribute('data-section');
            showSection(section);
        });
    });

    // Функция для переключения между вкладками
    function showSection(section) {
        // Скрываем все секции
        const sections = document.querySelectorAll('.section');
        sections.forEach(s => s.classList.remove('active'));
        
        // Показываем выбранную секцию
        const activeSection = document.getElementById('section-' + section);
        if (activeSection) activeSection.classList.add('active');
        
        // Обновляем активный элемент меню
        const navItems = document.querySelectorAll('.nav-item');
        navItems.forEach(item => {
            item.classList.remove('active');
            if (item.getAttribute('data-section') === section) {
                item.classList.add('active');
            }
        });
        
        // Загружаем данные для выбранной секции при необходимости
        if (section === 'users') {
            loadUsers();
            loadStats();
        } else if (section === 'shifts') {
            loadShifts();
        }
    }
    
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
                const roleCell = row.cells[2];
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
                    <td colspan="7" class="no-results">
                        Ничего не найдено по выбранным фильтрам
                    </td>
                `;
                tbody.appendChild(tr);
            }
        } else if (noResults) {
            noResults.remove();
        }
    }

    // Функция фильтрации для секции "Смены"
    function filterShifts(type, value) {
        const rows = document.querySelectorAll('#section-shifts .shift-table tbody tr:not(.no-results)');
        let hasVisibleRows = false;
        
        if (value === 'all') {
            rows.forEach(row => {
                row.style.display = '';
            });
            return;
        }
        
        const today = new Date().toLocaleDateString('ru-RU');
        
        rows.forEach(row => {
            if (type === 'date') {
                // Предполагаем, что дата находится во втором столбце
                let visible = false;
                
                switch (value) {
                    case 'today':
                        // Предполагаем, что в столбце находится строка вида дд.мм.гггг
                        visible = row.cells[1].textContent === today;
                        break;
                    case 'tomorrow':
                        const tomorrow = new Date();
                        tomorrow.setDate(tomorrow.getDate() + 1);
                        visible = row.cells[1].textContent === tomorrow.toLocaleDateString('ru-RU');
                        break;
                    case 'week':
                        // Реализация фильтра по неделе
                        // Это упрощённая реализация
                        visible = true;
                        break;
                    case 'month':
                        // Реализация фильтра по месяцу
                        // Это упрощённая реализация
                        visible = true;
                        break;
                    default:
                        visible = true;
                }
                
                row.style.display = visible ? '' : 'none';
                if (visible) hasVisibleRows = true;
            }
        });
        
        // Обновляем сообщение "не найдено"
        const tbody = document.querySelector('#section-shifts .shift-table tbody');
        const noResults = tbody.querySelector('.no-results');
        
        if (!hasVisibleRows) {
            if (!noResults) {
                const tr = document.createElement('tr');
                tr.className = 'no-results';
                tr.innerHTML = `
                    <td colspan="3" class="no-results">
                        Ничего не найдено по выбранным фильтрам
                    </td>
                `;
                tbody.appendChild(tr);
            }
        } else if (noResults) {
            noResults.remove();
        }
    }

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

    // Инициализация обработчиков событий для смен
    // Обработчики для фильтров смен
    document.getElementById('applyShiftFilters')?.addEventListener('click', function() {
        currentPage = 1; // Сбрасываем на первую страницу при применении фильтров
        loadShifts();
    });
    
    document.getElementById('resetShiftFilters')?.addEventListener('click', function() {
        document.getElementById('shift-date-filter').value = '';
        currentPage = 1;
        loadShifts();
    });
    
    // Обработчики для модального окна смен
    document.getElementById('addShiftBtn')?.addEventListener('click', showAddShiftModal);
    
    document.getElementById('shiftForm')?.addEventListener('submit', saveShift);
    
    document.getElementById('cancelShiftBtn')?.addEventListener('click', function() {
        document.getElementById('shiftModal').style.display = 'none';
    });
    
    // Закрытие модального окна по клику на крестик
    const shiftModalCloseBtn = document.querySelector('#shiftModal .close-modal-btn');
    if (shiftModalCloseBtn) {
        shiftModalCloseBtn.addEventListener('click', function() {
            document.getElementById('shiftModal').style.display = 'none';
        });
    }
    
    // Закрытие модального окна по клику вне его
    window.addEventListener('click', function(event) {
        const shiftModal = document.getElementById('shiftModal');
        if (event.target === shiftModal) {
            shiftModal.style.display = 'none';
        }
    });
});

async function loadUsers() {
    try {
        const response = await fetch('/api/admin/users', {
            method: 'GET',
            credentials: 'include' // Add credentials to send cookies
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        console.log(data);
        let users = Array.isArray(data) ? data : (data.users || []);

        const tbody = document.querySelector('.user-table tbody');
        if (!tbody) return;

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
                console.log(user.last_active);
                console.log(user.created_at);
                console.log(formattedLastActive);
                console.log(formattedCreatedAt);
                
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

// Функция для форматирования дат пользователей
function formatUserDate(dateString) {
    if (!dateString) return '—';
    console.log(dateString);
    try {
        const date = new Date(dateString);
        console.log(date);
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
        
        // Отладка: проверим значения по отдельности
        const username = formData.get('username');
        const name = formData.get('name');
        const email = formData.get('email');
        const password = formData.get('password');
        const role = formData.get('role');
        
        console.log('Значения полей формы:');
        console.log('username:', username);
        console.log('name:', name);
        console.log('email:', email);
        console.log('password:', password ? '[MASKED]' : 'empty');
        console.log('role:', role);
        
        const userData = {
            username: username,
            name: name,
            email: email,
            password: password,
            role: role,
            status: 'active'
        };

        console.log('Создание пользователя:', JSON.stringify(userData));

        try {
            const response = await fetch('/api/admin/users', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(userData),
                credentials: 'include' // Add credentials to send cookies
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
    form.elements['name'].value = userRow.cells[1].textContent;
    form.elements['role'].value = userRow.cells[2].getAttribute('data-role');
    form.elements['status'].value = userRow.querySelector('.status-badge').classList.contains('active') ? 'active' : 'inactive';

    modal.style.display = 'block';

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
            const response = await fetch(`/api/admin/users/${id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(userData),
                credentials: 'include' // Add credentials to send cookies
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
            credentials: 'include' // Add credentials to send cookies
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
    
    // Проверяем, что это валидная дата
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
        console.error('Невалидная дата:', dateString);
        return dateString; // Возвращаем исходную строку, если дата невалидна
    }
    
    // Если это только дата (без времени)
    if (dateString.indexOf('T') === -1 && dateString.indexOf(' ') === -1) {
        return date.toLocaleDateString('ru-RU', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    }
    
    // Если это полная дата с временем
    return date.toLocaleString('ru-RU', {
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
    // Delete the auth cookie
    document.cookie = "auth_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
    
    // Also clear localStorage for backward compatibility
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    
    // Redirect to login page
    window.location.href = '/';
}

// Инициализация статистики
async function loadStats() {
    try {
        const response = await fetch('/api/admin/stats', {
            credentials: 'include' // Add credentials to send cookies
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

// Добавляем код для управления сменами
// Глобальные переменные для смен
let allManagers = [];
let allEmployees = [];
let currentShift = null;
let currentPage = 1;
const shiftsPerPage = 10;

// Функции для работы со сменами
async function loadShifts() {
    try {
        // Загружаем список менеджеров, если он еще не загружен
        if (allManagers.length === 0) {
            await loadManagers();
        }
        
        // Попытка загрузки смен через API
        const response = await fetch('/api/admin/shifts', {
            method: 'GET',
            credentials: 'include',
        }).catch(e => {
            console.log('API недоступен, используем моки данных');
            return { ok: false };
        });

        let shifts;
        
        if (!response.ok) {
            // Если API недоступно, используем моки данных
            console.log('Используем моки данных для демонстрации');
            shifts = [
                { 
                    id: 1, 
                    date: new Date().toISOString().split('T')[0], 
                    start_time: '09:00', 
                    end_time: '18:00',
                    manager_name: 'Айгуль Нурланова',
                    manager_id: 1,
                    status: 'active',
                    employees_count: 5
                },
                { 
                    id: 2, 
                    date: new Date(Date.now() + 86400000).toISOString().split('T')[0], 
                    start_time: '08:00', 
                    end_time: '17:00',
                    manager_name: 'Нургуль Байжанова',
                    manager_id: 7,
                    status: 'pending',
                    employees_count: 3
                }
            ];
        } else {
            // Если API доступно, используем данные из него
            const data = await response.json();
            shifts = data.shifts || [];
        }

        const tbody = document.getElementById('shifts-tbody');
        tbody.innerHTML = '';

        if (shifts.length === 0) {
            tbody.innerHTML = `<tr><td colspan="5" class="no-results">Нет смен</td></tr>`;
            return;
        }

        // Обновляем счетчик смен
        const totalShiftsElem = document.getElementById('total-shifts');
        if (totalShiftsElem) {
            totalShiftsElem.textContent = shifts.length;
        }

        console.log('Загруженные смены для отображения:', shifts);
        
        // Добавляем строки для каждой смены
        shifts.forEach(shift => {
            // Форматируем дату
            const formattedDate = formatShiftDate(shift.date);
            
            // Форматируем время
            const startTime = formatShiftTime(shift.start_time);
            const endTime = formatShiftTime(shift.end_time);
            
            console.log(`Смена ${shift.id}, время: ${startTime}-${endTime}`);
            
            // Определяем менеджера
            let managerName = shift.manager_name || 'Не назначен';
            
            // Если есть manager_id, но нет manager_name, пытаемся найти менеджера в списке
            if (shift.manager_id && !shift.manager_name && allManagers.length > 0) {
                const manager = allManagers.find(m => m.id === parseInt(shift.manager_id));
                if (manager) {
                    managerName = manager.name || manager.username;
                }
            }
            
            // Определяем статус
            const status = shift.status || 'active';
            
            tbody.innerHTML += `
                <tr data-shift-id="${shift.id}">
                    <td>${formattedDate}</td>
                    <td>${startTime} - ${endTime}</td>
                    <td>${managerName}</td>
                    <td><span class="status-badge ${status}">${translateShiftStatus(status)}</span></td>
                    <td class="actions">
                        <button onclick="editShift(${shift.id})" class="edit-btn" title="Редактировать">
                            <img src="/static/images/edit.svg" alt="Редактировать" class="icon">
                        </button>
                        <button onclick="confirmDeleteShift(${shift.id})" class="delete-btn" title="Удалить">
                            <img src="/static/images/delete.svg" alt="Удалить" class="icon">
                        </button>
                    </td>
                </tr>
            `;
        });
    } catch (error) {
        console.error('Failed to load shifts:', error);
        showNotification('Ошибка при загрузке смен', 'error');
    }
}

// Функция для форматирования даты смены
function formatShiftDate(dateString) {
    if (!dateString) return '—';
    
    try {
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            console.error('Невалидная дата смены:', dateString);
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

async function loadManagers() {
    try {
        // При недоступности API используем моки данных
        const mockManagers = [
            { id: 1, name: 'Айгуль Нурланова', role: 'manager' }
        ];
        
        try {
            const response = await fetch('/api/admin/users?role=manager', {
                method: 'GET',
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                const managers = Array.isArray(data) ? data : (data.users || []);
                // Фильтруем, чтобы получить только менеджеров и только активных
                allManagers = managers.filter(user => 
                    user.role === 'manager' && (!user.status || user.status === 'active')
                );
                return allManagers;
            } else {
                console.log('API недоступен для загрузки менеджеров, используем моки данных');
                allManagers = mockManagers;
                return allManagers;
            }
        } catch (error) {
            console.error('Failed to load managers:', error);
            allManagers = mockManagers;
            return allManagers;
        }
    } catch (error) {
        console.error('Unexpected error loading managers:', error);
        showNotification('Ошибка при загрузке списка менеджеров', 'error');
        return [];
    }
}

async function loadEmployees() {
    try {
        // Загружаем только официантов и поваров
        const roleTypes = ['waiter', 'cook'];
        let allEmployeesList = [];
        
        // Временная карта для отслеживания уже добавленных сотрудников по ID
        const employeeMap = new Map();
        
        for (const role of roleTypes) {
            try {
                const response = await fetch(`/api/admin/users?role=${role}&status=active`, {
                    method: 'GET',
                    credentials: 'include'
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
        
        if (allEmployeesList.length === 0) {
            // Если API недоступно, используем моки данных
            allEmployeesList = [
                { id: 2, name: 'Аскар Сериков', role: 'cook' },
                { id: 3, name: 'Мадина Сабитова', role: 'waiter' },
                { id: 4, name: 'Руслан Куаныш', role: 'waiter' },
                { id: 5, name: 'Динара Нурлан', role: 'cook' }
            ];
        }
        
        allEmployees = allEmployeesList;
        return allEmployees;
    } catch (error) {
        console.error('Failed to load employees:', error);
        showNotification('Ошибка при загрузке списка сотрудников', 'error');
        return [];
    }
}

function populateManagerDropdown(selectedManagerId = null) {
    const select = document.getElementById('shift-manager');
    select.innerHTML = '<option value="">Выберите менеджера</option>';
    
    allManagers.forEach(manager => {
        const displayName = manager.name || manager.username;
        select.innerHTML += `<option value="${manager.id}" ${manager.id === selectedManagerId ? 'selected' : ''}>${displayName}</option>`;
    });
}

function populateEmployeeCheckboxes(selectedEmployeeIds = []) {
    const container = document.getElementById('shift-employees-container');
    container.innerHTML = '';
    
    // Группируем сотрудников по ролям для лучшей организации
    const groups = {
        'waiter': {title: 'Официанты', employees: []},
        'cook': {title: 'Повара', employees: []}
    };
    
    // Set для отслеживания добавленных ID, чтобы избежать дублирования
    const addedEmployeeIds = new Set();
    
    allEmployees.forEach(employee => {
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
    
    // Добавляем информацию о количестве сотрудников для отладки
    console.log('Добавлено сотрудников по категориям:', 
               Object.entries(groups).map(([role, group]) => `${role}: ${group.employees.length}`).join(', '));
}

async function showAddShiftModal() {
    currentShift = null;
    
    document.getElementById('shiftModalTitle').textContent = 'Добавить смену';
    document.getElementById('shiftForm').reset();
    document.getElementById('shift-id').value = '';
    
    // Устанавливаем текущую дату в поле даты
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('shift-date').value = today;
    
    // Загружаем списки менеджеров и сотрудников
    if (allManagers.length === 0) {
        await loadManagers();
    }
    
    if (allEmployees.length === 0) {
        await loadEmployees();
    }
    
    populateManagerDropdown();
    populateEmployeeCheckboxes();
    
    document.getElementById('shiftModal').style.display = 'block';
}

async function editShift(shiftId) {
    try {
        const response = await fetch(`/api/admin/shifts/${shiftId}`, {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const shift = await response.json();
        currentShift = shift;
        
        document.getElementById('shiftModalTitle').textContent = 'Редактировать смену';
        document.getElementById('shift-id').value = shift.id;
        
        // Форматируем дату для поля input
        const dateObj = new Date(shift.date);
        const formattedDate = dateObj.toISOString().split('T')[0]; // YYYY-MM-DD
        document.getElementById('shift-date').value = formattedDate;
        
        // Сохраняем оригинальные значения времени для логирования
        console.log('Исходные данные смены:', JSON.stringify(shift, null, 2));
        
        // Функция для очистки и форматирования времени
        const cleanTimeFormat = (timeStr) => {
            if (!timeStr) return '';
            
            // Если время включает 'T', это ISO формат даты-времени
            if (timeStr.includes('T')) {
                // Извлекаем только компонент времени (HH:MM)
                const timePart = timeStr.split('T')[1].substring(0, 5);
                return timePart;
            }
            
            // Остальная логика обработки...
        };
        
        // Если формат уже HH:MM, не трогаем его
        const startTime = shift.start_time && typeof shift.start_time === 'string' && 
                          shift.start_time.match(/^\d{1,2}:\d{2}$/) ? 
                          shift.start_time : cleanTimeFormat(shift.start_time);
                          
        const endTime = shift.end_time && typeof shift.end_time === 'string' && 
                        shift.end_time.match(/^\d{1,2}:\d{2}$/) ? 
                        shift.end_time : cleanTimeFormat(shift.end_time);
        
        console.log('Итоговое время начала:', startTime);
        console.log('Итоговое время окончания:', endTime);
        
        document.getElementById('shift-start-time').value = startTime;
        document.getElementById('shift-end-time').value = endTime;
        document.getElementById('shift-notes').value = shift.notes || '';
        
        // Устанавливаем статус смены
        const statusSelect = document.getElementById('shift-status');
        statusSelect.value = shift.status || 'active';
        
        // Загружаем списки менеджеров и сотрудников, если они еще не загружены
        if (allManagers.length === 0) {
            await loadManagers();
        }
        
        if (allEmployees.length === 0) {
            await loadEmployees();
        }
        
        populateManagerDropdown(shift.manager_id);
        
        // Получаем ID выбранных сотрудников
        const selectedEmployeeIds = shift.employees ? shift.employees.map(emp => emp.id) : [];
        populateEmployeeCheckboxes(selectedEmployeeIds);
        
        document.getElementById('shiftModal').style.display = 'block';

    } catch (error) {
        console.error('Failed to load shift details:', error);
        showNotification('Ошибка при загрузке информации о смене', 'error');
    }
}

async function saveShift(e) {
    e.preventDefault();
    
    try {
        console.log('Сохранение смены начато...');
        
        const shiftId = document.getElementById('shift-id').value;
        const date = document.getElementById('shift-date').value;
        const startTime = document.getElementById('shift-start-time').value;
        const endTime = document.getElementById('shift-end-time').value;
        const managerId = document.getElementById('shift-manager').value;
        const status = document.getElementById('shift-status').value;
        const notes = document.getElementById('shift-notes').value;
        
        console.log('Время начала:', startTime);
        console.log('Время окончания:', endTime);
        
        // Проверка обязательных полей
        if (!date) {
            showNotification('Укажите дату смены', 'error');
            return;
        }
        
        if (!startTime || !endTime) {
            showNotification('Укажите время начала и окончания смены', 'error');
            return;
        }
        
        if (!managerId) {
            showNotification('Выберите ответственного менеджера', 'error');
            return;
        }
        
        console.log('Проверка отмеченных сотрудников...');
        // Проверяем, есть ли контейнер для сотрудников
        const container = document.getElementById('shift-employees-container');
        if (!container) {
            console.error('Контейнер для сотрудников не найден!');
            showNotification('Ошибка в интерфейсе: контейнер сотрудников не найден', 'error');
            return;
        }
        
        // Получаем выбранных сотрудников
        const checkboxes = container.querySelectorAll('input[type="checkbox"]:checked');
        console.log('Найдено отмеченных чекбоксов:', checkboxes.length);
        
        const selectedEmployees = [];
        checkboxes.forEach(checkbox => {
            selectedEmployees.push(parseInt(checkbox.value, 10));
        });
        
        if (selectedEmployees.length === 0) {
            showNotification('Выберите хотя бы одного сотрудника для смены', 'error');
            return;
        }
        
        // Находим имя менеджера для отображения
        let managerName = '';
        if (allManagers.length > 0) {
            const selectedManager = allManagers.find(m => m.id === parseInt(managerId));
            if (selectedManager) {
                managerName = selectedManager.name || selectedManager.username;
            }
        }
        
        // Валидация формата времени (должно быть в формате HH:MM)
        const timeRegex = /^([01]?[0-9]|2[0-3]):[0-5][0-9]$/;
        if (!timeRegex.test(startTime) || !timeRegex.test(endTime)) {
            showNotification('Время должно быть в формате ЧЧ:ММ', 'error');
            return;
        }
        
        // Убедимся, что время в правильном формате (двузначные часы и минуты)
        const formatTimeString = (timeStr) => {
            const [hours, minutes] = timeStr.split(':');
            return `${hours.padStart(2, '0')}:${minutes.padStart(2, '0')}`;
        };
        
        const formattedStartTime = startTime.substring(0, 5); // Берём только HH:MM
        const formattedEndTime = endTime.substring(0, 5);     // Берём только HH:MM
        
        const shiftData = {
            date,
            start_time: formattedStartTime, // Отправляем время в формате HH:MM
            end_time: formattedEndTime, // Отправляем время в формате HH:MM
            manager_id: parseInt(managerId, 10),
            manager_name: managerName, // Добавляем имя менеджера
            status: status || 'active', // Используем выбранный статус или "active" по умолчанию
            notes,
            employee_ids: selectedEmployees
        };
        
        console.log('Отправляем данные смены:', shiftData);
        
        let response;
        
        if (shiftId) {
            // Обновляем существующую смену
            response = await fetch(`/api/admin/shifts/${shiftId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(shiftData),
                credentials: 'include'
            });
        } else {
            // Создаем новую смену
            response = await fetch('/api/admin/shifts', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(shiftData),
                credentials: 'include'
            });
        }
        
        // Обрабатываем ответ
        if (!response.ok) {
            let errorMessage = `Ошибка HTTP: ${response.status}`;
            try {
                const errorText = await response.text();
                console.error('Ответ сервера с ошибкой:', errorText);
                errorMessage += `, сообщение: ${errorText}`;
            } catch (err) {
                console.error('Не удалось прочитать текст ошибки:', err);
            }
            throw new Error(errorMessage);
        }
        
        showNotification(shiftId ? 'Смена успешно обновлена' : 'Смена успешно создана', 'success');
        document.getElementById('shiftModal').style.display = 'none';
        loadShifts();
    } catch (error) {
        console.error('Ошибка при сохранении смены:', error);
        showNotification('Ошибка при сохранении смены: ' + error.message, 'error');
    }
}

async function confirmDeleteShift(shiftId) {
    if (confirm('Вы уверены, что хотите удалить эту смену?')) {
        try {
            const response = await fetch(`/api/admin/shifts/${shiftId}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            showNotification('Смена успешно удалена', 'success');
            loadShifts();
        } catch (error) {
            console.error('Failed to delete shift:', error);
            showNotification('Ошибка при удалении смены', 'error');
        }
    }
}