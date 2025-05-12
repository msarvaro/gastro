document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.querySelector('.signup-form');

    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const usernameInput = document.getElementById('username');
            const passwordInput = document.getElementById('password');
            const rememberCheckbox = document.querySelector('.remember-checkbox');

            if (!usernameInput || !passwordInput) {
                alert('Ошибка: не найдены поля формы!');
                return;
            }

            const username = usernameInput.value;
            const password = passwordInput.value;
            const remember = rememberCheckbox ? rememberCheckbox.checked : false;

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
                    
                    // Сохраняем токен и роль
                    localStorage.setItem('token', data.token);
                    localStorage.setItem('role', data.role);
                    
                    // Перенаправляем на соответствующую страницу
                    if (data.redirect) {
                        window.location.href = data.redirect;
                    } else {
                        // Если redirect не указан, перенаправляем по роли
                        switch (data.role) {
                            case 'admin':
                                window.location.href = '/admin';
                                break;
                            case 'manager':
                                window.location.href = '/manager';
                                break;
                            case 'waiter':
                                window.location.href = '/waiter';
                                break;
                            default:
                                console.error('Unknown role:', data.role);
                                alert('Неизвестная роль пользователя');
                                window.location.href = '/';
                        }
                    }
                } else {
                    const errorData = await response.json().catch(() => ({}));
                    console.error('Login failed:', errorData);
                    alert(errorData.message || 'Неверные учетные данные');
                }
            } catch (error) {
                console.error('Error during login:', error);
                alert('Ошибка при входе в систему. Пожалуйста, попробуйте позже.');
            }
        });
    } else {
        console.error('Login form not found');
    }
});

// Функция для выхода
function logout() {
    // Удаляем куки
    document.cookie = 'auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
    // Удаляем данные из localStorage
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    // Перенаправляем на страницу входа
    window.location.href = '/';
}

document.addEventListener('DOMContentLoaded', function() {
    // Get the registration form
    const registerForm = document.querySelector('.signup-form');
    
    // Toggle password visibility
    const passwordToggles = document.querySelectorAll('.password-toggle');
    passwordToggles.forEach(toggle => {
        toggle.addEventListener('click', function() {
            const passwordInput = this.parentElement.querySelector('input');
            const toggleIcon = this.querySelector('.toggle-icon');
            
            // Toggle password visibility
            if (passwordInput.type === 'password') {
                passwordInput.type = 'text';
                toggleIcon.src = '../static/images/eye-off.svg'; // Assuming you have this icon
            } else {
                passwordInput.type = 'password';
                toggleIcon.src = '../static/images/eye.svg';
            }
        });
    });
    
    // Handle form submission
    if (registerForm) {
        registerForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            // Get form values
            const name = document.getElementById('name').value;
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            const confirmPassword = document.getElementById('confirmPassword').value;
            const termsAgreed = document.querySelector('.checkbox-input').checked;
            
            // Basic validation
            if (!name || !email || !password || !confirmPassword) {
                showError('Пожалуйста, заполните все поля.');
                return;
            }
            
            if (password !== confirmPassword) {
                showError('Пароли не совпадают.');
                return;
            }
            
            if (!termsAgreed) {
                showError('Вы должны согласиться с условиями и положениями.');
                return;
            }
            
            // Data to send to API
            const userData = {
                name,
                email,
                password
            };
            
            try {
                const response = await fetch('/api/register', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(userData)
                });
                
                if (response.ok) {
                    showSuccess('Регистрация успешна! Переадресация на страницу входа...');
                    setTimeout(() => {
                        window.location.href = 'login.html';
                    }, 2000);
                } else {
                    const errorData = await response.json();
                    showError(errorData.message || 'Ошибка при регистрации. Пожалуйста, попробуйте еще раз.');
                }
            } catch (error) {
                console.error('Error during registration:', error);
                showError('Ошибка при регистрации. Пожалуйста, попробуйте еще раз.');
            }
        });
    }
    
    // Function to show error message
    function showError(message) {
        // Remove any existing error message
        const existingError = document.querySelector('.error-message');
        if (existingError) {
            existingError.remove();
        }
        
        // Create and add new error message
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = message;
        
        const form = document.querySelector('.signup-form');
        form.insertBefore(errorDiv, form.firstChild);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            errorDiv.classList.add('fade-out');
            setTimeout(() => {
                errorDiv.remove();
            }, 500);
        }, 5000);
    }
    
    // Function to show success message
    function showSuccess(message) {
        // Remove any existing messages
        const existingMessages = document.querySelectorAll('.error-message, .success-message');
        existingMessages.forEach(msg => msg.remove());
        
        // Create and add success message
        const successDiv = document.createElement('div');
        successDiv.className = 'success-message';
        successDiv.textContent = message;
        
        const form = document.querySelector('.signup-form');
        form.insertBefore(successDiv, form.firstChild);
    }
});