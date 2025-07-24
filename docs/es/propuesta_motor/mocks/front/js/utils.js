// Utility Functions - Motor Contable
// Version: 1.0
// Last Updated: 2024-01-15

/**
 * Common utility functions used across the application
 */

// Format functions
function formatCurrency(amount) {
    return new Intl.NumberFormat('es-CO', {
        minimumFractionDigits: 0,
        maximumFractionDigits: 2
    }).format(amount);
}

function formatNumber(num) {
    return new Intl.NumberFormat('es-CO').format(num);
}

function formatDate(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleDateString('es-CO');
}

function formatDateTime(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleString('es-CO');
}

function formatTimeAgo(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    
    if (minutes < 1) return 'hace un momento';
    if (minutes < 60) return `hace ${minutes} minutos`;
    if (hours < 24) return `hace ${hours} horas`;
    return `hace ${days} días`;
}

// Debounce function for search and input handlers
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

// Throttle function for scroll and resize handlers
function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// Local storage helpers
const storage = {
    get: (key) => {
        try {
            const item = localStorage.getItem(key);
            return item ? JSON.parse(item) : null;
        } catch (error) {
            console.error('Error reading from localStorage:', error);
            return null;
        }
    },
    
    set: (key, value) => {
        try {
            localStorage.setItem(key, JSON.stringify(value));
            return true;
        } catch (error) {
            console.error('Error writing to localStorage:', error);
            return false;
        }
    },
    
    remove: (key) => {
        try {
            localStorage.removeItem(key);
            return true;
        } catch (error) {
            console.error('Error removing from localStorage:', error);
            return false;
        }
    },
    
    clear: () => {
        try {
            localStorage.clear();
            return true;
        } catch (error) {
            console.error('Error clearing localStorage:', error);
            return false;
        }
    }
};

// Cookie helpers
const cookies = {
    set: (name, value, days = 7) => {
        const date = new Date();
        date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
        const expires = `expires=${date.toUTCString()}`;
        document.cookie = `${name}=${value};${expires};path=/`;
    },
    
    get: (name) => {
        const nameEQ = name + "=";
        const ca = document.cookie.split(';');
        for(let i = 0; i < ca.length; i++) {
            let c = ca[i];
            while (c.charAt(0) === ' ') c = c.substring(1, c.length);
            if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
        }
        return null;
    },
    
    delete: (name) => {
        document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/;`;
    }
};

// URL parameter helpers
const urlParams = {
    get: (param) => {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get(param);
    },
    
    set: (param, value) => {
        const url = new URL(window.location);
        url.searchParams.set(param, value);
        window.history.pushState({}, '', url);
    },
    
    remove: (param) => {
        const url = new URL(window.location);
        url.searchParams.delete(param);
        window.history.pushState({}, '', url);
    },
    
    getAll: () => {
        const urlParams = new URLSearchParams(window.location.search);
        const params = {};
        for (const [key, value] of urlParams) {
            params[key] = value;
        }
        return params;
    }
};

// Validation helpers
const validate = {
    email: (email) => {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    },
    
    nit: (nit) => {
        // Colombian NIT validation
        const re = /^\d{8,10}-\d$/;
        return re.test(nit);
    },
    
    phone: (phone) => {
        const re = /^[\d\s\-\+\(\)]+$/;
        return re.test(phone) && phone.replace(/\D/g, '').length >= 7;
    },
    
    currency: (amount) => {
        return !isNaN(amount) && parseFloat(amount) >= 0;
    },
    
    required: (value) => {
        return value !== null && value !== undefined && value !== '';
    },
    
    minLength: (value, min) => {
        return value && value.length >= min;
    },
    
    maxLength: (value, max) => {
        return !value || value.length <= max;
    },
    
    between: (value, min, max) => {
        const num = parseFloat(value);
        return !isNaN(num) && num >= min && num <= max;
    }
};

// DOM helpers
const dom = {
    ready: (fn) => {
        if (document.readyState !== 'loading') {
            fn();
        } else {
            document.addEventListener('DOMContentLoaded', fn);
        }
    },
    
    createElement: (tag, attributes = {}, children = []) => {
        const element = document.createElement(tag);
        
        Object.entries(attributes).forEach(([key, value]) => {
            if (key === 'className') {
                element.className = value;
            } else if (key === 'style' && typeof value === 'object') {
                Object.assign(element.style, value);
            } else if (key.startsWith('on') && typeof value === 'function') {
                element.addEventListener(key.substring(2).toLowerCase(), value);
            } else {
                element.setAttribute(key, value);
            }
        });
        
        children.forEach(child => {
            if (typeof child === 'string') {
                element.appendChild(document.createTextNode(child));
            } else if (child instanceof Node) {
                element.appendChild(child);
            }
        });
        
        return element;
    },
    
    show: (element) => {
        if (element) element.style.display = '';
    },
    
    hide: (element) => {
        if (element) element.style.display = 'none';
    },
    
    toggle: (element) => {
        if (element) {
            element.style.display = element.style.display === 'none' ? '' : 'none';
        }
    }
};

// API helpers
const api = {
    get: async (url, options = {}) => {
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    },
    
    post: async (url, data, options = {}) => {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            body: JSON.stringify(data),
            ...options
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    },
    
    put: async (url, data, options = {}) => {
        const response = await fetch(url, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            body: JSON.stringify(data),
            ...options
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    },
    
    delete: async (url, options = {}) => {
        const response = await fetch(url, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    }
};

// Toast notifications
const toast = {
    container: null,
    
    init: () => {
        if (!toast.container) {
            toast.container = dom.createElement('div', {
                className: 'toast-container',
                style: {
                    position: 'fixed',
                    top: '20px',
                    right: '20px',
                    zIndex: '9999'
                }
            });
            document.body.appendChild(toast.container);
        }
    },
    
    show: (message, type = 'info', duration = 3000) => {
        toast.init();
        
        const toastElement = dom.createElement('div', {
            className: `toast toast-${type}`,
            style: {
                padding: '12px 24px',
                marginBottom: '10px',
                borderRadius: '4px',
                backgroundColor: type === 'success' ? '#52c41a' : 
                               type === 'error' ? '#f5222d' : 
                               type === 'warning' ? '#faad14' : '#1890ff',
                color: 'white',
                boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
                animation: 'slideIn 0.3s ease-out'
            }
        }, [message]);
        
        toast.container.appendChild(toastElement);
        
        setTimeout(() => {
            toastElement.style.animation = 'slideOut 0.3s ease-in';
            setTimeout(() => {
                toastElement.remove();
            }, 300);
        }, duration);
    },
    
    success: (message, duration) => toast.show(message, 'success', duration),
    error: (message, duration) => toast.show(message, 'error', duration),
    warning: (message, duration) => toast.show(message, 'warning', duration),
    info: (message, duration) => toast.show(message, 'info', duration)
};

// Loading overlay helpers
const loading = {
    show: (message = 'Cargando...') => {
        let overlay = document.getElementById('globalLoadingOverlay');
        
        if (!overlay) {
            overlay = dom.createElement('div', {
                id: 'globalLoadingOverlay',
                className: 'loading-overlay',
                style: {
                    display: 'flex',
                    position: 'fixed',
                    top: '0',
                    left: '0',
                    width: '100%',
                    height: '100%',
                    backgroundColor: 'rgba(255, 255, 255, 0.9)',
                    zIndex: '9998',
                    justifyContent: 'center',
                    alignItems: 'center',
                    flexDirection: 'column'
                }
            }, [
                dom.createElement('div', { className: 'spinner' }),
                dom.createElement('p', {}, [message])
            ]);
            
            document.body.appendChild(overlay);
        } else {
            overlay.querySelector('p').textContent = message;
            overlay.style.display = 'flex';
        }
    },
    
    hide: () => {
        const overlay = document.getElementById('globalLoadingOverlay');
        if (overlay) {
            overlay.style.display = 'none';
        }
    }
};

// Export utilities for use in other files
window.utils = {
    format: {
        currency: formatCurrency,
        number: formatNumber,
        date: formatDate,
        dateTime: formatDateTime,
        timeAgo: formatTimeAgo
    },
    debounce,
    throttle,
    storage,
    cookies,
    urlParams,
    validate,
    dom,
    api,
    toast,
    loading
};

// Global helper functions
window.showSuccess = (message) => toast.success(message);
window.showError = (message) => toast.error(message);
window.showWarning = (message) => toast.warning(message);
window.showInfo = (message) => toast.info(message);
window.showLoading = (message) => loading.show(message);
window.hideLoading = () => loading.hide();

// Sidebar toggle functionality
const initSidebarToggle = () => {
    const menuToggle = document.getElementById('menuToggle');
    const sidebar = document.getElementById('sidebar');
    const appContainer = document.querySelector('.app-container');
    
    if (menuToggle && sidebar) {
        // Check if sidebar state is saved
        const sidebarCollapsed = storage.get('sidebarCollapsed');
        if (sidebarCollapsed) {
            sidebar.classList.add('collapsed');
            appContainer.classList.add('sidebar-collapsed');
        }
        
        menuToggle.addEventListener('click', () => {
            sidebar.classList.toggle('collapsed');
            appContainer.classList.toggle('sidebar-collapsed');
            
            // Save state
            storage.set('sidebarCollapsed', sidebar.classList.contains('collapsed'));
        });
    }
};

// Initialize common functionality when DOM is ready
dom.ready(() => {
    // Initialize sidebar toggle
    initSidebarToggle();
    
    // Add CSS for animations
    const style = document.createElement('style');
    style.textContent = `
        @keyframes slideIn {
            from { transform: translateX(100%); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }
        
        @keyframes slideOut {
            from { transform: translateX(0); opacity: 1; }
            to { transform: translateX(100%); opacity: 0; }
        }
        
        .toast-container {
            pointer-events: none;
        }
        
        .toast {
            pointer-events: auto;
            cursor: pointer;
        }
        
        /* Additional collapsed styles from utils.js */
        .app-container.sidebar-collapsed .main-content {
            margin-left: 60px;
        }
    `;
    document.head.appendChild(style);
});