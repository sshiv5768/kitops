// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package unpack

import (
	"context"
	"errors"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
)

func getStoreForRef(ctx context.Context, opts *unpackOptions) (oras.Target, error) {
	storageHome := constants.StoragePath(opts.configHome)
	localRepo, err := local.NewLocalRepo(storageHome, opts.modelRef)
	if err != nil {
		return nil, fmt.Errorf("failed to read local storage: %s\n", err)
	}

	if _, err := localRepo.Resolve(ctx, opts.modelRef.Reference); err == nil {
		// Reference is present in local storage
		return localRepo, nil
	}

	if opts.modelRef.Registry == util.DefaultRegistry {
		return nil, fmt.Errorf("not found")
	}
	// Not in local storage, check remote
	remoteRegistry, err := remote.NewRegistry(opts.modelRef.Registry, &opts.NetworkOptions)
	if err != nil {
		return nil, fmt.Errorf("could not resolve registry %s: %w", opts.modelRef.Registry, err)
	}

	repo, err := remoteRegistry.Repository(ctx, opts.modelRef.Repository)
	if err != nil {
		return nil, fmt.Errorf("could not resolve repository %s in registry %s", opts.modelRef.Repository, opts.modelRef.Registry)
	}
	if _, err := repo.Resolve(ctx, opts.modelRef.Reference); err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return nil, fmt.Errorf("reference %s is not present in local storage and could not be found in remote", opts.modelRef.String())
		}
		return nil, fmt.Errorf("unexpected error retrieving reference from remote: %w", err)
	}

	return repo, nil
}
