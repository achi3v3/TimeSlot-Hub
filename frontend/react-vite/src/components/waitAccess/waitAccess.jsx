import "./waitAccess.css";
import { useEffect, useMemo, useState } from "react";
import api from "../../utils/api";
import { config } from "../../config/env";
import { getCurrentUser, setToken, setCurrentUser } from "../../utils/auth";
import { useNavigate } from "react-router-dom";

export default function WaitAccess() {
  const navigate = useNavigate();
  const user = useMemo(() => getCurrentUser(), []);
  const telegramId = user?.telegram_id;
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!telegramId) {
      setError("Не удалось определить telegram_id пользователя. Вернитесь и авторизуйтесь заново.");
      return;
    }
    const interval = setInterval(async () => {
      try {
        setChecking(true);
        // 1) Проверяем, есть ли ожидаемый токен для пользователя
        try {
          const checkRes = await api.get(`/user/check-login/${telegramId}`);
          const pending = Boolean(checkRes?.data?.pending);
          if (!pending) {
            return; // ещё не подтверждено в телеграме
          }
        } catch (_) {
          return;
        }

        // 2) Пытаемся забрать токен и пользователя
        const claimRes = await api.post(
          `/user/claim-token/${telegramId}`,
          {},
          {
            headers: {
              'X-Frontend-Secret': import.meta.env.VITE_FRONTEND_SECRET || config.FRONTEND_SECRET || '',
              // Optional fallback if backend expects INTERNAL_TOKEN
              'X-Internal-Token': import.meta.env.VITE_INTERNAL_TOKEN || config.INTERNAL_TOKEN || '',
            },
          }
        );
        const token = claimRes?.data?.token;
        const claimedUser = claimRes?.data?.user;
        const pending = claimRes?.data?.pending;
        if (pending) {
          return; // токен ещё не готов
        }
        if (token) {
          setToken(token);
          if (claimedUser) {
            setCurrentUser(claimedUser);
          }
          clearInterval(interval);
          navigate("/profile", { replace: true });
          return;
        }
      } catch (e) {
        // молча ждём подтверждения; попробуем ещё раз позже
      } finally {
        setChecking(false);
      }
    }, 2000);
    return () => clearInterval(interval);
  }, [telegramId, navigate]);

  return (
    <main className="wait-container">
    <div className="wait-block">
      <h2 className="header-wait">Авторизация</h2>
        <div className="text-wait">
          <p className="noselect">
            Подтвердите вход в свой аккаунт в телеграм боте
          </p>
        </div>
        <a href={config.TELEGRAM_BOT_LINK}>
        Перейти к боту
        </a>
        {checking && <p style={{ marginTop: 12 }}>Проверяем подтверждение…</p>}
        {error && <p className="Error">{error}</p>}
        </div>
        </main>
  );
}
