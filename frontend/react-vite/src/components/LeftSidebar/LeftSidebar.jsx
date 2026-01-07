import "./leftSidebar.css";
import { Link, useNavigate, useLocation } from "react-router-dom";
import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { clearCurrentUser, getCurrentUser, clearToken, isAuthenticated } from "../../utils/auth";
import { apiService } from "../../utils/api";

const LeftSidebar = ({ isMobileMenuOpen, toggleMobileMenu }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [, setUser] = useState(() => getCurrentUser());
  const [authed, setAuthed] = useState(() => isAuthenticated());

  // Запрос для получения количества непрочитанных уведомлений
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
    enabled: authed, // Запрашиваем только если пользователь авторизован
  });

  useEffect(() => {
    // Синхронизация при смене состояния авторизации
    const handler = () => {
      setUser(getCurrentUser());
      setAuthed(isAuthenticated());
    };
    window.addEventListener("storage", handler);
    window.addEventListener("auth-changed", handler);
    return () => {
      window.removeEventListener("storage", handler);
      window.removeEventListener("auth-changed", handler);
    };
  }, []);

  const handleLogout = () => {
    clearCurrentUser();
    clearToken();
    setUser(null);
    setAuthed(false);
    setIsOpen(false);
    navigate("/");
  };

  const handleLinkClick = () => {
    setIsOpen(false);
    if (toggleMobileMenu) {
      toggleMobileMenu();
    }
  };

  const isAuthed = authed;

  // Функция для определения активной вкладки
  const isActive = (path) => {
    return location.pathname === path;
  };

  // Получаем количество непрочитанных уведомлений
  const unreadCount = unreadQuery.data ?? 0;

  return (
    <aside className={`left-sidebar ${isMobileMenuOpen ? 'mobile-open' : ''}`}>
    <div className="sidebar-header">
    <h2>melot</h2>
    <button 
        className="mobile-close-btn"
        onClick={toggleMobileMenu}
        aria-label="Закрыть меню"
        >
        ×
    </button>
    </div>
<nav className="sidebar-nav">
  <ul className="nav-list">
    <li className="nav-item">
      <Link to="/home" className={`nav-link ${isActive('/home') ? 'active' : ''}`} onClick={handleLinkClick}>
        <span className="nav-icon"></span>
        Главная
      </Link>
    </li>
    <li className="nav-item">
      <Link to="/about" className={`nav-link ${isActive('/about') ? 'active' : ''}`} onClick={handleLinkClick}>
        <span className="nav-icon"></span>
        Инфо
      </Link>
    </li>
    {isAuthed && (
      <>
    <li className="nav-item">
      <Link to="/profile" className={`nav-link ${isActive('/profile') ? 'active' : ''}`} onClick={handleLinkClick}>
        <span className="nav-icon"></span>
        Профиль
      </Link>
    </li>
    <li className="nav-item">
      <Link to="/notifications" className={`nav-link ${isActive('/notifications') ? 'active' : ''}`} onClick={handleLinkClick}>
        <span className="nav-icon"></span>
        Уведомления
        {unreadCount > 0 && (
          <span className="notification-badge">{unreadCount}</span>
        )}
      </Link>
    </li>
    </>
    )}
    {!isAuthed ? (
    <li className="nav-item">
      <Link to="/login" className={`nav-link ${isActive('/login') ? 'active' : ''}`} onClick={handleLinkClick}>
        <span className="nav-icon"></span>
        Войти
      </Link>
    </li>
    ) : (
    <li className="nav-item">
      <Link to="/login" className={`nav-link ${isActive('/login') ? 'active' : ''}`} onClick={handleLogout}>
        <span className="nav-icon"></span>
        Выйти
      </Link>
    </li>
    )}
  </ul>
</nav>
</aside>
);
};

export default LeftSidebar;