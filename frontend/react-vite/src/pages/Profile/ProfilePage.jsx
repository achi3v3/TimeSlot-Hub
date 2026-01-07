import { useMemo, useState, useRef } from "react";
import LeftSidebar from "../../components/LeftSidebar/LeftSidebar";
import { getCurrentUser, setCurrentUser } from "../../utils/auth";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiService } from "../../utils/api";
import { useToast } from "../../components/Toast";
import { formatTimeInLocal, formatDateForDisplay, formatTimeForForm, formatDateForForm } from "../../utils/timeUtils";
import "./profile.css";
import { handleLogout } from '../../utils/auth';
import RecordStatusSelect from "./RecordStatusSelect";
import ApplicationsSection from "./ApplicationsSection";

function fetchSlots(masterId) {
  return apiService.slot.getByMaster(masterId).then((r) => r.data);
}

function createSlot(payload) {
  return apiService.slot.create(payload).then((r) => r.data);
}

function deleteAllSlots(masterId) {
  return apiService.slot.deleteAll(masterId).then((r) => r.data);
}

function fetchRecordsBySlot(slotId) {
  return apiService.record.getBySlotAll(slotId).then((r) => r.data?.data || r.data || []);
}

function confirmRecord(recordId) {
  return apiService.record.confirm(recordId).then((r) => r.data);
}

function rejectRecord(recordId) {
  return apiService.record.reject(recordId).then((r) => r.data);
}

function deleteSlot(slotId) {
  return apiService.slot.deleteOne(slotId).then((r) => r.data);
}

function updateSlot(payload) {
  return apiService.slot.update(payload).then((r) => r.data);
}

function fetchServices(masterId) {
  return apiService.service.getByMaster(masterId)
    .then((r) => {
      const arr = Array.isArray(r.data) ? r.data : (r.data?.services || []);
      return arr;
    })
    .catch(() => []);
}

function createService(payload) {
  console.log("Отправляем запрос на создание услуги с данными:", payload);
  return apiService.service.create(payload)
    .then((r) => {
      console.log("Ответ от сервера:", r.data);
      return r.data;
    })
    .catch((error) => {
      console.error("Ошибка API запроса:", error);
      throw error;
    });
}

export default function ProfilePage() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };
  const queryClient = useQueryClient();
  const { showSuccess, showError, showInfo } = useToast();
  const [user, setUser] = useState(getCurrentUser());
  const userId = user?.id;
  const telegramId = user?.telegram_id; // Оставляем для совместимости, но не используем для API
  
  const REFRESH_MS = 5000;
  const { data: slotsResponse, isLoading } = useQuery({
    queryKey: ["slots", userId],
    queryFn: () => fetchSlots(userId),
    enabled: Boolean(userId),
    staleTime: REFRESH_MS,
    refetchInterval: REFRESH_MS,
    refetchIntervalInBackground: true,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
  });
const copyToClipboard = async (text) => {
  try {
    // Пробуем современный способ
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    } else {
      // Fallback для HTTP
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.opacity = '0';
      document.body.appendChild(textArea);
      textArea.select();
      
      const successful = document.execCommand('copy');
      document.body.removeChild(textArea);
      return successful;
    }
  } catch (err) {
    return false;
  }
};
  const slots = useMemo(() => {
    const r = slotsResponse;
    const base = Array.isArray(r) ? r : (r?.slots ?? r?.data ?? []);
    return base || [];
  }, [slotsResponse]);

  // helpers for default date/time
  const todayStr = useMemo(() => {
    const d = new Date();
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
  }, []);

  const [form, setForm] = useState({
    date: todayStr,
    start: "10:00",
    end: "11:00",
    is_booked: false,
  });

  // keep last created duration and form in refs to compute next slot defaults
  const lastDurationRef = useRef(60);
  const lastFormRef = useRef(form);
  const lastEndRef = useRef(null);
  const [edit, setEdit] = useState({ date: "", start: "", end: "" });
  const [serviceId, setServiceId] = useState("");
  const [showServiceModal, setShowServiceModal] = useState(false);
  const [serviceForm, setServiceForm] = useState({
    name: "",
    description: "",
    price: "",
    duration: ""
  });
  const [serviceEditMap, setServiceEditMap] = useState({});

  // Reusable confirmation dialog
  const [confirmDlg, setConfirmDlg] = useState({ open: false, title: "", text: "", onConfirm: null });
  const openConfirm = (title, text, onConfirm, options = {}) => setConfirmDlg({ open: true, title, text, onConfirm, textStyle: options.textStyle});
  const closeConfirm = () => setConfirmDlg({ open: false, title: "", text: "", onConfirm: null });

  const [activeSlot, setActiveSlot] = useState(null);
  const [accountForm, setAccountForm] = useState({ first_name: user?.first_name || "", surname: user?.surname || "" });
  
  // Локальное состояние для изменений статусов заявок
  const [recordStatusChanges, setRecordStatusChanges] = useState({});
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const { data: slotRecords = [], refetch: refetchRecords } = useQuery({
    queryKey: ["slot-records", activeSlot?.id],
    queryFn: () => fetchRecordsBySlot(activeSlot.id),
    enabled: Boolean(activeSlot?.id),
    staleTime: REFRESH_MS,
    refetchInterval: REFRESH_MS,
    refetchIntervalInBackground: true,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
  });

  const slotRecordsDisplay = slotRecords;

  // Сброс локальных изменений при смене слота
  const handleSlotClick = (slot) => {
    if (hasUnsavedChanges) {
      openConfirm(
        'Несохраненные изменения',
        'У вас есть несохраненные изменения статусов заявок. Вы уверены, что хотите закрыть без сохранения?',
        () => {
          setRecordStatusChanges({});
          setHasUnsavedChanges(false);
          setActiveSlot(slot);
        }
      );
    } else {
      setActiveSlot(slot);
    }
  };

  // Функция для закрытия модального окна с проверкой несохраненных изменений
  const handleCloseSlotModal = () => {
    if (hasUnsavedChanges) {
      openConfirm(
        'Несохраненные изменения',
        'У вас есть несохраненные изменения статусов заявок. Вы уверены, что хотите закрыть без сохранения?',
        () => {
          setRecordStatusChanges({});
          setHasUnsavedChanges(false);
          setActiveSlot(null);
        }
      );
    } else {
      setActiveSlot(null);
    }
  };

  const confirmMut = useMutation({
    mutationFn: (id) => confirmRecord(id),
    onSuccess: () => {
      refetchRecords();
    },
  });

  const rejectMut = useMutation({
    mutationFn: (id) => rejectRecord(id),
    onSuccess: () => {
      refetchRecords();
    },
  });

  // Функция для сохранения всех изменений статусов заявок
  const saveRecordStatusChanges = () => {
    const changes = Object.entries(recordStatusChanges);
    if (changes.length === 0) return;

    // Выполняем все изменения последовательно
    const promises = changes.map(([recordId, newStatus]) => {
      if (newStatus === "confirm") {
        return confirmRecord(recordId);
      } else if (newStatus === "reject") {
        return rejectRecord(recordId);
      } else if (newStatus === "pending") {
        // Map pending to backend expected value if needed
        return apiService.record.updateStatus(recordId, "PENDING");
      }
      return Promise.resolve();
    });

    Promise.all(promises)
      .then(() => {
        setRecordStatusChanges({});
        setHasUnsavedChanges(false);
        refetchRecords();
        showSuccess("Изменения статусов заявок сохранены");
      })
      .catch((error) => {
        console.error("Ошибка сохранения изменений:", error);
        showError("Ошибка при сохранении изменений");
      });
  };

  // Функция для сброса изменений
  const resetRecordStatusChanges = () => {
    setRecordStatusChanges({});
    setHasUnsavedChanges(false);
  };

  const deleteSlotMut = useMutation({
    mutationFn: (slotId) => deleteSlot(slotId),
    onSuccess: () => {
      setActiveSlot(null);
      queryClient.invalidateQueries({ queryKey: ["slots", userId] });
    },
  });

  const updateSlotMut = useMutation({
    mutationFn: (payload) => updateSlot(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["slots", userId] });
      setActiveSlot(null);
    },
  });

  const createMutation = useMutation({
    mutationFn: (payload) => createSlot(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["slots", userId] });
      // After creating a slot: compute next defaults from absolute last end to avoid midnight wrap issues
      const durationMin = Number(lastDurationRef.current) || 60;
      const base = lastEndRef.current;
      if (base && !isNaN(base.getTime())) {
        const nextStart = new Date(base);
        const nextEnd = new Date(base.getTime() + durationMin * 60 * 1000);
        const y = nextStart.getFullYear();
        const m = String(nextStart.getMonth() + 1).padStart(2, "0");
        const d = String(nextStart.getDate()).padStart(2, "0");
        const sh = String(nextStart.getHours()).padStart(2, "0");
        const sm = String(nextStart.getMinutes()).padStart(2, "0");
        const eh = String(nextEnd.getHours()).padStart(2, "0");
        const em = String(nextEnd.getMinutes()).padStart(2, "0");
        setForm({ date: `${y}-${m}-${d}`, start: `${sh}:${sm}`, end: `${eh}:${em}`, is_booked: false });
      }
      showSuccess("Слот успешно создан! Время сдвинуто по предыдущей длительности.");
    },
    onError: (error) => {
      console.error("Ошибка создания слота:", error);
      const msg = error?.response?.data?.error || "Ошибка при создании слота. Попробуйте еще раз.";
      showError(msg);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async () => {
      // Используем только UUID для удаления
      if (userId) {
        return await deleteAllSlots(userId);
      }
      throw new Error("no-user-id");
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["slots", userId] });
      showSuccess("Все слоты успешно удалены!");
    },
    onError: (error) => {
      console.error("Ошибка удаления слотов:", error);
      showError("Ошибка при удалении слотов. Попробуйте еще раз.");
    },
  });

  const createServiceMutation = useMutation({
    mutationFn: (payload) => createService(payload),
    onSuccess: (data) => {
      console.log("Service created successfully:", data);
      queryClient.invalidateQueries({ queryKey: ["services", userId] });
      queryClient.refetchQueries({ queryKey: ["services", userId] });
      setShowServiceModal(false);
      setServiceForm({ name: "", description: "", price: "", duration: "" });
      // Автоматически выбираем созданную услугу
      if (data && data.id) {
        setServiceId(data.id.toString());
        console.log("Автоматически выбрана созданная услуга:", data.id);
      }
      showSuccess("Услуга успешно создана!");
    },
    onError: (error) => {
      console.error("Service creation failed:", error);
      showError("Ошибка при создании услуги: " + (error.message || "Неизвестная ошибка"));
    },
  });

  // Справочник услуг мастера
  const { data: services = [] } = useQuery({
    queryKey: ["services", userId],
    queryFn: () => fetchServices(userId),
    enabled: Boolean(userId),
    staleTime: REFRESH_MS,
    refetchInterval: REFRESH_MS,
    refetchIntervalInBackground: true,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
  });

  // Загружаем записи пользователя для раздела "Мои заявки"
  const { data: userRecordsData = [] } = useQuery({
    queryKey: ["user-records", userId],
    queryFn: () => apiService.record.getByClient(userId).then((r) => r.data?.data || r.data || []),
    enabled: Boolean(userId),
    staleTime: REFRESH_MS,
    refetchInterval: REFRESH_MS,
    refetchIntervalInBackground: true,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
  });

  const recordsSource = userRecordsData;

  // Фильтры для заявок
  const [statusFilter, setStatusFilter] = useState("ALL");
  const [showPastApplications, setShowPastApplications] = useState(false);
  
  // Фильтр для слотов в профиле
  const [showPastSlots, setShowPastSlots] = useState(false);

  // Обработка и фильтрация заявок
  const processedApplications = useMemo(() => {
    if (!recordsSource || recordsSource.length === 0) return [];

    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);

    return recordsSource
      .map((record) => {
        const slot = record.slot;
        const start = slot?.start_time ? new Date(slot.start_time) : null;
        const end = slot?.end_time ? new Date(slot.end_time) : null;
        
        return {
          ...record,
          slot,
          start,
          end,
          isPast: start ? start < oneHourAgo : false
        };
      })
      .filter((record) => {
        // Фильтр по статусу
        if (statusFilter !== "ALL") {
          const statusMap = {
            "CONFIRMED": "confirm",
            "REJECTED": "reject", 
            "PENDING": "pending"
          };
          if (record.status !== statusMap[statusFilter]) {
            return false;
          }
        }

        // Фильтр по прошедшим заявкам
        if (!showPastApplications && record.isPast) {
          return false;
        }

        return true;
      })
      .sort((a, b) => {
        // Сортировка: сначала ближайшие по времени начала
        if (!a.start && !b.start) return 0;
        if (!a.start) return 1;
        if (!b.start) return -1;
        return a.start.getTime() - b.start.getTime();
      });
  }, [recordsSource, statusFilter, showPastApplications]);
  
  console.log("Загруженные услуги:", services);
  console.log("Текущий serviceId:", serviceId);
  
  // Детальная информация об услугах
  if (services && services.length > 0) {
    console.log("Первая услуга:", services[0]);
    console.log("Все поля первой услуги:", Object.keys(services[0]));
  }

  const handleChange = (e) => {
    const { name, value } = e.target;
    setForm((prev) => {
      const newForm = { ...prev, [name]: value };
      
      // Автоматически вычисляем время окончания при изменении времени начала
      if (name === "start" && value && serviceId) {
        const selectedService = services.find(s => s.id === parseInt(serviceId));
        if (selectedService && selectedService.duration) {
          const [hours, minutes] = value.split(':').map(Number);
          const startTimeInMinutes = hours * 60 + minutes;
          const endTimeInMinutes = startTimeInMinutes + selectedService.duration;
          
          // Преобразуем обратно в формат HH:MM
          const endHours = Math.floor(endTimeInMinutes / 60);
          const endMinutes = endTimeInMinutes % 60;
          
          // Обрабатываем переход через полночь
          const finalEndHours = endHours >= 24 ? endHours - 24 : endHours;
          
          const endTime = `${String(finalEndHours).padStart(2, '0')}:${String(endMinutes).padStart(2, '0')}`;
          newForm.end = endTime;
        }
      }
      
      return newForm;
    });
  };

  const handleAccountChange = (e) => {
    const { name, value } = e.target;
    setAccountForm((prev) => ({ ...prev, [name]: value }));
  };

  const saveAccountMutation = useMutation({
    mutationFn: (payload) => apiService.user.update(payload).then((r) => r.data || r),
    onSuccess: () => {
      const updated = { ...user, first_name: accountForm.first_name, surname: accountForm.surname };
      setCurrentUser(updated);
      setUser(updated);
      showSuccess("Данные профиля обновлены");
    },
    onError: (err) => {
      console.error(err);
      showError("Не удалось сохранить профиль");
    },
  });

  const handleCreate = (e) => {
    e.preventDefault();
    if (!userId) return;
    if (!serviceId) { 
      showError("Пожалуйста, выберите услугу перед созданием слота");
      return; 
    }
    // Сконструируем ISO строки из date + time в локальной зоне устройства (браузер)
    // Build start and end as local Date, adjust end date if crosses midnight
    const startLocal = form.date && form.start ? new Date(`${form.date}T${form.start}:00`) : null;
    let endLocal = form.date && form.end ? new Date(`${form.date}T${form.end}:00`) : null;
    if (startLocal && endLocal && endLocal < startLocal) {
      // if end time is before start time, assume next day
      endLocal.setDate(endLocal.getDate() + 1);
    }
    const start_time = startLocal ? startLocal.toISOString() : "";
    const end_time = endLocal ? endLocal.toISOString() : "";
    // remember absolute last end for next defaults across midnight boundaries
    lastEndRef.current = endLocal ? new Date(endLocal) : null;
    // store last duration for next default
    try {
      const [sh, sm] = (form.start || "10:00").split(":").map((v) => parseInt(v) || 0);
      const [eh, em] = (form.end || "11:00").split(":").map((v) => parseInt(v) || 0);
      const startMin = sh * 60 + sm;
      const endMin = eh * 60 + em;
      let diff = endMin - startMin;
      if (diff <= 0) diff = 60; // fallback to 60 minutes
      lastDurationRef.current = diff;
      lastFormRef.current = { ...form };
    } catch {}
    
    const slotData = {
      master_id: userId,
      MasterID: userId,
      
      start_time,
      end_time,
      is_booked: Boolean(form.is_booked),
    };
    
    // Добавляем service_id только если услуга выбрана
    if (serviceId && serviceId !== "" && serviceId !== "0") {
      slotData.service_id = parseInt(serviceId);
    }
    
    console.log("Создаем слот с данными:", slotData);
    console.log("Выбранная услуга ID:", serviceId);
    
    createMutation.mutate(slotData);
  };

  const handleDeleteAll = () => {
    if (!userId) return;
    deleteMutation.mutate();
  };

  const handleServiceFormChange = (e) => {
    const { name, value } = e.target;
    setServiceForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleCreateService = (e) => {
    e.preventDefault();
    console.log("Creating service with userId:", userId);
    console.log("Service form data:", serviceForm);
    
    if (!userId) {
      console.error("No userId available");
      return;
    }
    
    // Валидация формы
    const name = serviceForm.name.trim();
    const description = serviceForm.description.trim();
    const price = parseFloat(serviceForm.price) || 0;
    const duration = parseInt(serviceForm.duration) || 60;
    
    // Проверка названия
    if (!name) {
      showError("Название услуги обязательно для заполнения");
      return;
    }
    
    if (name.length > 64) {
      showError("Название услуги не должно превышать 64 символа");
      return;
    }
    
    // Проверка цены
    if (price < 0) {
      showError("Цена не может быть отрицательной");
      return;
    }
    
    if (price > 1000000) {
      showError("Цена не может превышать 1 000 000 рублей");
      return;
    }
    
    // Проверка длительности
    if (duration < 15) {
      showError("Длительность должна быть не менее 15 минут");
      return;
    }
    
    if (duration > 360) { // 6 часов = 360 минут
      showError("Длительность не может превышать 6 часов (360 минут)");
      return;
    }
    
    const payload = {
      master_id: userId,
      name,
      description,
      price,
      duration,
    };
    
    console.log("Service creation payload:", payload);
    createServiceMutation.mutate(payload);
  };

  // Форматирование и группировка как на публичной странице
  const slotsByDate = useMemo(() => {
    const map = new Map();
    for (const s of slots) {
      const key = formatDateForDisplay(s.start_time);
      if (!map.has(key)) map.set(key, []);
      map.get(key).push(s);
    }
    return Array.from(map.entries());
  }, [slots]);

  // Фильтр по дате для области слотов (как на MasterPublic)
  const [dateFilter, setDateFilter] = useState("ALL");
  const dateOptions = useMemo(() => slotsByDate.map(([dateTitle]) => dateTitle), [slotsByDate]);
  
  // Фильтрация слотов с учетом прошедшего времени
  const filteredSlotsByDate = useMemo(() => {
    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    
    let filteredSlots = slotsByDate;
    
    // Фильтр по дате
    if (dateFilter !== "ALL") {
      filteredSlots = filteredSlots.filter(([dateTitle]) => dateTitle === dateFilter);
    }
    
    // Фильтр по прошедшим слотам (только в профиле)
    if (!showPastSlots) {
      filteredSlots = filteredSlots.map(([dateTitle, daySlots]) => [
        dateTitle,
        daySlots.filter(slot => {
          const slotStart = slot.start_time ? new Date(slot.start_time) : null;
          return !slotStart || slotStart >= oneHourAgo;
        })
      ]).filter(([dateTitle, daySlots]) => daySlots.length > 0);
    }
    
    return filteredSlots;
  }, [slotsByDate, dateFilter, showPastSlots]);

  return (
    <div className="profile-root">
            {/* Мобильная кнопка меню */}
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

      {/* Мобильный оверлей */}
      {isMobileMenuOpen && (
        <div 
          className="mobile-overlay"
          onClick={toggleMobileMenu}
        ></div>
      )}

       {/* Левый сайдбар навигации */}
       <LeftSidebar 
         isMobileMenuOpen={isMobileMenuOpen}
         toggleMobileMenu={toggleMobileMenu}
       />
      <div className="profile-main-content">
        <div className="profile-container">
          <header className="profile-header">
            <h1 className="profile-title">Профиль</h1>
            <p className="profile-subtitle">управляйте расписанием, заявками и данными</p>
          </header>
          {/* Расписание секция */}
          <div className="profile-section">
            <div className="profile-section-header">
              <h3 className="profile-section-title">Ваше расписание</h3>
            </div>
            <main className="profile-layout">
              <aside className="profile-card">
                <div className="profile-header-card">Данные пользователя</div>
                <div className="profile-user">
                  <div className="profile-row">
                    <b>Имя:</b>
                    <div className="profile-value">{user?.first_name || '—'}</div>
                  </div>
                  <div className="profile-row">
                    <b>Фамилия:</b>
                    <div className="profile-value">{user?.surname || '—'}</div>
                  </div>
                  <div className="profile-row">
                    <b>Ссылка:</b>
                    <div className="url-field-container">
                      <button 
                        className="btn-confirm copy-btn" 
                        type="button" 
                        onClick={async () => {
                          const url = `${window.location.origin}/master/${userId}`;
                          const success = await copyToClipboard(url);
                          
                          if (success) {
                            showSuccess("Ссылка скопирована!");
                          } else {
                            showError("Не удалось скопировать!");
                          }
                        }}
                      >
                        Скопировать
                      </button>
                    </div>
                  </div>
                </div>

                <div className="profile-service">
                  <div className="profile-service-title">Создать слот</div>
                  <div className="profile-service">
                    <div className="profile-service-header">
                      <span className="profile-service-pre-title">Услуга:</span>
                      <div className="service-selector">
                        <select
                          className="profile-select"
                          value={serviceId}
                          onChange={(e) => {
                            console.log("Выбрана услуга:", e.target.value);
                            setServiceId(e.target.value);
                            
                            // Автоматически пересчитываем время окончания при смене услуги
                            if (e.target.value && form.start) {
                              const selectedService = services.find(s => s.id === parseInt(e.target.value));
                              if (selectedService && selectedService.duration) {
                                const [hours, minutes] = form.start.split(':').map(Number);
                                const startTimeInMinutes = hours * 60 + minutes;
                                const endTimeInMinutes = startTimeInMinutes + selectedService.duration;
                                
                                // Преобразуем обратно в формат HH:MM
                                const endHours = Math.floor(endTimeInMinutes / 60);
                                const endMinutes = endTimeInMinutes % 60;
                                
                                // Обрабатываем переход через полночь
                                const finalEndHours = endHours >= 24 ? endHours - 24 : endHours;
                                
                                const endTime = `${String(finalEndHours).padStart(2, '0')}:${String(endMinutes).padStart(2, '0')}`;
                                setForm(prev => ({ ...prev, end: endTime }));
                              }
                            }
                          }}
                        >
                          <option value="">Не выбрано</option>
                          {(services || []).map((svc, index) => (
                            <option key={svc.id || `service-${index}`} value={svc.id}>
                              {svc.name || svc.Name || svc.code || `Услуга ${svc.id || index + 1}`} 
                            </option>
                          ))}
                        </select>
                        <button 
                          type="button"
                          className="add-service-btn"
                          onClick={() => setShowServiceModal(true)}
                          title="Добавить услугу"
                        >
                          +
                        </button>
                      </div>
                    </div>
              
                    {/* Показываем информацию о выбранной услуге */}
                    {serviceId && (() => {
                      const selectedService = services.find(s => s.id === parseInt(serviceId));
                      return selectedService ? (
                        <div className="profile-service-info">
                          <div>
                            <strong>Длительность:</strong> {selectedService.duration} мин
                            {selectedService.price && (
                              <span style={{ marginLeft: '15px' }}>
                                <strong>Цена:</strong> {selectedService.price} ₽
                              </span>
                            )}
                          </div>
                          {selectedService.description && (
                            <div className="service-description-box" title={selectedService.description}>
                              {selectedService.description}
                            </div>
                          )}
                        </div>
                      ) : null;
                    })()}
                  </div>

                  <form className="profile-form" onSubmit={handleCreate}>
                    <div className="profile-row">
                      <div className="profile-field">
                        <label>Начало</label>
                        <input className="profile-input" type="time" name="start" value={form.start} onChange={handleChange} />
                      </div>
                      <div className="profile-field">
                        <label>Конец</label>
                        <input 
                          className="profile-input" 
                          type="time" 
                          name="end" 
                          value={form.end} 
                          onChange={handleChange}
                          style={{ 
                            backgroundColor: serviceId ? '#f8f9fa' : 'white',
                            color: serviceId ? '#666' : 'black'
                          }}
                        />
                      </div>
                    </div>
                    <div className="profile-field">
                      <label>Дата</label>
                      <input className="profile-input" type="date" name="date" value={form.date} onChange={handleChange} />
                    </div>
                    <div className="profile-actions">
                      <button className="btn-primary" type="submit" disabled={createMutation.isPending || !userId}>Создать</button>
                    </div>
                  </form>
                </div>
              </aside>

              <section className="profile-center">
                {/* Верхняя панель фильтра по дате */}
                <div className="profile-filter">
                  <div className="profile-filter-label">Показывать</div>
                  <select className="profile-select" value={dateFilter} onChange={(e) => setDateFilter(e.target.value)}>
                    <option value="ALL">ВСЕ</option>
                    {dateOptions.map((d) => (
                      <option key={d} value={d}>{d}</option>
                    ))}
                  </select>
                  
                  {/* Чекбокс для показа прошедших слотов */}
                  <div className="filter-group">
                    <input 
                      type="checkbox" 
                      id="showPastSlots"
                      checked={showPastSlots}
                      onChange={(e) => setShowPastSlots(e.target.checked)}
                    />
                    <label htmlFor="showPastSlots">
                      Показать прошедшие слоты
                    </label>
                  </div>
                </div>

                <div className="profile-slots">
                  {isLoading ? (
                    <div>Загрузка...</div>
                  ) : filteredSlotsByDate.length > 0 ? (
                    filteredSlotsByDate.map(([dateTitle, daySlots]) => (
                      <div key={dateTitle} className="profile-day">
                        <div className="profile-day-title">{dateTitle}</div>
                        {daySlots.map((slot) => {
                          const isBooked = Boolean(slot.is_booked);
                          // Find service name by service_id
                          const serviceName = slot.service_id ? 
                            (services || []).find(s => s.id === slot.service_id)?.name || 
                            (services || []).find(s => s.id === slot.service_id)?.Name || "Услуга" : 
                            "Услуга";
                          return (
                            <div key={slot.id || `${slot.start_time}-${slot.end_time}`}
                                 className="profile-slot-card" onClick={() => {
                                   handleSlotClick(slot);
                                   setEdit({ 
                                     date: formatDateForForm(slot.start_time), 
                                     start: formatTimeForForm(slot.start_time), 
                                     end: formatTimeForForm(slot.end_time) 
                                   });
                                 }}>
                              <div className="profile-slot-top">
                                <div className="profile-slot-time">{formatTimeInLocal(slot.start_time)} — {formatTimeInLocal(slot.end_time)}</div>
                                <div className="profile-slot-service">{serviceName}</div>
                                <div className={`profile-slot-badge ${isBooked ? "booked" : "free"}`}>
                                  {isBooked ? "Забронирован" : "Свободен"}
                                </div>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    ))
                  ) : (
                    <div className="profile-day">Слотов пока нет.</div>
                  )}
                </div>
              </section>
            </main>
          </div>

      {/* Модальные окна */}
      {confirmDlg.open && (
        <div className="profile-modal-backdrop confirm-modal-backdrop" onClick={closeConfirm}>
          <div className="profile-modal" onClick={(e)=> e.stopPropagation()}>
            <div className="profile-modal-title">{confirmDlg.title}</div>
            <div className="profile-modal-sub">{confirmDlg.text}</div>
            <div className="profile-modal-actions">
              <button className="btn-danger-close" onClick={closeConfirm}>Отмена</button>
              <button className="btn-primary" onClick={() => { 
                const fn = confirmDlg.onConfirm;
                if (typeof fn === 'function') fn();
                closeConfirm();
              }}>Подтвердить</button>
            </div>
          </div>
        </div>
      )}

          {/* Мои заявки секция */}
          <ApplicationsSection
            processedApplications={processedApplications}
            statusFilter={statusFilter}
            showPastApplications={showPastApplications}
            setStatusFilter={setStatusFilter}
            setShowPastApplications={setShowPastApplications}
          />

          {/* Аккаунт секция */}
          <div className="profile-section">
            <div className="profile-section-header">
              <h3 className="profile-section-title">Публичные данные</h3>
            </div>
            <div className="account-layout">
              <div className="account-card">
                <h4 className="account-card-title">Информация</h4>
                <div className="account-card-content">
                  <div className="account-fields">
                    <div className="account-field">
                      <label>Имя</label>
                      <input className="account-input" name="first_name" type="text" value={accountForm.first_name} onChange={handleAccountChange} placeholder="Введите имя" />
                    </div>
                    <div className="account-field">
                      <label>Фамилия</label>
                      <input className="account-input" name="surname" type="text" value={accountForm.surname} onChange={handleAccountChange} placeholder="Введите фамилию" />
                    </div>
                    <div className="profile-actions">
                      <button className="btn-primary" type="button" onClick={() => saveAccountMutation.mutate({ user_id: userId, first_name: accountForm.first_name, surname: accountForm.surname })} disabled={saveAccountMutation.isPending || !userId}>Сохранить</button>
                    </div>
                  </div>
                </div>
              </div>
              
              <div className="account-card">
                <h4 className="account-card-title">Услуги</h4>
                <div className="account-card-content">
                  <div className="account-field">
                    <label>Мои услуги</label>
                    <div className="account-input" style={{ background: 'transparent', border: 'none', padding: 0 }}>
                      {(services || []).length === 0 ? (
                        <div style={{ color:'#666', fontWeight:'500' }}>Пока нет услуг</div>
                      ) : (
                        <div className="services-scrollable">
                          {(services || []).map((svc) => {
                            const edit = serviceEditMap[svc.id] || { name: svc.name || '', description: svc.description || '', price: svc.price || 0, duration: svc.duration || 60 };
                            return (
                              <div key={svc.id} style={{background: 'white', display: 'grid', gap: 6, border: '1px solid #e6e3f1', borderRadius: 8, padding: 10 }}>
                                <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap: 8 }}>
                                  <div className="account-field">
                                    <label className="account-sublabel">Название</label>
                                    <input className="account-input" placeholder="Название" value={edit.name} onChange={(e)=> setServiceEditMap((m)=> ({...m, [svc.id]: { ...edit, name: e.target.value }}))} />
                                  </div>
                                  <div className="account-field">
                                    <label className="account-sublabel">Длительность (мин)</label>
                                    <input className="account-input" placeholder="Длительность (мин)" type="number" value={edit.duration} onChange={(e)=> setServiceEditMap((m)=> ({...m, [svc.id]: { ...edit, duration: parseInt(e.target.value)||0 }}))} />
                                  </div>
                                </div>
                                <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap: 8 }}>
                                  <div className="account-field">
                                    <label className="account-sublabel">Цена (₽)</label>
                                    <input className="account-input" placeholder="Цена (₽)" type="number" step="0.01" value={edit.price} onChange={(e)=> setServiceEditMap((m)=> ({...m, [svc.id]: { ...edit, price: parseFloat(e.target.value)||0 }}))} />
                                  </div>
                                  <div className="account-field">
                                    <label className="account-sublabel">Описание</label>
                                    <input className="account-input" placeholder="Описание" value={edit.description} onChange={(e)=> setServiceEditMap((m)=> ({...m, [svc.id]: { ...edit, description: e.target.value }}))} />
                                  </div>
                                </div>
                                <div className="profile-actions" style={{ justifyContent: 'flex-end' }}>
                                  <button className="btn-primary" style={{ minWidth: '150px' }}type="button" onClick={() => {
                                    openConfirm(
                                      'Сохранить изменения услуги',
                                      'Вы уверены, что хотите сохранить изменения услуги?',
                                      () => apiService.service.update({ id: svc.id, master_id: userId, name: edit.name, description: edit.description, price: Number(edit.price)||0, duration: Number(edit.duration)||0 })
                                        .then(()=>{ 
                                          showSuccess('Услуга обновлена'); 
                                          queryClient.invalidateQueries({ queryKey: ['services', userId] });
                                          queryClient.refetchQueries({ queryKey: ['services', userId] });
                                        })
                                        .catch(()=> showError('Не удалось обновить услугу'))
                                    );
                                  }}>Сохранить</button>
                                  <button className="btn-danger" type="button" onClick={() => {
                                    openConfirm(
                                      'Удалить услугу',
                                      'Удалить услугу без возможности восстановления?',
                                      () => apiService.service.delete(svc.id)
                                        .then(()=>{ 
                                          showSuccess('Услуга удалена'); 
                                          queryClient.invalidateQueries({ queryKey: ['services', userId] });
                                          queryClient.refetchQueries({ queryKey: ['services', userId] });
                                        })
                                        .catch(()=> showError('Не удалось удалить услугу'))
                                    );
                                  }}>Удалить</button>
                                </div>
                              </div>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          {/* Приватные данные секция */}
          <div className="profile-section">
            <div className="profile-section-header">
              <h3 className="profile-section-title">Приватные данные</h3>
            </div>
            <div className="private-card">
              <h4 className="account-card-title">Информация</h4>
              <div className="account-card-content">
                <div className="account-fields">
                    <div className="account-field">
                      <label>Идентификатор пользователя</label>
                      <input className="account-input" type="text" value={user?.id || ""} disabled />
                    </div>
                    <div className="account-field">
                      <label>Телефон</label>
                      <input className="account-input" type="text" value={user?.phone || ""} disabled />
                    </div>
                  <div className="account-field">
                    <label>Telegram ID</label>
                    <input className="account-input" type="text" value={user?.telegram_id || ""} disabled />
                  </div>
                  <div className="account-field">
                    <label>Роли</label>
                    <div className="account-input" style={{ background:'#f3f3f3', borderRadius:'8px' }}>
                      {(user?.roles || []).map((r, i) => (
                        <span key={`${r.role || r}-${i}`} style={{ marginRight: 8 }}>{r.role || r}</span>
                      ))}
                      {(!user?.roles || user.roles.length === 0) && <span>—</span>}
                    </div>
                  </div>
                </div>
              </div>
              <div className="account-field" style={{marginTop: '30px', paddingTop: '20px', borderTop: '1px solid #e6e3f1' }}>
                <div style={{ marginTop: '10px'}}>
                  <button 
                    className="btn-delete" 
                    type="button"
                    onClick={() => {
                      openConfirm(
                        'Удалить аккаунт',
                        'Вы уверены, что хотите удалить свой аккаунт? Это действие нельзя отменить. Все ваши данные, слоты, услуги и записи будут безвозвратно удалены.',
                        () => {
                          apiService.user.requestDeletion()
                            .then(() => {
                              showSuccess('Запрос на удаление аккаунта отправлен в Telegram. Проверьте уведомления для подтверждения. Вы будете автоматически разлогинены через 5 секунд.');
                              // Автоматический логаут через 5 секунд
                              setTimeout(() => {
                                handleLogout();
                              }, 5000);
                            })
                            .catch((error) => {
                              console.error('Ошибка запроса удаления:', error);
                              showError('Не удалось отправить запрос на удаление. Попробуйте еще раз.');
                            });
                        },
                        { 
                          textStyle: {
                            fontSize: '10px',
                          }
                        }
                      );
                    }}
                  >
                    Удалить аккаунт
                  </button>
                  <p style={{ fontSize: '12px', color: '#666', marginTop: '8px', marginBottom: 0 }}>
                    Для подтверждения удаления проверьте уведомления в Telegram
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Модальные окна */}
        {confirmDlg.open && (
          <div className="profile-modal-backdrop confirm-modal-backdrop" onClick={closeConfirm}>
            <div className="profile-modal" onClick={(e)=> e.stopPropagation()}>
              <div className="profile-modal-title">{confirmDlg.title}</div>
              <div className="profile-modal-sub">{confirmDlg.text}</div>
              <div className="profile-modal-actions">
                <button className="btn-delete" onClick={closeConfirm}>Отмена</button>
                <button className="btn-confirm" onClick={() => { 
                  const fn = confirmDlg.onConfirm;
                  if (typeof fn === 'function') fn();
                  closeConfirm();
                }}>Подтвердить</button>
              </div>
            </div>
          </div>
        )}

        {activeSlot && (
          <div className="profile-modal-backdrop" onClick={handleCloseSlotModal}>
            <div className="profile-modal" onClick={(e) => e.stopPropagation()}>
              <div className="profile-modal-title">ПОДРОБНОСТИ</div>
              <div className="profile-modal-sub">{activeSlot.is_booked ? "Слот забронирован" : "Слот свободен"}</div>
              <div className="profile-modal-content">
              <div className="profile-edit">
              <div className="profile-row">
                <div className="profile-field">
                  <label>Начало</label>
                  <input className="profile-input" type="time" value={edit.start} disabled />
                </div>
                <div className="profile-field">
                  <label>Конец</label>
                  <input className="profile-input" type="time" value={edit.end} disabled />
                </div>
              </div>
              <div className="profile-field">
                <label>Дата</label>
                <input className="profile-input" type="date" value={edit.date} disabled />
              </div>
              <div className="profile-actions">
                <button className="btn-delete" onClick={() => openConfirm(
                  'Удалить слот',
                  'Удалить выбранный слот без возможности восстановления?',
                  () => deleteSlotMut.mutate(activeSlot.id)
                )} disabled={deleteSlotMut.isPending}>Удалить слот</button>
              </div>
            </div>
            <div className="records-list-title">Заявки</div>

            <div className="records-list">
              {Array.isArray(slotRecordsDisplay) && slotRecordsDisplay.length > 0 ? (
                slotRecordsDisplay.map((rec) => (
                  <div key={rec.id} className="record-row">
                    <div className="record-client">
                      <div className="record-name">{rec.client?.first_name || rec.client?.name || "Клиент"}</div>
                      <div className="record-phone">{rec.client?.phone || "—"}</div>
                    </div>
                    <div className={`record-status ${recordStatusChanges[rec.id] || rec.status} ${recordStatusChanges[rec.id] ? 'modified' : ''}`}>
                      {recordStatusChanges[rec.id] || rec.status}
                      {recordStatusChanges[rec.id] && <span className="status-indicator">*</span>}
                    </div>
                  <div className="record-actions">
                      <RecordStatusSelect
                        currentStatus={recordStatusChanges[rec.id] || rec.status}
                        onChange={(newStatus) => {
                          const currentChanges = { ...recordStatusChanges, [rec.id]: newStatus };
                          
                          setRecordStatusChanges(currentChanges);
                          setHasUnsavedChanges(true);
                        }}
                      />
                    </div>
                  </div>
                ))
              ) : (
                <div className="profile-modal-sub">Заявок нет</div>
              )}
            </div>
            </div>
            <div className="profile-modal-actions">
              <button 
                className="btn-confirm" 
                onClick={saveRecordStatusChanges}
                disabled={!hasUnsavedChanges}
              >
                Сохранить изменения {hasUnsavedChanges && `(${Object.keys(recordStatusChanges).length})`}
              </button>
              <button className="btn-delete" onClick={handleCloseSlotModal}>Закрыть</button>
            </div>
          </div>
        </div>
      )}

      {/* Service Creation Modal */}
      {showServiceModal && (
        <div className="profile-modal-backdrop" onClick={() => setShowServiceModal(false)}>
          <div className="profile-modal" onClick={(e) => e.stopPropagation()}>
            <div className="profile-modal-title">Создание услуги</div>
            
            <form onSubmit={(e) => {
              console.log("Form submitted");
              handleCreateService(e);
            }}>
              <div className="profile-field">
                <label>Название услуги</label>
                <input 
                  className="profile-input" 
                  type="text" 
                  name="name" 
                  value={serviceForm.name} 
                  onChange={(e) => {
                    // Удаляем HTML теги и ограничиваем длину
                    let cleanValue = e.target.value.replace(/[<>]/g, '');
                    if (cleanValue.length > 64) {
                      cleanValue = cleanValue.substring(0, 64);
                    }
                    handleServiceFormChange({
                      target: {
                        name: e.target.name,
                        value: cleanValue
                      }
                    });
                  }}
                  placeholder="Например: Маникюр"
                  maxLength={64}
                  required
                />
              </div>
              
              <div className="profile-field">
                <label>Описание</label>
                <textarea 
                  className="profile-input" 
                  name="description" 
                  value={serviceForm.description} 
                  onChange={handleServiceFormChange}
                  placeholder="Описание услуги"
                  rows="3"
                />
              </div>
              
              <div className="profile-row">
                <div className="profile-field">
                  <label>Цена (₽)</label>
                  <input 
                    className="profile-input" 
                    type="number" 
                    name="price" 
                    value={serviceForm.price} 
                    onChange={handleServiceFormChange}
                    placeholder="1000"
                    min="0"
                    max="1000000"
                    step="0.01"
                  />
                </div>
                <div className="profile-field">
                  <label>Длительность (мин)</label>
                  <input 
                    className="profile-input" 
                    type="number" 
                    name="duration" 
                    value={serviceForm.duration} 
                    onChange={handleServiceFormChange}
                    placeholder="60"
                    min="15"
                    max="360"
                    step="15"
                  />
                </div>
              </div>
              
              <div className="profile-actions">
                <button 
                  className="btn-confirm" 
                  type="submit" 
                  disabled={createServiceMutation.isPending}
                  onClick={() => console.log("Button clicked")}
                >
                  {createServiceMutation.isPending ? "Создание..." : "Создать услугу"}
                </button>
                <button 
                  type="button" 
                  className="btn-delete" 
                  onClick={() => setShowServiceModal(false)}
                >
                  Отмена
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
