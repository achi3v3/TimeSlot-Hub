import React from 'react';
import "./rightSidebar.css";
import { config } from '../../config/env';
import { FaTelegramPlane } from "react-icons/fa"; // Font Awesome (телеграм)
import { AiOutlineInfoCircle } from "react-icons/ai";

const RightSidebar = () => {
  const trackAdClick = (slot) => {
    const url = '/api/metrics/ad-click';
    const payload = JSON.stringify({ slot });
    try {
      if (navigator.sendBeacon) {
        const blob = new Blob([payload], { type: 'application/json' });
        navigator.sendBeacon(url, blob);
      } else {
        fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: payload, keepalive: true }).catch(() => {});
      }
    } catch {}
  };
  return (
<aside className="right-sidebar">
<div className="sidebar-widget">
  <h3>Новости</h3>
  <div className="news-list">
    <div className="news-item">
      <h4>Telegram канал</h4>
      <p>Запустили официальный Telegram канал для быстрой связи и уведомлений об обновлениях</p>
      <span className="news-date">03.11.2025</span>
      <a href={config.TELEGRAM_CHANNEL_LINK} className="quick-link" onClick={() => trackAdClick(1)}>Перейти</a>
    </div>
    <div className="news-item">
      <div className="news-badge">ОБНОВЛЕНИЕ</div>
      <h4>Улучшена система записи</h4>
      <p>Добавлено напоминание о визите за 1 час до начала записи</p>
      <span className="news-date">02.11.2025</span>
      </div>
  </div>
</div>

 {/* Быстрые ссылки */}
{/*
<div className="sidebar-widget">
  <h3>Быстрые ссылки</h3>
  <div className="quick-links">
    <a href="/booking" className="quick-link">Онлайн запись</a>
    <a href="/masters" className="quick-link">Наши мастера</a>
    <a href="/services" className="quick-link">Все услуги</a>
    <a href="/gallery" className="quick-link">Галерея работ</a>
  </div>
</div> */}

</aside>
);
};

export default RightSidebar;