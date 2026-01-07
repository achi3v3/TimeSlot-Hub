import React, { useState } from 'react';
import './home.css';
import './border.css';
import LeftSidebar from '../../components/LeftSidebar/LeftSidebar';  
import RightSidebar from '../../components/RightSidebar/RightSidebar';  
import Footer from '../../components/Footer/Footer';  
import heroImage2 from './main_image_2.jpg';
import { 
  AiOutlineThunderbolt,
  AiOutlineGift,
  AiOutlineFieldTime,
  AiOutlineRocket
} from "react-icons/ai";
import { 
  FaCut,
  FaHandSparkles,
  FaTools,
  FaHeadSideVirus,
  FaCamera,
  FaChalkboardTeacher,
  FaRunning
} from "react-icons/fa";


const Home = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  return (
    <div className="app">
    <div className="main-container">
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

      {/* Основной контент */}
      <main className="main-content">
        {/* Заголовок страницы */}
        <header className="page-header">
          <h1>melot</h1>
          <p className="page-subtitle">УПРОСТИ РА<span class="accent">БОТ</span>У С КЛИЕНТАМИ</p>
        </header>

        {/* Героический блок */}
        <section className="hero-section">
          <div className="hero-content">
            <h2>Ваше расписание в онлайне</h2>
            <p>
            Бесплатная платформа для управления записями через сайт и Telegram.
            Для частных специалистов, мастеров и малого бизнеса
            </p>
            <div className="hero-actions">
              <a href="/profile" className="btn btn-begin-work">Начать работу</a>
              <a href="#how-it-work-section" className="btn btn-secondary">Как устроено?</a>
            </div>
          </div>
          <div className="hero-image">
            
            <div className="placeholder-image">
            <img 
              src={heroImage2} 
              alt="TimeSlot Hub - управление расписанием" 
              className="hero-photo"
            />
            </div>
          </div>
        </section>

        {/* Преимущества */}
        <section className="features-section">
          <h3>Почему выбирают нас?</h3>
          <div className="features-grid">
            <div className="feature-card">
              <div className="feature-icon">
                <AiOutlineThunderbolt />
              </div>
              <h4>Удобство</h4>
              <p>Управляйте своим расписанием в любом удобном формате</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
              <AiOutlineFieldTime />
              </div>
              <h4>Экономия времени</h4>
              <p>Автоматизируйте процесс записи, меньше звонков и сообщений</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
              <AiOutlineGift />
              </div>
              <h4>Бесплатно</h4>
              <p>Никаких скрытых платежей или подписок</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
              <AiOutlineRocket />
              </div>
              <h4>Актуальность информации</h4>
              <p>Клиенты видят только свободные окна</p>
            </div>
          </div>
        </section>

        {/* Контакты */}
        <section className="contact-section">
          <h3>Кому подойдет?</h3>
          <div className="contact-info">
            <div className="contact-item">
              <span className="contact-icon">
              <FaCut />
              </span>
              <div>
                <h4>Парикмахеры и барберы</h4>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">
              <FaHandSparkles /> 
              </span>
              <div>
                <h4>Косметологи и мастера маникюра</h4>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">
              <FaTools />
              </span>
              <div>
                <h4>Ремонтники и IT-специалисты</h4>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">
              <FaHeadSideVirus />
              </span>
              <div>
                <h4>Психологи и коучи</h4>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">
              <FaCamera />
              </span>
              <div>
                <h4>Фотографы и операторы</h4>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">
              <FaChalkboardTeacher />
              </span>
              <div>
                <h4>Частные преподаватели</h4>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">
              <FaRunning />
              </span>
              <div>
                <h4>Тренеры и многие др.</h4>
              </div>
            </div>
          </div>
        </section>
        
        {/* Как это работает */}
        <section id="how-it-work-section" className="how-it-works-section">
          <h3>Как работать с сервисом?</h3>
          <div className="steps-container">
            <div className="step">
              <div className="step-number">1</div>
              <h4>Создай расписание</h4>
              <p>Настройте своё рабочее время, услуги. Система создаст "окна" для записи</p>
            </div>
            <div className="step">
              <div className="step-number">2</div>
              <h4>Размести у себя</h4>
              <p>Встройте ссылку своего расписания в профиль, отправьте ссылку в соцсетях</p>
            </div>
            <div className="step">
              <div className="step-number">3</div>
              <h4>Получай запросы</h4>
              <p>Клиенты сами выбирают удобное время. Вы мгновенно получаете уведомление о новой записи</p>
            </div>
            <div className="step">
              <div className="step-number">4</div>
              <h4>Подтверждайте 1 кликом</h4>
              <p>Подтвердите или отмените запись в профиле или в Телеграм боте. Клиент получит уведомление</p>
            </div>
          </div>
        </section>

        {/* Популярные услуги */}
        <section className="services-section">
          <h3>Популярные деятельности</h3>
          <div className="services-grid">
          <div className="service-card">
              <div className="service-image">
                <span>Маникюр</span>
              </div>
              <div className="service-info">
                <h4>Маникюр</h4>
                <p>Классический и аппаратный маникюр</p>
                {/* <div className="service-price">от 1200 ₽</div> */}
              </div>
            </div>
            <div className="service-card">
              <div className="service-image">
                <span>Фотосет</span>
              </div>
              <div className="service-info">
                <h4>Фотосет</h4>
                <p>Фотосессия в разных цветовых гаммах</p>
                {/* <div className="service-price">от 1500 ₽</div> */}
              </div>
            </div>
            <div className="service-card">
              <div className="service-image">
                <span>Английский язык</span>
              </div>
              <div className="service-info">
                <h4>Английский язык</h4>
                <p>Репетиторство английского языка 1 на 1</p>
                {/* <div className="service-price">от 2000 ₽</div> */}
              </div>
            </div>
            
          </div>
        </section>

      </main>

      <RightSidebar />
    </div>

    <Footer />
    </div>
    
  );
};

export default Home;
