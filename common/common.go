package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/lo"
	"math"
	"reflect"
)

func HmacSha256Sign(message, secretKey string) ([]byte, error) {
	// 创建 HMAC-SHA256 哈希
	h := hmac.New(sha256.New, lo.UnsafeBytes(secretKey))
	_, err := h.Write(lo.UnsafeBytes(message))
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC: %v", err)
	}

	return h.Sum(nil), nil
}

// ExtraPointsForInitialDecay 计算给定周期（period）下，使初始影响衰减至 0.1% 以下
// 所需的额外采样点数量
func ExtraPointsForInitialDecay(period int) int {
	// 目标衰减阈值（0.1%）
	threshold := 0.001
	base := float64(period-1) / float64(period)
	// 解方程 (base)^k <= threshold => k >= log(threshold) / log(base)
	k := math.Log(threshold) / math.Log(base)
	return int(math.Ceil(k)) + period
}

func CheckEmptyFields(value interface{}, skip ...string) error {
	rv := reflect.ValueOf(value)
	rt := reflect.TypeOf(value)

	// 处理指针指向的值
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}

	// 将跳过列表转换为 map 以提高查找效率
	skipMap := make(map[string]struct{}, len(skip))
	for _, name := range skip {
		skipMap[name] = struct{}{}
	}

	// 遍历所有字段
	for i := 0; i < rv.NumField(); i++ {
		fieldVal := rv.Field(i)
		fieldType := rt.Field(i)

		// 跳过非导出字段
		if !fieldType.IsExported() {
			continue
		}

		// 跳过需要排除的字段
		if _, ok := skipMap[fieldType.Name]; ok {
			continue
		}

		// 递归检查结构体和结构体指针
		var err error
		switch {
		case fieldVal.Kind() == reflect.Struct:
			// 直接处理结构体类型
			err = CheckEmptyFields(fieldVal.Interface(), skip...)
		case fieldVal.Kind() == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct && !fieldVal.IsNil():
			// 处理非 nil 的结构体指针
			err = CheckEmptyFields(fieldVal.Elem().Interface(), skip...)
		default:
			// 使用 IsZero 判断零值
			if fieldVal.IsZero() {
				// 错误提示
				err = fmt.Errorf("字段[%s]不可为空", fieldType.Name)
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}
