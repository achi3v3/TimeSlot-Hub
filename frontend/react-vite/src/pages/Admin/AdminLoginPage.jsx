import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './AdminLoginPage.css';
import axios from 'axios';

const AdminLoginPage = () => {
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const res = await axios.post('/api/admin/login', { password });
      if (res.data?.success) {
        // Save token and timestamp
        if (res.data?.token) {
          localStorage.setItem('admin_token', res.data.token);
        }
        const expire = Date.now() + 10 * 60 * 1000;
        localStorage.setItem('admin_authed_until', expire.toString());
        navigate('/admin/dashboard');
      } else {
        setError('Неверный пароль или ошибка сервера');
      }
    } catch (err) {
      if (err.response && err.response.status === 401) {
        setError('Неверный пароль');
      } else {
        setError('Ошибка соединения с сервером.');
      }
    }
    setLoading(false);
  };

  // Сброс ошибки при изменении поля
  const onChangePassword = (e) => {
    setPassword(e.target.value);
    setError('');
  };

  return (
    <div className="admin-login-container">
      <div className="admin-login-card">
        <div className="admin-login-header">
          <h1>Панель администратора</h1>
        </div>
        <form onSubmit={handleSubmit} className="admin-login-form">
          <div className="form-group">
            <label htmlFor="password">Пароль администратора</label>
            <input
              type="password"
              id="password"
              className="form-input"
              placeholder="Введите пароль"
              value={password}
              onChange={onChangePassword}
              required
            />
          </div>
          {error && <div className="error-message">{error}</div>}
          <button type="submit" disabled={loading} className="submit-button">
            {loading ? 'Вход...' : 'Войти в админку'}
          </button>
        </form>
        <div className="admin-login-footer">
          <p>Каждая попытка входа регистрируется, не пробуйте осуществить вход, если у вас нет доступа</p>
        </div>
      </div>
    </div>
  );
};

export default AdminLoginPage;
