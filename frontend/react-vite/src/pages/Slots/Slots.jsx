import React, { useMemo, useState, useEffect } from 'react';
import './slots.css';
import LeftSidebar from '../../components/LeftSidebar/LeftSidebar';
import RightSidebar from '../../components/RightSidebar/RightSidebar';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query';
import { apiService } from '../../utils/api';
import { formatTimeInLocal, formatDateForDisplay } from '../../utils/timeUtils';
import { isAuthenticated, getCurrentUser } from '../../utils/auth';
import { useToast } from '../../components/Toast';

function fetchSlots(telegramId, serviceId, dateFilter) {
  return apiService.slot.getByMaster(telegramId).then((r) => {
    let slots = Array.isArray(r.data) ? r.data : (r.data?.slots || r.data?.data || []);
    
    // Фильтр по времени: скрываем слоты, которые прошли более часа назад
    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    slots = slots.filter((s) => {
      const slotStart = s.start_time ? new Date(s.start_time) : null;
      return !slotStart || slotStart >= oneHourAgo;
    });
    
    // Фильтр по услуге
    if (serviceId && serviceId !== "ALL") {
      slots = slots.filter((s) => s.service_id === parseInt(serviceId));
    }
    
    // Фильтр по дате: YYYY-MM-DD
    if (dateFilter && dateFilter !== "ALL") {
      slots = slots.filter((s) => {
        const d = s.start_time ? new Date(s.start_time) : null;
        if (!d) return false;
        const y = d.getFullYear();
        const m = String(d.getMonth() + 1).padStart(2, "0");
        const day = String(d.getDate()).padStart(2, "0");
        return `${y}-${m}-${day}` === dateFilter;
      });
    }
    
    return slots;
  }).catch(() => []);
}

function fetchServices(telegramId) {
  return apiService.service.getByMaster(telegramId).then((r) => {
    return Array.isArray(r.data) ? r.data : (r.data?.services || []);
  }).catch(() => []);
}

function createRecord(payload) {
  return apiService.record.create(payload).then((r) => r.data);
}

export default function Slots() {
  const params = useParams();
  const telegramId = params.telegramId;
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { showSuccess, showError } = useToast();
  
  const authed = isAuthenticated();
  const user = useMemo(() => getCurrentUser(), []);
  const userId = user?.id;
  
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const toggleMobileMenu = () => setIsMobileMenuOpen(!isMobileMenuOpen);

  const [serviceId, setServiceId] = useState('ALL');
  const [dateFilter, setDateFilter] = useState('ALL');
  const [pendingSlot, setPendingSlot] = useState(null);
  const [needAuth, setNeedAuth] = useState(false);
  const [publicMaster, setPublicMaster] = useState(null);
  const [cachedMasterId, setCachedMasterId] = useState(null);
  const [userRecords, setUserRecords] = useState(new Map());

  // Получаем master_id из первого слота или используем кэшированный
  const { data: slots = [], isLoading } = useQuery({
    queryKey: ["public-slots", telegramId, serviceId, dateFilter],
    queryFn: () => fetchSlots(telegramId, serviceId, dateFilter),
    enabled: Boolean(telegramId),
    refetchInterval: 30000,
    refetchIntervalInBackground: true,
  });

  const masterId = slots.length > 0 ? slots[0].master_id : cachedMasterId;

  useEffect(() => {
    if (slots.length > 0 && slots[0].master_id) {
      setCachedMasterId(slots[0].master_id);
    }
    if (slots.length > 0) {
      const s0 = slots[0];
      setPublicMaster((prev) => prev || {
        id: s0.master_id,
        first_name: s0.master_name || undefined,
        surname: s0.master_surname || undefined,
        telegram_id: s0.master_telegram_id || undefined,
      });
    }
  }, [slots]);

  // Загружаем публичные данные мастера
  useEffect(() => {
    const idToLoad = masterId || telegramId;
    if (!idToLoad || !apiService.user?.getPublic) return;
    apiService.user.getPublic(idToLoad)
      .then((r) => {
        const data = r.data?.user ? r.data.user : r.data;
        setPublicMaster(data);
      })
      .catch(() => setPublicMaster(null));
  }, [masterId, telegramId]);

  // Получаем все слоты без фильтрации для dateOptions
  const { data: allSlots = [] } = useQuery({
    queryKey: ["public-slots-all", telegramId, serviceId],
    queryFn: () => fetchSlots(telegramId, serviceId, "ALL"),
    enabled: Boolean(telegramId),
    refetchInterval: 30000,
    refetchIntervalInBackground: true,
  });

  const { data: services = [] } = useQuery({
    queryKey: ["services", masterId],
    queryFn: () => fetchServices(masterId),
    enabled: Boolean(masterId),
    staleTime: 60000,
    refetchInterval: 60000,
    refetchIntervalInBackground: true,
  });

  // Загружаем записи пользователя
  const { data: userRecordsData = [] } = useQuery({
    queryKey: ["user-records", userId],
    queryFn: () => apiService.record.getByClient(userId).then((r) => r.data?.data || r.data || []),
    enabled: Boolean(userId),
    staleTime: 30000,
    refetchInterval: 30000,
    refetchIntervalInBackground: true,
  });

  useEffect(() => {
    if (userRecordsData && Array.isArray(userRecordsData)) {
      const recordMap = new Map();
      userRecordsData.forEach(record => {
        recordMap.set(record.slot_id, record);
      });
      setUserRecords(recordMap);
    }
  }, [userRecordsData]);

  const recordMutation = useMutation({
    mutationFn: (payload) => createRecord(payload),
    onSuccess: () => {
      setPendingSlot(null);
      queryClient.invalidateQueries({ queryKey: ["public-slots", telegramId, serviceId, dateFilter] });
      queryClient.invalidateQueries({ queryKey: ["public-slots-all", telegramId, serviceId] });
      queryClient.invalidateQueries({ queryKey: ["user-records", userId] });
      queryClient.invalidateQueries({ queryKey: ["services", masterId] });
      showSuccess("Запись успешно создана! Мастер получит уведомление.");
    },
    onError: (error) => {
      const errorMessage = error.response?.data?.error || error.message || "";
      if (
        errorMessage.includes("user already has a record for this slot") ||
        errorMessage.includes("already has a record")
      ) {
        showError("Вы уже записаны на этот слот. На один слот возможна только одна запись от одного пользователя.");
      } else {
        showError("Ошибка при создании записи. Попробуйте еще раз.");
      }
    },
  });

  const handleBookClick = (slot) => {
    if (!authed) {
      setNeedAuth(true);
      return;
    }
    setPendingSlot(slot);
  };

  const confirmBooking = () => {
    if (!pendingSlot || !userId) return;
    recordMutation.mutate({ slot_id: pendingSlot.id, client_id: userId });
  };

  // Группировка по дате
  const slotsByDate = useMemo(() => {
    const map = new Map();
    for (const s of slots) {
      const key = formatDateForDisplay(s.start_time);
      if (!map.has(key)) map.set(key, []);
      map.get(key).push(s);
    }
    return Array.from(map.entries());
  }, [slots]);

  // Даты для фильтра
  const dateOptions = useMemo(() => {
    const set = new Set();
    for (const s of allSlots) {
      const d = s.start_time ? new Date(s.start_time) : null;
      if (!d) continue;
      const y = d.getFullYear();
      const m = String(d.getMonth() + 1).padStart(2, "0");
      const day = String(d.getDate()).padStart(2, "0");
      set.add(`${y}-${m}-${day}`);
    }
    return Array.from(set).sort();
  }, [allSlots]);

  const selectedService = serviceId !== 'ALL' ? services.find(s => String(s.id) === String(serviceId)) : null;

  return (
    <div className="slots-container">
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

      {isMobileMenuOpen && <div className="mobile-overlay" onClick={toggleMobileMenu}></div>}

      <LeftSidebar isMobileMenuOpen={isMobileMenuOpen} toggleMobileMenu={toggleMobileMenu} />

      <main className="slots-main">
        <header className="slots-header">
          <h1>
            расписание 
            {publicMaster && (
              <span className="master-name">
                {publicMaster.first_name} {publicMaster.surname}
              </span>
            )}
          </h1>
          <p className="slots-subtitle">выберите подходящее время</p>
        </header>

        <section className="slots-layout">
          <aside className="slots-card">
            <div className="slots-card-title"><b>ФИЛЬТРЫ</b></div>

            <div className="slots-field">
              <div className="slots-label">Услуга</div>
              <select className="slots-select" value={serviceId} onChange={e => setServiceId(e.target.value)}>
                <option value="ALL">Все</option>
                {services.map(s => (
                  <option key={s.id} value={s.id}>{s.name}</option>
                ))}
              </select>
            </div>

            {selectedService && (
              <div className="slots-service-info">
                <div className="slots-service-pair">
                  <div className="slots-service-key">Цена</div>
                  <div className="slots-service-val">{selectedService.price} ₽</div>
                </div>
                <div className="slots-service-pair">
                  <div className="slots-service-key">Длительность</div>
                  <div className="slots-service-val">{selectedService.duration} мин</div>
                </div>
                {selectedService.description && (
                  <div className="slots-service-desc">
                    {selectedService.description}
                  </div>
                )}
              </div>
            )}

            <div className="slots-field">
              <div className="slots-label">Дата</div>
              <select className="slots-select" value={dateFilter} onChange={e => setDateFilter(e.target.value)}>
                <option value="ALL">Все</option>
                {dateOptions.map(d => (
                  <option key={d} value={d}>{d}</option>
                ))}
              </select>
            </div>

          <section className="slots-center">
            <div className="slots-list">
              {isLoading ? (
                <div className="slots-empty">Загрузка...</div>
              ) : slotsByDate.length === 0 ? (
                <div className="slots-empty">Нет доступных окон.</div>
              ) : (
                slotsByDate.map(([dateTitle, daySlots]) => (
                  <div key={dateTitle} className="slots-day">
                    <div className="slots-day-title">{dateTitle}</div>
                    {daySlots.map(slot => {
                      const serviceName = slot.service_id ? 
                        services.find(s => s.id === slot.service_id)?.name || "Услуга" : 
                        "Услуга";
                      
                      const isBooked = Boolean(slot.is_booked);
                      const userRecord = userRecords.get(slot.id);
                      const userHasRecord = Boolean(userRecord);
                      
                      let statusClass = "free";
                      let statusText = "";
                      let showButton = true;
                      
                      if (isBooked) {
                        statusClass = "booked";
                        statusText = "Забронирован";
                        showButton = false;
                      } else if (userHasRecord) {
                        const recordStatus = userRecord.status;
                        switch (recordStatus) {
                          case "pending":
                            statusClass = "user-pending";
                            statusText = "Заявка отправлена";
                            showButton = false;
                            break;
                          case "confirm":
                            statusClass = "user-confirmed";
                            statusText = "Одобрено";
                            showButton = false;
                            break;
                          case "reject":
                            statusClass = "user-rejected";
                            statusText = "Отклонено";
                            showButton = true;
                            break;
                          default:
                            statusClass = "user-pending";
                            statusText = "Заявка отправлена";
                            showButton = false;
                        }
                      }
                      
                      return (
                        <div key={slot.id || `${slot.start_time}-${slot.end_time}`} className="slots-item">
                          <div className="slots-item-time">{formatTimeInLocal(slot.start_time)} — {formatTimeInLocal(slot.end_time)}</div>
                          <div className="slots-item-service">{serviceName}</div>
                          <div className={`slots-item-cta ${statusClass}`}>
                            {showButton ? (
                              <button className="btn btn-primary" onClick={() => handleBookClick(slot)}>Записаться</button>
                            ) : (
                              <span>{statusText}</span>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                ))
              )}
            </div>
          </section>
          </aside>

        </section>
      </main>

      <RightSidebar />

      {pendingSlot && (
        <div className="mp-modal-backdrop" onClick={() => setPendingSlot(null)}>
          <div className="mp-modal" onClick={(e) => e.stopPropagation()}>
            <div className="mp-modal-title">Подтверждаем<br/>запись?</div>
            <div className="mp-modal-grid">
              <div className="mp-field-label">УСЛУГА</div>
              <div className="mp-field-value">
                {pendingSlot.service_id ? 
                  services.find(s => s.id === pendingSlot.service_id)?.name || "Услуга" : 
                  "—"
                }
              </div>
              <div className="mp-field-label">ДАТА</div>
              <div className="mp-field-value">
                {pendingSlot.start_time ? new Date(pendingSlot.start_time).toLocaleDateString("ru-RU") : "—"}
              </div>
              <div className="mp-field-label">ВРЕМЯ</div>
              <div className="mp-field-value">
                {pendingSlot.start_time ? formatTimeInLocal(pendingSlot.start_time) : "—"}
                {" "}–{" "}
                {pendingSlot.end_time ? formatTimeInLocal(pendingSlot.end_time) : "—"}
              </div>
            </div>
            <div className="mp-modal-actions">
              <button className="btn-primary-confirm" onClick={confirmBooking} disabled={recordMutation.isPending}>Подтвердить</button>
              <button className="btn-danger" onClick={() => setPendingSlot(null)} disabled={recordMutation.isPending}>Отмена</button>
            </div>
          </div>
        </div>
      )}

      {needAuth && (
        <div className="mp-modal-backdrop" onClick={() => setNeedAuth(false)}>
          <div className="mp-modal" onClick={(e) => e.stopPropagation()}>
            <div className="mp-modal-title">ВОЙДИ В ОДИН<br/> КЛИК</div>
            <div className="mp-modal-note">Для записи необходимо<br/>авторизоваться</div>
            <div className="mp-modal-actions">
              <button className="btn-primary-login" onClick={() => navigate("/login")}>Войти</button>
              <button className="btn-danger" onClick={() => setNeedAuth(false)}>Отмена</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
