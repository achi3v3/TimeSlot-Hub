import { useState } from "react";

// Компонент выпадающего списка для статуса записи (с поддержкой мобильных)
export default function RecordStatusSelect({ currentStatus, onChange }) {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);

  const handleSelectStatus = (newStatus) => {
    onChange(newStatus);
    setIsDropdownOpen(false);
  };

  const statusOptions = [
    { value: "confirm", label: "Подтвердить" },
    { value: "reject", label: "Отклонить" },
  ];

  return (
    <div className="record-status-select-wrapper">
      {/* Нативный select для десктопа */}
      <select
        className="profile-select desktop-only-select"
        value={currentStatus}
        onChange={(e) => onChange(e.target.value)}
      >
        {statusOptions.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>

      {/* Кнопка-иконка для мобильных */}
      <button
        type="button"
        className="mobile-status-toggle"
        onClick={(e) => {
          e.stopPropagation();
          setIsDropdownOpen(!isDropdownOpen);
        }}
        aria-label="Изменить статус"
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
          <path d="M4 6l4 4 4-4H4z" />
        </svg>
      </button>

      {/* Кастомное выпадающее меню для мобильных */}
      {isDropdownOpen && (
        <>
          <div
            className="mobile-dropdown-backdrop"
            onClick={() => setIsDropdownOpen(false)}
          />
          <div className="mobile-status-dropdown">
            {statusOptions.map((opt) => (
              <button
                key={opt.value}
                type="button"
                className={`mobile-status-option ${
                  opt.value === currentStatus ? "active" : ""
                }`}
                onClick={(e) => {
                  e.stopPropagation();
                  handleSelectStatus(opt.value);
                }}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}


