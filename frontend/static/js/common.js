// common.js - Utility functions for the application

// Function to make API calls with proper authentication and business ID
async function apiCall(endpoint, method = 'GET', data = null) {
    const token = localStorage.getItem('token');
    
    // Get the business ID from cookie
    const businessId = getBusinessIdFromCookie();
    
    const options = {
        method,
        headers: {
            'Content-Type': 'application/json',
        },
        credentials: 'include' // Include cookies in the request
    };
    
    // Add token if available
    if (token) {
        options.headers['Authorization'] = `Bearer ${token}`;
    }
    
    // Add business ID to header if available
    if (businessId) {
        options.headers['X-Business-ID'] = businessId;
        console.log(`API call to ${endpoint} with business ID: ${businessId}`);
    } else {
        console.log(`API call to ${endpoint} without business ID`);
    }
    
    if (data) {
        options.body = JSON.stringify(data);
    }
    
    try {
        const response = await fetch(endpoint, options);
        
        // Check if unauthorized (401) or forbidden (403)
        if (response.status === 401) {
            console.log("Unauthorized access (401), redirecting to login");
            // Redirect to login page
            localStorage.removeItem('token');
            localStorage.removeItem('role');
            window.location.href = '/';
            return null;
        }
        
        // Check if no business selected (needs special handling)
        // We need to clone the response before reading text since response body can only be read once
        const responseClone = response.clone();
        if (response.status === 400) {
            const responseText = await responseClone.text();
            if (responseText.includes('business_id')) {
                console.log("Missing business ID in API request");
                const userRole = localStorage.getItem('role');
                
                // Only redirect admins to business selection page
                if (userRole === 'admin') {
                    console.log("User is admin, redirecting to business selection");
                    window.location.href = '/select-business';
                    return null;
                } else {
                    console.log("User is not admin, not redirecting to business selection");
                    // Let the caller handle the error for non-admin users
                    throw new Error("Business ID required but user can't select business");
                }
            }
        }
        
        if (!response.ok) {
            throw new Error(`API error: ${response.status} - ${response.statusText}`);
        }
        
        // For non-JSON responses
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            return await response.json();
        } else {
            return await response.text();
        }
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
}

// Get business ID from cookie
function getBusinessIdFromCookie() {
    const cookies = document.cookie.split(';');
    for (let i = 0; i < cookies.length; i++) {
        const cookie = cookies[i].trim();
        if (cookie.startsWith('business_id=')) {
            const businessId = cookie.substring('business_id='.length);
            console.log(`Found business_id cookie: ${businessId}`);
            return businessId;
        }
    }
    console.log('No business_id cookie found');
    return null;
}

// Set business ID cookie (for testing or manual setting)
function setBusinessIdCookie(businessId) {
    if (!businessId) {
        console.error('Cannot set empty business ID');
        return false;
    }
    
    const expiryDate = new Date();
    expiryDate.setDate(expiryDate.getDate() + 30); // 30 days expiry
    
    document.cookie = `business_id=${businessId};path=/;expires=${expiryDate.toUTCString()};SameSite=Lax`;
    console.log(`Set business_id cookie: ${businessId}`);
    return true;
}

// Check if business is selected, redirect if not
function checkBusinessSelected() {
    const businessId = getBusinessIdFromCookie();
    const userRole = localStorage.getItem('role');
    
    // Only redirect to business selection if user is admin
    if (!businessId && userRole === 'admin') {
        console.log('No business selected and user is admin, redirecting to selection page');
        window.location.href = '/select-business';
        return false;
    } else if (!businessId) {
        console.log('No business selected but user is not admin, not redirecting');
        // For non-admin users, we'll let the backend handle business assignment
        return true;
    }
    return true;
}

// Export for use in other scripts
window.api = {
    call: apiCall,
    getBusinessId: getBusinessIdFromCookie,
    setBusinessId: setBusinessIdCookie,
    checkBusinessSelected: checkBusinessSelected
}; 