import React, { useState, useEffect } from 'react';
import './Toast.css';

const Toast = ({ 
  message, 
  type = 'success', 
  duration = 3000, 
  onClose,
  position = 'top-right'
}) => {
  const [isVisible, setIsVisible] = useState(false);
  const [isLeaving, setIsLeaving] = useState(false);

  useEffect(() => {
    // Показываем toast с небольшой задержкой для анимации
    const showTimer = setTimeout(() => {
      setIsVisible(true);
    }, 10);

    // Автоматически скрываем через duration
    const hideTimer = setTimeout(() => {
      handleClose();
    }, duration);

    return () => {
      clearTimeout(showTimer);
      clearTimeout(hideTimer);
    };
  }, [duration]);

  const handleClose = () => {
    setIsLeaving(true);
    setTimeout(() => {
      setIsVisible(false);
      if (onClose) onClose();
    }, 300); // Время анимации выхода
  };

  if (!isVisible && !isLeaving) return null;

  const getIcon = () => {
    switch (type) {
      case 'success':
        return '✓';
      case 'error':
        return '✕';
      case 'warning':
        return '⚠';
      case 'info':
        return 'ℹ';
      default:
        return '✓';
    }
  };

  return (
    <div 
      className={`toast toast-${type} toast-${position} ${isLeaving ? 'toast-leaving' : ''}`}
      onClick={handleClose}
    >
      <div className="toast-content">
        <div className="toast-icon">
          {getIcon()}
        </div>
        <div className="toast-message">
          {message}
        </div>
        <button 
          className="toast-close"
          onClick={(e) => {
            e.stopPropagation();
            handleClose();
          }}
        >
          ×
        </button>
      </div>
    </div>
  );
};

export default Toast;
