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

package probo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/probo/coredata"
	"go.gearno.de/kit/migrator"
	"go.gearno.de/kit/pg"
)

type (
	Service struct {
		pg     *pg.Client
		scope  *coredata.Scope
		s3     *s3.Client
		bucket string
	}
)

func NewService(ctx context.Context, pgClient *pg.Client, s3Client *s3.Client, bucket string) (*Service, error) {
	err := migrator.NewMigrator(pgClient, coredata.Migrations).Run(ctx, "migrations")
	if err != nil {
		return nil, fmt.Errorf("cannot migrate database schema: %w", err)
	}

	if bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	return &Service{
		pg:     pgClient,
		s3:     s3Client,
		scope:  coredata.NewScope(), // must be created from auth
		bucket: bucket,
	}, nil
}
