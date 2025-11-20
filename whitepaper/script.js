// Ducros Chain Whitepaper - Interactive Features
// Author: Claude AI for Aquí oï
// Version: 1.0.0

// ========================================
// Smooth Scrolling for Navigation Links
// ========================================

document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
            const navHeight = document.querySelector('#navbar').offsetHeight;
            const targetPosition = target.offsetTop - navHeight;

            window.scrollTo({
                top: targetPosition,
                behavior: 'smooth'
            });
        }
    });
});

// ========================================
// Active Navigation State on Scroll
// ========================================

window.addEventListener('scroll', () => {
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('.nav-menu a');
    const navbar = document.querySelector('#navbar');

    // Add/remove shadow to navbar on scroll
    if (window.scrollY > 50) {
        navbar.classList.add('scrolled');
    } else {
        navbar.classList.remove('scrolled');
    }

    // Update active link based on scroll position
    let current = '';
    sections.forEach(section => {
        const sectionTop = section.offsetTop;
        const sectionHeight = section.clientHeight;
        const navHeight = navbar.offsetHeight;

        if (window.scrollY >= (sectionTop - navHeight - 100)) {
            current = section.getAttribute('id');
        }
    });

    navLinks.forEach(link => {
        link.classList.remove('active');
        if (link.getAttribute('href') === `#${current}`) {
            link.classList.add('active');
        }
    });
});

// ========================================
// Scroll Animations (Intersection Observer)
// ========================================

const observerOptions = {
    threshold: 0.1,
    rootMargin: '0px 0px -100px 0px'
};

const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            entry.target.classList.add('visible');
        }
    });
}, observerOptions);

// Observe all sections and cards for animation
document.querySelectorAll('.content-section, .executive-card, .problem-card, .feature-card, .phase-item').forEach(el => {
    el.classList.add('fade-in');
    observer.observe(el);
});

// ========================================
// PDF Download Functionality
// ========================================

document.querySelector('.download-btn').addEventListener('click', () => {
    // Show alert for now (PDF generation would require backend)
    alert('📄 PDF Download\n\nLa version PDF du whitepaper sera disponible prochainement.\n\nVous pouvez sauvegarder cette page en utilisant:\nCtrl+P (Windows) ou Cmd+P (Mac) puis "Enregistrer en PDF".');

    // Optional: Trigger browser print dialog for PDF save
    // window.print();
});

// ========================================
// Mobile Menu Toggle
// ========================================

// Create mobile menu toggle button if not exists
const navbar = document.querySelector('#navbar .nav-container');
const navMenu = document.querySelector('.nav-menu');

// Create hamburger button
const hamburger = document.createElement('button');
hamburger.classList.add('hamburger');
hamburger.innerHTML = `
    <span></span>
    <span></span>
    <span></span>
`;
hamburger.setAttribute('aria-label', 'Toggle navigation menu');

// Insert hamburger before nav menu
if (window.innerWidth <= 768) {
    navbar.insertBefore(hamburger, navMenu);
}

// Toggle mobile menu
hamburger.addEventListener('click', () => {
    navMenu.classList.toggle('active');
    hamburger.classList.toggle('active');
});

// Close mobile menu when clicking outside
document.addEventListener('click', (e) => {
    if (!navbar.contains(e.target) && navMenu.classList.contains('active')) {
        navMenu.classList.remove('active');
        hamburger.classList.remove('active');
    }
});

// Close mobile menu when clicking a link
document.querySelectorAll('.nav-menu a').forEach(link => {
    link.addEventListener('click', () => {
        navMenu.classList.remove('active');
        hamburger.classList.remove('active');
    });
});

// ========================================
// Scroll Indicator Animation
// ========================================

const scrollIndicator = document.querySelector('.scroll-indicator');
if (scrollIndicator) {
    window.addEventListener('scroll', () => {
        if (window.scrollY > 200) {
            scrollIndicator.style.opacity = '0';
        } else {
            scrollIndicator.style.opacity = '1';
        }
    });
}

// ========================================
// Number Animation (Count Up Effect)
// ========================================

function animateNumber(element, target, duration = 2000) {
    const start = 0;
    const increment = target / (duration / 16); // 60fps
    let current = start;

    const timer = setInterval(() => {
        current += increment;
        if (current >= target) {
            current = target;
            clearInterval(timer);
        }
        element.textContent = Math.floor(current).toLocaleString();
    }, 16);
}

// Observe stat numbers for animation
const statsObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
        if (entry.isIntersecting && !entry.target.classList.contains('animated')) {
            const statNumber = entry.target.querySelector('.stat-number');
            if (statNumber) {
                const text = statNumber.textContent;
                const number = parseInt(text.replace(/[^0-9]/g, ''));

                if (!isNaN(number)) {
                    statNumber.textContent = '0';
                    animateNumber(statNumber, number, 1500);
                }

                entry.target.classList.add('animated');
            }
        }
    });
}, { threshold: 0.5 });

document.querySelectorAll('.stat-card').forEach(card => {
    statsObserver.observe(card);
});

// ========================================
// Table Responsive Wrapper
// ========================================

// Wrap all tables in responsive containers for mobile scroll
document.querySelectorAll('table').forEach(table => {
    if (!table.parentElement.classList.contains('table-responsive')) {
        const wrapper = document.createElement('div');
        wrapper.classList.add('table-responsive');
        table.parentNode.insertBefore(wrapper, table);
        wrapper.appendChild(table);
    }
});

// ========================================
// Progress Bar for Allocation Charts
// ========================================

const allocationObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            const bars = entry.target.querySelectorAll('.chart-bar, .alloc-bar');
            bars.forEach((bar, index) => {
                setTimeout(() => {
                    bar.style.animation = `growBar 1s ease-out forwards`;
                }, index * 100);
            });
        }
    });
}, { threshold: 0.3 });

document.querySelectorAll('.distribution-chart, .allocation-chart').forEach(chart => {
    allocationObserver.observe(chart);
});

// ========================================
// Copy to Clipboard for Addresses
// ========================================

document.querySelectorAll('code').forEach(codeElement => {
    // Check if it looks like an address (starts with 0x)
    if (codeElement.textContent.startsWith('0x')) {
        codeElement.style.cursor = 'pointer';
        codeElement.title = 'Cliquer pour copier';

        codeElement.addEventListener('click', () => {
            const text = codeElement.textContent;
            navigator.clipboard.writeText(text).then(() => {
                // Visual feedback
                const original = codeElement.textContent;
                codeElement.textContent = '✓ Copié!';
                codeElement.style.color = '#4caf50';

                setTimeout(() => {
                    codeElement.textContent = original;
                    codeElement.style.color = '';
                }, 1500);
            });
        });
    }
});

// ========================================
// Back to Top Button
// ========================================

const backToTopBtn = document.createElement('button');
backToTopBtn.classList.add('back-to-top');
backToTopBtn.innerHTML = '↑';
backToTopBtn.setAttribute('aria-label', 'Retour en haut');
document.body.appendChild(backToTopBtn);

window.addEventListener('scroll', () => {
    if (window.scrollY > 500) {
        backToTopBtn.classList.add('visible');
    } else {
        backToTopBtn.classList.remove('visible');
    }
});

backToTopBtn.addEventListener('click', () => {
    window.scrollTo({
        top: 0,
        behavior: 'smooth'
    });
});

// ========================================
// Lazy Loading for Images (if any added later)
// ========================================

if ('IntersectionObserver' in window) {
    const imageObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const img = entry.target;
                img.src = img.dataset.src;
                img.classList.add('loaded');
                imageObserver.unobserve(img);
            }
        });
    });

    document.querySelectorAll('img[data-src]').forEach(img => {
        imageObserver.observe(img);
    });
}

// ========================================
// Timeline Animation
// ========================================

const timelineObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            entry.target.style.animation = 'slideInLeft 0.6s ease-out forwards';
        }
    });
}, { threshold: 0.2 });

document.querySelectorAll('.timeline-item, .timeline-phase').forEach(item => {
    timelineObserver.observe(item);
});

// ========================================
// Risk Level Color Coding
// ========================================

document.querySelectorAll('.risk-item').forEach(item => {
    const level = item.classList.contains('high') ? 'high' :
                  item.classList.contains('medium') ? 'medium' : 'low';

    const levelBadge = item.querySelector('.risk-level');
    if (levelBadge) {
        switch(level) {
            case 'high':
                levelBadge.style.backgroundColor = '#e74c3c';
                break;
            case 'medium':
                levelBadge.style.backgroundColor = '#f39c12';
                break;
            case 'low':
                levelBadge.style.backgroundColor = '#27ae60';
                break;
        }
    }
});

// ========================================
// Console Easter Egg
// ========================================

console.log(`
%c██████╗ ██╗   ██╗ ██████╗██████╗  ██████╗ ███████╗
%c██╔══██╗██║   ██║██╔════╝██╔══██╗██╔═══██╗██╔════╝
%c██║  ██║██║   ██║██║     ██████╔╝██║   ██║███████╗
%c██║  ██║██║   ██║██║     ██╔══██╗██║   ██║╚════██║
%c██████╔╝╚██████╔╝╚██████╗██║  ██║╚██████╔╝███████║
%c╚═════╝  ╚═════╝  ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝

%c🚀 The CPU-First French Blockchain
%c⚡ Mining accessible à tous
%c🔗 Website: ducroschain.io
%c📧 Contact: contact@ducroschain.io

%c👨‍💻 Interested in the code? Check out our GitHub:
%c   github.com/Aqui-oi/go-Ducros
`,
'color: #e94560; font-weight: bold;',
'color: #e94560; font-weight: bold;',
'color: #e94560; font-weight: bold;',
'color: #e94560; font-weight: bold;',
'color: #e94560; font-weight: bold;',
'color: #e94560; font-weight: bold;',
'color: #0f3460; font-size: 16px; font-weight: bold;',
'color: #0f3460; font-size: 14px;',
'color: #0f3460; font-size: 14px;',
'color: #0f3460; font-size: 14px;',
'color: #16213e; font-size: 13px;',
'color: #16213e; font-size: 13px; font-weight: bold;'
);

// ========================================
// Initialize on DOM Load
// ========================================

document.addEventListener('DOMContentLoaded', () => {
    console.log('🎉 Ducros Chain Whitepaper loaded successfully!');
    console.log('📊 Version: 1.0.0');
    console.log('🏢 SASU Aquí oï - France');

    // Add loaded class to body for CSS transitions
    document.body.classList.add('loaded');
});

// ========================================
// Performance Monitoring (Development)
// ========================================

if (window.performance) {
    window.addEventListener('load', () => {
        const perfData = window.performance.timing;
        const pageLoadTime = perfData.loadEventEnd - perfData.navigationStart;
        console.log(`⏱️ Page load time: ${pageLoadTime}ms`);
    });
}
