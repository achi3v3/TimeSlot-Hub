import { useEffect, useState } from "react";
import "./login.css";
import tglogo from "/telegram_icon.svg";
import { usePhoneInput } from "../../utils/phoneUtils";
import { useMutation } from "@tanstack/react-query";
import { apiService } from "../../utils/api";
import { useNavigate } from "react-router-dom";
import { setCurrentUser } from "../../utils/auth";
import RightSidebar from '../../components/RightSidebar/RightSidebar';  
import { config } from "../../config/env";

async function authUser(phone) {
  return await apiService.user.login(phone);
}

export default function Login() {
  const { displayValue, handlePhoneChange, isPhoneValid, phone, reset } =
    usePhoneInput();

  const navigate = useNavigate();
  const [errorText, setErrorText] = useState("");
  const [cooldownUntil, setCooldownUntil] = useState(0);
  const COOLDOWN_MS = 30000; // 30s cooldown between attempts

  // tick to re-render button during cooldown
  const [, setTick] = useState(0);
  useEffect(() => {
    const t = setInterval(() => setTick((v) => v + 1), 1000);
    return () => clearInterval(t);
  }, []);

  const mutation = useMutation({
    mutationFn: (newUser) => authUser(newUser),
    onError: (error) => {
      console.error("Authentication error:", error);
      // Пытаемся показать понятное сообщение пользователю
      const status = error?.response?.status;
      const serverMsg = error?.response?.data?.error || error?.response?.data?.message;
      if (status === 400) {
        setErrorText(serverMsg || "Пользователь с таким номером не найден");
      } else {
        setErrorText(serverMsg || "Ошибка авторизации. Повторите попытку позже");
      }
    },
    onSuccess: (response) => {
      const data = response?.data;
      if (data?.user) {
        setCurrentUser(data.user);
      }
      navigate("/login/wait");
    },
  });

  const handleClick = (e) => {
    e.preventDefault();
    const now = Date.now();
    if (now < cooldownUntil) {
      return;
    }
    mutation.mutate(phone);
    setCooldownUntil(now + COOLDOWN_MS);
    reset();
  };

  return (
    <div className="authorization-container">
    <main className="authorization-container">
      <form id="block-authorization" onSubmit={handleClick}>
        <h2>Авторизация</h2>
        {errorText && (
          <div className="error-text" role="alert" style={{ color: "#b00020", marginBottom: 8 }}>
            {errorText}
          </div>
        )}
        <label htmlFor="phone">Номер телефона</label>
        <input
          type="tel"
          id="phone"
          className="phone"
          placeholder="+7 (XXX) XXX-XX-XX"
          title="Введите 10 цифр без кода страны"
          value={displayValue}
          onChange={handlePhoneChange}
          required
        />
        
        <button
          type="submit"
          className="button"
          disabled={!isPhoneValid || mutation.isPending || Date.now() < cooldownUntil}
        >
          <p className="noselect">
            {Date.now() < cooldownUntil
              ? `Подождите ${Math.max(1, Math.ceil((cooldownUntil - Date.now()) / 1000))}с`
              : 'Войти через Telegram'}
            <img className="tg-logo" src={tglogo} alt="Telegram" />
          </p>
        </button>
        <a
          href={`${config.TELEGRAM_BOT_LINK}?start`}
          className="register-link"
          title="Создай аккаунт"
        >
          Не зарегистрированы?
        </a>
      </form>
    </main>
      <RightSidebar />
      </div>
  );
}
