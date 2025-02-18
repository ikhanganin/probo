// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"database/sql/driver"
	"fmt"
)

type ServiceCriticality string

const (
	ServiceCriticalityLow    ServiceCriticality = "LOW"
	ServiceCriticalityMedium ServiceCriticality = "MEDIUM"
	ServiceCriticalityHigh   ServiceCriticality = "HIGH"
)

func (sc ServiceCriticality) MarshalText() ([]byte, error) {
	return []byte(sc.String()), nil
}

func (sc *ServiceCriticality) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case ServiceCriticalityLow.String():
		*sc = ServiceCriticalityLow
	case ServiceCriticalityMedium.String():
		*sc = ServiceCriticalityMedium
	case ServiceCriticalityHigh.String():
		*sc = ServiceCriticalityHigh
	default:
		return fmt.Errorf("invalid ServiceCriticality value: %q", val)
	}

	return nil
}

func (sc ServiceCriticality) String() string {
	return string(sc)
}

func (sc *ServiceCriticality) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for ServiceCriticality, expected string got %T", value)
	}

	return sc.UnmarshalText([]byte(val))
}

func (sc ServiceCriticality) Value() (driver.Value, error) {
	return sc.String(), nil
}
