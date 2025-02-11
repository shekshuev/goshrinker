package models

// ShortURLCreateDTO представляет структуру запроса на создание сокращённого URL.
type ShortURLCreateDTO struct {
	URL string `json:"url"` // Исходный URL, который нужно сократить.
}

// ShortURLReadDTO содержит результат успешного создания сокращённого URL.
type ShortURLReadDTO struct {
	Result string `json:"result"` // Сокращённый URL.
}

// SerializeData представляет структуру данных для сериализации URL пользователя.
type SerializeData struct {
	UserID      string `json:"user_id"`      // Уникальный идентификатор пользователя.
	ShortURL    string `json:"short_url"`    // Сокращённый URL.
	OriginalURL string `json:"original_url"` // Исходный URL.
}

// BatchShortURLCreateDTO представляет структуру для пакетного создания сокращённых URL.
type BatchShortURLCreateDTO struct {
	CorrelationID string `json:"correlation_id"` // Уникальный идентификатор запроса (используется клиентом для сопоставления).
	OriginalURL   string `json:"original_url"`   // Исходный URL.
	ShortURL      string // Сокращённый URL (не сериализуется в JSON).
}

// BatchShortURLReadDTO содержит результат пакетного создания сокращённых URL.
type BatchShortURLReadDTO struct {
	CorrelationID string `json:"correlation_id"` // Уникальный идентификатор запроса.
	ShortURL      string `json:"short_url"`      // Сокращённый URL.
}

// UserShortURLReadDTO содержит данные о сокращённом URL, привязанном к пользователю.
type UserShortURLReadDTO struct {
	ShortURL    string `json:"short_url"`    // Сокращённый URL.
	OriginalURL string `json:"original_url"` // Исходный URL.
}
