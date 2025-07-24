// Sidebar functionality
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM loaded - initializing sidebar...');
    
    // Initialize sidebar immediately
    initializeSidebar();
    
    // Also try with a small delay as backup
    setTimeout(initializeSidebar, 100);
});

function initializeSidebar() {
    const sidebar = document.getElementById('sidebar');
    const menuToggle = document.querySelector('.menu-toggle') || document.getElementById('menuToggle');
    const mainContent = document.querySelector('.main-content');
    
    console.log('Sidebar initialization attempt:');
    console.log('- Sidebar found:', !!sidebar);
    console.log('- Menu toggle found:', !!menuToggle);
    console.log('- Main content found:', !!mainContent);
    
    if (!sidebar || !menuToggle) {
        console.error('Required elements not found:', {
            sidebar: !!sidebar,
            menuToggle: !!menuToggle
        });
        return;
    }
    
    // Remove any existing listeners to avoid duplicates
    const newMenuToggle = menuToggle.cloneNode(true);
    menuToggle.parentNode.replaceChild(newMenuToggle, menuToggle);
    
    // Add click listener to the new element
    newMenuToggle.addEventListener('click', function(e) {
        e.preventDefault();
        e.stopPropagation();
        
        console.log('Menu toggle clicked!');
        console.log('Current collapsed state:', sidebar.classList.contains('collapsed'));
        
        // Toggle the collapsed class
        sidebar.classList.toggle('collapsed');
        
        // Log the new state
        console.log('New collapsed state:', sidebar.classList.contains('collapsed'));
        console.log('Sidebar classList:', sidebar.classList.toString());
        
        // Save state to localStorage
        const isCollapsed = sidebar.classList.contains('collapsed');
        localStorage.setItem('sidebarCollapsed', isCollapsed);
        
        // Update tooltips
        updateTooltips();
        
        // Manually trigger layout recalculation
        if (mainContent) {
            if (isCollapsed) {
                mainContent.style.marginLeft = '80px';
            } else {
                mainContent.style.marginLeft = '250px';
            }
        }
    });
    
    console.log('Click listener attached to menu toggle');
    
    // Restore sidebar state from localStorage
    const savedState = localStorage.getItem('sidebarCollapsed');
    if (savedState === 'true') {
        sidebar.classList.add('collapsed');
        if (mainContent) {
            mainContent.style.marginLeft = '80px';
        }
        console.log('Restored collapsed state from localStorage');
    }
    
    // Mobile sidebar handling
    if (window.innerWidth < 768) {
        sidebar.classList.add('collapsed');
        if (mainContent) {
            mainContent.style.marginLeft = '80px';
        }
    }
    
    // Handle window resize
    window.addEventListener('resize', function() {
        const sidebar = document.getElementById('sidebar');
        const mainContent = document.querySelector('.main-content');
        if (window.innerWidth < 768 && sidebar) {
            sidebar.classList.add('collapsed');
            if (mainContent) {
                mainContent.style.marginLeft = '80px';
            }
        }
    });
    
    // Add tooltips to collapsed menu
    function updateTooltips() {
        const sidebar = document.getElementById('sidebar');
        const isCollapsed = sidebar && sidebar.classList.contains('collapsed');
        const menuItems = document.querySelectorAll('.sidebar-nav a');
        
        menuItems.forEach(item => {
            const text = item.querySelector('.menu-text');
            if (isCollapsed && text) {
                item.setAttribute('title', text.textContent.trim());
            } else {
                item.removeAttribute('title');
            }
        });
    }
    
    // Initial tooltip setup
    updateTooltips();
}

// Export for debugging
window.initializeSidebar = initializeSidebar;