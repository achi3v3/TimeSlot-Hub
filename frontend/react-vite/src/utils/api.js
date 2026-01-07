// API Configuration and utilities
import axios from 'axios';

// Base API configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || '';

// Create axios instance with default config
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});
// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);
// API endpoints configuration
export const API_ENDPOINTS = {
  // User endpoints
  USER: {
    REGISTER: '/user/register',
    LOGIN: '/user/login',
    LOGOUT: '/user/logout',
    UPDATE: '/user/update',
    GET_PUBLIC: (userId) => `/user/public/${userId}`,
    CHECK_AUTH: (telegramId) => `/user/check/${telegramId}`,
    CHECK_LOGIN: (telegramId) => `/user/check-login/${telegramId}`,
    CONFIRM_LOGIN: (telegramId) => `/user/confirm-login/${telegramId}`,
    CLAIM_TOKEN: (telegramId) => `/user/claim-token/${telegramId}`,
    PUBLIC: (uuid) => `/user/public/${uuid}`,
    CLEAR: '/user/clear',
    REQUEST_DELETION: '/user/request-deletion',
    CONFIRM_DELETION: '/user/confirm-deletion',
  },
  
  // Slot endpoints
  SLOT: {
    CREATE: '/slot/master/create',
    GET_BY_MASTER: (masterId) => `/slot/${masterId}`,
    DELETE_ALL: (masterId) => `/slot/master/${masterId}`,
    DELETE_ONE: (slotId) => `/slot/master/one/${slotId}`,
  },
  
  // Record endpoints
  RECORD: {
    CREATE: '/record/master/create',
    GET_BY_CLIENT: (clientId) => `/record/${clientId}`,
    GET_BY_SLOT: (slotId) => `/record/master/${slotId}`,
    GET_BY_SLOT_ALL: (slotId) => `/record/master/${slotId}`,
    DELETE: (recordId) => `/record/master/${recordId}`,
    CONFIRM: (recordId) => `/record/master/confirm/${recordId}`,
    REJECT: (recordId) => `/record/master/reject/${recordId}`,
    UPDATE_STATUS: '/record/master/status',
  },
  
  // Service endpoints
  SERVICE: {
    CREATE: '/service/create',
    GET_BY_MASTER: (masterId) => `/service/master/${masterId}`,
    GET_BY_ID: (serviceId) => `/service/${serviceId}`,
    UPDATE: '/service/update',
    DELETE: (serviceId) => `/service/${serviceId}`,
  },

  // Notification endpoints
  NOTIFICATION: {
    LIST: '/notification/',
    UNREAD_COUNT: '/notification/unread-count',
    MARK_READ: (id) => `/notification/${id}/mark-read`,
    MARK_ALL_READ: '/notification/mark-all-read',
  },

  // Admin endpoints
  ADMIN: {
    LOGIN: '/admin/login',
    STATS: '/admin/stats',
    USERS: '/admin/users',
    USER_DETAIL: (id) => `/admin/users/${id}`,
    SLOT_DETAIL: (id) => `/admin/slots/${id}`,
    SERVICE_DETAIL: (id) => `/admin/services/${id}`,
    RECORD_DETAIL: (id) => `/admin/records/${id}`,
    DELETE_USER: (id) => `/admin/users/${id}`,
    TOGGLE_USER_ACTIVE: (id) => `/admin/users/${id}/toggle-active`,
    ROLES: '/admin/roles',
    USER_ROLES: (id) => `/admin/users/${id}/roles`,
    CHECK_USER_ROLE: (id, role) => `/admin/users/${id}/roles/${role}`,
  },
};

// API service functions
export const apiService = {
  // User services
  user: {
    register: (userData) => api.post(API_ENDPOINTS.USER.REGISTER, userData),
    login: (phone) => api.post(API_ENDPOINTS.USER.LOGIN, { phone }),
    logout: () => api.post(API_ENDPOINTS.USER.LOGOUT),
    update: (payload) => api.put(API_ENDPOINTS.USER.UPDATE, payload),
    getPublic: (userId) => api.get(API_ENDPOINTS.USER.GET_PUBLIC(userId)),
    checkAuth: (telegramId) => api.get(API_ENDPOINTS.USER.CHECK_AUTH(telegramId)),
    checkLogin: (telegramId) => api.get(API_ENDPOINTS.USER.CHECK_LOGIN(telegramId)),
    confirmLogin: (telegramId) => api.post(API_ENDPOINTS.USER.CONFIRM_LOGIN(telegramId)),
    claimToken: (telegramId) => api.post(API_ENDPOINTS.USER.CLAIM_TOKEN(telegramId)),
    getPublic: (uuid) => api.get(API_ENDPOINTS.USER.PUBLIC(uuid)),
    clear: () => api.delete(API_ENDPOINTS.USER.CLEAR),
    requestDeletion: () => api.post(API_ENDPOINTS.USER.REQUEST_DELETION),
    confirmDeletion: () => api.post(API_ENDPOINTS.USER.CONFIRM_DELETION),
  },
  
  // Slot services
  slot: {
    create: (slotData) => api.post(API_ENDPOINTS.SLOT.CREATE, slotData),
    getByMaster: (masterId) => api.get(API_ENDPOINTS.SLOT.GET_BY_MASTER(masterId)),
    deleteAll: (masterId) => api.delete(API_ENDPOINTS.SLOT.DELETE_ALL(masterId)),
    deleteOne: (slotId) => api.delete(API_ENDPOINTS.SLOT.DELETE_ONE(slotId)),
  },
  
  // Record services
  record: {
    create: (recordData) => api.post(API_ENDPOINTS.RECORD.CREATE, recordData),
    getByClient: (clientId) => api.get(API_ENDPOINTS.RECORD.GET_BY_CLIENT(clientId)),
    getBySlot: (slotId) => api.get(API_ENDPOINTS.RECORD.GET_BY_SLOT(slotId)),
    getBySlotAll: (slotId) => api.get(API_ENDPOINTS.RECORD.GET_BY_SLOT_ALL(slotId)),
    delete: (recordId) => api.delete(API_ENDPOINTS.RECORD.DELETE(recordId)),
    confirm: (recordId) => api.post(API_ENDPOINTS.RECORD.CONFIRM(recordId)),
    reject: (recordId) => api.post(API_ENDPOINTS.RECORD.REJECT(recordId)),
    updateStatus: (recordId, status) => api.post(API_ENDPOINTS.RECORD.UPDATE_STATUS, { record_id: recordId, status }),
  },
  
  // Service services
  service: {
    create: (serviceData) => api.post(API_ENDPOINTS.SERVICE.CREATE, serviceData),
    getByMaster: (masterId) => api.get(API_ENDPOINTS.SERVICE.GET_BY_MASTER(masterId)),
    getById: (serviceId) => api.get(API_ENDPOINTS.SERVICE.GET_BY_ID(serviceId)),
    update: (serviceData) => api.put(API_ENDPOINTS.SERVICE.UPDATE, serviceData),
    delete: (serviceId) => api.delete(API_ENDPOINTS.SERVICE.DELETE(serviceId)),
  },

  // Notification services
  notification: {
    list: () => api.get(API_ENDPOINTS.NOTIFICATION.LIST),
    unreadCount: () => api.get(API_ENDPOINTS.NOTIFICATION.UNREAD_COUNT),
    markRead: (id) => api.post(API_ENDPOINTS.NOTIFICATION.MARK_READ(id)),
    markAllRead: () => api.post(API_ENDPOINTS.NOTIFICATION.MARK_ALL_READ),
  },

  // Admin services
  admin: {
    login: (phone) => api.post(API_ENDPOINTS.ADMIN.LOGIN, { phone }),
    getStats: () => api.get(API_ENDPOINTS.ADMIN.STATS),
    getUsers: (page = 1, limit = 20) => api.get(`${API_ENDPOINTS.ADMIN.USERS}?page=${page}&limit=${limit}`),
    getUserDetail: (id) => api.get(API_ENDPOINTS.ADMIN.USER_DETAIL(id)),
    getSlotDetail: (id) => api.get(API_ENDPOINTS.ADMIN.SLOT_DETAIL(id)),
    getServiceDetail: (id) => api.get(API_ENDPOINTS.ADMIN.SERVICE_DETAIL(id)),
    getRecordDetail: (id) => api.get(API_ENDPOINTS.ADMIN.RECORD_DETAIL(id)),
    deleteUser: (id) => api.delete(API_ENDPOINTS.ADMIN.DELETE_USER(id)),
    toggleUserActive: (id) => api.post(API_ENDPOINTS.ADMIN.TOGGLE_USER_ACTIVE(id)),
    createRole: (user_id, role) => api.post(API_ENDPOINTS.ADMIN.ROLES, { user_id, role }),
    deleteRole: (user_id, role) => api.delete(API_ENDPOINTS.ADMIN.ROLES, { data: { user_id, role } }),
    getAllRoles: () => api.get(API_ENDPOINTS.ADMIN.ROLES),
    getUserRoles: (id) => api.get(API_ENDPOINTS.ADMIN.USER_ROLES(id)),
    checkUserRole: (id, role) => api.get(API_ENDPOINTS.ADMIN.CHECK_USER_ROLE(id, role)),
  },
};

export default api;
