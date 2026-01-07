package formatter

import (
	"fmt"
	"sort"
	"strings"
	"telegram-bot/internal/utils"
	mymodels "telegram-bot/pkg/models"
	"time"

	"github.com/go-telegram/bot/models"
)

// SlotPaginationData содержит данные для пагинации слотов
type SlotPaginationData struct {
	Slots        []mymodels.SlotResponse
	CurrentDate  string
	CurrentPage  int
	TotalPages   int
	HasNextDate  bool
	HasPrevDate  bool
	NextDate     string
	PrevDate     string
	TimeType     time.Time
	MasterInfo   MasterInfo
	IsClientView bool
}

// MasterInfo содержит информацию о мастере
type MasterInfo struct {
	TelegramID int64
	Name       string
	Surname    string
	Phone      string
}

// DateGroup представляет группу слотов по дате
type DateGroup struct {
	Date  string
	Slots []mymodels.SlotResponse
}

// GroupSlotsByDate группирует слоты по датам и сортирует их
func GroupSlotsByDate(slots []mymodels.SlotResponse) []DateGroup {
	if len(slots) == 0 {
		return []DateGroup{}
	}

	// Сортируем слоты по времени
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].StartTime.Before(slots[j].StartTime)
	})

	// Группируем по датам
	dateMap := make(map[string][]mymodels.SlotResponse)
	for _, slot := range slots {
		date := utils.FormatDateInMoscow(slot.StartTime)
		dateMap[date] = append(dateMap[date], slot)
	}

	// Преобразуем в слайс и сортируем по дате
	var groups []DateGroup
	for date, slots := range dateMap {
		groups = append(groups, DateGroup{
			Date:  date,
			Slots: slots,
		})
	}

	// Сортируем группы по дате
	sort.Slice(groups, func(i, j int) bool {
		dateI, _ := time.Parse("02-01-2006", groups[i].Date)
		dateJ, _ := time.Parse("02-01-2006", groups[j].Date)
		return dateI.Before(dateJ)
	})

	return groups
}

// CreatePaginatedSlots создает пагинированные слоты для конкретной даты
func CreatePaginatedSlots(allSlots []mymodels.SlotResponse, targetDate string, page int, slotsPerPage int, restrictNavigation bool) *SlotPaginationData {
	if len(allSlots) == 0 {
		return nil
	}

	groups := GroupSlotsByDate(allSlots)

	// Находим нужную дату
	var targetGroup *DateGroup
	var currentIndex int
	for i, group := range groups {
		if group.Date == targetDate {
			targetGroup = &group
			currentIndex = i
			break
		}
	}

	if targetGroup == nil {
		// Если дата не найдена, берем первую
		if len(groups) > 0 {
			targetGroup = &groups[0]
			currentIndex = 0
		} else {
			return nil
		}
	}

	// Вычисляем пагинацию для слотов
	totalSlots := len(targetGroup.Slots)
	totalPages := (totalSlots + slotsPerPage - 1) / slotsPerPage

	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * slotsPerPage
	end := start + slotsPerPage
	if end > totalSlots {
		end = totalSlots
	}

	paginatedSlots := targetGroup.Slots[start:end]

	// Определяем навигацию по датам
	hasNextDate := currentIndex < len(groups)-1
	hasPrevDate := currentIndex > 0

	var nextDate, prevDate string

	if restrictNavigation {
		todayStr := utils.FormatDateInMoscow(time.Now())
		todayParsed, _ := time.Parse("02-01-2006", todayStr)

		if currentIndex > 0 {
			prevGroup := groups[currentIndex-1]
			if pd, err := time.Parse("02-01-2006", prevGroup.Date); err == nil {
				if !pd.Before(todayParsed) {
					hasPrevDate = true
					prevDate = prevGroup.Date
				} else {
					hasPrevDate = false
					prevDate = ""
				}
			}
		} else {
			hasPrevDate = false
			prevDate = ""
		}

		if hasNextDate {
			nextGroup := groups[currentIndex+1]
			nextDateParsed, err := time.Parse("02-01-2006", nextGroup.Date)
			if err != nil || nextDateParsed.Before(todayParsed) {
				hasNextDate = false
				nextDate = ""
			} else {
				nextDate = nextGroup.Date
			}
		}
	} else {
		if hasNextDate {
			nextDate = groups[currentIndex+1].Date
		}
		if hasPrevDate {
			prevDate = groups[currentIndex-1].Date
		}
	}

	return &SlotPaginationData{
		Slots:       paginatedSlots,
		CurrentDate: targetGroup.Date,
		CurrentPage: page,
		TotalPages:  totalPages,
		HasNextDate: hasNextDate,
		HasPrevDate: hasPrevDate,
		NextDate:    nextDate,
		PrevDate:    prevDate,
		MasterInfo: MasterInfo{
			TelegramID: allSlots[0].MasterTelegramID,
			Name:       allSlots[0].MasterName,
			Surname:    allSlots[0].MasterSurname,
		},
		IsClientView: restrictNavigation,
	}
}

func ParseSlots(slots []mymodels.SlotResponse) string {
	if len(slots) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>ID Пользователя: </b><code>%d</code>\n\n", slots[0].MasterTelegramID))
	b.WriteString(fmt.Sprintf("Имя: <b>%s %s</b>\n", slots[0].MasterName, slots[0].MasterSurname))
	b.WriteString("🟩 [ Свободен ]\n🟥 [ Забронирован ]\n\n")

	currentDate := utils.FormatDateInLocation(slots[0].MasterTimezone, slots[0].StartTime)
	tzLabel := slots[0].MasterTimezone
	if tzLabel == "" {
		tzLabel = "Europe/Moscow"
	}
	b.WriteString(fmt.Sprintf("Дата: <code>%s</code>  TZ: <code>%s</code>\n", currentDate, tzLabel))

	for _, s := range slots {
		date := utils.FormatDateInLocation(s.MasterTimezone, s.StartTime)
		if date != currentDate {
			b.WriteString(fmt.Sprintf("Дата: <code>%s</code>\n", date))
			currentDate = date
		}
		startTime := utils.FormatTimeOnlyInLocation(s.MasterTimezone, s.StartTime)
		endTime := utils.FormatTimeOnlyInLocation(s.MasterTimezone, s.EndTime)
		color := "🟩"
		if s.IsBooked {
			color = "🟥"
		}
		b.WriteString(fmt.Sprintf("<blockquote><code>🕒 %s – %s</code>. [ %s ] %s</blockquote>\n", startTime, endTime, s.ServiceName, color))
	}
	return b.String()
}
func CreateInlineKeyboardSlots(slots []mymodels.SlotResponse) (*models.InlineKeyboardMarkup, string) {
	if len(slots) == 0 {
		return nil, ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>ID Пользователя: </b><code>%d</code>\n\n", slots[0].MasterTelegramID))
	b.WriteString(fmt.Sprintf("Имя: <b>%s %s</b>\n", slots[0].MasterName, slots[0].MasterSurname))
	b.WriteString("🟩 [ Свободен ]\n🟥 [ Забронирован ]\n\n")

	currentDate := utils.FormatDateInLocation(slots[0].MasterTimezone, slots[0].StartTime)
	tzLabel := slots[0].MasterTimezone
	if tzLabel == "" {
		tzLabel = "Europe/Moscow"
	}
	b.WriteString(fmt.Sprintf("Дата: <code>%s</code>  TZ: <code>%s</code>\n", currentDate, tzLabel))

	result := [][]models.InlineKeyboardButton{}
	for _, s := range slots {
		date := utils.FormatDateInLocation(s.MasterTimezone, s.StartTime)
		if date != currentDate {
			b.WriteString(fmt.Sprintf("Дата: <code>%s</code>\n", date))
			currentDate = date
		}
		startTime := utils.FormatTimeOnlyInLocation(s.MasterTimezone, s.StartTime)
		endTime := utils.FormatTimeOnlyInLocation(s.MasterTimezone, s.EndTime)
		color := "🟩"
		if s.IsBooked {
			color = "🟥"
		}

		result = append(result, []models.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("🕒 %s – %s. [ %s ] %s", startTime, endTime, s.ServiceName, color),
				CallbackData: fmt.Sprintf("slot/%d", s.ID),
			},
		})
	}
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: result,
	}, b.String()
}

// CreatePaginatedInlineKeyboard создает пагинированную клавиатуру с навигацией
func CreatePaginatedInlineKeyboard(paginationData *SlotPaginationData) (*models.InlineKeyboardMarkup, string) {
	if paginationData == nil || len(paginationData.Slots) == 0 {
		return nil, ""
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("<b>ID Пользователя: </b><code>%d</code>\n\n", paginationData.MasterInfo.TelegramID))
	b.WriteString(fmt.Sprintf("Имя: <b>%s %s</b>\n", paginationData.MasterInfo.Name, paginationData.MasterInfo.Surname))
	b.WriteString("🟩 [ Свободен ]\n🟥 [ Забронирован ]\n\n")

	// Примечание: CurrentDate уже сформирована в нужной TZ на этапе группировки
	tzLabel := ""
	if len(paginationData.Slots) > 0 {
		tzLabel = paginationData.Slots[0].MasterTimezone
	}
	if tzLabel == "" {
		tzLabel = "Europe/Moscow"
	}
	b.WriteString(fmt.Sprintf("Дата: <code>%s</code>  TZ: <code>%s</code>\n", paginationData.CurrentDate, tzLabel))

	if paginationData.TotalPages > 1 {
		b.WriteString(fmt.Sprintf("Страница: <code>%d/%d</code>\n\n", paginationData.CurrentPage, paginationData.TotalPages))
	} else {
		b.WriteString("\n")
	}

	var keyboard [][]models.InlineKeyboardButton

	for _, slot := range paginationData.Slots {
		startTime := utils.FormatTimeOnlyInLocation(slot.MasterTimezone, slot.StartTime)
		endTime := utils.FormatTimeOnlyInLocation(slot.MasterTimezone, slot.EndTime)
		color := "🟩"
		if slot.IsBooked {
			color = "🟥"
		}

		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("🕒 %s – %s. [ %s ] %s", startTime, endTime, slot.ServiceName, color),
				CallbackData: fmt.Sprintf("slot/%d", slot.ID),
			},
		})
	}

	// Навигационные кнопки
	var navButtons []models.InlineKeyboardButton

	if paginationData.HasPrevDate {
		navButtons = append(navButtons, models.InlineKeyboardButton{
			Text:         "⬅️ " + paginationData.PrevDate,
			CallbackData: fmt.Sprintf("%s/%d/%s/1", tern(paginationData.IsClientView, "client_date", "date"), paginationData.MasterInfo.TelegramID, paginationData.PrevDate),
		})
	}

	if paginationData.TotalPages > 1 {
		if paginationData.CurrentPage > 1 {
			navButtons = append(navButtons, models.InlineKeyboardButton{
				Text:         "◀️",
				CallbackData: fmt.Sprintf("%s/%d/%s/%d", tern(paginationData.IsClientView, "client_date", "date"), paginationData.MasterInfo.TelegramID, paginationData.CurrentDate, paginationData.CurrentPage-1),
			})
		}

		navButtons = append(navButtons, models.InlineKeyboardButton{
			Text:         fmt.Sprintf("%d/%d", paginationData.CurrentPage, paginationData.TotalPages),
			CallbackData: "noop", // Неактивная кнопка для отображения текущей страницы
		})

		if paginationData.CurrentPage < paginationData.TotalPages {
			navButtons = append(navButtons, models.InlineKeyboardButton{
				Text:         "▶️",
				CallbackData: fmt.Sprintf("%s/%d/%s/%d", tern(paginationData.IsClientView, "client_date", "date"), paginationData.MasterInfo.TelegramID, paginationData.CurrentDate, paginationData.CurrentPage+1),
			})
		}
	}

	if paginationData.HasNextDate {
		navButtons = append(navButtons, models.InlineKeyboardButton{
			Text:         paginationData.NextDate + " ➡️",
			CallbackData: fmt.Sprintf("%s/%d/%s/1", tern(paginationData.IsClientView, "client_date", "date"), paginationData.MasterInfo.TelegramID, paginationData.NextDate),
		})
	}

	// Добавляем навигационные кнопки в клавиатуру
	if len(navButtons) > 0 {
		keyboard = append(keyboard, navButtons)
	}

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}, b.String()
}

func tern(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
