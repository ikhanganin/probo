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

package gid

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"go.gearno.de/crypto/uuid"
)

type (
	GID uuid.UUID
)

func NewGID(et uint32) (GID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return GID(uuid.Nil), err
	}

	binary.BigEndian.PutUint32(id[10:14], et)

	return GID(id), nil
}

func (gid GID) Value() (driver.Value, error) {
	return gid[:], nil
}

func (gid GID) EntityType() uint32 {
	return binary.BigEndian.Uint32(gid[10:14])
}

func (gid *GID) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		*gid = GID(id)
	case []byte:
		id, err := uuid.ParseBytes(v)
		if err != nil {
			return err
		}
		*gid = GID(id)
	default:
		return fmt.Errorf("invalid type for GID: expected string or []byte, got %T", value)
	}
	return nil
}

func (gid GID) String() string {
	return base64.RawURLEncoding.EncodeToString(gid[:])
}

func (gid GID) MarshalText() ([]byte, error) {
	enc := base64.RawURLEncoding

	buf := make([]byte, enc.EncodedLen(len(gid)))
	enc.Encode(buf, gid[:])

	return buf, nil
}

func (gid *GID) UnmarshalText(encoded []byte) error {
	enc := base64.RawURLEncoding

	_, err := enc.Decode(gid[:], encoded)
	if err != nil {
		return err
	}

	return nil
}
