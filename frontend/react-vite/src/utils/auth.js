export function getCurrentUser() {
  try {
    const raw = localStorage.getItem("user");
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function setCurrentUser(user) {
  try {
    localStorage.setItem("user", JSON.stringify(user));
    window.dispatchEvent(new Event("auth-changed"));
  } catch {
    // ignore
  }
}

export function clearCurrentUser() {
  localStorage.removeItem("user");
  window.dispatchEvent(new Event("auth-changed"));
}

export function getToken() {
  return localStorage.getItem("token") || null;
}

export function setToken(token) {
  if (token) {
    localStorage.setItem("token", token);
    window.dispatchEvent(new Event("auth-changed"));
  }
}

export function clearToken() {
  localStorage.removeItem("token");
  window.dispatchEvent(new Event("auth-changed"));
}

export function isAuthenticated() {
  return Boolean(getToken());
}


export const handleLogout = () => {
  clearCurrentUser();
  clearToken();
  // Дополнительно: диспатч события для синхронизации
  window.dispatchEvent(new Event('auth-changed'));
  window.location.href = '/'; // принудительный редирект
};