package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

type OrderStatus int

const (
	StatusNew OrderStatus = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
)

func (s OrderStatus) String() string {
	switch s {
	case StatusNew:
		return "NEW"
	case StatusProcessing:
		return "PROCESSING"
	case StatusInvalid:
		return "INVALID"
	case StatusProcessed:
		return "PROCESSED"
	default:
		panic(fmt.Sprintf("unexpected domain.OrderStatus: %#v", s))
	}
}

func FromString(s string) (OrderStatus, error) {
	switch s {
	case "NEW":
		return StatusNew, nil
	case "PROCESSING":
		return StatusProcessing, nil
	case "INVALID":
		return StatusInvalid, nil
	case "PROCESSED":
		return StatusProcessed, nil
	default:
		return 0, errors.New(fmt.Sprintf("unknown value: %s", s))
	}
}

func (s OrderStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		return errors.New("nil values not supported for OrderStatus")
	}

	sv, err := driver.String.ConvertValue(value)
	if err != nil {
		return fmt.Errorf("cannot scan value. %w", err)
	}

	v, ok := sv.(string)
	if !ok {
		return errors.New("cannot scan value. cannot convert value to string")
	}
	result, err := FromString(v)
	if err != nil {
		return err
	}
	*s = result

	return nil
}
