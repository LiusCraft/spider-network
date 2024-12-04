package api

import (
    "fmt"
    "html/template"
)

// FormatBytes 将字节数转换为人类可读的格式
func FormatBytes(bytes int64) string {
    const (
        B  = 1
        KB = 1024 * B
        MB = 1024 * KB
        GB = 1024 * MB
    )

    var val float64
    var unit string

    switch {
    case bytes >= GB:
        val = float64(bytes) / float64(GB)
        unit = "GB"
    case bytes >= MB:
        val = float64(bytes) / float64(MB)
        unit = "MB"
    case bytes >= KB:
        val = float64(bytes) / float64(KB)
        unit = "KB"
    default:
        val = float64(bytes)
        unit = "B"
    }

    return fmt.Sprintf("%.2f%s", val, unit)
}

// GetTemplateFuncs 返回模板函数映射
func GetTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "formatBytes": FormatBytes,
    }
}
