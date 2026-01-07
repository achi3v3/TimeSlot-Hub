export default function ApplicationsSection({
  processedApplications,
  statusFilter,
  showPastApplications,
  setStatusFilter,
  setShowPastApplications,
}) {
  return (
    <div className="profile-section">
      <div className="profile-section-header">
        <h3 className="profile-section-title">Заявки</h3>
      </div>
      <div className="applications-layout">
        <div className="applications-card">
          <h4 className="applications-card-title">К кому вы записывались?</h4>

          {/* Фильтры */}
          <div className="applications-filters">
            <div className="filter-group">
              <label>Статус:</label>
              <select
                className="profile-select"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
              >
                <option value="ALL">Все</option>
                <option value="PENDING">В ожидании</option>
                <option value="CONFIRMED">Подтвержденная</option>
                <option value="REJECTED">Отклоненная</option>
              </select>
            </div>

            <div className="filter-group">
              <input
                type="checkbox"
                id="showPastApplications"
                checked={showPastApplications}
                onChange={(e) => setShowPastApplications(e.target.checked)}
              />
              <label htmlFor="showPastApplications">
                Показать прошедшие заявки
              </label>
            </div>
          </div>

          <div className="applications-content">
            {processedApplications && processedApplications.length > 0 ? (
              <div className="applications-list">
                {processedApplications.map((record) => {
                  const service = record.slot?.service;
                  const start = record.start;
                  const end = record.end;

                  const masterInfo = {
                    id: record.slot?.master?.id || record?.master_id,
                    name:
                      record.slot?.master?.first_name ||
                      record.slot?.master?.name ||
                      "Неизвестный мастер",
                    surname:
                      record.slot?.master?.surname ||
                      record.slot?.master?.surname ||
                      "",
                    telegram_id: record.slot?.master?.telegram_id,
                  };

                  const getStatusInfo = (status) => {
                    switch (status) {
                      case "pending":
                        return {
                          text: "Ожидает рассмотрения",
                          class: "status-pending",
                          icon: "⏳",
                        };
                      case "confirm":
                        return {
                          text: "Одобрено",
                          class: "status-confirmed",
                          icon: "✅",
                        };
                      case "reject":
                        return {
                          text: "Отклонено",
                          class: "status-rejected",
                          icon: "❌",
                        };
                      default:
                        return {
                          text: "Неизвестно",
                          class: "status-unknown",
                          icon: "❓",
                        };
                    }
                  };

                  const statusInfo = getStatusInfo(record.status);

                  return (
                    <div
                      key={record.id}
                      className={`application-item ${
                        record.isPast ? "application-past" : ""
                      }`}
                    >
                      <div className="application-main">
                        <div className="application-header">
                          <div className="application-service">
                            <span className="service-name">
                              {service?.name || "Услуга"}
                            </span>
                            {service?.price && (
                              <span className="service-price">
                                {service.price} ₽
                              </span>
                            )}
                          </div>
                          <div
                            className={`application-status ${statusInfo.class}`}
                          >
                            <span className="status-icon">
                              {statusInfo.icon}
                            </span>
                            <span className="status-text">
                              {statusInfo.text}
                            </span>
                          </div>
                        </div>

                        <div className="application-details">
                          <div className="detail-row">
                            <span className="detail-label">Мастер:</span>
                            <span className="detail-value">
                              {masterInfo.name} {masterInfo.surname}
                            </span>
                          </div>

                          <div className="detail-row">
                            <span className="detail-label">Дата и время:</span>
                            <span className="detail-value">
                              {start
                                ? start.toLocaleDateString("ru-RU", {
                                    day: "2-digit",
                                    month: "long",
                                    year: "numeric",
                                  })
                                : "—"}{" "}
                              в{" "}
                              {start
                                ? start.toLocaleTimeString("ru-RU", {
                                    hour: "2-digit",
                                    minute: "2-digit",
                                  })
                                : "—"}
                              {end &&
                                ` - ${end.toLocaleTimeString("ru-RU", {
                                  hour: "2-digit",
                                  minute: "2-digit",
                                })}`}
                            </span>
                          </div>

                          {service?.duration && (
                            <div className="detail-row">
                              <span className="detail-label">Длительность:</span>
                              <span className="detail-value">
                                {service.duration} мин
                              </span>
                            </div>
                          )}

                          {service?.description && (
                            <div className="detail-row">
                              <span className="detail-label">Описание:</span>
                              <span
                                className="detail-value service-description"
                                title={service.description}
                              >
                                {service.description.length > 100
                                  ? service.description.slice(0, 100) + "…"
                                  : service.description}
                              </span>
                            </div>
                          )}

                          <div className="detail-row">
                            <span className="detail-label">
                              Заявка подана:
                            </span>
                            <span className="detail-value">
                              {new Date(
                                record.created_at || Date.now()
                              ).toLocaleDateString("ru-RU", {
                                day: "2-digit",
                                month: "long",
                                year: "numeric",
                                hour: "2-digit",
                                minute: "2-digit",
                              })}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="applications-empty">
                <p>У вас пока нет заявок на записи</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}


