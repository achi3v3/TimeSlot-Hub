import React, { useState } from 'react';
import LeftSidebar from '../../components/LeftSidebar/LeftSidebar';
import RightSidebar from '../../components/RightSidebar/RightSidebar';
import Footer from '../../components/Footer/Footer';  
import './help.css';
import { config } from '../../config/env';
import { AiOutlineMail } from "react-icons/ai"; // Ant Design Icons (почта)
import { FaTelegramPlane } from "react-icons/fa"; // Font Awesome (телеграм)
import { FaTelegram } from "react-icons/fa"; // Font Awesome (телеграм)

const HelpPage = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const toggleMobileMenu = () => setIsMobileMenuOpen(!isMobileMenuOpen);

  return (
    <div className="app">
    <div className="help-container">
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

      {/* Левый сайдбар */}
      <LeftSidebar 
        isMobileMenuOpen={isMobileMenuOpen}
        toggleMobileMenu={toggleMobileMenu}
      />

      {/* Основной контент */}
      <main className="help-main">
        <header className="help-header">
          <h1>поддержка</h1>
          <p className="help-subtitle">получите помощь и свяжитесь с нами</p>
        </header>

        <section className="help-content">
          
          {/* Контактная информация */}
          <div className="help-section">
            <div className="help-section-header">
              <div className="help-icon">📞</div>
              <h2 className="help-section-title">Свяжитесь с нами</h2>
            </div>
            <div className="help-description">
              Если у вас есть вопросы или предложения, мы всегда готовы помочь
            </div>
            
            <div className="contact-cards">
              <div className="contact-card">
                <div className="contact-card-icon">
                <FaTelegramPlane />
                </div>
                <h3>Telegram</h3>
                <p>Быстрая поддержка через Telegram</p>
                <a href={config.TELEGRAM_SUPPORT_LINK} className="contact-link" target="_blank" rel="noopener">
                Написать в Telegram 
                </a>
                  
              </div>
              
              <div className="contact-card">
                <div className="contact-card-icon">
                <AiOutlineMail />
                </div>
                <h3>Email</h3>
                <p>Отправьте нам письмо</p>
                <a href="mailto:__" className="contact-link">
                  Написать на почту
                </a>
              </div>
            </div>
          </div>

          {/* Дополнительная информация */}
          <div className="help-section">
            <div className="help-section-header">
              <div className="help-icon">📢</div>
              <h2 className="help-section-title">Дополнительная информация</h2>
            </div>
            <div className="help-description">
              Следите за обновлениями и новостями проекта
            </div>
            
            <div className="info-cards">
              <div className="info-card">
                <div className="info-card-icon">
                <FaTelegram />
                </div>
                <h3>Telegram канал</h3>
                <p>Подписывайтесь на наш канал для получения новостей и обновлений</p>
                <a href={config.TELEGRAM_CHANNEL_LINK} className="info-link" target="_blank" rel="noopener">
                  Подписаться на канал
                </a>
              </div>
            </div>
          </div>

        </section>
      </main>

      {/* Правый сайдбар */}
      <RightSidebar />
    </div>
    
    <Footer />
    </div>
  );
};

export default HelpPage;