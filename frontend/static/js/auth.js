
document.addEventListener('DOMContentLoaded', function() {
    console.log("Auth.js script loaded");
    
    // Add token to all fetch requests
    const originalFetch = window.fetch;
    window.fetch = function(url, options = {}) {
        console.log("Fetch called for URL:", url);
        const token = localStorage.getItem('token');
        if (token) {
            console.log("Adding token to request for:", url);
            options.headers = {
                ...options.headers,
                'Authorization': `Bearer ${token}`
            };
        } else {
            console.log("No token found for request:", url);
        }
        return originalFetch(url, options);
    };

    // Get the login form
    const loginForm = document.querySelector('.signup-form');
    console.log("Login form found:", !!loginForm);
 
    if (loginForm) {
        console.log("Adding submit event listener to login form");
        loginForm.addEventListener('submit', async (e) => {
            console.log("Login form submitted");
            e.preventDefault();
            
            const usernameInput = document.getElementById('username');
            const passwordInput = document.getElementById('password');
            const rememberCheckbox = document.querySelector('.remember-checkbox');

            console.log("Form elements found:", {
                username: !!usernameInput,
                password: !!passwordInput,
                remember: !!rememberCheckbox
            });
            alert("Form elements found - Username: " + !!usernameInput + ", Password: " + !!passwordInput); // Test alert

            if (!usernameInput || !passwordInput) {
                console.error("Form elements not found!");
                alert('Ошибка: не найдены поля формы!');
                return;
            }

            const username = usernameInput.value;
            const password = passwordInput.value;
            const remember = rememberCheckbox ? rememberCheckbox.checked : false;

            console.log("Attempting login with username:", username);
            alert("Attempting login with username: " + username); // Test alert
            const requestData = { 
                username, 
                password,
                remember
            };
  
            try {
                console.log("Sending login request...");
                alert("Sending login request..."); // Test alert
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(requestData)
                });
                
                console.log("Login response received, status:", response.status);
                alert("Login response status: " + response.status); // Test alert
                const data = await response.json();
                console.log("Login response data:", data);
                alert("Login response data: " + JSON.stringify(data)); // Test alert
                
                if (response.ok) {
                    // Store token and role in localStorage only
                    localStorage.setItem('token', data.token);
                    localStorage.setItem('role', data.role);
                    
                    console.log("Auth.js: Login successful. Role:", data.role);
                    console.log("Auth.js: Token stored:", data.token ? "Yes" : "No");
                    if (data.redirect) {
                        console.log(data.redirect);
                        alert("йоу йоу йоу");
                        window.location.href = data.redirect;
                    } else {
                        // Optionally, handle other roles here
                    }
                } else {
                    console.error("Login failed:", data);
                    alert("Login failed: " + (data.message || 'Неверные учетные данные'));
                }
            } catch (error) {
                console.error("Error during login:", error);
                alert("Error during login: " + error.message);
            }
        });
    } else {
        console.error("Login form not found on the page!");
        alert("Login form not found on the page!");
    }
});

// Функция для выхода
function logout() {
    // Remove only localStorage items
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    // Redirect to login page
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
            
            // Get form valuesimage.png
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


