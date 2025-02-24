document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.querySelector('.signup-form');
    
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const remember = document.querySelector('input[type="checkbox"]').checked;

            // Локальные данные официантов для тестирования
            const waiters = [
                { id: 1, username: 'waiter1', password: 'password123', role: 'waiter' },
                { id: 2, username: 'waiter2', password: 'password123', role: 'waiter' }
            ];

            const waiter = waiters.find(w => 
                w.username === username && 
                w.password === password
            );

            if (waiter) {
                localStorage.setItem('currentUser', JSON.stringify(waiter));
                window.location.href = 'index.html';
                return;
            }

            // Если не официант, пробуем через бэкенд
            const requestData = { 
                username, 
                password,
                remember
            };
            
            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(requestData)
                });
                
                if (response.ok) {
                    const data = await response.json();
                    localStorage.setItem('token', data.token);
                    localStorage.setItem('role', data.role);
                    window.location.href = '/admin';
                } else {
                    const errorText = await response.text();
                    console.error('Login failed:', errorText);
                    alert('Неверные учетные данные');
                }
            } catch (error) {
                console.error('Error during login:', error);
                // Если сервер недоступен, просто показываем общую ошибку
                alert('Неверные учетные данные');
            }
        });
    } else {
        console.error('Login form not found');
    }
});

// Функция для выхода
function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    localStorage.removeItem('currentUser');
    window.location.href = '/';
}