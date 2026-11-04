package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/utils"
	validatorv10 "github.com/go-playground/validator/v10"
)

const (
	// CodePasswordCharsetEmpty 表示没有勾选任何字符种类。
	CodePasswordCharsetEmpty = "PASSWORD_CHARSET_EMPTY"
	// CodePasswordCharsetUnusable 表示勾选的字符在排除易混字符后已空。
	CodePasswordCharsetUnusable = "PASSWORD_CHARSET_UNUSABLE"
	// CodePasswordTooLong 表示明文超过 bcrypt 能接受的字节数。
	CodePasswordTooLong = "PASSWORD_TOO_LONG"

	passwordUpper     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	passwordLower     = "abcdefghijklmnopqrstuvwxyz"
	passwordDigit     = "0123456789"
	passwordSymbol    = "!@#$%^&*()-_=+[]{}:,.?"
	passwordAmbiguous = "0OoIl1"
	passwordBcryptMax = 72
)

// PasswordCharsetEmpty 表示生成时一种字符都没选。
func PasswordCharsetEmpty() *errContract.BizError {
	return errContract.NewBizError(CodePasswordCharsetEmpty, "请至少选择一种字符")
}

// PasswordCharsetUnusable 表示排除易混字符后没有剩余字符。
func PasswordCharsetUnusable() *errContract.BizError {
	return errContract.NewBizError(CodePasswordCharsetUnusable, "排除易混字符后没有可用字符")
}

// PasswordTooLong 表示待计算密文的明文超过 72 字节。
func PasswordTooLong() *errContract.BizError {
	return errContract.NewBizError(CodePasswordTooLong, "密码不能超过 72 个字节")
}

// GeneratePasswordReq 是随机密码的生成条件。
type GeneratePasswordReq struct {
	Length           int  `json:"length" validate:"required,min=8,max=64" label:"长度"`
	Count            int  `json:"count" validate:"required,min=1,max=20" label:"条数"`
	Upper            bool `json:"upper" label:"大写字母"`
	Lower            bool `json:"lower" label:"小写字母"`
	Digit            bool `json:"digit" label:"数字"`
	Symbol           bool `json:"symbol" label:"符号"`
	ExcludeAmbiguous bool `json:"excludeAmbiguous" label:"排除易混字符"`
}

// GeneratePasswordResult 是一次生成的明文列表，不落库。
type GeneratePasswordResult struct {
	Passwords []string `json:"passwords"`
}

// HashPasswordReq 是待计算 bcrypt 的明文。
type HashPasswordReq struct {
	Password string `json:"password" validate:"required,min=1,max=72" label:"明文"`
}

// HashPasswordResult 是 bcrypt 密文。
type HashPasswordResult struct {
	Hash string `json:"hash"`
}

// PasswordToolService 生成随机密码并计算 bcrypt。
type PasswordToolService struct {
	validate *validatorv10.Validate
}

// NewPasswordToolService 创建密码工具服务。
func NewPasswordToolService(validate *validatorv10.Validate) *PasswordToolService {
	return &PasswordToolService{validate: validate}
}

// Generate 按字符种类生成若干条密码。明文只出现在返回值里。
func (i *PasswordToolService) Generate(ctx context.Context, req *GeneratePasswordReq) (GeneratePasswordResult, error) {
	if err := ctx.Err(); err != nil {
		return GeneratePasswordResult{}, err
	}
	if err := i.validate.Struct(req); err != nil {
		return GeneratePasswordResult{}, err
	}
	classes, err := passwordClasses(req)
	if err != nil {
		return GeneratePasswordResult{}, err
	}
	passwords := make([]string, 0, req.Count)
	for range req.Count {
		password, err := newPassword(req.Length, classes)
		if err != nil {
			return GeneratePasswordResult{}, err
		}
		passwords = append(passwords, password)
	}
	return GeneratePasswordResult{Passwords: passwords}, nil
}

// Hash 用与命令 admin:gen-pwd 相同的 bcrypt 计算密文。
func (i *PasswordToolService) Hash(ctx context.Context, req *HashPasswordReq) (HashPasswordResult, error) {
	if err := ctx.Err(); err != nil {
		return HashPasswordResult{}, err
	}
	if err := i.validate.Struct(req); err != nil {
		return HashPasswordResult{}, err
	}
	if len(req.Password) > passwordBcryptMax {
		return HashPasswordResult{}, PasswordTooLong()
	}
	return HashPasswordResult{Hash: utils.BcryptHash(req.Password)}, nil
}

func passwordClasses(req *GeneratePasswordReq) ([][]rune, error) {
	if !req.Upper && !req.Lower && !req.Digit && !req.Symbol {
		return nil, PasswordCharsetEmpty()
	}
	var classes [][]rune
	var err error
	classes, err = appendPasswordClass(classes, req.Upper, passwordUpper, req.ExcludeAmbiguous)
	if err != nil {
		return nil, err
	}
	classes, err = appendPasswordClass(classes, req.Lower, passwordLower, req.ExcludeAmbiguous)
	if err != nil {
		return nil, err
	}
	classes, err = appendPasswordClass(classes, req.Digit, passwordDigit, req.ExcludeAmbiguous)
	if err != nil {
		return nil, err
	}
	classes, err = appendPasswordClass(classes, req.Symbol, passwordSymbol, req.ExcludeAmbiguous)
	if err != nil {
		return nil, err
	}
	if len(classes) == 0 {
		return nil, PasswordCharsetUnusable()
	}
	return classes, nil
}

func appendPasswordClass(classes [][]rune, enabled bool, alphabet string, excludeAmbiguous bool) ([][]rune, error) {
	if !enabled {
		return classes, nil
	}
	chars := []rune(alphabet)
	if excludeAmbiguous {
		chars = withoutAmbiguous(chars)
	}
	if len(chars) == 0 {
		return nil, PasswordCharsetUnusable()
	}
	return append(classes, chars), nil
}

func withoutAmbiguous(chars []rune) []rune {
	kept := make([]rune, 0, len(chars))
	for _, ch := range chars {
		if strings.ContainsRune(passwordAmbiguous, ch) {
			continue
		}
		kept = append(kept, ch)
	}
	return kept
}

func newPassword(length int, classes [][]rune) (string, error) {
	if length < len(classes) {
		return "", PasswordCharsetUnusable()
	}
	buf := make([]rune, 0, length)
	for _, class := range classes {
		ch, err := pickRune(class)
		if err != nil {
			return "", err
		}
		buf = append(buf, ch)
	}
	all := combineClasses(classes)
	for len(buf) < length {
		ch, err := pickRune(all)
		if err != nil {
			return "", err
		}
		buf = append(buf, ch)
	}
	if err := shuffleRunes(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func combineClasses(classes [][]rune) []rune {
	var all []rune
	for _, class := range classes {
		all = append(all, class...)
	}
	return all
}

func pickRune(chars []rune) (rune, error) {
	if len(chars) == 0 {
		return 0, fmt.Errorf("password charset is empty")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		return 0, fmt.Errorf("generate password: %w", err)
	}
	return chars[n.Int64()], nil
}

func shuffleRunes(chars []rune) error {
	for i := len(chars) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return fmt.Errorf("generate password: %w", err)
		}
		j := int(n.Int64())
		chars[i], chars[j] = chars[j], chars[i]
	}
	return nil
}
