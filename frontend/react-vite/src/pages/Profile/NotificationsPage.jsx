import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiService } from "../../utils/api";
import { useNavigate } from "react-router-dom";
import { isAuthenticated } from "../../utils/auth";
import LeftSidebar from "../../components/LeftSidebar/LeftSidebar";
import RightSidebar from "../../components/RightSidebar/RightSidebar";
import { FaBell } from "react-icons/fa";
import "./notifications.css";
import { useState } from 'react';

export default function NotificationsPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };
  const listQuery = useQuery({
    queryKey: ["notifications", "list"],
    queryFn: async () => {
      const r = await apiService.notification.list();
      const payload = r?.data;
      const list = Array.isArray(payload) ? payload : (payload?.data || []);
      return Array.isArray(list) ? list : [];
    },
    retry: 1,
    refetchOnWindowFocus: true,
    staleTime: 60 * 1000,
    refetchInterval: 2 * 60 * 1000,
  });

  const unreadQuery = useQuery({
    queryKey: ["notifications", "unreadCount"],
    queryFn: async () => {
      const r = await apiService.notification.unreadCount();
      return r?.data?.count ?? 0;
    },
    retry: 1,
    refetchOnWindowFocus: true,
    staleTime: 30 * 1000,
    refetchInterval: 30 * 1000,
  });

  const markReadMutation = useMutation({
    mutationFn: (id) => apiService.notification.markRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const markAllMutation = useMutation({
    mutationFn: () => apiService.notification.markAllRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  // Проверка авторизации убрана - она уже есть в App.jsx

  if (listQuery.isLoading) {
    return (
      <div className="notifications-root">
              {/* Мобильная кнопка меню */}
      <button 
        className="mobile-menu-toggle"
        onClick={toggleMobileMenu}
        aria-label="Открыть меню"
      >
        <span className={`hamburger ${isMobileMenuOpen ? 'active' : ''}`}>
          <span></span>
          <span></span>
          <span></span>
        </span>
      </button>

      {/* Мобильный оверлей */}
      {isMobileMenuOpen && (
        <div 
          className="mobile-overlay"
          onClick={toggleMobileMenu}
        ></div>
      )}

       {/* Левый сайдбар навигации */}
       <LeftSidebar 
         isMobileMenuOpen={isMobileMenuOpen}
         toggleMobileMenu={toggleMobileMenu}
       />
        <div className="notifications-main-content">
          <div className="notifications-container">
            <div className="notifications-header">
              <h2 className="notifications-title">уведомления</h2>
            </div>
            <div className="notifications-loading">Загрузка уведомлений...</div>
          </div>
        </div>
        <RightSidebar />
      </div>
    );
  }

  if (listQuery.isError) {
    const status = listQuery.error?.response?.status;
    // Убираем автоматическое перенаправление - пусть App.jsx сам решает
    return (
      <div className="notifications-root">
              {/* Мобильная кнопка меню */}
      <button 
        className="mobile-menu-toggle"
        onClick={toggleMobileMenu}
        aria-label="Открыть меню"
      >
        <span className={`hamburger ${isMobileMenuOpen ? 'active' : ''}`}>
          <span></span>
          <span></span>
          <span></span>
        </span>
      </button>

      {/* Мобильный оверлей */}
      {isMobileMenuOpen && (
        <div 
          className="mobile-overlay"
          onClick={toggleMobileMenu}
        ></div>
      )}

       {/* Левый сайдбар навигации */}
       <LeftSidebar 
         isMobileMenuOpen={isMobileMenuOpen}
         toggleMobileMenu={toggleMobileMenu}
       />
        <div className="notifications-main-content">
          <div className="notifications-container">
            <div className="notifications-header">
              <h2 className="notifications-title">уведомления</h2>
            </div>
            <div className="notifications-error">
              <h3>Ошибка загрузки уведомлений</h3>
              <p>Статус ошибки: {status || 'Неизвестно'}</p>
              <button className="btn-retry" onClick={() => listQuery.refetch()}>
                Повторить попытку
              </button>
            </div>
          </div>
        </div>
        <RightSidebar />
      </div>
    );
  }

  const notifications = Array.isArray(listQuery.data) ? listQuery.data : [];
  const unreadCount = unreadQuery.data ?? 0;

  // Сортируем уведомления: новые вверху, разделяем на прочитанные и непрочитанные
  const sortedNotifications = [...notifications].sort((a, b) => {
    // Сначала по статусу прочтения (непрочитанные вверху)
    if (a.is_read !== b.is_read) {
      return a.is_read ? 1 : -1;
    }
    // Затем по дате создания (новые вверху)
    return new Date(b.created_at || b.timestamp || Date.now()) - new Date(a.created_at || a.timestamp || Date.now());
  });

  const unreadNotifications = sortedNotifications.filter(n => !n.is_read);
  const readNotifications = sortedNotifications.filter(n => n.is_read);

  return (
    <div className="notifications-root">
            {/* Мобильная кнопка меню */}
            <button 
        className="mobile-menu-toggle"
        onClick={toggleMobileMenu}
        aria-label="Открыть меню"
      >
        <span className={`hamburger ${isMobileMenuOpen ? 'active' : ''}`}>
          <span></span>
          <span></span>
          <span></span>
        </span>
      </button>

      {/* Мобильный оверлей */}
      {isMobileMenuOpen && (
        <div 
          className="mobile-overlay"
          onClick={toggleMobileMenu}
        ></div>
      )}

       {/* Левый сайдбар навигации */}
       <LeftSidebar 
         isMobileMenuOpen={isMobileMenuOpen}
         toggleMobileMenu={toggleMobileMenu}
       />
      <div className="notifications-main-content">
        <div className="notifications-container">
          <div className="notifications-header">
            <h2 className="notifications-title">уведомления</h2>
            <div className="notifications-stats">
              <span className="notifications-count">
                Непрочитанных: <strong>{unreadCount}</strong>
              </span>
            </div>
          </div>
          <div className="notifications-actions">
            <button 
              className="btn-mark-all" 
              onClick={() => markAllMutation.mutate()} 
              disabled={markAllMutation.isPending || unreadCount === 0}
            >
              {markAllMutation.isPending ? "Обрабатываем..." : "Пометить все как прочитанные"}
            </button>
            <button 
              className="btn-refresh" 
              onClick={() => listQuery.refetch()} 
              disabled={listQuery.isFetching}
            >
              {listQuery.isFetching ? "Обновляем..." : "Обновить"}
            </button>
          </div>


          <div className="notifications-content">
            {notifications.length === 0 ? (
              <div className="notifications-empty">
                <div className="empty-icon">
                  <FaBell />
                </div>
                <h3>Уведомлений пока нет</h3>
                <p>Здесь будут появляться уведомления о ваших записях и других важных событиях</p>
              </div>
            ) : (
              <div className="notifications-list">
                {/* Непрочитанные уведомления */}
                {unreadNotifications.length > 0 && (
                  <div className="notifications-section">
                    <div className="section-header">
                      <h3 className="section-title">Новые уведомления</h3>
                      <span className="section-count">{unreadNotifications.length}</span>
                    </div>
                    {unreadNotifications.map((notification) => (
                      <div key={notification.id} className="notification-item unread">
                        <div className="notification-indicator"></div>
                        <div className="notification-content">
                          <div className="notification-header">
                            <h4 className="notification-title">{notification.title}</h4>
                            <span className="notification-time">
                              {new Date(notification.created_at || notification.timestamp || Date.now())
                                .toLocaleString("ru-RU", {
                                  day: "2-digit",
                                  month: "short",
                                  hour: "2-digit",
                                  minute: "2-digit"
                                })}
                            </span>
                          </div>
                          <p className="notification-message">{notification.message}</p>
                          <button 
                            className="btn-mark-read" 
                            onClick={() => markReadMutation.mutate(notification.id)}
                            disabled={markReadMutation.isPending}
                          >
                            Пометить как прочитанное
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {/* Разделитель, если есть и прочитанные, и непрочитанные */}
                {unreadNotifications.length > 0 && readNotifications.length > 0 && (
                  <div className="notifications-divider">
                    <span>Прочитанные уведомления</span>
                  </div>
                )}

                {/* Прочитанные уведомления */}
                {readNotifications.length > 0 && (
                  <div className="notifications-section read-section">
                    {readNotifications.map((notification) => (
                      <div key={notification.id} className="notification-item read">
                        <div className="notification-content">
                          <div className="notification-header">
                            <h4 className="notification-title">{notification.title}</h4>
                            <span className="notification-time">
                              {new Date(notification.created_at || notification.timestamp || Date.now())
                                .toLocaleString("ru-RU", {
                                  day: "2-digit",
                                  month: "short",
                                  hour: "2-digit",
                                  minute: "2-digit"
                                })}
                            </span>
                          </div>
                          <p className="notification-message">{notification.message}</p>
                          <div className="notification-read-mark">✓ Прочитано</div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
      <RightSidebar />
    </div>
  );
}


