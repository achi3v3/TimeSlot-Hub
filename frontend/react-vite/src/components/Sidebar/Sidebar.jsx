import "./sidebar.css";
import { useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import {
  clearCurrentUser,
  getCurrentUser,
  clearToken,
  isAuthenticated,
} from "../../utils/auth";
import SidebarToggle from "./SidebarToggle";
import SidebarNav from "./SidebarNav";

export default function Sidebar() {
  const navigate = useNavigate();
  const [, setUser] = useState(() => getCurrentUser());
  const [authed, setAuthed] = useState(() => isAuthenticated());
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
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
  };

  const toggleSidebar = () => {
    setIsOpen((prev) => !prev);
  };

  const isAuthed = authed;

  return (
    <>
      <SidebarToggle isOpen={isOpen} onToggle={toggleSidebar} />

      {isOpen && (
        <div
          className="sidebar-overlay"
          onClick={() => setIsOpen(false)}
        ></div>
      )}

      <div className={`sidebar ${isOpen ? "sidebar-open" : ""}`}>
        <div className="sidebar-header">
          <div className="sidebar-logo">
            <h2 className="noselect">melot</h2>
          </div>
        </div>

        <SidebarNav
          isAuthed={isAuthed}
          onLinkClick={handleLinkClick}
          onLogout={handleLogout}
        />
      </div>
    </>
  );
}
