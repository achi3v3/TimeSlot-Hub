// components/Footer.jsx
import React from 'react';
import "./footer.css";
import { config } from '../../config/env';
const Footer = () => {
  return (
  <footer className="home-footer">
    {/* MAIN CONTENT */}
  <div className="footer-content">
    <div className="footer-section">
      <div className="footer-logo">melot</div>
      <p className="footer-description">
        Бесплатная платформа для упрощения взаимодействия с клиентами. 
      </p>
    </div>
    {/* PROJECT INFO */}
    <div className="footer-section">
      <h4 className="footer-title">Проект</h4>
      <ul className="footer-links">
        <ul className="footer-links">
        <li><a href="/help">Помощь и контакты</a></li>
        <li><a href={config.TELEGRAM_CHANNEL_LINK} target="_blank" rel="noopener">Telegram-канал</a></li>
      </ul>
      </ul>
    </div>
    {/* LEGAL INFO */}
    <div className="footer-section">
      <h4 className="footer-title">Правовая информация</h4>
      <ul className="footer-links">
        <li><a href="/privacy">Политика конфиденциальности</a></li>
        <li><a href="/terms">Условия использования</a></li>
      </ul>
    </div>
  </div>
    {/* BOTTOM CONTENT */}
  <div className="footer-bottom">
    <div className="footer-copyright">
      © 2025 melot tech.
    </div>
    <div className="footer-made-with">
      Сделано с ❤️
    </div>
  </div>
</footer>
  );
};

export default Footer;