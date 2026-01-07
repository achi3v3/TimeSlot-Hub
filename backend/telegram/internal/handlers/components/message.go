package components

import (
	"fmt"
	"strings"
	"telegram-bot/internal/config"
)

func Header() string {
	return "<code>🫟 melot</code>\n\n"
}
func Info() string {
	information :=
		fmt.Sprintf(
			"%s"+
				"ℹ️ Информация\n"+
				"<i>Данное приложение разработано для упрощения взаимодействия пользователей в области предоставления услуг</i>"+
				""+
				"", Header())
	return information
}
func HelpAccount() string {
	cfg := config.Load()
	if strings.TrimSpace(cfg.SupportContact) == "" {
		return "@support"
	}
	return cfg.SupportContact
}
