import "./header.css";
import { Link, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { clearCurrentUser, getCurrentUser, clearToken, isAuthenticated } from "../../utils/auth";

export default function Header() {
  const navigate = useNavigate();
  const [user, setUser] = useState(() => getCurrentUser());
  const [authed, setAuthed] = useState(() => isAuthenticated());

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
    navigate("/");
  };

  const isAuthed = authed;

  return (
    <div className="highbar">
      <div className="highbar-leftsection">
        <div className="logo">
          <h2
            className="noselect"
            style={{ color: "inherit", backgroundColor: "transparent" }}
          >
            W
          </h2>
        </div>
        <ul>
          <li className="list">
            <Link to="/">ГЛАВНАЯ</Link>
          </li>
          <li className="list">
            <Link to="/info">ИНФО</Link>
          </li>
          <li className="list">
            <Link to="/about">КОМПАНИЯ</Link>
          </li>
        </ul>
      </div>
      <div className="right-list">
      {isAuthed && (
            <>
              <li className="list">
                <Link to="/profile">ПРОФИЛЬ</Link>
              </li>
              <li className="list">
                <Link to="/notifications">УВЕДОМЛЕНИЯ</Link>
              </li>
            </>
          )}
      {!isAuthed ? (
          <ul>
            <li className="list">
              <Link to="/login">ВОЙТИ</Link>
            </li>
          </ul>
        ) : (
          <button className="list-btn" onClick={handleLogout}>
            ВЫЙТИ
          </button>
        )}
        {/* <div className="switch">
          <span className="noselect">ТЕМА</span>
          <input type="checkbox" id="toggle" />
          <label className="switch-box" htmlFor="toggle"></label>
        </div> */}

      </div>
    </div>
  );
}
