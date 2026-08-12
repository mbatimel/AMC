package service

import (
	"errors"
	"strings"
)


func validate(inn string) (bool, error) {
    inn = strings.ReplaceAll(inn, " ", "")
    inn = strings.ReplaceAll(inn, "-", "")

    if len(inn) != 10 && len(inn) != 12 {
        return false, errors.New("INN should be have 10 or 12 lenght")
    }


    for _, ch := range inn {
        if ch < '0' || ch > '9' {
            return false, errors.New("the INN must consist of numbers only.")
        }
    }


    digits := make([]int, len(inn))
    for i, ch := range inn {
        digits[i] = int(ch - '0')
    }

    if len(inn) == 10 {
        return validate10(digits), nil
    }
    return validate12(digits), nil
}


func validate10(d []int) bool {

    coeffs := []int{2, 4, 10, 3, 5, 9, 4, 6, 8}
    sum := 0
    for i := 0; i < 9; i++ {
        sum += d[i] * coeffs[i]
    }
    control := (sum % 11) % 10
    return control == d[9]
}


func validate12(d []int) bool {

    coeffs11 := []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
    sum11 := 0
    for i := 0; i < 10; i++ {
        sum11 += d[i] * coeffs11[i]
    }
    control11 := (sum11 % 11) % 10
    if control11 != d[10] {
        return false
    }


    coeffs12 := []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
    sum12 := 0
    for i := 0; i < 11; i++ {
        sum12 += d[i] * coeffs12[i]
    }
    control12 := (sum12 % 11) % 10
    return control12 == d[11]
}