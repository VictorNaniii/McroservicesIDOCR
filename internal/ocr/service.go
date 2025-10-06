package ocr

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/otiai10/gosseract/v2"
	"github.com/sirupsen/logrus"
	"github.com/victornani/id-ocr-service/internal/models"
)

type Service struct {
	logger       *logrus.Logger
	tesseractLang string
	tempDir      string
}

func NewService(logger *logrus.Logger, tesseractLang, tempDir string) (*Service, error) {
	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &Service{
		logger:       logger,
		tesseractLang: tesseractLang,
		tempDir:      tempDir,
	}, nil
}

// ProcessImage performs OCR on an image and extracts ID data
func (s *Service) ProcessImage(imageData []byte, imagePath string) (*models.IDData, error) {
	var pathToProcess string

	// Handle image data or path
	if len(imageData) > 0 {
		// Save image data to temp file
		tempFile := filepath.Join(s.tempDir, fmt.Sprintf("scan_%d.jpg", time.Now().UnixNano()))

		// Check if data is base64 encoded
		if decoded, err := base64.StdEncoding.DecodeString(string(imageData)); err == nil {
			imageData = decoded
		}

		if err := os.WriteFile(tempFile, imageData, 0644); err != nil {
			return nil, fmt.Errorf("failed to save temp image: %w", err)
		}
		pathToProcess = tempFile
		defer os.Remove(tempFile) // Clean up
	} else if imagePath != "" {
		pathToProcess = imagePath
	} else {
		return nil, fmt.Errorf("no image data or path provided")
	}

	// Perform OCR
	rawText, err := s.extractText(pathToProcess)
	if err != nil {
		return nil, fmt.Errorf("OCR failed: %w", err)
	}

	s.logger.Debugf("Extracted text: %s", rawText)

	// Parse ID data from text
	idData := s.parseIDData(rawText)
	idData.Timestamp = time.Now()

	return idData, nil
}

// extractText uses Tesseract to extract text from image
func (s *Service) extractText(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	client.SetLanguage(s.tesseractLang)
	client.SetImage(imagePath)

	text, err := client.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}

// parseIDData extracts structured data from raw OCR text
func (s *Service) parseIDData(rawText string) *models.IDData {
	data := &models.IDData{
		RawText: rawText,
	}

	lines := strings.Split(rawText, "\n")
	text := strings.Join(lines, " ")

	// Extract IDNP (Personal identification number - 13 digits for Moldova)
	// Pattern: 13 digits, possibly with spaces
	idnpRegex := regexp.MustCompile(`\b(\d{13}|\d{4}\s?\d{4}\s?\d{5})\b`)
	if matches := idnpRegex.FindStringSubmatch(text); len(matches) > 0 {
		data.IDNP = strings.ReplaceAll(matches[0], " ", "")
	}

	// Extract birth date (various formats: DD.MM.YYYY, DD/MM/YYYY, YYYY-MM-DD)
	dateRegex := regexp.MustCompile(`\b(\d{2}[./]\d{2}[./]\d{4}|\d{4}-\d{2}-\d{2})\b`)
	if matches := dateRegex.FindStringSubmatch(text); len(matches) > 0 {
		data.BirthDate = matches[0]
	}

	// Extract names - look for common ID patterns
	// Pattern: Look for lines with "Name", "Nume", "Prenume" keywords
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lineUpper := strings.ToUpper(line)

		// First name patterns
		if (strings.Contains(lineUpper, "PRENUME") || strings.Contains(lineUpper, "FIRST NAME") ||
		    strings.Contains(lineUpper, "GIVEN NAME")) && data.FirstName == "" {
			// Extract text after the label
			parts := regexp.MustCompile(`[:]\s*`).Split(line, 2)
			if len(parts) > 1 {
				data.FirstName = strings.TrimSpace(parts[1])
			}
		}

		// Last name patterns
		if (strings.Contains(lineUpper, "NUME") || strings.Contains(lineUpper, "LAST NAME") ||
		    strings.Contains(lineUpper, "SURNAME") || strings.Contains(lineUpper, "FAMILY NAME")) &&
		    data.LastName == "" {
			parts := regexp.MustCompile(`[:]\s*`).Split(line, 2)
			if len(parts) > 1 {
				data.LastName = strings.TrimSpace(parts[1])
			}
		}
	}

	// Fallback: extract capitalized words as potential names
	if data.FirstName == "" || data.LastName == "" {
		nameRegex := regexp.MustCompile(`\b[A-Z][A-Z]+\b`)
		names := nameRegex.FindAllString(text, -1)

		validNames := []string{}
		for _, name := range names {
			// Filter out common ID document words
			if !s.isCommonIDWord(name) && len(name) > 2 {
				validNames = append(validNames, name)
			}
		}

		if len(validNames) >= 2 {
			if data.LastName == "" {
				data.LastName = validNames[0]
			}
			if data.FirstName == "" && len(validNames) > 1 {
				data.FirstName = validNames[1]
			}
		}
	}

	return data
}

// isCommonIDWord checks if a word is a common ID document keyword
func (s *Service) isCommonIDWord(word string) bool {
	commonWords := map[string]bool{
		"IDENTITY": true, "CARD": true, "PASSPORT": true, "REPUBLIC": true,
		"MOLDOVA": true, "ROMANIA": true, "ISSUED": true, "DATE": true,
		"BIRTH": true, "SEX": true, "NATIONALITY": true, "VALID": true,
		"BULETINUL": true, "ACTE": true, "IDENTITATE": true,
	}
	return commonWords[strings.ToUpper(word)]
}

// Cleanup removes old temporary files
func (s *Service) Cleanup() error {
	files, err := filepath.Glob(filepath.Join(s.tempDir, "scan_*.jpg"))
	if err != nil {
		return err
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		// Remove files older than 1 hour
		if time.Since(info.ModTime()) > time.Hour {
			os.Remove(file)
		}
	}

	return nil
}
