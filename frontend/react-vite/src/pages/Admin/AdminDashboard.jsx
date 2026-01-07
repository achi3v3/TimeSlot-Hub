import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import './AdminDashboard.css';
import { useToast } from '../../components/Toast';

const AdminDashboard = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { showError } = useToast();
  const [stats, setStats] = useState(null);
  const [users, setUsers] = useState([]);
  const [slots, setSlots] = useState([]);
  const [services, setServices] = useState([]);
  const [records, setRecords] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [userDetail, setUserDetail] = useState(null);
  const [slotDetail, setSlotDetail] = useState(null);
  const [serviceDetail, setServiceDetail] = useState(null);
  const [recordDetail, setRecordDetail] = useState(null);
  const [showRoleModal, setShowRoleModal] = useState(false);
  const [roleInput, setRoleInput] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [activeList, setActiveList] = useState(null); // 'users' | 'slots' | 'services' | 'records' | null
  const [sortConfig, setSortConfig] = useState({ field: '', direction: 'asc' });

  // отдельные пагинации для слотов, услуг, записей
  const [currentPageSlots, setCurrentPageSlots] = useState(1);
  const [totalPagesSlots, setTotalPagesSlots] = useState(1);
  const [currentPageServices, setCurrentPageServices] = useState(1);
  const [totalPagesServices, setTotalPagesServices] = useState(1);
  const [currentPageRecords, setCurrentPageRecords] = useState(1);
  const [totalPagesRecords, setTotalPagesRecords] = useState(1);
  const [activeListSlots, setActiveListSlots] = useState(null); // 'slots' | null
  const [activeListServices, setActiveListServices] = useState(null); // 'services' | null
  const [activeListRecords, setActiveListRecords] = useState(null); // 'records' | null

  // Page size for all admin lists
  const PAGE_SIZE = 10;

  // Проверка сессии админа по таймеру
  useEffect(() => {
    const until = Number(localStorage.getItem('admin_authed_until'));
    if (!until || Date.now() > until) {
      localStorage.removeItem('admin_token');
      localStorage.removeItem('admin_authed_until');
      navigate('/admin/login');
      return;
    }
    // Таймер для автологаута
    const msLeft = until - Date.now();
    const timeout = setTimeout(() => {
      localStorage.removeItem('admin_token');
      localStorage.removeItem('admin_authed_until');
      navigate('/admin/login');
    }, msLeft);
    setLoading(false); // только после проверки и таймера!
    return () => clearTimeout(timeout);
  }, [navigate]);

  useEffect(() => {
    if (error) showError(error);
  }, [error, showError]);

  useEffect(() => {
      fetchStats();
  }, []);

  // Управление прокруткой страницы при открытии деталей
  useEffect(() => {
    if (userDetail || slotDetail || serviceDetail || recordDetail) {
      document.body.classList.add('detail-open');
    } else {
      document.body.classList.remove('detail-open');
    }
    
    return () => {
      document.body.classList.remove('detail-open');
    };
  }, [userDetail, slotDetail, serviceDetail, recordDetail]);

  const fetchStats = async () => {
    try {
      const response = await fetch('/api/admin/stats', {
        headers: adminHeaders(),
      });
      if (response.ok) {
        const data = await response.json();
        setStats(data.stats || data);
      } else {
        const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
        setError(`Ошибка загрузки статистики: ${errorData.error || response.statusText}`);
      }
    } catch (err) {
      setError('Ошибка соединения с сервером при загрузке статистики');
    } finally {
      setLoading(false);
    }
  };

  
  const refreshAllData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Обновляем статистику
      await fetchStats();
      
      // Обновляем текущий список в зависимости от активного списка
      if (activeList === 'users') {
        await fetchUsers(currentPage);
      } else if (activeList === 'slots') {
        await fetchAllSlots();
      } else if (activeList === 'services') {
        await fetchAllServices();
      } else if (activeList === 'records') {
        await fetchAllRecords();
      }
      
      // Если открыты детали пользователя, обновляем их тоже
      if (selectedUser) {
        await fetchUserDetail(selectedUser);
      }
    } catch (err) {
      setError('Ошибка обновления данных');
    } finally {
      setLoading(false);
    }
  };

  const fetchUsers = async (page = 1) => {
    try {
      const response = await fetch(`/api/admin/users?page=${page}&limit=${PAGE_SIZE}`, {
        headers: adminHeaders(),
      });

      if (response.ok) {
        const data = await response.json();
        setUsers(Array.isArray(data.users) ? data.users : (Array.isArray(data.items) ? data.items : []));
        const total = Number(data.total) || (Array.isArray(data.users) ? data.users.length : 0);
        setTotalPages(Math.max(1, Math.ceil(total / PAGE_SIZE)));
        setCurrentPage(page);
      } else {
        setError('Ошибка загрузки пользователей');
      }
    } catch (err) {
      setError('Ошибка соединения с сервером');
    } finally {
      setLoading(false);
    }
  };
  const fetchAllSlots = async (page = 1) => {
    try {
      const response = await fetch(`/api/admin/slots?page=${page}&limit=${PAGE_SIZE}`, { headers: adminHeaders() });
      if (response.ok) {
        const data = await response.json();
        // ожидаем массив; если бек вернёт total - обновим, иначе считаем 1 страницу
        const list = Array.isArray(data) ? data : (data.items || []);
        const total = data.total || list.length;
        setSlots(list);
        setTotalPagesSlots(Math.max(1, Math.ceil(total / PAGE_SIZE)));
        setCurrentPageSlots(page);
      } else {
        setError('Ошибка загрузки слотов');
      }
    } catch (err) {
      setError('Ошибка соединения с сервером');
    }
  };

  const fetchAllServices = async (page = 1) => {
    try {
      const response = await fetch(`/api/admin/services?page=${page}&limit=${PAGE_SIZE}`, { headers: adminHeaders() });
      if (response.ok) {
        const data = await response.json();
        const list = Array.isArray(data) ? data : (data.items || []);
        const total = data.total || list.length;
        setServices(list);
        setTotalPagesServices(Math.max(1, Math.ceil(total / PAGE_SIZE)));
        setCurrentPageServices(page);
      } else {
        setError('Ошибка загрузки услуг');
      }
    } catch (err) {
      setError('Ошибка соединения с сервером');
    }
  };

  const fetchAllRecords = async (page = 1) => {
    try {
      const response = await fetch(`/api/admin/records?page=${page}&limit=${PAGE_SIZE}`, { headers: adminHeaders() });
      if (response.ok) {
        const data = await response.json();
        const list = Array.isArray(data) ? data : (data.items || []);
        const total = data.total || list.length;
        setRecords(list);
        setTotalPagesRecords(Math.max(1, Math.ceil(total / PAGE_SIZE)));
        setCurrentPageRecords(page);
      } else {
        setError('Ошибка загрузки записей');
      }
    } catch (err) {
      setError('Ошибка соединения с сервером');
    }
  };
  const openList = (key) => {
    setActiveList(key);
    setSelectedUser(null);
    setUserDetail(null);
    if (key === 'users') fetchUsers(1);
    if (key === 'slots') {
      fetchAllSlots(currentPageSlots);
    }
    if (key === 'services') {
      fetchAllServices(currentPageServices);
    }
    if (key === 'records') {
      fetchAllRecords(currentPageRecords);
    }
  };

  // Проверка активности админ-сессии по таймеру
  const isAdminSessionActive = () => {
    const until = Number(localStorage.getItem('admin_authed_until'));
    return Boolean(until && Date.now() < until);
  };

  const adminHeaders = () => {
    const headers = {};
    headers['Content-Type'] = 'application/json';
    // Используем JWT токен для авторизации в админке
    const token = localStorage.getItem('admin_token');
    if (token && isAdminSessionActive()) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
  };

  // Переключение табов: УДАЛЕНО. Рендерим по activeList
  // function handleTabChange(tabKey) {
  //   setActiveTab(tabKey);
  //   openList(tabKey);
  // }

  const deleteUser = async (userId) => {
    try {
      const res = await fetch(`/api/admin/users/${userId}`, {
        method: 'DELETE',
        headers: adminHeaders(),
      });
      if (!res.ok) throw new Error('delete-failed');
      if (activeList === 'users') fetchUsers(currentPage);
      if (selectedUser === userId) setUserDetail(null);
      fetchStats();
    } catch (e) {
      setError('Не удалось удалить пользователя');
    }
  };

  const toggleUserActive = async (userId) => {
    try {
      const res = await fetch(`/api/admin/users/${userId}/toggle-active`, {
        method: 'POST',
        headers: adminHeaders(),
      });
      if (!res.ok) throw new Error('toggle-failed');
      if (activeList === 'users') fetchUsers(currentPage);
      if (selectedUser === userId) fetchUserDetail(userId);
      fetchStats();
    } catch (e) {
      setError('Не удалось изменить статус пользователя');
    }
  };

  const addUserRole = async (userId, role) => {
    try {
      const body = JSON.stringify({ user_id: userId, role: role });
      const res = await fetch(`/api/admin/roles`, {
        method: 'POST',
        headers: adminHeaders(),
        body,
      });
      if (!res.ok) throw new Error('role-failed');
      if (activeList === 'users') fetchUsers(currentPage);
      if (selectedUser === userId) fetchUserDetail(userId);
    } catch (e) {
      setError('Не удалось добавить роль');
    }
  };

  const removeUserRole = async (userId, role) => {
    try {
      const body = JSON.stringify({ user_id: userId, role: role });
      const res = await fetch(`/api/admin/roles`, {
        method: 'DELETE',
        headers: adminHeaders(),
        body,
      });
      if (!res.ok) throw new Error('role-delete-failed');
      if (activeList === 'users') fetchUsers(currentPage);
      if (selectedUser === userId) fetchUserDetail(userId);
    } catch (e) {
      setError('Не удалось удалить роль');
    }
  };

  const deleteSlot = async (slotId) => {
    try {
      const res = await fetch(`/api/admin/slots/${slotId}`, {
        method: 'DELETE',
        headers: adminHeaders(),
      });
      if (!res.ok) throw new Error('delete-slot-failed');
      if (activeList === 'slots') fetchAllSlots();
      if (userDetail) fetchUserDetail(userDetail.user.id);
      fetchStats();
    } catch (e) {
      setError('Не удалось удалить слот');
    }
  };

  const fetchUserDetail = async (userId) => {
    try {
      const response = await fetch(`/api/admin/users/${userId}`, {
        headers: adminHeaders(),
      });

      if (response.ok) {
        const data = await response.json();
        console.log('User detail response:', data); // Добавим логирование для отладки
        // Бэкенд возвращает данные в правильном формате
        setUserDetail(data);
        setSelectedUser(userId);
        // Закрываем другие детали
        setSlotDetail(null);
        setServiceDetail(null);
        setRecordDetail(null);
      } else {
        setError('Ошибка загрузки деталей пользователя');
      }
    } catch (err) {
      setError('Ошибка соединения с сервером');
    }
  };

  const fetchSlotDetail = async (slotId) => {
    try {
      const response = await fetch(`/api/admin/slots/${slotId}`, {
        headers: adminHeaders(),
      });
      if (response.ok) {
        const slotData = await response.json();
        console.log('Slot detail response:', slotData); 
        // Закрываем другие детали
        setUserDetail(null);
        setServiceDetail(null);
        setRecordDetail(null);
        setSlotDetail(slotData);
      } else{
        setError('Ошибка загрузки деталей слота');
      } 
      
    } catch (error) {
      setError('Ошибка соединения с сервером');
    }
  };

  const fetchServiceDetail = async (serviceId) => {
    try {
      const response = await fetch(`/api/admin/services/${serviceId}`, {
        headers: adminHeaders(),
      });
      if (response.ok) {
        const serviceData = await response.json();
        // Закрываем другие детали
        setUserDetail(null);
        setSlotDetail(null);
        setRecordDetail(null);
        setServiceDetail(serviceData);
      } else {
      setError('Не удалось загрузить детали услуги');
    }
    }catch (error) {
      setError('Ошибка соединения с сервером');
    }
  };

  const fetchRecordDetail = async (recordId) => {
    try {
      const response = await fetch(`/api/admin/records/${recordId}`, {
        headers: adminHeaders(),
      });
      if (response.ok) {
      const recordData = await response.json();
      // Закрываем другие детали
      setUserDetail(null);
      setSlotDetail(null);
      setServiceDetail(null);
      setRecordDetail(recordData);
      } else {
        setError('Ошибка загрузки деталей записей');
      }
    } catch (error) {
      setError('Ошибка соединения с сервером');
    }
  };
  const handleAddRole = () => {
    if (roleInput.trim()) {
      addUserRole(userDetail.user.id, roleInput.trim());
      setShowRoleModal(false);
      setRoleInput('');
    }
  };

  const handleCloseRoleModal = () => {
    setShowRoleModal(false);
    setRoleInput('');
  };

  const adminLogout = () => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_authed_until');
    navigate('/admin/login');
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  function sortData(data, field) {
    if (!field) return data;
    const { direction } = sortConfig;
    return [...data].sort((a, b) => {
      if (a[field] == null) return 1;
      if (b[field] == null) return -1;
      if (typeof a[field] === 'number') {
        return direction === 'asc' ? a[field] - b[field] : b[field] - a[field];
      }
      // строковое сравнение
      return direction === 'asc'
        ? String(a[field]).localeCompare(String(b[field]), 'ru')
        : String(b[field]).localeCompare(String(a[field]), 'ru');
    });
  }

  function handleSort(field) {
    if (sortConfig.field === field) {
      setSortConfig({ field, direction: sortConfig.direction === 'asc' ? 'desc' : 'asc' });
    } else {
      setSortConfig({ field, direction: 'asc' });
    }
  }

  function headerClass(field) {
    const classes = ['sortable'];
    if (sortConfig.field === field) {
      classes.push('active');
      classes.push(sortConfig.direction);
    }
    return classes.join(' ');
  }

  if (loading) {
    return (
      <div className="admin-dashboard">
        <div className="loading">Загрузка...</div>
      </div>
    );
  }

  // Build paginated visible lists for rendering (client-side fallback)
  const visibleUsers = sortData(users, sortConfig.field).slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE);
  const visibleSlots = sortData(slots, sortConfig.field).slice((currentPageSlots - 1) * PAGE_SIZE, currentPageSlots * PAGE_SIZE);
  const visibleServices = sortData(services, sortConfig.field).slice((currentPageServices - 1) * PAGE_SIZE, currentPageServices * PAGE_SIZE);
  const visibleRecords = sortData(records, sortConfig.field).slice((currentPageRecords - 1) * PAGE_SIZE, currentPageRecords * PAGE_SIZE);

  return (
    <div className="admin-dashboard">
      <header className="admin-header">
        <h1>ПАНЕЛЬ АДМИНИСТРАТОРА</h1>
        <div className="header-actions">
          <button onClick={refreshAllData} className="refresh-button" title="Обновить">
            Обновить
          </button>
        </div>
      </header>

      {/* Статистика */}
      {stats && (
        <section className="stats-section">
          <h2>Общая статистика</h2>
          <div className="stats-grid">
            <div className="stat-card" onClick={() => openList('users')} style={{ cursor: 'pointer' }}>
              <div className="stat-number">{stats.total_users}</div>
              <div className="stat-label">Всего пользователей</div>
            </div>
            <div className="stat-card" onClick={() => openList('users')} style={{ cursor: 'pointer' }}>
              <div className="stat-number">{stats.active_users}</div>
              <div className="stat-label">Активных пользователей</div>
            </div>
            <div className="stat-card" onClick={() => openList('slots')} style={{ cursor: 'pointer' }}>
              <div className="stat-number">{stats.total_slots}</div>
              <div className="stat-label">Всего слотов</div>
            </div>
            <div className="stat-card" onClick={() => openList('slots')} style={{ cursor: 'pointer' }}>
              <div className="stat-number">{stats.booked_slots}</div>
              <div className="stat-label">Забронированных слотов</div>
            </div>
            <div className="stat-card" onClick={() => openList('records')} style={{ cursor: 'pointer' }}>
              <div className="stat-number">{stats.total_records}</div>
              <div className="stat-label">Всего записей</div>
            </div>
            <div className="stat-card" onClick={() => openList('records')} style={{ cursor: 'pointer' }}>
              <div className="stat-number">{stats.pending_records}</div>
              <div className="stat-label">Ожидающих подтверждения</div>
            </div>
            <div className="stat-card" onClick={() => openList('records')} style={{ cursor: 'pointer' }}>
              <div className="stat-number">{stats.confirmed_records}</div>
              <div className="stat-label">Подтвержденных записей</div>
            </div>
            <div className="stat-card" onClick={() => openList('services')} style={{ cursor: 'pointer' }}>
              <div className="stat-number">{stats.total_services}</div>
              <div className="stat-label">Всего услуг</div>
            </div>
            <div className="stat-card">
              <div className="stat-number">{stats.ad_clicks_1 ?? 0}</div>
              <div className="stat-label">Клики по рекламе 1</div>
            </div>
            <div className="stat-card">
              <div className="stat-number">{stats.ad_clicks_2 ?? 0}</div>
              <div className="stat-label">Клики по рекламе 2</div>
            </div>
          </div>
        </section>
      )}

      {/* Удалены admin-tabs. Рендерим списки исходя из activeList */}
      {activeList==='users' && (
      <section className="users-section">
        <h2>Пользователи</h2>
        <div className="users-table-container">
          <table className="users-table">
            <thead>
              <tr>
                  <th className={headerClass('first_name')} onClick={()=>handleSort('first_name')}>Имя</th>
                  <th className={headerClass('phone')} onClick={()=>handleSort('phone')}>Телефон</th>
                  <th className={headerClass('telegram_id')} onClick={()=>handleSort('telegram_id')}>Telegram ID</th>
                <th>Админка</th>
                  <th className={headerClass('roles')} onClick={()=>handleSort('roles')}>Роли</th>
                  <th className={headerClass('created_at')} onClick={()=>handleSort('created_at')}>Дата регистрации</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
                {visibleUsers.map((user) => (
                <tr key={user.id}>
                  <td>{user.first_name} {user.surname}</td>
                  <td>{user.phone}</td>
                  <td>{user.telegram_id}</td>
                  <td>
                    <span className={`status-badge ${user.is_active ? 'active' : 'inactive'}`}>
                      {user.is_active ? 'Активна' : 'Неактивна'}
                    </span>
                  </td>
                    <td>{Array.isArray(user.roles) ? user.roles.join(', ') : ''}</td>
                  <td>{formatDate(user.created_at)}</td>
                  <td>
                    <button
                      onClick={() => fetchUserDetail(user.id)}
                      className="detail-button"
                    >
                      Подробнее
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="pagination">
            <button onClick={() => fetchUsers(currentPage - 1)} disabled={currentPage === 1} className="page-button">Назад</button>
            <span className="page-info">Страница {currentPage} из {totalPages}</span>
            <button onClick={() => fetchUsers(currentPage + 1)} disabled={currentPage === totalPages} className="page-button">Вперед</button>
          </div>
        </section>
      )}

      {activeList==='slots' && (
        <section className="users-section">
          <h2>Слоты</h2>
          <div className="users-table-container">
            <table className="slots-table">
              <thead>
                <tr>
                  <th className={headerClass('start_time')} onClick={()=>handleSort('start_time')}>Дата и Время начала</th>
                  <th className={headerClass('end_time')} onClick={()=>handleSort('end_time')}>Время окончания</th>
                  <th className={headerClass('is_booked')} onClick={()=>handleSort('is_booked')}>Статус</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {visibleSlots.map((slot) => (
                  <tr key={slot.id}>
                    <td>{new Date(slot.start_time).toLocaleDateString('ru-RU')}</td>
                    <td>
                      {new Date(slot.start_time).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })} - 
                      {new Date(slot.end_time).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                    </td>
                    <td>
                      <span className={`slot-status ${slot.is_booked ? 'booked' : 'available'}`}>
                        {slot.is_booked ? 'Забронирован' : 'Свободен'}
                      </span>
                    </td>
                    <td>
          <button
                        onClick={() => fetchSlotDetail(slot.id)}
                      className="detail-button"
          >
                      Подробнее
          </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
          <div className="pagination">
            <button onClick={() => fetchAllSlots(currentPageSlots - 1)} disabled={currentPageSlots === 1} className="page-button">Назад</button>
            <span className="page-info">Страница {currentPageSlots} из {totalPagesSlots}</span>
            <button onClick={() => fetchAllSlots(currentPageSlots + 1)} disabled={currentPageSlots === totalPagesSlots} className="page-button">Вперед</button>
        </div>
        </section>
      )}

      {activeList==='services' && (
        <section className="users-section">
          <h2>Услуги</h2>
          <div className="services-table-container">
            <table className="services-table">
              <thead>
                <tr>
                  <th className={headerClass('name')} onClick={()=>handleSort('name')}>Название</th>
                  <th className={headerClass('price')} onClick={()=>handleSort('price')}>Цена</th>
                  <th className={headerClass('duration')} onClick={()=>handleSort('duration')}>Длительность</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {visibleServices.map((service) => (
                  <tr key={service.id}>
                    <td>{service.name}</td>
                    <td>{service.price} руб.</td>
                    <td>{service.duration} мин.</td>
                    <td>
          <button
                        onClick={() => fetchServiceDetail(service.id)}
                        className="detail-button"
          >
                        Подробнее
          </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="pagination">
            <button onClick={() => fetchAllServices(currentPageServices - 1)} disabled={currentPageServices === 1} className="page-button">Назад</button>
            <span className="page-info">Страница {currentPageServices} из {totalPagesServices}</span>
            <button onClick={() => fetchAllServices(currentPageServices + 1)} disabled={currentPageServices === totalPagesServices} className="page-button">Вперед</button>
          </div>
        </section>
      )}

      {activeList==='records' && (
        <section className="users-section">
          <h2>Записи</h2>
          <div className="records-table-container">
            <table className="records-table">
              <thead>
                <tr>
                  <th className={headerClass('id')} onClick={()=>handleSort('id')}>ID</th>
                  <th className={headerClass('status')} onClick={()=>handleSort('status')}>Статус</th>
                  <th className={headerClass('created_at')} onClick={()=>handleSort('created_at')}>Дата создания</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {visibleRecords.map((record) => (
                  <tr key={record.id}>
                    <td>{record.id}</td>
                    <td>
                      <span className={`record-status ${record.status}`}>
                        {record.status === 'confirm' ? 'Подтверждена' : record.status === 'pending' ? 'Ожидает' : 'Отклонена'}
          </span>
                    </td>
                    <td>{new Date(record.created_at).toLocaleString('ru-RU')}</td>
                    <td>
          <button
                        onClick={() => fetchRecordDetail(record.id)}
                        className="detail-button"
          >
                        Подробнее
          </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
        </div>
          <div className="pagination">
            <button onClick={() => fetchAllRecords(currentPageRecords - 1)} disabled={currentPageRecords === 1} className="page-button">Назад</button>
            <span className="page-info">Страница {currentPageRecords} из {totalPagesRecords}</span>
            <button onClick={() => fetchAllRecords(currentPageRecords + 1)} disabled={currentPageRecords === totalPagesRecords} className="page-button">Вперед</button>
        </div>
      </section>
      )}

      {/* Детали пользователя: нижний выезжающий дроуер */}
      {userDetail && (
        <section className="user-detail-drawer">
          <div className="user-detail-header">
          <h2>Детали пользователя: {userDetail.user.first_name} {userDetail.user.surname}</h2>
          
            <button onClick={() => { setUserDetail(null); setSelectedUser(null); }} className="close-detail-button">Закрыть</button>
          </div>
          
          <div className="user-detail-grid">
            {/* Основная информация */}
            <div className="detail-card compact-card">
              <h3>Основная информация</h3>
              
              <div className="detail-item-row">
                <span className="detail-label">Имя:</span>
                <span className="detail-value">{userDetail.user.first_name} {userDetail.user.surname}</span>
              </div>
              
              <div className="detail-item-row">
                <span className="detail-label">ID:</span>
                <div className="detail-value-with-copy">
                  <span className="detail-value copyable" title={userDetail.user.id}>{userDetail.user.id}</span>
                  <button 
                    className="copy-button"
                    onClick={() => {
                      navigator.clipboard.writeText(userDetail.user.id);
                      // Можно добавить уведомление о копировании
                    }}
                    title="Копировать ID"
                  >
                    📋
                  </button>
              </div>
              </div>
              
              <div className="detail-item-row">
                <span className="detail-label">Телефон:</span>
                <span className="detail-value">{userDetail.user.phone}</span>
              </div>
              
              <div className="detail-item-row">
                <span className="detail-label">Telegram ID:</span>
                <span className="detail-value">{userDetail.user.telegram_id}</span>
              </div>
              
              <div className="detail-item-row">
                <span className="detail-label">Статус:</span>
                <div className="detail-value-with-action">
                <span className={`status-badge ${userDetail.user.is_active ? 'active' : 'inactive'}`}>
                  {userDetail.user.is_active ? 'Активен' : 'Неактивен'}
                </span>
                  <button
                    onClick={() => toggleUserActive(userDetail.user.id)}
                    className="action-button"
                    title={userDetail.user.is_active ? 'Деактивировать' : 'Активировать'}
                  >
                    {userDetail.user.is_active ? '-' : '+'}
                  </button>
              </div>
              </div>
              
              <div className="detail-item-row">
                <span className="detail-label">Роли:</span>
                <div className="detail-value-with-action">
                  <div className="roles-container">
                    {userDetail.user.roles && userDetail.user.roles.length > 0 ? (
                      userDetail.user.roles.map((role, index) => (
                        <span key={index} className="role-badge">
                          {role}
                          <button 
                            onClick={() => removeUserRole(userDetail.user.id, role)}
                            className="role-remove-button"
                            title="Удалить роль"
                          >
                            ×
                          </button>
                        </span>
                      ))
                    ) : (
                      <span className="no-roles">Нет ролей</span>
                    )}
                  </div>
                  <button
                    onClick={() => setShowRoleModal(true)}
                    className="action-button"
                    title="Добавить роль"
                  >
                    +
                  </button>
                </div>
              </div>
            {/* Кнопка удаления пользователя - отдельно */}
            <div className="detail-card danger-card">
              <button 
                onClick={() => { 
                  if (window.confirm('Удалить пользователя без возможности восстановления?')) 
                    deleteUser(userDetail.user.id); 
                }} 
                className="delete-user-button"
              >
                Удалить пользователя
              </button>
              </div>
            </div>
            
            

            {/* Слоты */}
            <div className="detail-card">
              <h3>Слоты ({userDetail.slots?.length || 0})</h3>
              {userDetail.slots && Array.isArray(userDetail.slots) && userDetail.slots.length > 0 ? (
                <div className="user-slots-table-container">
                  <table className="user-slots-table">
                    <thead>
                      <tr>
                        <th>Дата</th>
                        <th>Время</th>
                        <th>Статус</th>
                      </tr>
                    </thead>
                    <tbody>
                      {userDetail.slots.map((slot) => (
                        <tr key={slot.id}>
                          <td>{new Date(slot.start_time).toLocaleDateString('ru-RU')}</td>
                          <td>
                            {new Date(slot.start_time).toLocaleTimeString('ru-RU', { 
                              hour: '2-digit', 
                              minute: '2-digit' 
                            })} - 
                            {new Date(slot.end_time).toLocaleTimeString('ru-RU', { 
                              hour: '2-digit', 
                              minute: '2-digit' 
                            })}
                          </td>
                          <td>
                            <span className={`slot-status ${slot.is_booked ? 'booked' : 'available'}`}>
                              {slot.is_booked ? 'Забронирован' : 'Свободен'}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="no-data">Нет слотов</div>
              )}
            </div>

            {/* Услуги */}
            <div className="detail-card">
              <h3>Услуги ({userDetail.services?.length || 0})</h3>
              {userDetail.services && Array.isArray(userDetail.services) && userDetail.services.length > 0 ? (
                <div className="user-services-table-container">
                  <table className="user-services-table">
                    <thead>
                      <tr>
                        <th>Название</th>
                        <th>Цена</th>
                        <th>Длительность</th>
                      </tr>
                    </thead>
                    <tbody>
                      {userDetail.services.map((service) => (
                        <tr key={service.id}>
                          <td>{service.name}</td>
                          <td>{service.price} руб.</td>
                          <td>{service.duration} мин.</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="no-data">Нет услуг</div>
              )}
            </div>

            {/* Записи */}
            <div className="detail-card">
              <h3>Записи ({userDetail.records?.length || 0})</h3>
              {userDetail.records && Array.isArray(userDetail.records) && userDetail.records.length > 0 ? (
                <div className="user-records-table-container">
                  <table className="user-records-table">
                    <thead>
                      <tr>
                        <th>ID</th>
                        <th>Статус</th>
                        <th>Дата создания</th>
                      </tr>
                    </thead>
                    <tbody>
                      {userDetail.records.map((record) => (
                        <tr key={record.id}>
                          <td>{record.id}</td>
                          <td>
                            <span className={`record-status ${record.status}`}>
                        {record.status === 'confirm' ? 'Подтверждена' : 
                         record.status === 'pending' ? 'Ожидает' : 'Отклонена'}
                            </span>
                          </td>
                          <td>{new Date(record.created_at).toLocaleString('ru-RU')}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="no-data">Нет записей</div>
              )}
            </div>
          </div>

        </section>
      )}

      {/* Детали слота: нижний выезжающий дроуер */}
      {slotDetail && (
        <section className="user-detail-drawer">
          <div className="user-detail-header">
            <h2>Детали слота: {slotDetail.id}</h2>
            <button onClick={() => { setSlotDetail(null); }} className="close-detail-button">Закрыть</button>
          </div>
          
          <div className="user-detail-grid">
            {/* Основная информация о слоте */}
            <div className="detail-card">
              <h3>Информация о слоте</h3>
              <div className="detail-item">
                <strong>ID:</strong> {slotDetail.id}
              </div>
              <div className="detail-item">
                <strong>Время начала:</strong> {new Date(slotDetail.start_time).toLocaleString('ru-RU')}
              </div>
              <div className="detail-item">
                <strong>Время окончания:</strong> {new Date(slotDetail.end_time).toLocaleString('ru-RU')}
              </div>
              <div className="detail-item">
                <strong>Статус:</strong> 
                <span className={`status-badge ${slotDetail.is_booked ? 'booked' : 'available'}`}>
                  {slotDetail.is_booked ? 'Забронирован' : 'Свободен'}
                </span>
              </div>
            </div>

            {/* Информация об услуге */}
            <div className="detail-card">
              <h3>Услуга</h3>
              <div className="detail-item">
                <strong>Название:</strong> {slotDetail.service_name}
              </div>
              <div className="detail-item">
                <strong>Описание:</strong> {slotDetail.service_description}
              </div>
              <div className="detail-item">
                <strong>Цена:</strong> {slotDetail.service_price} руб.
              </div>
              <div className="detail-item">
                <strong>Длительность:</strong> {slotDetail.service_duration} мин.
              </div>
            </div>

            {/* Информация о мастере */}
            <div className="detail-card">
              <h3>Мастер</h3>
              <div className="detail-item">
                <strong>Имя:</strong> {slotDetail.master_name} {slotDetail.master_surname}
              </div>
              <div className="detail-item">
                <strong>Телефон:</strong> {slotDetail.master_phone}
              </div>
              <div className="detail-item">
                <strong>Telegram ID:</strong> {slotDetail.master_telegram_id}
              </div>
            </div>
          </div>
        </section>
      )}

      {/* Детали услуги: нижний выезжающий дроуер */}
      {serviceDetail && (
        <section className="user-detail-drawer">
          <div className="user-detail-header">
            <h2>Детали услуги: {serviceDetail.name}</h2>
            <button onClick={() => { setServiceDetail(null); }} className="close-detail-button">Закрыть</button>
          </div>
          
          <div className="user-detail-grid">
            {/* Основная информация об услуге */}
            <div className="detail-card">
              <h3>Информация об услуге</h3>
              <div className="detail-item">
                <strong>ID:</strong> {serviceDetail.id}
              </div>
              <div className="detail-item">
                <strong>Название:</strong> {serviceDetail.name}
              </div>
              <div className="detail-item">
                <strong>Цена:</strong> {serviceDetail.price} руб.
              </div>
              <div className="detail-item">
                <strong>Длительность:</strong> {serviceDetail.duration} мин.
              </div>
              <div className="detail-item">
                <strong>Описание:</strong> {serviceDetail.description}
              </div>
            </div>

            {/* Информация о мастере */}
            <div className="detail-card">
              <h3>Мастер</h3>
              <div className="detail-item">
                <strong>Имя:</strong> {serviceDetail.master_name} {serviceDetail.master_surname}
              </div>
              <div className="detail-item">
                <strong>Телефон:</strong> {serviceDetail.master_phone}
              </div>
              <div className="detail-item">
                <strong>Telegram ID:</strong> {serviceDetail.master_telegram_id}
              </div>
            </div>
          </div>
        </section>
      )}

      {/* Детали записи: нижний выезжающий дроуер */}
      {recordDetail && (
        <section className="user-detail-drawer">
          <div className="user-detail-header">
            <h2>Детали записи: {recordDetail.id}</h2>
            <button onClick={() => { setRecordDetail(null); }} className="close-detail-button">Закрыть</button>
          </div>
          
          <div className="user-detail-grid">
            {/* Основная информация о записи */}
            <div className="detail-card">
              <h3>Информация о записи</h3>
              <div className="detail-item">
                <strong>ID:</strong> {recordDetail.id}
              </div>
              <div className="detail-item">
                <strong>Статус:</strong> 
                <span className={`status-badge ${recordDetail.status === 'confirm' ? 'active' : recordDetail.status === 'pending' ? 'inactive' : 'inactive'}`}>
                  {recordDetail.status === 'confirm' ? 'Подтверждена' : recordDetail.status === 'pending' ? 'Ожидает' : 'Отклонена'}
                </span>
              </div>
              <div className="detail-item">
                <strong>Дата создания:</strong> {new Date(recordDetail.created_at).toLocaleString('ru-RU')}
              </div>
            </div>

            {/* Информация о клиенте */}
            <div className="detail-card">
              <h3>Клиент</h3>
              <div className="detail-item">
                <strong>Имя:</strong> {recordDetail.client_name} {recordDetail.client_surname}
              </div>
              <div className="detail-item">
                <strong>Телефон:</strong> {recordDetail.client_phone}
              </div>
              <div className="detail-item">
                <strong>Telegram ID:</strong> {recordDetail.client_telegram_id}
              </div>
            </div>

            {/* Информация об услуге */}
            <div className="detail-card">
              <h3>Услуга</h3>
              <div className="detail-item">
                <strong>Название:</strong> {recordDetail.slot_name}
              </div>
              <div className="detail-item">
                <strong>Цена:</strong> {recordDetail.slot_price} руб.
              </div>
              <div className="detail-item">
                <strong>Длительность:</strong> {recordDetail.slot_duration} мин.
              </div>
            </div>

            {/* Информация о мастере */}
            <div className="detail-card">
              <h3>Мастер</h3>
              <div className="detail-item">
                <strong>Имя:</strong> {recordDetail.master_name} {recordDetail.master_surname}
              </div>
              <div className="detail-item">
                <strong>Телефон:</strong> {recordDetail.master_phone}
              </div>
              <div className="detail-item">
                <strong>Telegram ID:</strong> {recordDetail.master_telegram_id}
              </div>
            </div>
          </div>
        </section>
      )}

      {/* Модальное окно для добавления роли */}
      {showRoleModal && (
        <div className="admin-modal-backdrop" onClick={handleCloseRoleModal}>
          <div className="admin-modal" onClick={(e) => e.stopPropagation()}>
            <div className="admin-modal-title">Добавить роль</div>
            <div className="admin-modal-sub">Введите название роли для пользователя</div>
            
            <div style={{ marginBottom: '30px' }}>
              <input
                type="text"
                value={roleInput}
                onChange={(e) => setRoleInput(e.target.value)}
                placeholder="Например: ADMIN, MASTER, USER"
                style={{
                  width: '100%',
                  padding: '14px',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                  fontSize: '12px',
                  outline: 'none',
                  boxSizing: 'border-box'
                }}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    handleAddRole();
                  }
                }}
                autoFocus
              />
            </div>
            
            <div className="admin-modal-actions">
          <button
                className="btn-primary" 
                onClick={handleAddRole}
                disabled={!roleInput.trim()}
          >
                Добавить роль
          </button>
              <button 
                className="btn-danger" 
                onClick={handleCloseRoleModal}
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminDashboard;
