import { Link } from "react-router-dom";

export default function SidebarNav({ isAuthed, onLinkClick, onLogout }) {
  return (
    <>
      <nav className="sidebar-nav">
        <ul className="sidebar-menu">
          <li className="sidebar-item">
            <Link to="/" className="sidebar-link" onClick={onLinkClick}>
              <span className="sidebar-text">ГЛАВНАЯ</span>
            </Link>
          </li>
          <li className="sidebar-item">
            <Link to="/about" className="sidebar-link" onClick={onLinkClick}>
              <span className="sidebar-text">ИНФО</span>
            </Link>
          </li>
          {isAuthed && (
            <>
              <li className="sidebar-item">
                <Link
                  to="/profile"
                  className="sidebar-link"
                  onClick={onLinkClick}
                >
                  <span className="sidebar-text">ПРОФИЛЬ</span>
                </Link>
              </li>
              <li className="sidebar-item">
                <Link
                  to="/notifications"
                  className="sidebar-link"
                  onClick={onLinkClick}
                >
                  <span className="sidebar-text">УВЕДОМЛЕНИЯ</span>
                </Link>
              </li>
            </>
          )}
        </ul>
      </nav>

      <div className="sidebar-bottom">
        {!isAuthed ? (
          <Link
            to="/login"
            className="sidebar-auth-btn sidebar-login"
            onClick={onLinkClick}
          >
            <span className="sidebar-text">ВОЙТИ</span>
          </Link>
        ) : (
          <button
            className="sidebar-auth-btn sidebar-logout"
            onClick={onLogout}
          >
            <span className="sidebar-text">ВЫЙТИ</span>
          </button>
        )}
      </div>
    </>
  );
}


