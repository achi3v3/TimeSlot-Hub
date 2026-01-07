import { LoginPage } from "./pages/Login/LoginPage";
import { Routes, Route, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import axios from "axios";
import { WaitAccessPage } from "./pages/Login/WaitAccessPage";
import Home from "./pages/Home/Home";
import Slots from "./pages/Slots/Slots";
import HelpPage from "./pages/Help/Help";
import ProfilePage from "./pages/Profile/ProfilePage";
import NotificationsPage from "./pages/Profile/NotificationsPage";
import AboutPage from "./pages/About/AboutPage";
import PrivacyPage from "./pages/Legal/PrivacyPage";
import TermsPage from "./pages/Legal/TermsPage";
import NotFound from "./pages/NotFound/NotFound";
import AdminLoginPage from "./pages/Admin/AdminLoginPage";
import AdminDashboard from "./pages/Admin/AdminDashboard";
import { isAuthenticated } from "./utils/auth";
import { ToastProvider } from "./components/Toast";
import { useEffect, useState } from "react";

// В дев-режиме используем прокси Vite, поэтому baseURL не нужен
delete axios.defaults.baseURL;

// Создаём экземпляр QueryClient
const queryClient = new QueryClient();

export default function App() {
  const [authed, setAuthed] = useState(() => isAuthenticated());

  useEffect(() => {
    const handler = () => setAuthed(isAuthenticated());
    window.addEventListener("auth-changed", handler);
    window.addEventListener("storage", handler);
    return () => {
      window.removeEventListener("auth-changed", handler);
      window.removeEventListener("storage", handler);
    };
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <Routes>
          <Route path="/" element={<Navigate to="/home" replace />} />
          <Route path="/home" element={<Home />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/help" element={<HelpPage />} />
          <Route path="/login/wait" element={<WaitAccessPage />} />
          <Route path="/about" element={<AboutPage />} />
          <Route path="/privacy" element={<PrivacyPage />} />
          <Route path="/terms" element={<TermsPage />} />
          <Route path="/master/:telegramId" element={<Slots />} />
          <Route
            path="/profile"
            element={authed ? <ProfilePage /> : <Navigate to="/login" replace />}
          />
          <Route
            path="/notifications"
            element={authed ? <NotificationsPage /> : <Navigate to="/login" replace />}
          />
          {/* Админские маршруты */}
          <Route path="/admin/login" element={<AdminLoginPage />} />
          <Route path="/admin/dashboard" element={<AdminDashboard />} />
          {/* 404 - catch all unmatched routes */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </ToastProvider>
    </QueryClientProvider>
  );
}
