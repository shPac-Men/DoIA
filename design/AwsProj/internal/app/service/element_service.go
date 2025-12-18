package service

import (
	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/repository"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type ElementService struct {
	repo        *repository.Repository
	minioClient *minio.Client
	bucketName  string
}

func NewElementService(repo *repository.Repository, minioEndpoint, minioAccessKey, minioSecretKey, bucketName string) (*ElementService, error) {
	// #region agent log
	logFile, _ := os.OpenFile("/home/artem/Desktop/RIP2/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()
	logEntry := map[string]interface{}{
		"sessionId": "debug-session",
		"runId": "post-fix",
		"hypothesisId": "B",
		"location": "element_service.go:28",
		"message": "NewElementService entry - attempting optional MinIO init",
		"data": map[string]interface{}{"minioEndpoint": minioEndpoint, "bucketName": bucketName},
		"timestamp": time.Now().UnixMilli(),
	}
	json.NewEncoder(logFile).Encode(logEntry)
	// #endregion
	
	var minioClient *minio.Client
	
	// Пытаемся инициализировать MinIO клиента, но не блокируем запуск приложения при ошибке
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false, // используйте true для HTTPS
	})
	
	// #region agent log
	logEntry2 := map[string]interface{}{
		"sessionId": "debug-session",
		"runId": "post-fix",
		"hypothesisId": "B",
		"location": "element_service.go:37",
		"message": "MinIO client creation result",
		"data": map[string]interface{}{"clientCreated": err == nil, "error": fmt.Sprintf("%v", err)},
		"timestamp": time.Now().UnixMilli(),
	}
	json.NewEncoder(logFile).Encode(logEntry2)
	// #endregion
	
	if err == nil {
		// Проверяем существование бакета, если нет - создаем (но не критично, если не получится)
		ctx := context.Background()
		exists, err := minioClient.BucketExists(ctx, bucketName)
		
		// #region agent log
		logEntry3 := map[string]interface{}{
			"sessionId": "debug-session",
			"runId": "post-fix",
			"hypothesisId": "B",
			"location": "element_service.go:45",
			"message": "BucketExists check result",
			"data": map[string]interface{}{"exists": exists, "error": fmt.Sprintf("%v", err)},
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(logFile).Encode(logEntry3)
		// #endregion
		
		if err == nil && !exists {
			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				// #region agent log
				logEntry4 := map[string]interface{}{
					"sessionId": "debug-session",
					"runId": "post-fix",
					"hypothesisId": "B",
					"location": "element_service.go:52",
					"message": "MakeBucket failed, continuing without MinIO",
					"data": map[string]interface{}{"error": fmt.Sprintf("%v", err)},
					"timestamp": time.Now().UnixMilli(),
				}
				json.NewEncoder(logFile).Encode(logEntry4)
				// #endregion
				
				logrus.Warnf("⚠️ MinIO недоступен (ошибка создания бакета: %v), приложение продолжит работу без загрузки изображений", err)
				minioClient = nil // Отключаем MinIO если не удалось создать бакет
			}
		} else if err != nil {
			// #region agent log
			logEntry5 := map[string]interface{}{
				"sessionId": "debug-session",
				"runId": "post-fix",
				"hypothesisId": "B",
				"location": "element_service.go:61",
				"message": "BucketExists failed, continuing without MinIO",
				"data": map[string]interface{}{"error": fmt.Sprintf("%v", err)},
				"timestamp": time.Now().UnixMilli(),
			}
			json.NewEncoder(logFile).Encode(logEntry5)
			// #endregion
			
			logrus.Warnf("⚠️ MinIO недоступен (ошибка проверки бакета: %v), приложение продолжит работу без загрузки изображений", err)
			minioClient = nil // Отключаем MinIO если не удалось проверить бакет
		} else {
			logrus.Info("✅ MinIO инициализирован успешно")
		}
	} else {
		// #region agent log
		logEntry6 := map[string]interface{}{
			"sessionId": "debug-session",
			"runId": "post-fix",
			"hypothesisId": "B",
			"location": "element_service.go:71",
			"message": "MinIO client creation failed, continuing without MinIO",
			"data": map[string]interface{}{"error": fmt.Sprintf("%v", err)},
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(logFile).Encode(logEntry6)
		// #endregion
		
		logrus.Warnf("⚠️ MinIO недоступен (ошибка инициализации: %v), приложение продолжит работу без загрузки изображений", err)
		minioClient = nil // Продолжаем без MinIO
	}

	return &ElementService{
		repo:        repo,
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}

// uploadToMinIO загружает файл в MinIO
func (s *ElementService) uploadToMinIO(file io.Reader, fileName string, fileSize int64, contentType string) (string, error) {
	if s.minioClient == nil {
		return "", errors.New("MinIO недоступен, загрузка изображений отключена")
	}

	ctx := context.Background()

	// Загружаем файл в MinIO
	_, err := s.minioClient.PutObject(ctx, s.bucketName, fileName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки в MinIO: %v", err)
	}

	// Генерируем URL для доступа к файлу
	imageURL := fmt.Sprintf("/%s/%s", s.bucketName, fileName)

	logrus.Infof("Изображение загружено в MinIO: %s", fileName)
	return imageURL, nil
}

func (s *ElementService) GetAllElements(search string) (*ElementsResponse, error) {
	var elements []ds.Elements
	var err error

	if search == "" {
		elements, err = s.repo.GetAllElements()
	} else {
		elements, err = s.repo.SearchElementByName(search)
	}

	if err != nil {
		return nil, err
	}

	var items []ElementItem
	for _, elem := range elements {
		items = append(items, ElementItem{
			ID:            elem.ID,
			Image:         elem.Img,
			Title:         elem.Name,
			Concentration: s.formatConcentration(elem.Concentration),
			PH:            s.formatPH(elem.Ph),
		})
	}

	return &ElementsResponse{
		Items:   items,
		Total:   len(items),
		Query:   search,
		HasMore: false,
	}, nil
}

func (s *ElementService) CreateElement(req *CreateElementRequest) (*CreateElementResponse, error) {
	// Валидация бизнес-правил
	if s == nil {
		return nil, errors.New("ElementService is nil")
	}
	if s.repo == nil {
		return nil, errors.New("repository is nil in ElementService")
	}

	if req.Name == "" {
		return nil, errors.New("название элемента обязательно")
	}

	if req.Ph < 0 || req.Ph > 14 {
		return nil, errors.New("pH должен быть в диапазоне 0-14")
	}

	if req.Concentration < 0 {
		return nil, errors.New("концентрация не может быть отрицательной")
	}

	// Проверка на уникальность названия
	exists, err := s.repo.CheckElementExists(req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("элемент с таким названием уже существует")
	}

	// Создаем элемент
	element := &ds.Elements{
		Name:          req.Name,
		Description:   req.Description,
		Ph:            req.Ph,
		Concentration: req.Concentration,
		Img:           "", // изображение будет добавлено отдельно
		IsDelete:      false,
	}

	// Сохраняем в БД
	err = s.repo.CreateElement(element)
	if err != nil {
		return nil, err
	}

	// Возвращаем ответ
	return &CreateElementResponse{
		ID:            element.ID,
		Name:          element.Name,
		Description:   element.Description,
		Ph:            element.Ph,
		Concentration: element.Concentration,
		Message:       "Элемент успешно создан",
	}, nil
}

func (s *ElementService) formatConcentration(conc float32) string {
	return fmt.Sprintf("%.2f%%", conc)
}

func (s *ElementService) formatPH(ph float32) string {
	return fmt.Sprintf("%.1f", ph)
}

func (s *ElementService) UpdateElement(elementID int, req *UpdateElementRequest) (*ElementResponse, error) {
	// Проверяем существование элемента
	existingElement, err := s.repo.GetElementByID(elementID) // ← теперь int
	if err != nil {
		return nil, errors.New("элемент не найден")
	}

	// Валидация бизнес-правил
	if req.Name != nil {
		if *req.Name == "" {
			return nil, errors.New("название элемента не может быть пустым")
		}
		// Проверка уникальности названия (кроме текущего элемента)
		exists, err := s.repo.CheckElementExists(*req.Name)
		if err != nil {
			return nil, err
		}
		if exists && *req.Name != existingElement.Name {
			return nil, errors.New("элемент с таким названием уже существует")
		}
	}

	if req.Ph != nil && (*req.Ph < 0 || *req.Ph > 14) {
		return nil, errors.New("pH должен быть в диапазоне 0-14")
	}

	if req.Concentration != nil && *req.Concentration < 0 {
		return nil, errors.New("концентрация не может быть отрицательной")
	}

	// Подготавливаем обновления
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Ph != nil {
		updates["ph"] = *req.Ph
	}
	if req.Concentration != nil {
		updates["concentration"] = *req.Concentration
	}

	// Если нет полей для обновления
	if len(updates) == 0 {
		return nil, errors.New("нет данных для обновления")
	}

	// Обновляем элемент в БД
	err = s.repo.UpdateElement(elementID, updates)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления элемента: %v", err)
	}

	// Получаем обновленный элемент
	updatedElement, err := s.repo.GetElementByID(elementID)
	if err != nil {
		return nil, err
	}

	// Возвращаем ответ
	return &ElementResponse{
		ID:            updatedElement.ID,
		Name:          updatedElement.Name,
		Description:   updatedElement.Description,
		Ph:            updatedElement.Ph,
		Concentration: updatedElement.Concentration,
		Image:         updatedElement.Img,
		Message:       "Элемент успешно обновлен",
	}, nil
}

// DeleteElement удаляет элемент и его изображение
func (s *ElementService) DeleteElement(elementID int) (*DeleteElementResponse, error) {
	// Проверяем существование элемента
	_, err := s.repo.GetElementByID(elementID)
	if err != nil {
		return nil, errors.New("элемент не найден")
	}

	// Удаляем элемент и получаем путь к изображению
	imagePath, err := s.repo.DeleteElement(elementID)
	if err != nil {
		return nil, fmt.Errorf("ошибка удаления элемента: %v", err)
	}

	// Удаляем изображение из MinIO если оно есть
	if imagePath != "" {
		// Просто вызываем без логирования ошибки
		_ = s.deleteImageFromMinIO(imagePath)
	}

	return &DeleteElementResponse{
		ID:      elementID,
		Message: "Элемент успешно удален",
	}, nil
}

// validateImageFile валидирует файл изображения
func (s *ElementService) validateImageFile(fileHeader *multipart.FileHeader) error {
	// Проверяем размер файла (макс 5MB)
	if fileHeader.Size > 5<<20 {
		return errors.New("размер файла не должен превышать 5MB")
	}

	// Проверяем MIME тип
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
	}

	fileType := fileHeader.Header.Get("Content-Type")
	if !allowedTypes[fileType] {
		return errors.New("поддерживаются только JPEG, PNG и GIF изображения")
	}

	return nil
}

// generateFileName генерирует уникальное имя файла на латинице
func (s *ElementService) generateFileName(originalName string) string {
	ext := filepath.Ext(originalName) // сохраняем расширение
	name := strings.TrimSuffix(originalName, ext)

	// Транслитерация кириллицы в латиницу
	name = s.transliterate(name)

	// Оставляем только латинские буквы, цифры и дефисы
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	name = reg.ReplaceAllString(name, "-")

	// Добавляем timestamp и случайную строку для уникальности
	timestamp := time.Now().Format("20060102150405")
	randomStr := s.generateRandomString(8)

	return fmt.Sprintf("%s-%s-%s%s", name, timestamp, randomStr, ext)
}

// transliterate транслитерирует кириллицу в латиницу
func (s *ElementService) transliterate(text string) string {
	translitMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch", 'ъ': "",
		'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	}

	var result strings.Builder
	for _, char := range strings.ToLower(text) {
		if lat, exists := translitMap[char]; exists {
			result.WriteString(lat)
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// generateRandomString генерирует случайную строку
func (s *ElementService) generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes)
}

// deleteImageFromMinIO удаляет изображение из MinIO
func (s *ElementService) deleteImageFromMinIO(imageURL string) error {
	if imageURL == "" {
		return nil
	}

	if s.minioClient == nil {
		// MinIO недоступен, просто логируем и продолжаем
		logrus.Warn("MinIO недоступен, пропускаем удаление изображения")
		return nil
	}

	// Извлекаем имя файла из URL
	fileName := filepath.Base(imageURL)
	if fileName == "" || fileName == "." {
		return errors.New("неверный URL изображения")
	}

	ctx := context.Background()
	err := s.minioClient.RemoveObject(ctx, s.bucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка удаления из MinIO: %v", err)
	}

	logrus.Infof("Изображение удалено из MinIO: %s", fileName)
	return nil
}

// UploadImage загружает изображение для элемента
func (s *ElementService) UploadImage(elementID int, fileHeader *multipart.FileHeader) (*UploadImageResponse, error) {
	// Проверяем существование элемента
	element, err := s.repo.GetElementByID(elementID)
	if err != nil {
		return nil, errors.New("элемент не найден")
	}

	// Валидация файла
	if err := s.validateImageFile(fileHeader); err != nil {
		return nil, err
	}

	// Открываем файл
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	// Генерируем уникальное имя файла на латинице
	fileName := s.generateFileName(fileHeader.Filename)
	contentType := fileHeader.Header.Get("Content-Type")

	// Удаляем старое изображение если есть
	if element.Img != "" {
		if err := s.deleteImageFromMinIO(element.Img); err != nil {
			logrus.Warnf("Не удалось удалить старое изображение: %v", err)
		}
	}

	// Загружаем в MinIO
	imageURL, err := s.uploadToMinIO(file, fileName, fileHeader.Size, contentType)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки в MinIO: %v", err)
	}

	// Обновляем путь к изображению в БД
	if err := s.repo.UpdateElementImage(elementID, imageURL); err != nil {
		// Если не удалось обновить БД, удаляем загруженное изображение
		_ = s.deleteImageFromMinIO(imageURL)
		return nil, fmt.Errorf("ошибка обновления элемента: %v", err)
	}

	return &UploadImageResponse{
		ID:      elementID,
		Image:   imageURL,
		Message: "Изображение успешно загружено",
	}, nil
}
