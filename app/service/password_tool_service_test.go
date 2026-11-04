package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/utils"
	"github.com/go-playground/validator/v10"
)

func TestGeneratePassword(t *testing.T) {
	svc := NewPasswordToolService(validator.New())
	ctx := context.Background()
	got, err := svc.Generate(ctx, &GeneratePasswordReq{
		Length:           16,
		Count:            5,
		Upper:            true,
		Lower:            true,
		Digit:            true,
		Symbol:           true,
		ExcludeAmbiguous: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Passwords) != 5 {
		t.Fatalf("count = %d", len(got.Passwords))
	}
	for _, password := range got.Passwords {
		if utf8.RuneCountInString(password) != 16 {
			t.Fatalf("length = %d", utf8.RuneCountInString(password))
		}
		if strings.ContainsAny(password, passwordAmbiguous) {
			t.Fatalf("ambiguous char in %q", password)
		}
		if !containsAny(password, withoutAmbiguous([]rune(passwordUpper))) {
			t.Fatalf("missing upper in %q", password)
		}
		if !containsAny(password, withoutAmbiguous([]rune(passwordLower))) {
			t.Fatalf("missing lower in %q", password)
		}
		if !containsAny(password, withoutAmbiguous([]rune(passwordDigit))) {
			t.Fatalf("missing digit in %q", password)
		}
		if !containsAny(password, []rune(passwordSymbol)) {
			t.Fatalf("missing symbol in %q", password)
		}
	}
}

func TestGeneratePasswordRejectsEmptyCharset(t *testing.T) {
	svc := NewPasswordToolService(validator.New())
	_, err := svc.Generate(context.Background(), &GeneratePasswordReq{Length: 12, Count: 1})
	var biz *errContract.BizError
	if !asBiz(err, &biz) || biz.Code != CodePasswordCharsetEmpty {
		t.Fatalf("err = %v", err)
	}
}

func TestGeneratePasswordRejectsShortLength(t *testing.T) {
	svc := NewPasswordToolService(validator.New())
	_, err := svc.Generate(context.Background(), &GeneratePasswordReq{Length: 3, Count: 1, Lower: true})
	var biz *errContract.BizError
	if err == nil || asBiz(err, &biz) {
		t.Fatalf("err = %v", err)
	}
}

func TestGeneratePasswordDigitsExcludeAmbiguous(t *testing.T) {
	svc := NewPasswordToolService(validator.New())
	got, err := svc.Generate(context.Background(), &GeneratePasswordReq{
		Length:           12,
		Count:            3,
		Digit:            true,
		ExcludeAmbiguous: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	allowed := string(withoutAmbiguous([]rune(passwordDigit)))
	for _, password := range got.Passwords {
		for _, ch := range password {
			if !strings.ContainsRune(allowed, ch) {
				t.Fatalf("unexpected %q in %q", ch, password)
			}
		}
	}
}

func TestHashPassword(t *testing.T) {
	svc := NewPasswordToolService(validator.New())
	const plain = "secret-pass"
	got, err := svc.Hash(context.Background(), &HashPasswordReq{Password: plain})
	if err != nil {
		t.Fatal(err)
	}
	if !utils.BcryptCheck(plain, got.Hash) {
		t.Fatalf("hash does not match")
	}
}

func TestHashPasswordRejectsTooLong(t *testing.T) {
	svc := NewPasswordToolService(validator.New())
	_, err := svc.Hash(context.Background(), &HashPasswordReq{Password: strings.Repeat("a", passwordBcryptMax+1)})
	if err == nil {
		t.Fatal("expected error")
	}
	plain := strings.Repeat("你", 30)
	_, err = svc.Hash(context.Background(), &HashPasswordReq{Password: plain})
	var biz *errContract.BizError
	if !asBiz(err, &biz) || biz.Code != CodePasswordTooLong {
		t.Fatalf("err = %v", err)
	}
}

func containsAny(password string, chars []rune) bool {
	for _, ch := range chars {
		if strings.ContainsRune(password, ch) {
			return true
		}
	}
	return false
}

func asBiz(err error, target **errContract.BizError) bool {
	return errors.As(err, target)
}
