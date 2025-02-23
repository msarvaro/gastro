document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.querySelector('.signup-form');
    
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const remember = document.querySelector('input[type="checkbox"]').checked;
            
            const requestData = { 
                username, 
                password,
                remember
            };
            
            console.log('Sending login request:', requestData);
            
            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(requestData)
                });
                
                console.log('Response status:', response.status);
                
                if (response.ok) {
                    const data = await response.json();
                    console.log('Login successful:', data);
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
                alert('Произошла ошибка при входе');
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
    window.location.href = '/';
}