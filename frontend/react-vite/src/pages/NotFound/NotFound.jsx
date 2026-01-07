import LeftSidebar from "../../components/LeftSidebar/LeftSidebar";
import "./notFound.css";
import Footer from '../../components/Footer/Footer';
import React, { useState } from 'react';

export default function NotFound() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const toggleMobileMenu = () => setIsMobileMenuOpen(!isMobileMenuOpen);
  return (
    <div className="app">
      <div className="notfound-root">
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
        {isMobileMenuOpen && (
          <div className="mobile-overlay" onClick={toggleMobileMenu}></div>
        )}
        <LeftSidebar isMobileMenuOpen={isMobileMenuOpen} toggleMobileMenu={toggleMobileMenu} />
        <div className="notfound-main-content">
          <div className="notfound-card">
            <h1 className="notfound-title">Страница не найдена</h1>
            <p className="notfound-subtitle">Запрошенный адрес не существует или был перемещён.</p>
            <a className="notfound-btn" href="/home">На главную</a>
          </div>
        </div>
      </div>
      <Footer />
    </div>
  );
}

