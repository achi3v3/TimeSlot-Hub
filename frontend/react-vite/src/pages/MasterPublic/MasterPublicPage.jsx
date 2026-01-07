import { useMemo, useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiService } from "../../utils/api";
import Sidebar from "../../components/Sidebar/Sidebar";
import { getCurrentUser, isAuthenticated } from "../../utils/auth";
import { useToast } from "../../components/Toast";
import { formatTimeInLocal, formatDateForDisplay } from "../../utils/timeUtils";
import "./masterPublic.css";

function fetchSlots(telegramId, serviceId, dateFilter) {
  console.log("MasterPublic - Загружаем слоты для telegramId:", telegramId, "serviceId:", serviceId, "dateFilter:", dateFilter);
  // Базовый эндпоинт слотов для мастера (анонимно)
  return apiService.slot.getByMaster(telegramId).then((r) => {
    console.log("MasterPublic - Ответ от сервера слотов:", r.data);
    let slots = Array.isArray(r.data) ? r.data : (r.data?.slots || r.data?.data || []);
    console.log("MasterPublic - Исходные слоты:", slots);
    
    // Фильтр по времени: скрываем слоты, которые прошли более часа назад
    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    slots = slots.filter((s) => {
      const slotStart = s.start_time ? new Date(s.start_time) : null;
      return !slotStart || slotStart >= oneHourAgo;
    });
    console.log("MasterPublic - Слоты после фильтрации по времени:", slots);
    
    // Фильтр по услуге: фильтруем по service_id
    if (serviceId && serviceId !== "ALL") {
      console.log("MasterPublic - Фильтруем по serviceId:", serviceId);
      slots = slots.filter((s) => s.service_id === parseInt(serviceId));
      console.log("MasterPublic - Слоты после фильтрации по услуге:", slots);
    }
    
    // Фильтр по дате: YYYY-MM-DD
    if (dateFilter && dateFilter !== "ALL") {
      console.log("MasterPublic - Фильтруем по дате:", dateFilter);
      slots = slots.filter((s) => {
        const d = s.start_time ? new Date(s.start_time) : null;
        if (!d) return false;
        const y = d.getFullYear();
        const m = String(d.getMonth() + 1).padStart(2, "0");
        const day = String(d.getDate()).padStart(2, "0");
        return `${y}-${m}-${day}` === dateFilter;
      });
      console.log("MasterPublic - Слоты после фильтрации по дате:", slots);
    }
    
    console.log("MasterPublic - Итоговые слоты:", slots);
    return slots;
  }).catch((error) => {
    console.error("MasterPublic - Ошибка загрузки слотов:", error);
    return [];
  });
}

function fetchServices(telegramId) {
  console.log("MasterPublic - Загружаем услуги для telegramId:", telegramId);
  // Используем правильный эндпоинт для услуг мастера
  return apiService.service.getByMaster(telegramId).then((r) => {
    console.log("MasterPublic - Ответ от сервера услуг:", r.data);
    const arr = Array.isArray(r.data) ? r.data : (r.data?.services || []);
    console.log("MasterPublic - Обработанные услуги:", arr);
    return arr;
  }).catch((error) => {
    console.error("MasterPublic - Ошибка загрузки услуг:", error);
    console.error("MasterPublic - Статус ошибки:", error.response?.status);
    console.error("MasterPublic - Данные ошибки:", error.response?.data);
    // Фолбек до пустого списка
    return [];
  });
}

function createRecord(payload) {
  return apiService.record.create(payload).then((r) => r.data);
}

function checkUserRecord(slotId, userId) {
  return apiService.record.getByClient(userId).then((r) => {
    const records = r.data?.data || r.data || [];
    return records.some(record => record.slot_id === slotId);
  }).catch(() => false);
}


export default function MasterPublicPage() {
  const params = useParams();
  const telegramId = params.telegramId; // фактически сюда передаем UUID мастера
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { showSuccess, showError } = useToast();

  const authed = isAuthenticated();
  const user = useMemo(() => getCurrentUser(), []);
  const userId = user?.id; // UUID пользователя

  const [serviceId, setServiceId] = useState("ALL");
  const [dateFilter, setDateFilter] = useState("ALL");
  const [pendingSlot, setPendingSlot] = useState(null);
  const [needAuth, setNeedAuth] = useState(false);
  const [contactOpen, setContactOpen] = useState(false);
  const [userRecords, setUserRecords] = useState(new Set());
  const [publicMaster, setPublicMaster] = useState(null);

  const { data: slots = [], isLoading } = useQuery({
    queryKey: ["public-slots", telegramId, serviceId, dateFilter],
    queryFn: () => fetchSlots(telegramId, serviceId, dateFilter),
    enabled: Boolean(telegramId),
    refetchInterval: 30000, // Автообновление каждые 30 секунд
    refetchIntervalInBackground: true, // Обновлять даже когда вкладка неактивна
  });

  // Отдельный запрос для получения всех слотов без фильтрации по дате (для dateOptions)
  const { data: allSlots = [] } = useQuery({
    queryKey: ["public-slots-all", telegramId, serviceId],
    queryFn: () => fetchSlots(telegramId, serviceId, "ALL"), // Всегда получаем все слоты
    enabled: Boolean(telegramId),
    refetchInterval: 30000, // Автообновление каждые 30 секунд
    refetchIntervalInBackground: true,
  });

  // Получаем master_id из первого слота или сохраняем предыдущий
  const [cachedMasterId, setCachedMasterId] = useState(null);
  const masterId = slots.length > 0 ? slots[0].master_id : cachedMasterId;
  
  // Сохраняем masterId в кэш
  useEffect(() => {
    if (slots.length > 0 && slots[0].master_id) {
      setCachedMasterId(slots[0].master_id);
    }
    // Если пришли данные по слотам — используем их для публичных полей
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

  // Загружаем публичные данные мастера по UUID, когда известен masterId
  useEffect(() => {
    const idToLoad = masterId || telegramId; // пробуем masterId; если еще нет — используем параметр
    if (!idToLoad) return;
    if (!apiService.user?.getPublic) return;
    apiService.user.getPublic(idToLoad)
      .then((r) => {
        const data = r.data?.user ? r.data.user : r.data;
        setPublicMaster(data);
      })
      .catch(() => setPublicMaster(null));
  }, [masterId, telegramId]);
  
  const { data: services = [] } = useQuery({
    queryKey: ["services", masterId],
    queryFn: () => fetchServices(masterId),
    enabled: Boolean(masterId),
    staleTime: 60_000,
    refetchInterval: 60000, // Автообновление каждые 60 секунд
    refetchIntervalInBackground: true,
  });

  // Загружаем записи пользователя
  const { data: userRecordsData = [] } = useQuery({
    queryKey: ["user-records", userId],
    queryFn: () => apiService.record.getByClient(userId).then((r) => r.data?.data || r.data || []),
    enabled: Boolean(userId),
    staleTime: 30_000,
    refetchInterval: 30000, // Автообновление каждые 30 секунд
    refetchIntervalInBackground: true,
  });

  // Обновляем состояние записей пользователя с учетом статусов
  useEffect(() => {
    if (userRecordsData && Array.isArray(userRecordsData)) {
      const recordMap = new Map();
      userRecordsData.forEach(record => {
        recordMap.set(record.slot_id, record);
      });
      setUserRecords(recordMap);
    }
  }, [userRecordsData]);
  
  console.log("MasterPublic - Загруженные услуги:", services);
  console.log("MasterPublic - TelegramId:", telegramId);
  console.log("MasterPublic - MasterId из слотов:", masterId);
  console.log("MasterPublic - Кэшированный MasterId:", cachedMasterId);

  const recordMutation = useMutation({
    mutationFn: (payload) => createRecord(payload),
    onSuccess: () => {
      setPendingSlot(null);
      // Инвалидируем все связанные запросы для обновления данных
      queryClient.invalidateQueries({ queryKey: ["public-slots", telegramId, serviceId, dateFilter] });
      queryClient.invalidateQueries({ queryKey: ["public-slots-all", telegramId, serviceId] });
      queryClient.invalidateQueries({ queryKey: ["user-records", userId] });
      queryClient.invalidateQueries({ queryKey: ["services", masterId] });
      showSuccess("Запись успешно создана! Мастер получит уведомление.");
    },
    onError: (error) => {
      console.error("Ошибка создания записи:", error);
      
      // Проверяем, является ли ошибка дублированием записи
      const errorMessage = error.response?.data?.error || error.message || "";
      if (errorMessage.includes("user already has a record for this slot") || 
          errorMessage.includes("already has a record")) {
        showError("Вы уже записаны на этот слот. На один слот возможна только одна запись от одного пользователя.");
      } else {
        showError("Ошибка при создании записи. Попробуйте еще раз.");
      }
    },
  });


  const handleWriteMaster = () => {
    setContactOpen(true);
  };

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


  // Группировка по дате для рендера
  const slotsByDate = useMemo(() => {
    const map = new Map();
    for (const s of slots) {
      const key = formatDateForDisplay(s.start_time);
      if (!map.has(key)) map.set(key, []);
      map.get(key).push(s);
    }
    return Array.from(map.entries());
  }, [slots]);


  // Для фильтра дат составим список уникальных дат YYYY-MM-DD из всех слотов
  const dateOptions = useMemo(() => {
    const set = new Set();
    for (const s of allSlots) { // Используем allSlots вместо slots
      const d = s.start_time ? new Date(s.start_time) : null;
      if (!d) continue;
      const y = d.getFullYear();
      const m = String(d.getMonth() + 1).padStart(2, "0");
      const day = String(d.getDate()).padStart(2, "0");
      set.add(`${y}-${m}-${day}`);
    }
    return Array.from(set).sort(); // Сортируем даты
  }, [allSlots]); // Зависимость от allSlots

  return (
    <div className="master-public-root">
      <Sidebar />
      <div className="master-public-main-content">
      <div className="mp-main">
          <h3 className="mp-title">РАСПИСАНИЕ</h3>
      <main className="mp-layout">
        <aside className="mp-card">
            <div className="mp-header-card"><b>ДАННЫЕ ПОЛЬЗОВАТЕЛЯ</b></div>
          <div className="mp-user">
            <div className="mp-row"><b>ИМЯ:</b><div className="mp-value">{publicMaster?.first_name || "—"}</div></div>
            <div className="mp-row"><b>ФАМИЛИЯ:</b><div className="mp-value">{publicMaster?.surname || "—"}</div></div>
          </div>

          <div className="mp-service">
            <div className="mp-service-header">
              <span className="mp-service-title">УСЛУГА:</span>
              <div className="service-selector">
                <select
                  className="mp-select mp-value"
                  value={serviceId}
                  onChange={(e) => setServiceId(e.target.value)}
                >
                  <option value="ALL">ВСЕ</option>
                  {(services || []).map((svc, index) => (
                    <option key={svc.id || `service-${index}`} value={svc.id}>
                      {svc.name || `Услуга ${svc.id || index + 1}`}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            {serviceId !== "ALL" && (() => {
              const selectedService = (services || []).find(s => s.id === parseInt(serviceId));
              return selectedService ? (
                <div className="mp-service-info">
                  <div className="mp-service-price-duration">
                    <div className="mp-service-price">
                      <span className="mp-service-label">Цена:</span>
                      <span className="mp-service-value">{selectedService.price} ₽</span>
                    </div>
                    <div className="mp-service-duration">
                      <span className="mp-service-label">Длительность:</span>
                      <span className="mp-service-value">{selectedService.duration} мин</span>
                    </div>
                  </div>
                {selectedService.description && (
                  <div className="mp-service-desc">
                    <div className="mp-service-desc-label">Описание:</div>
                    <div className="mp-service-desc-text" style={{ whiteSpace: 'pre-wrap', overflowY: 'auto', maxHeight: 160 }}>
                      {selectedService.description}
                    </div>
                  </div>
                )}
                </div>
              ) : null;
            })()}
          </div>
          {/* <button className="mp-btn-send" onClick={handleWriteMaster}>Написать мастеру</button> */}
        </aside>

        <section className="mp-center">
        <div className="mp-filter" style={{ display:'flex', alignItems:'center', gap: 10, margin: '0 10% 10px' }}>
              <div className="mp-filter-label">ПОКАЗЫВАТЬ</div>
              <select className="mp-select" value={dateFilter} onChange={(e) => setDateFilter(e.target.value)}>
                <option value="ALL">ВСЕ</option>
                {dateOptions.map((d) => (
                  <option key={d} value={d}>{d}</option>
                ))}
              </select>
            </div>
            {/* Селект услуги перенесён в левую колонку */}

          <div className="mp-slots">
            {isLoading ? (
              <div>Загрузка...</div>
            ) : slotsByDate.length === 0 ? (
              <div>Нет доступных окон.</div>
            ) : (
              slotsByDate.map(([dateTitle, daySlots]) => (
                <div key={dateTitle} className="mp-day">
                  <div className="mp-day-title">{dateTitle}</div>
                  {daySlots.map((slot) => {
                    const isBooked = Boolean(slot.is_booked);
                    const userRecord = userRecords.get(slot.id);
                    const userHasRecord = Boolean(userRecord);
                    // Find service name by service_id
                    const serviceName = slot.service_id ? 
                      (services || []).find(s => s.id === slot.service_id)?.name || "Услуга" : 
                      "Услуга";
                    
                    // Определяем статус и текст для отображения
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
                          showButton = true; // Можно подать заявку повторно
                          break;
                        default:
                          statusClass = "user-pending";
                          statusText = "Заявка отправлена";
                          showButton = false;
                      }
                    }
                    
                    console.log("MasterPublic - Слот:", slot);
                    console.log("MasterPublic - User record:", userRecord);
                    console.log("MasterPublic - Status:", statusClass);
                    
                    return (
                      <div key={slot.id || `${slot.start_time}-${slot.end_time}`} className="mp-slot-card">
                        <div className="mp-slot-top">
                          <div className="mp-slot-time">{formatTimeInLocal(slot.start_time)} — {formatTimeInLocal(slot.end_time)}</div>
                          <div className="mp-slot-service">{serviceName}</div>
                          <div className={`mp-slot-badge ${statusClass}`}>
                            {showButton ? (
                              <button className="mp-btn-follow" onClick={() => handleBookClick(slot)}>Записаться</button>
                            ) : (
                              statusText
                            )}
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              ))
            )}
          </div>
        </section>

        {/* Правая колонка удалена */}
      </main>
      </div>

      {pendingSlot && (
        <div className="mp-modal-backdrop" onClick={() => setPendingSlot(null)}>
          <div className="mp-modal" onClick={(e) => e.stopPropagation()}>
            <div className="mp-modal-title">Подтверждаем<br/>запись?</div>
            <div className="mp-modal-grid">
              <div className="mp-field-label">УСЛУГА</div>
              <div className="mp-field-value">
                {pendingSlot.service_id ? 
                  (services || []).find(s => s.id === pendingSlot.service_id)?.name || "Услуга" : 
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

      {contactOpen && (
        <div className="mp-modal-backdrop" onClick={() => setContactOpen(false)}>
          <div className="mp-modal" onClick={(e) => e.stopPropagation()}>
          <div className="mp-modal-title">Связаться с<br/> мастером</div>
          <div className="mp-modal-note">Напишите пользователю в телеграм, чтобы уточнить детали.</div>
            <div className="mp-modal-actions">
              <a
                className="btn-primary"
                href={publicMaster?.telegram_id ? `https://t.me/${publicMaster.telegram_id}` : `https://t.me/`}
                target="_blank"
                rel="noreferrer"
              >
                Открыть диалог
              </a>
              <button className="btn-danger" onClick={() => setContactOpen(false)}>Отмена</button>
            </div>
          </div>
        </div>
      )}

      </div>
    </div>
  );
}


