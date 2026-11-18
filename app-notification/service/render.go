package service

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	texttemplate "text/template"
)

// renderText 用文本模板渲染。模板为空时返回空字符串，由调用方决定是否回退。
func renderText(src string, vars map[string]any) (string, error) {
	if strings.TrimSpace(src) == "" {
		return "", nil
	}
	tpl, err := texttemplate.New("notify").Option("missingkey=error").Parse(src)
	if err != nil {
		return "", fmt.Errorf("模板语法错误: %w", err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, templateVars(vars)); err != nil {
		return "", fmt.Errorf("模板变量缺失: %w", err)
	}
	return buf.String(), nil
}

// renderHTML 用 HTML 模板渲染邮件正文，变量会按 HTML 转义。
func renderHTML(src string, vars map[string]any) (string, error) {
	if strings.TrimSpace(src) == "" {
		return "", nil
	}
	tpl, err := template.New("notify").Option("missingkey=error").Parse(src)
	if err != nil {
		return "", fmt.Errorf("模板语法错误: %w", err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, templateVars(vars)); err != nil {
		return "", fmt.Errorf("模板变量缺失: %w", err)
	}
	return buf.String(), nil
}

func templateVars(vars map[string]any) map[string]any {
	if vars == nil {
		return map[string]any{}
	}
	return vars
}

// applyText 在模板为空时使用回退文案。
func applyText(src, fallback string, vars map[string]any) (string, error) {
	if strings.TrimSpace(src) == "" {
		return fallback, nil
	}
	return renderText(src, vars)
}

// applyHTML 在模板为空时直接用回退正文。回退正文已经是清洗过的 HTML，不能再转义。
func applyHTML(src, fallback string, vars map[string]any) (string, error) {
	if strings.TrimSpace(src) == "" {
		return fallback, nil
	}
	return renderHTML(src, vars)
}
