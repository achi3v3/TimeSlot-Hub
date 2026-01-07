// Утилиты для работы с временем и часовыми поясами

/**
 * Конвертирует время из UTC в локальное время пользователя
 * @param {string|Date} timeString - время в UTC формате
 * @returns {Date|null} - время в локальном часовом поясе пользователя
 */
export function toLocalTime(timeString) {
  if (!timeString) return null;
  
  // Если это уже Date объект, используем его
  if (timeString instanceof Date) {
    return timeString;
  }
  
  // Парсим как UTC время - JavaScript автоматически конвертирует в локальное время
  const localTime = new Date(timeString);
  if (isNaN(localTime.getTime())) return null;
  
  return localTime;
}

/**
 * Форматирует время в локальном часовом поясе пользователя
 * @param {string|Date} timeString - время в UTC формате
 * @returns {string} - отформатированное время (HH:MM)
 */
export function formatTimeInLocal(timeString) {
  const localTime = toLocalTime(timeString);
  if (!localTime) return "—";
  
  return localTime.toLocaleTimeString("ru-RU", { 
    hour12: false, 
    hour: "2-digit", 
    minute: "2-digit"
  });
}

/**
 * Форматирует дату в локальном часовом поясе пользователя
 * @param {string|Date} timeString - время в UTC формате
 * @returns {string} - отформатированная дата (DD.MM.YYYY)
 */
export function formatDateInLocal(timeString) {
  const localTime = toLocalTime(timeString);
  if (!localTime) return "";
  
  return localTime.toLocaleDateString("ru-RU", {
    day: "2-digit",
    month: "2-digit", 
    year: "numeric"
  });
}

/**
 * Форматирует дату для отображения в списке (день месяц год)
 * @param {string|Date} timeString - время в UTC формате
 * @returns {string} - отформатированная дата (день месяц год)
 */
export function formatDateForDisplay(timeString) {
  const localTime = toLocalTime(timeString);
  if (!localTime) return "";
  
  return localTime.toLocaleDateString("ru-RU", {
    day: "2-digit",
    month: "long",
    year: "numeric"
  });
}

/**
 * Форматирует время для формы (HH:MM)
 * @param {string|Date} timeString - время в UTC формате
 * @returns {string} - время в формате HH:MM
 */
export function formatTimeForForm(timeString) {
  const localTime = toLocalTime(timeString);
  if (!localTime) return "";
  
  const fmt = new Intl.DateTimeFormat('ru-RU', { hour: '2-digit', minute: '2-digit', hour12: false });
  const parts = fmt.formatToParts(localTime);
  const hours = parts.find(p => p.type === 'hour')?.value?.padStart(2, '0') || '00';
  const minutes = parts.find(p => p.type === 'minute')?.value?.padStart(2, '0') || '00';
  return `${hours}:${minutes}`;
}

/**
 * Форматирует дату для формы (YYYY-MM-DD)
 * @param {string|Date} timeString - время в UTC формате
 * @returns {string} - дата в формате YYYY-MM-DD
 */
export function formatDateForForm(timeString) {
  const localTime = toLocalTime(timeString);
  if (!localTime) return "";
  const fmt = new Intl.DateTimeFormat('ru-RU', { year: 'numeric', month: '2-digit', day: '2-digit' });
  const parts = fmt.formatToParts(localTime);
  const year = parts.find(p => p.type === 'year')?.value || '1970';
  const month = parts.find(p => p.type === 'month')?.value?.padStart(2, '0') || '01';
  const day = parts.find(p => p.type === 'day')?.value?.padStart(2, '0') || '01';
  return `${year}-${month}-${day}`;
}

/**
 * Формирует ISO-строку (UTC) из даты (YYYY-MM-DD) и времени (HH:MM),
 * считая, что введенное время — это Europe/Moscow (UTC+3, без переходов).
 */
export function toIsoFromMoscow(dateStr, timeStr) {
  if (!dateStr || !timeStr) return "";
  const [y, m, d] = dateStr.split('-').map((v) => parseInt(v, 10));
  const [hh, mm] = timeStr.split(':').map((v) => parseInt(v, 10));
  if (!y || !m || !d || isNaN(hh) || isNaN(mm)) return "";
  // Moscow is UTC+3 → UTC time = Moscow time - 3 hours
  const utcDate = new Date(Date.UTC(y, m - 1, d, hh - 3, mm, 0, 0));
  return utcDate.toISOString();
}
