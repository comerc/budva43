package media

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"

	"github.com/comerc/budva43/entity"
	"github.com/zelenin/go-tdlib/client"
)

//go:generate mockery --name=mediaProcessor --exported
type mediaProcessor interface {
	ProcessMedia(mediaContent []byte, mediaType string) ([]byte, error)
}

//go:generate mockery --name=mediaStorage --exported
type mediaStorage interface {
	SaveMedia(key string, content []byte) error
	GetMedia(key string) ([]byte, bool, error)
}

// MediaService предоставляет методы для работы с медиа-файлами
type MediaService struct {
	processor mediaProcessor
	storage   mediaStorage
}

// NewMediaService создает новый экземпляр сервиса для работы с медиа-файлами
func NewMediaService(processor mediaProcessor, storage mediaStorage) *MediaService {
	return &MediaService{
		processor: processor,
		storage:   storage,
	}
}

// GetMediaContent извлекает содержимое медиа из сообщения
func (s *MediaService) GetMediaContent(message *entity.Message) ([]byte, string, error) {
	if message == nil || message.Content == nil {
		return nil, "", errors.New("empty message")
	}

	var mediaType string

	switch content := message.Content.(type) {
	case *client.MessagePhoto:
		if len(content.Photo.Sizes) > 0 {
			// Берем самую большую версию изображения
			largestSize := content.Photo.Sizes[0]
			for _, size := range content.Photo.Sizes {
				if size.Photo.Size > largestSize.Photo.Size {
					largestSize = size
				}
			}
			mediaType = "photo"
		}
	case *client.MessageVideo:
		mediaType = "video"
	case *client.MessageDocument:
		mediaType = filepath.Ext(content.Document.FileName)
		if mediaType == "" {
			mediaType = "document"
		} else {
			mediaType = mediaType[1:] // Убираем точку в начале
		}
	case *client.MessageAudio:
		mediaType = "audio"
	case *client.MessageAnimation:
		mediaType = "animation"
	case *client.MessageVoiceNote:
		mediaType = "voice"
	default:
		return nil, "", errors.New("unsupported media type")
	}

	// Здесь должна быть реализация загрузки файла по его ID
	// Например, через TDLib API

	// Заглушка
	content := []byte("Media content placeholder")
	return content, mediaType, nil
}

// ProcessAndCacheMedia обрабатывает и кэширует медиа-файл
func (s *MediaService) ProcessAndCacheMedia(content []byte, mediaType, cacheKey string) ([]byte, error) {
	// Сначала пытаемся получить из кэша
	if s.storage != nil {
		cachedContent, exists, err := s.storage.GetMedia(cacheKey)
		if err == nil && exists {
			return cachedContent, nil
		}
	}

	// Если в кэше нет, обрабатываем
	var processedContent []byte
	var err error

	if s.processor != nil {
		processedContent, err = s.processor.ProcessMedia(content, mediaType)
		if err != nil {
			return nil, err
		}
	} else {
		processedContent = content
	}

	// Сохраняем в кэш
	if s.storage != nil {
		err = s.storage.SaveMedia(cacheKey, processedContent)
		if err != nil {
			// Ошибка кэширования не критична
			// можно продолжить работу
		}
	}

	return processedContent, nil
}

// OptimizeImage оптимизирует изображение
func (s *MediaService) OptimizeImage(content []byte, format string, quality int) ([]byte, error) {
	// Здесь должна быть реализация оптимизации изображения
	// Например, с использованием библиотек для обработки изображений

	// Заглушка
	return content, nil
}

// GetMediaType определяет тип медиа-файла по его содержимому
func (s *MediaService) GetMediaType(content []byte) string {
	if len(content) < 4 {
		return "unknown"
	}

	// Простая эвристика по сигнатурам файлов
	header := content[:4]

	// JPEG: FF D8 FF
	if bytes.Equal(header[:3], []byte{0xFF, 0xD8, 0xFF}) {
		return "jpeg"
	}

	// PNG: 89 50 4E 47
	if bytes.Equal(header, []byte{0x89, 0x50, 0x4E, 0x47}) {
		return "png"
	}

	// GIF: 47 49 46 38
	if bytes.Equal(header, []byte{0x47, 0x49, 0x46, 0x38}) {
		return "gif"
	}

	// MP4/M4A/etc: 66 74 79 70
	if bytes.Equal(content[4:8], []byte{0x66, 0x74, 0x79, 0x70}) {
		return "mp4"
	}

	// Для других форматов могут быть добавлены другие сигнатуры

	return "unknown"
}

// CopyMedia копирует медиа-файл из одного сообщения в другое
func (s *MediaService) CopyMedia(sourceMessage, targetMessage *entity.Message) error {
	// Заглушка для функции копирования медиа
	// В реальной реализации это может быть сложная логика
	// по копированию медиа-файлов между сообщениями
	return nil
}

// GetMediaFileSize возвращает размер медиа-файла
func (s *MediaService) GetMediaFileSize(message *entity.Message) (int64, error) {
	if message == nil || message.Content == nil {
		return 0, errors.New("empty message")
	}

	var fileSize int64

	switch content := message.Content.(type) {
	case *client.MessagePhoto:
		if len(content.Photo.Sizes) > 0 {
			largestSize := content.Photo.Sizes[0]
			for _, size := range content.Photo.Sizes {
				if size.Photo.Size > largestSize.Photo.Size {
					largestSize = size
				}
			}
			fileSize = int64(largestSize.Photo.Size)
		}
	case *client.MessageVideo:
		if content.Video != nil && content.Video.Video != nil {
			fileSize = int64(content.Video.Video.Size)
		}
	case *client.MessageDocument:
		fileSize = int64(content.Document.Document.Size)
	case *client.MessageAudio:
		fileSize = int64(content.Audio.Audio.Size)
	case *client.MessageAnimation:
		fileSize = int64(content.Animation.Animation.Size)
	case *client.MessageVoiceNote:
		fileSize = int64(content.VoiceNote.Voice.Size)
	default:
		return 0, errors.New("unsupported media type")
	}

	return fileSize, nil
}

// StreamMedia потоковая передача медиа-файла
func (s *MediaService) StreamMedia(message *entity.Message, writer io.Writer) error {
	content, _, err := s.GetMediaContent(message)
	if err != nil {
		return err
	}

	_, err = writer.Write(content)
	return err
}
