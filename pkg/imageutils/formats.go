package imageutils

import (
	"strings"
)

// SupportedImageFormats は対応している画像形式のリストを返します
func SupportedImageFormats() []string {
	return []string{
		"jpg", "jpeg", "png", "gif", "webp", "avif", "heic", "heif",
	}
}

// GetSupportedImageExtensions は対応している画像拡張子のリストを返します
func GetSupportedImageExtensions() []string {
	formats := SupportedImageFormats()
	extensions := make([]string, len(formats))

	for i, format := range formats {
		extensions[i] = "." + format
	}

	return extensions
}

// IsImageExt は拡張子が画像ファイルかどうかを判断します
func IsImageExt(ext string) bool {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	for _, supportedExt := range GetSupportedImageExtensions() {
		if ext == supportedExt {
			return true
		}
	}

	return false
}

// IsWebPExt は拡張子がWebPファイルかどうかを判断します
func IsWebPExt(ext string) bool {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	return ext == ".webp"
}

// IsAVIFExt は拡張子がAVIFファイルかどうかを判断します
func IsAVIFExt(ext string) bool {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	return ext == ".avif"
}

// IsHEICExt は拡張子がHEIC/HEIFファイルかどうかを判断します
func IsHEICExt(ext string) bool {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	return ext == ".heic" || ext == ".heif"
}

// GetFormatFromExt は拡張子から画像形式を取得します
func GetFormatFromExt(ext string) string {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	case ".webp":
		return "webp"
	case ".avif":
		return "avif"
	case ".heic", ".heif":
		return "heif"
	default:
		return ""
	}
}

// GetExtFromFormat は画像形式から拡張子を取得します
func GetExtFromFormat(format string) string {
	format = strings.ToLower(format)

	switch format {
	case "jpeg":
		return ".jpg"
	case "png", "gif", "webp", "avif":
		return "." + format
	case "heif":
		return ".heic"
	default:
		return ""
	}
}

// IsSupportedFormat は画像形式がサポートされているかどうかを判断します
func IsSupportedFormat(format string) bool {
	format = strings.ToLower(format)

	for _, supportedFormat := range SupportedImageFormats() {
		if format == supportedFormat ||
			(format == "jpeg" && supportedFormat == "jpg") ||
			(format == "heif" && supportedFormat == "heic") {
			return true
		}
	}

	return false
}
