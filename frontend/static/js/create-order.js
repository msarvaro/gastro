let currentOrderData = {
    tableId: null,
    items: [],
    comment: '',
    total: 0
};

document.addEventListener('DOMContentLoaded', async function() {
    // Проверка авторизации
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');
    
    if (!token || role !== 'waiter') {
        window.location.href = '/';
        return;
    }

    // Код для выбора стола
    const selectTableBtn = document.getElementById('selectTableBtn');
    const tableModal = document.getElementById('tableModal');
    const closeModalBtn = document.querySelector('.close-modal-btn');
    const selectedTableText = document.getElementById('selectedTableText');
    
    // Временные данные о столах
    const tables = [
        { id: 1, number: 1, seats: 4, status: 'free' },
        { id: 2, number: 2, seats: 4, status: 'free' },
        { id: 3, number: 3, seats: 6, status: 'free' },
        { id: 4, number: 4, seats: 2, status: 'free' },
        { id: 5, number: 5, seats: 4, status: 'free' },
        { id: 6, number: 6, seats: 8, status: 'free' }
    ];

    // Отрисовка столов в модальном окне
    function renderTables() {
        const grid = document.querySelector('.table-modal__grid');
        grid.innerHTML = tables.map(table => `
            <div class="table-option ${table.status === 'occupied' ? 'occupied' : ''}" 
                 data-table-id="${table.id}">
                <div class="table-number">Стол ${table.number}</div>
                <div class="table-seats">${table.seats} мест</div>
            </div>
        `).join('');

        // Добавляем обработчики для каждого стола
        grid.querySelectorAll('.table-option:not(.occupied)').forEach(tableEl => {
            tableEl.addEventListener('click', () => {
                const tableId = tableEl.dataset.tableId;
                const table = tables.find(t => t.id === parseInt(tableId));
                selectedTableText.textContent = `Стол №${table.number}`;
                tableModal.classList.remove('active');
                
                // Активируем форму создания заказа
                document.querySelector('.create-order-section').style.opacity = '1';
                document.querySelector('.create-order-section').style.pointerEvents = 'auto';
            });
        });
    }

    // Обработчики модального окна
    selectTableBtn.addEventListener('click', () => {
        tableModal.classList.add('active');
        renderTables();
    });

    closeModalBtn.addEventListener('click', () => {
        tableModal.classList.remove('active');
    });

    tableModal.addEventListener('click', (e) => {
        if (e.target === tableModal) {
            tableModal.classList.remove('active');
        }
    });

    // Изначально форма создания заказа неактивна
    document.querySelector('.create-order-section').style.opacity = '0.5';
    document.querySelector('.create-order-section').style.pointerEvents = 'none';

    // Существующий код для меню и заказов
    loadMenu();
    
    // Обработчики категорий
    const categoryButtons = document.querySelectorAll('.category-btn');
    categoryButtons.forEach(button => {
        button.addEventListener('click', () => {
            categoryButtons.forEach(btn => btn.classList.remove('active'));
            button.classList.add('active');
            filterDishesByCategory(button.textContent);
        });
    });

    // Поиск блюд
    const searchInput = document.querySelector('.search-input');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            searchDishes(e.target.value);
        });
    }

    // Обновляем обработчик для кнопки создания заказа
    const createOrderBtn = document.querySelector('.create-order-btn');
    createOrderBtn.addEventListener('click', () => {
        if (currentOrderData.items.length > 0) {
            showConfirmOrderModal();
        }
    });

    const confirmOrderBtn = document.getElementById('confirmOrderBtn');
    confirmOrderBtn.addEventListener('click', () => {
        hideConfirmOrderModal();
        createOrder();
    });

    const cancelOrderBtn = document.getElementById('cancelOrderBtn');
    cancelOrderBtn.addEventListener('click', hideConfirmOrderModal);

    // Закрытие модального окна по клику вне него
    document.getElementById('confirmOrderModal').addEventListener('click', (e) => {
        if (e.target === e.currentTarget) {
            hideConfirmOrderModal();
        }
    });

    // Добавляем обработчик для кнопки "Назад"
    const backBtn = document.querySelector('.back-btn');
    backBtn.addEventListener('click', () => {
        // Если есть несохраненные изменения, показываем подтверждение
        if (currentOrderData.items.length > 0) {
            if (confirm('Вы уверены, что хотите вернуться? Несохраненные изменения будут потеряны.')) {
                window.location.href = 'orders.html';
            }
        } else {
            window.location.href = 'orders.html';
        }
    });

    // Добавляем обработчик для кнопки очистки заказа
    const clearOrderBtn = document.getElementById('clearOrderBtn');
    clearOrderBtn.addEventListener('click', () => {
        if (currentOrderData.items.length > 0) {
            if (confirm('Вы уверены, что хотите очистить заказ?')) {
                clearOrder();
            }
        }
    });

    // Загружаем статус столов
    try {
        const tableStatus = await waiterApi.getTableStatus();
        const tables = await waiterApi.getTables();
        
        // Обновляем статус столов в модальном окне
        tables.forEach(table => {
            const tableEl = document.querySelector(`.table-option[data-table-id="${table.id}"]`);
            if (tableEl) {
                if (tableStatus[table.id] === 'occupied') {
                    tableEl.classList.add('occupied');
                } else {
                    tableEl.classList.remove('occupied');
                }
            }
        });
    } catch (error) {
        console.error('Error loading table status:', error);
    }
});

// Загрузка меню
async function loadMenu() {
    try {
        const dishes = await menuApi.getMenuItems();
        renderMenu(dishes);
    } catch (error) {
        console.error('Error loading menu:', error);
        // Show error notification to user
        const dishesList = document.querySelector('.dishes-list');
        dishesList.innerHTML = '<div class="error-message">Ошибка загрузки меню. Пожалуйста, попробуйте позже.</div>';
    }
}

// Отрисовка меню
function renderMenu(dishes) {
    const dishesList = document.querySelector('.dishes-list');
    dishesList.innerHTML = dishes.map(dish => `
        <div class="dish-card" data-id="${dish.id}" data-category="${dish.category}">
            <div class="dish-card__info">
                <div class="dish-card__title">${dish.name}</div>
                <div class="dish-card__price">${dish.price} KZT</div>
            </div>
            <div class="dish-card__actions">
                <button class="quantity-btn minus" onclick="decreaseQuantity(${dish.id})">-</button>
                <span class="quantity" id="quantity-${dish.id}">0</span>
                <button class="quantity-btn plus" onclick="increaseQuantity(${dish.id})">+</button>
            </div>
        </div>
    `).join('');
}

// Функции управления количеством
function increaseQuantity(dishId) {
    const quantityElement = document.getElementById(`quantity-${dishId}`);
    const dishCard = document.querySelector(`.dish-card[data-id="${dishId}"]`);
    
    // Добавляем анимацию нажатия
    dishCard.classList.add('adding');
    setTimeout(() => dishCard.classList.remove('adding'), 200);

    let quantity = parseInt(quantityElement.textContent);
    quantityElement.textContent = quantity + 1;
    updateOrder();
}

function decreaseQuantity(dishId) {
    const quantityElement = document.getElementById(`quantity-${dishId}`);
    let quantity = parseInt(quantityElement.textContent);
    
    if (quantity > 0) {
        quantityElement.textContent = quantity - 1;
        
        // Если количество стало 0, добавляем анимацию удаления
        if (quantity === 1) {
            const orderItem = document.querySelector(`.order-item[data-id="${dishId}"]`);
            if (orderItem) {
                orderItem.classList.add('removing');
                setTimeout(() => {
                    updateOrder();
                }, 200);
                return;
            }
        }
        updateOrder();
    }
}

// Обновление заказа
function updateOrder() {
    const currentOrder = [];
    let total = 0;

    document.querySelectorAll('.dish-card').forEach(card => {
        const dishId = card.dataset.id;
        const quantity = parseInt(document.getElementById(`quantity-${dishId}`).textContent);
        const price = parseInt(card.querySelector('.dish-card__price').textContent);
        
        if (quantity > 0) {
            const item = {
                id: dishId,
                name: card.querySelector('.dish-card__title').textContent,
                quantity: quantity,
                price: price
            };
            currentOrder.push(item);
            total += price * quantity;
        }
    });

    // Обновляем глобальные данные заказа
    currentOrderData.items = currentOrder;
    currentOrderData.total = total;
    
    // Обновляем UI
    document.querySelector('.current-order__total').textContent = `${total} KZT`;
    
    const orderItems = document.querySelector('.current-order__items');
    orderItems.innerHTML = currentOrder.map(item => `
        <div class="order-item" data-id="${item.id}">
            <span class="order-item__name">${item.name} x ${item.quantity}</span>
            <span class="order-item__price">${item.price * item.quantity} KZT</span>
        </div>
    `).join('');

    // Обновляем состояние кнопки создания заказа
    const createOrderBtn = document.querySelector('.create-order-btn');
    if (currentOrder.length === 0) {
        createOrderBtn.disabled = true;
        createOrderBtn.classList.add('disabled');
    } else {
        createOrderBtn.disabled = false;
        createOrderBtn.classList.remove('disabled');
    }
}

// Фильтрация по категориям
async function filterDishesByCategory(category) {
    try {
        const dishes = category === 'Все' 
            ? await menuApi.getMenuItems()
            : await menuApi.getMenuItemsByCategory(category);
        renderMenu(dishes);
    } catch (error) {
        console.error('Error filtering dishes:', error);
        // Show error notification to user
        const dishesList = document.querySelector('.dishes-list');
        dishesList.innerHTML = '<div class="error-message">Ошибка фильтрации меню. Пожалуйста, попробуйте позже.</div>';
    }
}

// Поиск блюд
async function searchDishes(query) {
    try {
        const dishes = await menuApi.getMenuItems();
        const filteredDishes = dishes.filter(dish => 
            dish.name.toLowerCase().includes(query.toLowerCase())
        );
        renderMenu(filteredDishes);
    } catch (error) {
        console.error('Error searching dishes:', error);
        // Show error notification to user
        const dishesList = document.querySelector('.dishes-list');
        dishesList.innerHTML = '<div class="error-message">Ошибка поиска блюд. Пожалуйста, попробуйте позже.</div>';
    }
}

function showConfirmOrderModal() {
    // Проверяем, есть ли блюда в заказе
    if (currentOrderData.items.length === 0) {
        return; // Не показываем модальное окно, если заказ пустой
    }

    const modal = document.getElementById('confirmOrderModal');
    const orderSummary = modal.querySelector('.order-summary__table');
    const commentSection = modal.querySelector('.order-summary__comment');
    const totalSection = modal.querySelector('.order-summary__total');
    
    // Заполняем данные заказа
    orderSummary.innerHTML = currentOrderData.items.map(item => `
        <div class="order-summary__item">
            <span>${item.name} x ${item.quantity}</span>
            <span>${item.price * item.quantity} KZT</span>
        </div>
    `).join('');
    
    // Добавляем комментарий если есть
    const comment = document.querySelector('.current-order__comment textarea').value;
    if (comment.trim()) {
        commentSection.textContent = comment;
        currentOrderData.comment = comment;
    } else {
        commentSection.style.display = 'none';
    }
    
    totalSection.textContent = `Итого: ${currentOrderData.total} KZT`;
    modal.classList.add('active');
}

function hideConfirmOrderModal() {
    document.getElementById('confirmOrderModal').classList.remove('active');
}

async function createOrder() {
    try {
        // Проверяем, выбран ли стол
        const selectedTableText = document.getElementById('selectedTableText');
        if (!selectedTableText || !selectedTableText.textContent) {
            alert('Пожалуйста, выберите стол');
            return;
        }

        // Получаем ID стола из текста (например, "Стол №2" -> 2)
        const tableId = parseInt(selectedTableText.textContent.match(/\d+/)[0]);

        // Проверяем, есть ли выбранные блюда
        if (currentOrderData.items.length === 0) {
            alert('Добавьте хотя бы одно блюдо в заказ');
            return;
        }
        
        // Создаем новый заказ
        const newOrder = {
            tableId: tableId,
            waiterId: parseInt(localStorage.getItem('userId')),
            status: 'new',
            items: currentOrderData.items,
            comment: currentOrderData.comment || '',
            total: currentOrderData.total
        };

        // Отправляем заказ на сервер
        await waiterApi.createOrder(newOrder);
        
        // Показываем уведомление об успехе
        showSuccessNotification();
        
        // Перенаправляем на страницу заказов
        setTimeout(() => {
            window.location.href = 'orders.html';
        }, 1500);
        
    } catch (error) {
        console.error('Error creating order:', error);
        alert('Ошибка при создании заказа: ' + error.message);
    }
}

function showSuccessNotification() {
    const notification = document.getElementById('successNotification');
    notification.classList.add('active');
    setTimeout(() => {
        notification.classList.remove('active');
    }, 3000);
}

function resetOrderForm() {
    // Очищаем количества
    document.querySelectorAll('.quantity').forEach(el => {
        el.textContent = '0';
    });
    
    // Очищаем комментарий
    document.querySelector('.current-order__comment textarea').value = '';
    
    // Обновляем заказ
    updateOrder();
}

// Функция очистки заказа
function clearOrder() {
    // Сбрасываем все количества
    document.querySelectorAll('.quantity').forEach(el => {
        el.textContent = '0';
    });
    
    // Очищаем комментарий
    document.querySelector('.current-order__comment textarea').value = '';
    
    // Очищаем данные заказа
    currentOrderData.items = [];
    currentOrderData.total = 0;
    
    // Обновляем UI
    updateOrder();
}