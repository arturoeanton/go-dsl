// API Service - Motor Contable
// Version: 1.0
// Centralizado para facilitar la migración a API real

// Configuración global
const API_CONFIG = {
    USE_MOCK: false, // Usando backend real del POC
    BASE_URL: '/api/v1',
    MOCK_BASE: '../../mocks/api',
    TIMEOUT: 30000
};

// Cache de catálogo de API
let apiCatalog = null;

// Mapping de endpoints para backend real
const REAL_ENDPOINTS = {
    dashboard: '/dashboard/stats',
    'dashboard.activity': '/dashboard/activity',
    'vouchers.list': '/vouchers',
    'vouchers.detail': '/vouchers/:id',
    'vouchers.create': '/vouchers',
    'vouchers.types': '/vouchers/types',
    'journal_entries.list': '/journal-entries',
    'journal_entries.create': '/journal-entries',
    'accounts.tree': '/accounts/tree',
    'accounts.list': '/accounts',
    'accounts.detail': '/accounts/:code',
    'accounts.types': '/accounts/types',
    'reports.types': '/reports/types',
    'reports.generate': '/reports/generate',
    'reports.recent': '/reports/recent',
    'dsl.templates': '/dsl/templates',
    'dsl.template_detail': '/dsl/templates/:id',
    'dsl.validate': '/dsl/validate',
    'dsl.test': '/dsl/test',
    'organizations.list': '/organizations',
    'organizations.current': '/organizations/current',
    'users.profile': '/users/profile',
    'users.preferences': '/users/preferences',
    'lookups.countries': '/lookups/countries',
    'lookups.currencies': '/lookups/currencies',
    'lookups.tax_types': '/lookups/tax-types',
    'lookups.document_types': '/lookups/document-types',
    'third_parties.search': '/third-parties/search',
    'third_parties.detail': '/third-parties/:id'
};

/**
 * Cargar catálogo de API
 */
async function loadApiCatalog() {
    if (!apiCatalog && API_CONFIG.USE_MOCK) {
        const response = await fetch(`${API_CONFIG.MOCK_BASE}/catalog_api.json`);
        apiCatalog = await response.json();
    }
    return apiCatalog;
}

/**
 * Servicio API centralizado
 */
class ApiService {
    constructor() {
        this.catalog = null;
        this.cache = new Map();
        this.cacheTimeout = 5 * 60 * 1000; // 5 minutos
    }

    /**
     * Inicializar servicio
     */
    async init() {
        if (API_CONFIG.USE_MOCK) {
            this.catalog = await loadApiCatalog();
        }
        return this;
    }

    /**
     * Obtener endpoint del catálogo o mapping real
     */
    getEndpoint(path) {
        if (API_CONFIG.USE_MOCK) {
            const parts = path.split('.');
            let endpoint = this.catalog.endpoints;
            
            for (const part of parts) {
                endpoint = endpoint[part];
                if (!endpoint) {
                    throw new Error(`Endpoint no encontrado: ${path}`);
                }
            }
            
            return endpoint;
        } else {
            // Usar mapping directo para backend real
            const url = REAL_ENDPOINTS[path];
            if (!url) {
                throw new Error(`Endpoint no encontrado: ${path}`);
            }
            return { url, method: 'GET' };
        }
    }

    /**
     * Construir URL
     */
    buildUrl(endpoint, params = {}) {
        if (API_CONFIG.USE_MOCK) {
            return `${API_CONFIG.MOCK_BASE}/${endpoint.mock_file}`;
        }
        
        let url = `${API_CONFIG.BASE_URL}${endpoint.url}`;
        
        // Reemplazar parámetros de ruta
        Object.keys(params).forEach(key => {
            url = url.replace(`:${key}`, params[key]);
        });
        
        return url;
    }

    /**
     * Verificar cache
     */
    checkCache(key) {
        const cached = this.cache.get(key);
        if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
            return cached.data;
        }
        return null;
    }

    /**
     * Guardar en cache
     */
    saveCache(key, data) {
        this.cache.set(key, {
            data,
            timestamp: Date.now()
        });
    }

    /**
     * Realizar petición GET
     */
    async get(endpointPath, params = {}, useCache = true) {
        const endpoint = this.getEndpoint(endpointPath);
        const url = this.buildUrl(endpoint, params);
        const cacheKey = `${endpointPath}:${JSON.stringify(params)}`;
        
        // Verificar cache
        if (useCache) {
            const cached = this.checkCache(cacheKey);
            if (cached) {
                return cached;
            }
        }
        
        try {
            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            // Guardar en cache
            if (useCache) {
                this.saveCache(cacheKey, data);
            }
            
            return data;
        } catch (error) {
            console.error(`Error en GET ${endpointPath}:`, error);
            throw error;
        }
    }

    /**
     * Realizar petición POST
     */
    async post(endpointPath, body = {}, params = {}) {
        const endpoint = this.getEndpoint(endpointPath);
        const url = this.buildUrl(endpoint, params);
        
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                },
                body: JSON.stringify(body)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error(`Error en POST ${endpointPath}:`, error);
            throw error;
        }
    }

    /**
     * Obtener token de autenticación
     */
    getAuthToken() {
        // En producción, obtener de localStorage o sessionStorage
        return 'mock-token-12345';
    }

    /**
     * Limpiar cache
     */
    clearCache() {
        this.cache.clear();
    }
}

// Instancia singleton
let apiServiceInstance = null;

/**
 * Obtener instancia del servicio API
 */
async function getApiService() {
    if (!apiServiceInstance) {
        apiServiceInstance = new ApiService();
        await apiServiceInstance.init();
    }
    return apiServiceInstance;
}

// API específicas para cada módulo

/**
 * Dashboard API
 */
const dashboardApi = {
    async getStats() {
        const api = await getApiService();
        return api.get('dashboard');
    },
    
    async getActivity() {
        const api = await getApiService();
        return api.get('dashboard.activity');
    }
};

/**
 * Vouchers API
 */
const vouchersApi = {
    async getList(filters = {}) {
        const api = await getApiService();
        return api.get('vouchers.list', filters);
    },
    
    async getDetail(id) {
        const api = await getApiService();
        return api.get('vouchers.detail', { id });
    },
    
    async create(data) {
        const api = await getApiService();
        return api.post('vouchers.create', data);
    },
    
    async getTypes() {
        const api = await getApiService();
        return api.get('vouchers.types');
    }
};

/**
 * Journal Entries API
 */
const journalEntriesApi = {
    async getList(filters = {}) {
        const api = await getApiService();
        return api.get('journal_entries.list', filters);
    },
    
    async create(data) {
        const api = await getApiService();
        return api.post('journal_entries.create', data);
    }
};

/**
 * Accounts API
 */
const accountsApi = {
    async getTree() {
        const api = await getApiService();
        return api.get('accounts.tree');
    },
    
    async getList(filters = {}) {
        const api = await getApiService();
        return api.get('accounts.list', filters);
    },
    
    async getDetail(code) {
        const api = await getApiService();
        return api.get('accounts.detail', { code });
    },
    
    async getTypes() {
        const api = await getApiService();
        return api.get('accounts.types');
    }
};

/**
 * Reports API
 */
const reportsApi = {
    async getTypes() {
        const api = await getApiService();
        return api.get('reports.types');
    },
    
    async generate(config) {
        const api = await getApiService();
        return api.post('reports.generate', config);
    },
    
    async getRecent() {
        const api = await getApiService();
        return api.get('reports.recent');
    }
};

/**
 * DSL API
 */
const dslApi = {
    async getTemplates(filters = {}) {
        const api = await getApiService();
        return api.get('dsl.templates', filters);
    },
    
    async getTemplateDetail(id) {
        const api = await getApiService();
        return api.get('dsl.template_detail', { id });
    },
    
    async validate(code) {
        const api = await getApiService();
        return api.post('dsl.validate', { code });
    },
    
    async test(templateId, testData) {
        const api = await getApiService();
        return api.post('dsl.test', { templateId, testData });
    }
};

/**
 * Organizations API
 */
const organizationsApi = {
    async getList() {
        const api = await getApiService();
        return api.get('organizations.list');
    },
    
    async getCurrent() {
        const api = await getApiService();
        return api.get('organizations.current');
    }
};

/**
 * Users API
 */
const usersApi = {
    async getProfile() {
        const api = await getApiService();
        return api.get('users.profile');
    },
    
    async getPreferences() {
        const api = await getApiService();
        return api.get('users.preferences');
    }
};

/**
 * Lookups API
 */
const lookupsApi = {
    async getCountries() {
        const api = await getApiService();
        return api.get('lookups.countries');
    },
    
    async getCurrencies() {
        const api = await getApiService();
        return api.get('lookups.currencies');
    },
    
    async getTaxTypes() {
        const api = await getApiService();
        return api.get('lookups.tax_types');
    },
    
    async getDocumentTypes() {
        const api = await getApiService();
        return api.get('lookups.document_types');
    }
};

/**
 * Third Parties API
 */
const thirdPartiesApi = {
    async search(query) {
        const api = await getApiService();
        return api.get('third_parties.search', { q: query });
    },
    
    async getDetail(id) {
        const api = await getApiService();
        return api.get('third_parties.detail', { id });
    }
};

// Exportar APIs
window.motorContableApi = {
    dashboard: dashboardApi,
    vouchers: vouchersApi,
    journalEntries: journalEntriesApi,
    accounts: accountsApi,
    reports: reportsApi,
    dsl: dslApi,
    organizations: organizationsApi,
    users: usersApi,
    lookups: lookupsApi,
    thirdParties: thirdPartiesApi,
    
    // Utilidades
    clearCache: async () => {
        const api = await getApiService();
        api.clearCache();
    },
    
    setMockMode: (useMock) => {
        API_CONFIG.USE_MOCK = useMock;
    }
};