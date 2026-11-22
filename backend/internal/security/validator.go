package security

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Standard error messages for business logic & overflow protection
var (
	ErrNegativeNumber = errors.New("nilai numerik tidak boleh bernilai negatif")
	ErrOverflow       = errors.New("nilai numerik melebihi batas wajar sistem (integer/float overflow protection)")
	ErrInvalidQuantity = errors.New("jumlah (qty) harus bernilai minimal 1 dan tidak melebihi batas wajar")
	ErrInvalidPrice    = errors.New("harga/nominal tidak valid atau bernilai negatif")
	ErrInvalidDiscount = errors.New("diskon tidak boleh melebihi subtotal transaksi")
)

// ValidateQuantity ensures quantities are strictly positive and within safe upper bounds
// to prevent integer overflow and allocation Denial-of-Service attacks.
func ValidateQuantity(qty int, maxQty int) error {
	if qty <= 0 {
		return fmt.Errorf("%w: jumlah (%d) harus >= 1", ErrInvalidQuantity, qty)
	}
	if maxQty > 0 && qty > maxQty {
		return fmt.Errorf("%w: jumlah (%d) melebihi batas maksimum (%d)", ErrOverflow, qty, maxQty)
	}
	return nil
}

// ValidatePrice ensures monetary values are positive, finite (non-NaN, non-Inf),
// and within realistic financial limits (e.g. 100 Billion IDR).
func ValidatePrice(amount float64, maxAmount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return fmt.Errorf("%w: nominal tidak boleh NaN atau Tak Hingga", ErrInvalidPrice)
	}
	if amount < 0 {
		return fmt.Errorf("%w: nominal (%.2f) tidak boleh negatif", ErrNegativeNumber, amount)
	}
	if maxAmount > 0 && amount > maxAmount {
		return fmt.Errorf("%w: nominal (%.2f) melebihi batas maksimum (%.2f)", ErrOverflow, amount, maxAmount)
	}
	return nil
}

// ValidateDiscount ensures discount does not exceed the subtotal, preventing negative invoices.
func ValidateDiscount(subtotal, discount float64) error {
	if err := ValidatePrice(discount, subtotal); err != nil {
		return fmt.Errorf("%w: nilai diskon tidak valid", err)
	}
	if discount > subtotal {
		return fmt.Errorf("%w: diskon (%.2f) melebihi subtotal (%.2f)", ErrInvalidDiscount, discount, subtotal)
	}
	return nil
}

// ValidateVitalSigns performs medical plausibility checks on vital signs.
func ValidateVitalSigns(tdSistole, tdDiastole *int, hr, rr, temp, spo2 *float64) error {
	if tdSistole != nil {
		if *tdSistole < 30 || *tdSistole > 320 {
			return errors.New("nilai TD Sistole di luar rentang klinis wajar (30 - 320 mmHg)")
		}
	}
	if tdDiastole != nil {
		if *tdDiastole < 20 || *tdDiastole > 220 {
			return errors.New("nilai TD Diastole di luar rentang klinis wajar (20 - 220 mmHg)")
		}
	}
	if hr != nil {
		if *hr < 20 || *hr > 300 {
			return errors.New("nilai Heart Rate di luar rentang klinis wajar (20 - 300 bpm)")
		}
	}
	if rr != nil {
		if *rr < 4 || *rr > 100 {
			return errors.New("nilai Laju Napas (RR) di luar rentang klinis wajar (4 - 100 x/mnt)")
		}
	}
	if temp != nil {
		if *temp < 25.0 || *temp > 46.0 {
			return errors.New("nilai Suhu Tubuh di luar rentang fisiologis wajar (25.0 - 46.0 °C)")
		}
	}
	if spo2 != nil {
		if *spo2 < 0.0 || *spo2 > 100.0 {
			return errors.New("nilai SpO2 harus berada dalam rentang 0.0 - 100.0 %")
		}
	}
	return nil
}

// ValidateStringLength checks that string inputs do not exceed maximum database field lengths.
func ValidateStringLength(val string, fieldName string, maxLen int) error {
	trimmed := strings.TrimSpace(val)
	if len(trimmed) > maxLen {
		return fmt.Errorf("panjang teks %s melebihi batas maksimum (%d karakter)", fieldName, maxLen)
	}
	return nil
}
