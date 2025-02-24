// Инициализация данных при первом запуске
if (!localStorage.getItem('tables')) {
    const initialTables = {
        tables: [
            {id: 1, number: 1, seats: 4},
            {id: 2, number: 2, seats: 6},
            {id: 3, number: 3, seats: 2},
            {id: 4, number: 4, seats: 4},
            {id: 5, number: 5, seats: 8}
        ]
    };
    localStorage.setItem('tables', JSON.stringify(initialTables));
}

if (!localStorage.getItem('orders')) {
    localStorage.setItem('orders', JSON.stringify({orders: []}));
}

if (!localStorage.getItem('orderHistory')) {
    localStorage.setItem('orderHistory', JSON.stringify({orders: []}));
}