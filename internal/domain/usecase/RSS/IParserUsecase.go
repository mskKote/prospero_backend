package RSS

type IParserUsecase interface {
	// Startup запускает постоянный RSS парсер
	Startup()

	// ParseJob функция, которая запускается раз в N времени
	ParseJob()
}
